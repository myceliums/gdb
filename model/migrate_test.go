package model

import (
	"database/sql"
	"os"
	"strings"
	"testing"

	"github.com/myceliums/assert"
	"github.com/myceliums/gdb/dialect"
)

func TestInitialSQL(t *testing.T) {
	x, as := initTest(t)

	dialect := dialect.GetByDriver(`postgres`)

	initial := InitialSQL(dialect, *x)
	as.Ne(``, initial)

	initial = checkAndTrimString(as, initial, `CREATE TYPE bond_type AS ENUM ('companion', 'fiance', 'spouce', 'friend')`)
	initial = checkAndTrimString(as, initial, `CREATE TABLE accounts();`)
	initial = checkAndTrimString(as, initial, `ALTER TABLE accounts ADD COLUMN id INT;`)
	//initial = checkAndTrimString(as, initial, `ALTER TABLE roles ALTER COLUMN id SET DATA TYPE SERIAL;`)
	initial = checkAndTrimString(as, initial, `CREATE SEQUENCE seq_roles_id;`)
	initial = checkAndTrimString(as, initial, `ALTER TABLE roles ALTER COLUMN id SET DEFAULT nextval('seq_roles_id'::regclass);`)
	initial = checkAndTrimString(as, initial, `ALTER TABLE relationships ADD CONSTRAINT fk_relationships_account_id FOREIGN KEY (account_id) REFERENCES accounts(id);`)
	initial = checkAndTrimString(as, initial, `ALTER TABLE relationships ADD COLUMN id INT;`)
	initial = checkAndTrimString(as, initial, `ALTER TABLE relationships ADD COLUMN verified_at TIMESTAMP;`)
	initial = checkAndTrimString(as, initial, `ALTER TABLE relationships ALTER COLUMN bond SET DEFAULT 'friend';`)
	initial = checkAndTrimString(as, initial, `ALTER TABLE relationships ADD COLUMN bond bond_type;`)

	if t.Failed() {
		t.Log(initial)
	}
}

func TestCompareSQL(t *testing.T) {
	x, as := initTest(t)

	nextMdl := initModel(t, testNextModel)
	as.Eq(5, len(nextMdl.Tables))

	var colCount int
	for t := range nextMdl.Tables {
		colCount += len(nextMdl.Tables[t])
	}

	as.Eq(22, colCount)

	dialect := dialect.GetByDriver(`postgres`)

	sq := UpgradeSQL(dialect, *x, *nextMdl)
	sq = checkAndTrimString(as, sq, `ALTER TABLE accounts ADD COLUMN bio TEXT;`)
	sq = checkAndTrimString(as, sq, `ALTER TABLE relationships DROP COLUMN bond;`)
	sq = checkAndTrimString(as, sq, `CREATE TABLE posts();`)
	sq = checkAndTrimString(as, sq, `ALTER TABLE posts ADD COLUMN id INT;`)
	sq = checkAndTrimString(as, sq, `ALTER TABLE posts ADD COLUMN type post_type;`)
	sq = checkAndTrimString(as, sq, `ALTER TABLE posts ALTER COLUMN type SET DEFAULT 'general';`)
	sq = checkAndTrimString(as, sq, `ALTER TABLE posts ADD COLUMN created_by INT;`)
	sq = checkAndTrimString(as, sq, `ALTER TABLE posts ADD CONSTRAINT fk_posts_created_by FOREIGN KEY (created_by) REFERENCES accounts(id);`)
	sq = checkAndTrimString(as, sq, `ALTER TABLE posts ALTER COLUMN created_by SET NOT NULL;`)
	sq = checkAndTrimString(as, sq, `ALTER TABLE posts ADD COLUMN created_at TIMESTAMP;`)
	sq = checkAndTrimString(as, sq, `ALTER TABLE posts ALTER COLUMN created_at SET DEFAULT NOW();`)
	sq = checkAndTrimString(as, sq, `ALTER TABLE posts ADD COLUMN context TEXT;`)
	sq = checkAndTrimString(as, sq, `ALTER TABLE posts ALTER COLUMN context SET NOT NULL;`)
	sq = checkAndTrimString(as, sq, `DROP TYPE bond_type;`)

	if t.Failed() {
		t.Log(sq)
	}
}

func TestMigrate(t *testing.T) {
	db, err := sql.Open(`postgres`, os.Getenv(`TEST_DB_CONNECTION_STRING`))
	if err != nil {
		t.Error(err)
		t.Fail()
	}

	x, as := initTest(t)
	dialect := dialect.GetByDriver(`postgres`)

	as.NoError(Migrate(dialect, db, *x))

	nextMdl := initModel(t, testNextModel)
	as.NoError(Migrate(dialect, db, *nextMdl))
}

func checkAndTrimString(as assert.Assert, str, check string) string {
	as.True(strings.Contains(str, check), "expected '", check, "' but not found")
	return strings.Trim(str, check)
}
