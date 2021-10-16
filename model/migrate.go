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
	if err := tx.QueryRow(q).Scan(&version, &storedConfig); err != nil && err != sql.ErrNoRows {
		return err
	}

	q = dialect.InsertVersion()
	if _, err := tx.Exec(q, version+1, mdl.Config()); err != nil {
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

		q = UpgradeSQL(dialect, *oldMdl, mdl)
	}

	log.Printf("Applying migration, version: %d:\n%s\n", version+1, q)

	if _, err := tx.Exec(q); err != nil {
		return err
	}

	return tx.Commit()
}

// InitialSQL returns the sql to model the database after the given configuration.
func InitialSQL(dialect dialect.Dialect, mdl Model) string {
	builder := &strings.Builder{}
	for _, enum := range mdl.Enums {
		dialect.AddEnum(builder, enum.Name, enum.Values)
	}

	for table, columns := range mdl.Tables {
		addTable(builder, dialect, table, columns)
	}

	for _, cols := range mdl.Primaries {
		var colNames []string
		for _, col := range cols {
			colNames = append(colNames, col.Name)
		}
		dialect.AddPrimaryKey(builder, cols[0].Table, colNames)
	}

	for id, cols := range mdl.Uniques {
		var colNames []string
		for _, col := range cols {
			colNames = append(colNames, col.Name)
		}
		dialect.AddUnique(builder, id, cols[0].Table, colNames)
	}

	for _, col := range mdl.Foreigns {
		dialect.AddForeignKey(builder, col.Table, col.Name, col.Ref.Table, col.Ref.Name)
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
			dialect.AddEnum(builder, enum.Name, enum.Values)
			goto ENUMLOOPEND
		}

		osval = strings.Join(oenum.Values, ` `)
		sval = strings.Join(enum.Values, ` `)

		if osval == sval {
			break
		}

		for _, val := range enum.Values {
			if !strings.Contains(osval, val) {
				dialect.AppendEnum(builder, n, val)
				osval = strings.Trim(osval, val)
				osval = strings.TrimSuffix(osval, ` `)
			}
		}

		delete(prev.Enums, n)

	ENUMLOOPEND:
	}

	compareTables(builder, dialect, curr.Tables, prev.Tables)

	for k, cols := range curr.Primaries {
		var names []string
		var update bool

		mnames := map[string]int{}
		moldNames := map[string]int{}

		for i, col := range cols {
			mnames[col.Name] = i
			names = append(names, col.Name)
		}

		oldCols, ok := prev.Primaries[k]
		if !ok {
			goto PRIMARYLOOPEND
		}

		for i, col := range oldCols {
			if _, ok := mnames[col.Name]; !ok {
				update = true
			}
			moldNames[col.Name] = i
		}

		for _, col := range cols {
			oldIndex, ok := moldNames[col.Name]
			if !ok {
				update = true
			}

			if l := len(prev.Primaries[k]); l > 1 {
				prev.Primaries[k] = append(prev.Primaries[k][:oldIndex], prev.Primaries[k][oldIndex+1:]...)
			} else if l == 1 {
				prev.Primaries[k] = []*Column{}
			}
		}

	PRIMARYLOOPEND:
		if !ok {
			dialect.AddPrimaryKey(builder, k, names)
		} else if update {
			dialect.UpdatePrimaryKey(builder, cols[0].Table, names)
		}

		if len(prev.Primaries[k]) < 1 {
			delete(prev.Primaries, k)
		}
	}

	for k, cols := range curr.Uniques {
		var names []string
		var update bool

		mnames := map[string]int{}
		moldNames := map[string]int{}

		for i, col := range cols {
			mnames[col.Name] = i
			names = append(names, col.Name)
		}

		oldCols, ok := prev.Uniques[k]
		if !ok {
			goto UNIQUELOOPEND
		}

		for i, col := range oldCols {
			if _, ok := mnames[col.Name]; !ok {
				update = true
			}
			moldNames[col.Name] = i
		}

		for _, col := range cols {
			oldIndex, ok := moldNames[col.Name]
			if !ok {
				update = true
			}

			if l := len(prev.Uniques[k]); l > 1 {
				prev.Uniques[k] = append(prev.Uniques[k][:oldIndex], prev.Uniques[k][oldIndex+1:]...)
			} else if l == 1 {
				prev.Uniques[k] = []*Column{}
			}
		}

	UNIQUELOOPEND:
		if !ok {
			dialect.AddUnique(builder, k, cols[0].Table, names)
		} else if update {
			dialect.UpdateUnique(builder, k, cols[0].Table, names)
		}

		if len(prev.Uniques[k]) < 1 {
			delete(prev.Uniques, k)
		}

	}

	for k, col := range curr.Foreigns {
		oldcol := prev.Foreigns[k]
		if oldcol == nil {
			dialect.AddForeignKey(builder, col.Table, col.Name, col.Ref.Table, col.Ref.Name)
			goto FOREIGNLOOPEND
		}

		if !(oldcol.Table == col.Table && oldcol.Name == col.Name) {
			dialect.UpdateForeignKey(builder, col.Table, col.Name, col.Ref.Table, col.Ref.Name)
		}

	FOREIGNLOOPEND:
		delete(prev.Foreigns, k)
	}

	for k, ocol := range prev.Foreigns {
		if curr.Foreigns[k] == nil {
			dialect.DropForeignKey(builder, ocol.Table, ocol.Name)
		}
	}

	for k, cols := range prev.Primaries {
		if ncols, ok := curr.Primaries[k]; !ok || len(ncols) < 1 {
			dialect.DropPrimaryKey(builder, cols[0].Table)
		}
	}

	for k, cols := range prev.Uniques {
		if ncols, ok := curr.Uniques[k]; !ok || len(ncols) < 1 {
			dialect.DropUnique(builder, k, cols[0].Table)
		}
	}

	for _, enum := range prev.Enums {
		dialect.DropEnum(builder, enum.Name)
	}

	return builder.String()
}

func compareTables(wr io.Writer, dialect dialect.Dialect, tables, old map[string]map[string]*Column) {
	for tname, cols := range tables {
		if old[tname] == nil {
			addTable(wr, dialect, tname, tables[tname])
			goto TABLELOOPEND
		}

		for cname, col := range cols {
			oldcol, ok := old[tname][cname]
			if !ok {
				dialect.AddColumn(wr, col.Table, col.Name, col.Datatype.Type(), col.Size)
			}

			if !ok || col.raw == oldcol.raw {
				delete(old[tname], cname)
				goto COLLOOPEND
			}

			if col.rawtype != oldcol.rawtype || col.Size != oldcol.Size {
				dialect.UpdateColumn(wr, tname, cname, col.Datatype.Type(), col.Size)
			}

			if col.AutoIncement && !oldcol.AutoIncement {
				dialect.SetAutoIncrement(wr, tname, cname)
			} else if !col.AutoIncement && oldcol.AutoIncement {
				dialect.UnsetAutoIncrement(wr, tname, cname)
			}

			if col.NotNull != oldcol.NotNull {
				if col.NotNull {
					dialect.SetNotNull(wr, col.Table, col.Name)
				} else {
					dialect.DeleteNotNull(wr, col.Table, col.Name)
				}
			}

			if col.Default != oldcol.Default {
				if col.Default == `` {
					dialect.DropDefault(wr, col.Table, col.Name)
				} else {
					dialect.SetDefault(wr, col.Table, col.Name, col.Default)
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
			dialect.DropTable(wr, table)
			goto OLDTABLELOOPEND
		}

		for i, col := range cols {
			if tables[table][i] == nil {
				dialect.DropColumn(wr, table, col.Name)
			}
		}
	OLDTABLELOOPEND:
	}
}

func addTable(wr io.Writer, dialect dialect.Dialect, table string, cols map[string]*Column) {
	dialect.AddTable(wr, table, false)
	for _, col := range cols {
		addColumn(wr, dialect, col)
	}
}

func addColumn(wr io.Writer, dialect dialect.Dialect, col *Column) {
	dialect.AddColumn(wr, col.Table, col.Name, col.Datatype.Type(), col.Size)

	if col.AutoIncement {
		dialect.SetAutoIncrement(wr, col.Table, col.Name)
	}

	if col.NotNull {
		dialect.SetNotNull(wr, col.Table, col.Name)
	}

	if col.Default != `` {
		dialect.SetDefault(wr, col.Table, col.Name, col.Default)
	}

	if col.Check != `` {
		dialect.AddCheck(wr, col.Table, col.Name, col.Check)
	}
}
