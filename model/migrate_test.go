package model

import (
	"strings"
	"testing"

	"github.com/myceliums/gdb/dialect"
)

func TestInitialSQL(t *testing.T) {
	x, as := initTest(t)

	dialect := dialect.GetByDriver(`postgres`)

	initial := InitialSQL(dialect, *x)
	as.Ne(``, initial)

	as.True(strings.Contains(initial, `CREATE TYPE bond_type AS ENUM ('companion', 'fiance', 'spouce', 'friend')`))
	as.True(strings.Contains(initial, `CREATE TABLE accounts();`))
	as.True(strings.Contains(initial, `ALTER TABLE accounts ADD COLUMN id INT;`))
	as.True(strings.Contains(initial, `ALTER TABLE roles ALTER COLUMN id SET DATA TYPE SERIAL;`))
	as.True(strings.Contains(initial, `ALTER TABLE relationships ADD CONSTRAINT fk_relationships_account_id FOREIGN KEY (account_id) REFERENCES accounts(id);`))
	as.True(strings.Contains(initial, `ALTER TABLE relationships ADD COLUMN id INT;`))
	as.True(strings.Contains(initial, `ALTER TABLE relationships ADD COLUMN verified_at TIMESTAMP;`))
	as.True(strings.Contains(initial, `ALTER TABLE relationships ALTER COLUMN bond SET DEFAULT 'friend';`))
	as.True(strings.Contains(initial, `ALTER TABLE relationships ADD COLUMN bond bond_type;`))

	if t.Failed() {
		t.Log(initial)
	}
}

func TestCompareSQL(t *testing.T) {
	x, as := initTest(t)

	nextMdl, err := New(testNextModel)
	as.NoError(err)
	if err != nil {
		t.FailNow()
	}

	dialect := dialect.GetByDriver(`postgres`)

	sq := UpgradeSQL(dialect, *x, *nextMdl)

	as.True(strings.Contains(sq, `CREATE TABLE posts();`))
	as.True(strings.Contains(sq, `ALTER TABLE accounts ADD COLUMN bio TEXT;`), "expected accounts.bio not added")
	as.True(strings.Contains(sq, `ALTER TABLE posts ADD CONSTRAINT fk_posts_created_by FOREIGN KEY (created_by) REFERENCES accounts(id);`))

	if t.Failed() {
		t.Log(sq)
	}
}
