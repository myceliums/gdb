package migrate

import (
	"database/sql"

	"github.com/myceliums/gdb/dialect"
	"github.com/myceliums/gdb/model"
)

func initialMigrate(dialect dialect.Dialect, tx *sql.Tx, mdl model.Model) error {
	var q, defaults, notnulls, checks string

	for table, columns := range mdl.Tables {
		q += dialect.AddTable(table, false)

		for _, col := range columns {
			q += dialect.AddColumn(table, col.Name, col.Datatype.Type(), col.Size)

			if col.NotNull {
				notnulls += dialect.AddNotNull(table, col.Name)
			}

			if col.Default != `` {
				defaults += dialect.AddDefault(table, column, col.Default)
			}

			if col.Check != `` {
				checks += dialect.AddCheck(table, column, col.Check)
			}
		}
	}

	return nil
}
