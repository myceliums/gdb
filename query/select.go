package query

import (
	"context"
	"database/sql"
)

type SelectQuery Query

func (x SelectQuery) Read(in []byte) (int, error) {
	return x.wr.Read(in)
}

func (x SelectQuery) Args() []interface{} {
	return x.args
}

// Limit adds a LIMIT statement to the query
func (x *SelectQuery) Limit(i int) *SelectQuery {
	x.dia.Limit(x.wr, i)

	return x
}

func (x SelectQuery) Query(qr Querier) (*sql.Rows, error) {
	return x.QueryContext(context.Background(), qr)
}

func (x SelectQuery) QueryContext(ctx context.Context, qr Querier) (*sql.Rows, error) {
	return qr.QueryContext(ctx, x.wr.String(), x.args...)
}

func (x SelectQuery) QueryRow(qr Querier) *sql.Row {
	return x.QueryRowContext(context.Background(), qr)
}

// QueryRowContext returns a row
func (x SelectQuery) QueryRowContext(ctx context.Context, qr Querier) *sql.Row {
	return qr.QueryRowContext(ctx, x.wr.String()+";", x.args...)
}

// Cursor creates a DECLARE CURSOR statment and excetutes the statement and returns the Cursor
func (x SelectQuery) Cursor(qr Querier) (*Cursor, error) {
	return newCursor(x.dia, qr, x)
}

func (x SelectQuery) CursorContext(ctx context.Context, qr Querier) (*Cursor, error) {
	return newCursorContext(ctx, x.dia, qr, x)
}
