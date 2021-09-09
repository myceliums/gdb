// Package migrate is responsible for the migrations using the datamodel
package migrate

import (
	"database/sql"

	"github.com/myceliums/gdb/dialect"
	"github.com/myceliums/gdb/model"
)

// TODO think about encrypting the config in the database and decrypting with a key

// Migrate runs the configured migration if there's any changes between the
// last inserted configuration in the database
func Migrate(dialect dialect.Dialect, db *sql.DB, cfg []byte) error {
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
	if err := tx.QueryRow(q).Scan(&verion, &storedConfig); err != nil && err != sql.ErrNoRows {
		return err
	}

	newMdl := model.New(cfg)
	if version == 0 {
		if err := initialMigrate(dialect, tx, newMdl); err != nil {
			return err
		}
	} else {
		oldMdl := model.New(storedConfig)
		if err := diffMigrate(dialect, tx, newMdl, oldMdl); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
