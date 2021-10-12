package model

import (
	"testing"

	_ "embed"

	_ "github.com/lib/pq"
	"github.com/myceliums/assert"
)

//go:embed testmodel.yml
var testModel []byte

//go:embed testnextmodel.yml
var testNextModel []byte

func TestNew(t *testing.T) {
	x, as := initTest(t)

	as.Eq(1, len(x.Enums))
	as.Eq(4, len(x.Tables))
	as.Eq(6, len(x.Tables[`accounts`]))

	for table, cols := range x.Tables {
		for _, col := range cols {
			switch table {
			case `accounts`:
				switch col.Name {
				case `id`:
					as.Eq(`serial`, col.rawtype)
					as.Eq(`int`, col.Datatype.Type())
					as.Eq(0, col.Size)
					as.Eq(false, col.NotNull)
				case `username`:
					as.Eq(`varchar`, col.rawtype)
					as.Eq(50, col.Size)
					as.Eq(true, col.NotNull)
				case `created_at`:
					as.Eq(`timestamp`, col.rawtype)
					as.Eq(`NOW()`, col.Default)
				}
			case `account_roles`:
				if col.Name == `account.id` {
					as.Eq(`accounts.id`, col.rawtype)
				}
			}
		}
	}

	as.Eq(1, len(x.Primaries[`accounts`]))
	as.Eq(4, len(x.Foreigns))
}

func initTest(t *testing.T) (*Model, assert.Assert) {
	as := assert.New(t)
	x := initModel(t, testModel)

	return x, as
}

func initModel(t *testing.T, in []byte) *Model {
	x, err := New(in)
	if err != nil {
		t.Errorf("error reading given test configuration: %v", err)
		t.FailNow()
	}

	return x
}
