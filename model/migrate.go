package model

import (
	"database/sql"
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

	var migration string
	if version == 0 {
		migration = InitialSQL(dialect, mdl)
	}

	if version > 0 {
		oldMdl, err := New(storedConfig)
		if err != nil {
			return err
		}

		migration = UpgradeSQL(dialect, *oldMdl, mdl)

		q = dialect.InsertVersion()
		if _, err := tx.Exec(q, version+1, mdl.conf); err != nil {
			return err
		}
	}

	if _, err := tx.Exec(migration); err != nil {
		return err
	}

	return tx.Commit()
}

// InitialSQL returns the sql to model the database after the given configuration.
func InitialSQL(dialect dialect.Dialect, mdl Model) (q string) {
	for _, enum := range mdl.Enums {
		q += dialect.AddEnum(enum.Name, enum.Values)
	}

	for table, columns := range mdl.Tables {
		q += addTable(dialect, table, columns)
	}

	for _, cols := range mdl.Primaries {
		var colNames []string
		for _, col := range cols {
			colNames = append(colNames, col.Name)
		}
		q += dialect.AddPrimaryKey(*cols[0].Table, colNames)
	}

	for id, cols := range mdl.Uniques {
		var colNames []string
		for _, col := range cols {
			colNames = append(colNames, col.Name)
		}
		q += dialect.AddUnique(id, *cols[0].Table, colNames)
	}

	for _, col := range mdl.Foreigns {
		q += dialect.AddForeignKey(*col.Table, col.Name, *col.Ref.Table, col.Ref.Name)
	}

	return q
}

// UpgradeSQL returns the sql which resolves the differential safely between 2 models
func UpgradeSQL(dialect dialect.Dialect, prev, curr Model) (q string) {
	for n, enum := range curr.Enums {
		osval := strings.Join(prev.Enums[n].Values, ` `)
		sval := strings.Join(enum.Values, ` `)

		if osval == sval {
			break
		}

		if sval == `` && sval != osval {
			q += dialect.DropEnum(prev.Enums[n].Name)
			break
		} else if sval != `` && osval == `` {
			q += dialect.AddEnum(enum.Name, enum.Values)
			break
		}

		for _, val := range enum.Values {
			if !strings.Contains(osval, val) {
				q += dialect.AppendEnum(n, val)
				osval = strings.Trim(osval, val)
				osval = strings.TrimSuffix(osval, ` `)
			}
		}

		for _, delval := range strings.Split(osval, ` `) {
			q += dialect.DeleteEnum(n, delval)
		}
	}

	q += compareTables(dialect, curr.Tables, prev.Tables)

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
			q += dialect.DropPrimaryKey(*cols[0].Table)
		} else if len(names) > 0 && len(oldnames) < 1 {
			q += dialect.AddPrimaryKey(*cols[0].Table, names)
		} else if len(names) > 0 && strings.Join(names, ` `) != strings.Join(oldnames, ` `) {
			q += dialect.UpdatePrimaryKey(*cols[0].Table, names)
		}

		if len(prev.Primaries[k]) < 1 {
			delete(prev.Primaries, k)
		}
	}

	for k, cols := range prev.Primaries {
		if len(curr.Primaries[k]) < 1 {
			q += dialect.DropPrimaryKey(*cols[0].Table)
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
			q += dialect.DropUnique(k, *cols[0].Table)
		} else if len(names) > 0 && len(oldnames) < 1 {
			q += dialect.AddUnique(k, *cols[0].Table, names)
		} else if len(names) > 0 && strings.Join(names, ` `) != strings.Join(oldnames, ` `) {
			q += dialect.UpdateUnique(k, *cols[0].Table, names)
		}

	}

	for k, cols := range prev.Uniques {
		if len(curr.Uniques[k]) < 1 {
			q += dialect.DropUnique(k, *cols[0].Table)
		}
	}

	for k, col := range curr.Foreigns {
		oldcol := prev.Foreigns[k]
		if oldcol == nil {
			q += dialect.AddForeignKey(*col.Table, col.Name, *col.Ref.Table, col.Ref.Name)
			break
		}

		if !(*oldcol.Table == *col.Table && oldcol.Name == col.Name) {
			q += dialect.UpdateForeignKey(*col.Table, col.Name, *col.Ref.Table, col.Ref.Name)
		}

		delete(prev.Foreigns, k)
	}

	for k, ocol := range prev.Foreigns {
		if curr.Foreigns[k] == nil {
			q += dialect.DropForeignKey(*ocol.Table, ocol.Name)
		}
	}

	return q
}

func compareTables(dialect dialect.Dialect, tables, old map[string]map[string]*Column) (q string) {
	for tname, cols := range tables {
		if old[tname] == nil {
			q += addTable(dialect, table, tables[tname])
			delete(old, tname)
			break
		}

		for cname, col := range cols {
			oldcol, ok := old[tname][cname]
			if !ok {
				q += dialect.AddColumn(*col.Table, col.Name, col.Datatype.Type(), col.Size)
			}

			if !ok || col.raw == oldcol.raw {
				delete(old[tname], cname)
				break
			}

			if col.rawtype != old[tname][cname].rawtype || col.Size != old[tname][cname].Size {
				q += dialect.UpdateColumn(tname, cname, col.Datatype.Type(), col.Size)
			}

			if col.AutoIncement && !old[tname][cname].AutoIncement {
				q += dialect.SetAutoIncrement(tname, cname)
			} else if !col.AutoIncement && old[tname][cname].AutoIncement {
				q += dialect.UnsetAutoIncrement(tname, cname)
			}

			if col.NotNull != old[tname][cname].NotNull {
				if col.NotNull {
					q += dialect.SetNotNull(*col.Table, col.Name)
				} else {
					q += dialect.DeleteNotNull(*col.Table, col.Name)
				}
			}

			if col.Default != old[tname][cname].Default {
				if col.Default == `` {
					q += dialect.DropDefault(*col.Table, col.Name)
				} else {
					q += dialect.SetDefault(*col.Table, col.Name, col.Default)
				}
			}

			delete(old[tname], cname)
		}

		if len(old[table]) < 1 {
			delete(old, table)
		}
	}

	for table, cols := range old {
		if len(tables[table]) == 0 {
			q += dialect.DropTable(table)
			break
		}

		for i, col := range cols {
			if tables[table][i] == nil {
				q += dialect.DropColumn(table, col.Name)
			}
		}
	}

	return
}

func addTable(dialect dialect.Dialect, table string, cols map[string]*Column) (q string) {
	q += dialect.AddTable(table, false)
	for _, col := range cols {
		q += addColumn(dialect, col)
	}

	return
}

func addColumn(dialect dialect.Dialect, col *Column) (q string) {
	q += dialect.AddColumn(*col.Table, col.Name, col.Datatype.Type(), col.Size)

	if col.AutoIncement {
		q += dialect.SetAutoIncrement(*col.Table, col.Name)
	}

	if col.NotNull {
		q += dialect.SetNotNull(*col.Table, col.Name)
	}

	if col.Default != `` {
		q += dialect.SetDefault(*col.Table, col.Name, col.Default)
	}

	if col.Check != `` {
		q += dialect.AddCheck(*col.Table, col.Name, col.Check)
	}

	return
}
