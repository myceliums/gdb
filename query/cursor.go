package query

import (
	"context"

	"github.com/google/uuid"
	"github.com/myceliums/gdb/dialect"
)

func newCursor(dialect dialect.Dialect, qr Querier, query Field) (*Cursor, error) {
	return newCursorContext(context.Background(), dialect, qr, query)
}

func newCursorContext(ctx context.Context, dialect dialect.Dialect, qr Querier, query Field) (*Cursor, error) {
	x := &Cursor{}
	x.qr = qr
	x.dia = dialect
	x.field = query
	x.key = uuid.New().String()

	if _, err := x.qr.ExecContext(ctx, x.dia.Cursor(x.field, x.key)); err != nil {
		return nil, err
	}

	return x, nil
}

// Cursor is a Query statment that holds its position on a particular row of the given results
type Cursor struct {
	key   string
	dia   dialect.Dialect
	field Field
	qr    Querier
}

func (x Cursor) Scan(args ...interface{}) error {
	return x.qr.QueryRow(x.dia.CursorSelect(x.key), x.field.Args()...).Scan(args...)
}

func (x Cursor) Close() error {
	_, err := x.qr.Exec(x.dia.CursorClose(x.key))
	return err
}
