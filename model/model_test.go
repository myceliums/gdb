package model

import (
	"testing"

	_ "embed"

	"github.com/myceliums/assert"
)

//go:embed testmodel.yml
var testModel []byte

func TestNew(t *testing.T) {
	as := assert.New(t)

	x, err := New(testModel)
	as.NoError(err)

	as.Eq(1, len(x.Enums))
	as.Eq(4, len(x.Tables))
	as.Eq(6, len(x.Tables[`accounts`]))

	for table, cols := range x.Tables {
		for _, col := range cols {
			switch table {
			case `accounts`:
				switch col.Name {
				case `id`:
					as.Eq(`int`, col.rawtype)
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
			}
		}
	}
}
