package migrate

import (
	"database/sql"

	"github.com/myceliums/gdb/dialect"
	"github.com/myceliums/gdb/model"
)

func diffMigrate(dialect dialect.Dialect, tx *sql.Tx, newMdl, oldMdl model.Model) error {

	return nil
}
