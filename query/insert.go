package query

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/myceliums/gdb/dialect"
)

func NewInsertQuery(dialect dialect.Dialect, table string, columns ...string) *InsertQuery {
	x := &InsertQuery{}
	x.dia = dialect
	x.wr = &bytes.Buffer{}
	x.dia.Insert(x.wr, table, columns)

	return x
}

type InsertQuery struct {
	wr   *bytes.Buffer
	dia  dialect.Dialect
	args []interface{}
}

// Read
func (x InsertQuery) Read(in []byte) (int, error) {
	return x.wr.Read(in)
}

// String in the implementation of fmt.Stringer
// and returns the full writen query
func (x InsertQuery) String() string {
	return x.wr.String()
}

// Args return all the arguments given in the query in order
func (x InsertQuery) Args() []interface{} {
	return x.args
}

func (x *InsertQuery) Values(vals ...interface{}) *InsertQuery {
	var s []string
	for _, val := range vals {
		switch t := val.(type) {
		case Field:
			s = append(s, fmt.Sprint(t))
			x.args = append(x.args, t.Args()...)
		default:
			s = append(s, "?")
			x.args = append(x.args, val)
		}
	}

	x.dia.Values(x.wr, strings.Join(s, `, `))

	return x
}

// ExecContext executes the query
func (x InsertQuery) ExecContext(ctx context.Context, qr Querier) (sql.Result, error) {
	return qr.ExecContext(ctx, x.String()+";", x.Args()...)
}

// Exec executes the query
func (x InsertQuery) Exec(qr Querier) (sql.Result, error) {
	return x.ExecContext(context.Background(), qr)
}

func (x InsertQuery) ReturningContext(ctx context.Context, qr Querier, columns ...Column) *sql.Row {
	var names []string
	for _, col := range columns {
		names = append(names, fmt.Sprintf("%s.%s", col.Table, col.Name))
	}

	x.dia.Returning(x.wr, names)
	return qr.QueryRowContext(ctx, x.String())
}

func (x InsertQuery) Returning(qr Querier, columns ...Column) *sql.Row {
	return x.ReturningContext(context.Background(), qr, columns...)
}
