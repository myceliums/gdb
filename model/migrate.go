package model

import (
	"database/sql"
	"io"
	"log"
	"strings"

	"github.com/myceliums/gdb/dialect"
)

// Migrate runs the configured migration model, when there's a differential between
// the given model and the last stored model in the database it will run a script
// that will settle the differences safely between the stored model and the given one.
func Migrate(dialect dialect.Dialect, db *sql.DB, mdl Model) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() // nolint: errcheck

	q := dialect.AddVersionTable()
	if _, err := tx.Exec(q); err != nil {
		return err
	}

	q = dialect.CheckVersion()
	var version int
	var storedConfig []byte
	if err := tx.QueryRow(q).Scan(&version, &storedConfig); err != nil {
		return err
	}

	if version == 0 {
		q = InitialSQL(dialect, mdl)
	}

	if version > 0 {
		oldMdl, err := New(storedConfig)
		if err != nil {
			return err
		}

		q = dialect.InsertVersion()
		if _, err := tx.Exec(q, version+1, mdl.conf); err != nil {
			return err
		}

		q = UpgradeSQL(dialect, *oldMdl, mdl)
	}

	if _, err := tx.Exec(q); err != nil {
		return err
	}

	return tx.Commit()
}

// InitialSQL returns the sql to model the database after the given configuration.
func InitialSQL(dialect dialect.Dialect, mdl Model) string {
	builder := &strings.Builder{}
	for _, enum := range mdl.Enums {
		builder.WriteString(dialect.AddEnum(enum.Name, enum.Values))
	}

	for table, columns := range mdl.Tables {
		addTable(builder, dialect, table, columns)
	}

	for _, cols := range mdl.Primaries {
		var colNames []string
		for _, col := range cols {
			colNames = append(colNames, col.Name)
		}
		builder.WriteString(dialect.AddPrimaryKey(cols[0].Table, colNames))
	}

	for id, cols := range mdl.Uniques {
		var colNames []string
		for _, col := range cols {
			colNames = append(colNames, col.Name)
		}
		builder.WriteString(dialect.AddUnique(id, cols[0].Table, colNames))
	}

	for _, col := range mdl.Foreigns {
		builder.WriteString(dialect.AddForeignKey(col.Table, col.Name, col.Ref.Table, col.Ref.Name))
	}

	return builder.String()
}

// UpgradeSQL returns the sql which resolves the differential safely between 2 models
func UpgradeSQL(dialect dialect.Dialect, prev, curr Model) (q string) {
	builder := &strings.Builder{}
	for n, enum := range curr.Enums {
		var osval, sval string

		oenum, ok := prev.Enums[n]
		if !ok {
			builder.WriteString(dialect.AddEnum(enum.Name, enum.Values)) // nolint: errcheck
			goto ENUMLOOPEND
		}

		osval = strings.Join(oenum.Values, ` `)
		sval = strings.Join(enum.Values, ` `)

		if osval == sval {
			break
		}

		for _, val := range enum.Values {
			if !strings.Contains(osval, val) {
				builder.WriteString(dialect.AppendEnum(n, val)) // nolint: errcheck
				osval = strings.Trim(osval, val)
				osval = strings.TrimSuffix(osval, ` `)
			}
		}

		for _, delval := range strings.Split(osval, ` `) {
			builder.WriteString(dialect.DeleteEnum(n, delval))
		}

	ENUMLOOPEND:
	}

	compareTables(builder, dialect, curr.Tables, prev.Tables)

	for k, cols := range curr.Primaries {
		oldCols := prev.Primaries[k]

		var names, oldnames []string
		for i, col := range cols {
			names = append(names, col.Name)
			prev.Primaries[k] = append(prev.Primaries[k][:i], prev.Primaries[k][i+1:]...)
		}

		for _, col := range oldCols {
			oldnames = append(oldnames, col.Name)
		}

		if len(names) < 1 && len(oldnames) > 0 {
			builder.WriteString(dialect.DropPrimaryKey(cols[0].Table)) // nolint: errcheck
		} else if len(names) > 0 && len(oldnames) < 1 {
			builder.WriteString(dialect.AddPrimaryKey(cols[0].Table, names)) // nolint: errcheck
		} else if len(names) > 0 && strings.Join(names, ` `) != strings.Join(oldnames, ` `) {
			builder.WriteString(dialect.UpdatePrimaryKey(cols[0].Table, names)) // nolint: errcheck
		}

		if len(prev.Primaries[k]) < 1 {
			delete(prev.Primaries, k)
		}
	}

	for k, cols := range prev.Primaries {
		if len(curr.Primaries[k]) < 1 {
			builder.WriteString(dialect.DropPrimaryKey(cols[0].Table)) // nolint: errcheck
		}
	}

	for k, cols := range curr.Uniques {
		oldCols := prev.Uniques[k]

		var names, oldnames []string
		for i, col := range cols {
			names = append(names, col.Name)

			un := prev.Uniques[k]
			if len(prev.Uniques[k]) == 1 {
				delete(prev.Uniques, k)
			} else {
				prev.Uniques[k] = append(un[:i], un[i+1:]...)
			}
		}

		for _, col := range oldCols {
			oldnames = append(oldnames, col.Name)
		}

		if len(names) < 1 && len(oldnames) > 0 {
			builder.WriteString(dialect.DropUnique(k, cols[0].Table)) // nolint: errcheck
		} else if len(names) > 0 && len(oldnames) < 1 {
			builder.WriteString(dialect.AddUnique(k, cols[0].Table, names)) // nolint: errcheck
		} else if len(names) > 0 && strings.Join(names, ` `) != strings.Join(oldnames, ` `) {
			builder.WriteString(dialect.UpdateUnique(k, cols[0].Table, names)) // nolint: errcheck
		}

	}

	for k, cols := range prev.Uniques {
		if len(curr.Uniques[k]) < 1 {
			builder.WriteString(dialect.DropUnique(k, cols[0].Table)) // nolint: errcheck
		}
	}

	for k, col := range curr.Foreigns {
		oldcol := prev.Foreigns[k]
		if oldcol == nil {
			builder.WriteString(dialect.AddForeignKey(col.Table, col.Name, col.Ref.Table, col.Ref.Name)) // nolint: errcheck
			goto FOREIGNLOOPEND
		}

		if !(oldcol.Table == col.Table && oldcol.Name == col.Name) {
			builder.WriteString(dialect.UpdateForeignKey(col.Table, col.Name, col.Ref.Table, col.Ref.Name)) // nolint: errcheck
		}

	FOREIGNLOOPEND:
		delete(prev.Foreigns, k)
	}

	for k, ocol := range prev.Foreigns {
		if curr.Foreigns[k] == nil {
			builder.WriteString(dialect.DropForeignKey(ocol.Table, ocol.Name)) // nolint: errcheck
		}
	}

	return builder.String()
}

func compareTables(wr io.StringWriter, dialect dialect.Dialect, tables, old map[string]map[string]*Column) {
	for tname, cols := range tables {
		if old[tname] == nil {
			addTable(wr, dialect, tname, tables[tname])
			goto TABLELOOPEND
		}

		for cname, col := range cols {
			oldcol, ok := old[tname][cname]
			if !ok {
				log.Printf("%s.%s\n", tname, cname)
				wr.WriteString(dialect.AddColumn(col.Table, col.Name, col.Datatype.Type(), col.Size)) // nolint: errcheck
			}

			if !ok || col.raw == oldcol.raw {
				delete(old[tname], cname)
				goto COLLOOPEND
			}

			if col.rawtype != oldcol.rawtype || col.Size != oldcol.Size {
				wr.WriteString(dialect.UpdateColumn(tname, cname, col.Datatype.Type(), col.Size)) // nolint: errcheck
			}

			if col.AutoIncement && !oldcol.AutoIncement {
				wr.WriteString(dialect.SetAutoIncrement(tname, cname)) // nolint: errcheck
			} else if !col.AutoIncement && oldcol.AutoIncement {
				wr.WriteString(dialect.UnsetAutoIncrement(tname, cname)) // nolint: errcheck
			}

			if col.NotNull != oldcol.NotNull {
				if col.NotNull {
					wr.WriteString(dialect.SetNotNull(col.Table, col.Name)) // nolint: errcheck
				} else {
					wr.WriteString(dialect.DeleteNotNull(col.Table, col.Name)) // nolint: errcheck
				}
			}

			if col.Default != oldcol.Default {
				if col.Default == `` {
					wr.WriteString(dialect.DropDefault(col.Table, col.Name)) // nolint: errcheck
				} else {
					wr.WriteString(dialect.SetDefault(col.Table, col.Name, col.Default)) // nolint: errcheck
				}
			}

		COLLOOPEND:
			delete(old[tname], cname)
		}

	TABLELOOPEND:
		if len(old[tname]) < 1 {
			delete(old, tname)
		}
	}

	for table, cols := range old {
		if len(tables[table]) == 0 {
			wr.WriteString(dialect.DropTable(table)) // nolint: errcheck
			goto OLDTABLELOOPEND
		}

		for i, col := range cols {
			if tables[table][i] == nil {
				wr.WriteString(dialect.DropColumn(table, col.Name)) // nolint: errcheck
			}
		}
	OLDTABLELOOPEND:
	}
}

func addTable(wr io.StringWriter, dialect dialect.Dialect, table string, cols map[string]*Column) {
	wr.WriteString(dialect.AddTable(table, false)) // nolint: errcheck
	for _, col := range cols {
		addColumn(wr, dialect, col)
	}
}

func addColumn(wr io.StringWriter, dialect dialect.Dialect, col *Column) {
	wr.WriteString(dialect.AddColumn(col.Table, col.Name, col.Datatype.Type(), col.Size)) // nolint: errcheck

	if col.AutoIncement {
		wr.WriteString(dialect.SetAutoIncrement(col.Table, col.Name)) // nolint: errcheck
	}

	if col.NotNull {
		wr.WriteString(dialect.SetNotNull(col.Table, col.Name)) // nolint: errcheck
	}

	if col.Default != `` {
		wr.WriteString(dialect.SetDefault(col.Table, col.Name, col.Default)) // nolint: errcheck
	}

	if col.Check != `` {
		wr.WriteString(dialect.AddCheck(col.Table, col.Name, col.Check)) // nolint: errcheck
	}
}
