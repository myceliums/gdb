package query

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"

	"github.com/myceliums/gdb/dialect"
)

const (
	JoinTypeInner = iota + 1
	JoinTypeOuter
	JoinTypeLeft
	JoinTypeRight
)

// Querier is an interface that accepts both *sql.DB and *sql.Tx
type Querier interface {
	Exec(string, ...interface{}) (sql.Result, error)
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)

	Query(string, ...interface{}) (*sql.Rows, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)

	QueryRow(string, ...interface{}) *sql.Row
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row

	Prepare(string) (*sql.Stmt, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
}

// Field contains the field identifiers
type Field interface {
	io.Reader
	Args() []interface{}
}

// Column describes a table column
type Column struct {
	Table string
	Name  string
}

// Read is an implementation of Field
func (x Column) Read(in bytes) (int, error) {
	return copy(in, []byte(x.Name)), nil
}

// Args is an implementation of Field
func (x Column) Args() []interface{} {
	return []interface{}{}
}

type where struct {
	col        Column
	comparitor string
	val        Field
}

func (x Column) Reader(in []byte) (int, error) {
	s := x.Table + "." + x.Name
	return copy(in, []byte(s)), nil
}

type UpdateQuery interface {
	Set(col Column, val interface{}) UpdateQuery
	Where(qc ...Condition) UpdateQuery
	Join(col, ref Column) UpdateQuery
	InnerJoin(col, ref Column) UpdateQuery
	OuterJoin(col, ref Column) UpdateQuery
	LeftJoin(col, ref Column) UpdateQuery
	RightJoin(col, ref Column) UpdateQuery

	Exec(qr Querier) (sql.Result, error)
	ExecContext(ctx context.Context, qr Querier) (sql.Result, error)
}

type DeleteQuery interface {
	Where(qc ...Condition) UpdateQuery
	Join(col, ref Column) UpdateQuery
	InnerJoin(col, ref Column) UpdateQuery
	OuterJoin(col, ref Column) UpdateQuery
	LeftJoin(col, ref Column) UpdateQuery
	RightJoin(col, ref Column) UpdateQuery

	Exec(qr Querier) (sql.Result, error)
	ExecContext(ctx context.Context, qr Querier) (sql.Result, error)
}

//func NewInsert(dialect dialect.Dialect, table string, columns ...string) InsertQuery {
//	x := newQuery(dialect)
//	x.dia.Insert(x.wr, table, columns)
//
//	return x
//}
//
//func NewUpdate(dialect dialect.Dialect, table string) UpdateQuery {
//	x := newQuery(dialect)
//	x.dia.Update(x.wr, table)
//
//	return x
//}
//
//func NewDelete(dialect dialect.Dialect, table string) DeleteQuery {
//	x := newQuery(dialect)
//	x.dia.Delete(x.wr, table)
//
//	return x
//}
//
//func newQuery(dialect dialect.Dialect) *Query {
//	x := &Query{}
//	x.dia = dialect
//	x.wr = &bytes.Buffer{}
//
//	return x
//}

// Query contains all the metadata to compile an sql query
type Query struct {
	wr   *bytes.Buffer
	dia  dialect.Dialect
	args []interface{}
}

// Read is an implementation of io.Reader
func (x Query) Read(in []byte) (int, error) {
	return x.wr.Read(in)
}

func (x Query) String() string {
	return x.wr.String()
}

// Join adds a JOIN statement to the query
func (x *Query) Join(col, ref Column) *Query {
	return x.InnerJoin(col, ref)
}

// InnerJoin adds a INNER JOIN statement to the query
func (x *Query) InnerJoin(col, ref Column) *Query {
	x.dia.InnerJoin(x.wr, col.Table, col.Name, ref.Table, ref.Name)

	return x
}

//// Where adds a WHERE statement to the query
//func (x *Query) Where(qc Condition) *Query {
//	x.dia.Where(x.wr, qc)
//
//	return x
//}

// Add adds your curstom piece of query to the query
func (x *Query) Add(q string, args ...interface{}) *Query {
	x.wr.WriteString(q)
	x.args = append(x.args, args...)

	return x
}

// ExecContext execute queries without returning rows.
func (x Query) ExecContext(ctx context.Context, qr Querier) (sql.Result, error) {
	return qr.ExecContext(ctx, fmt.Sprint(x), x.args...)
}

// Exec execute queries without returning rows.
func (x Query) Exec(qr Querier) (sql.Result, error) {
	return x.ExecContext(context.Background(), qr)
}

// QueryContext executes a query that returns rows, typically a SELECT.
func (x Query) QueryContext(ctx context.Context, qr Querier) (*sql.Rows, error) {
	return qr.QueryContext(ctx, fmt.Sprint(x), x.args...)
}

// Query executes a query that returns rows, typically a SELECT.
func (x Query) Query(qr Querier) (*sql.Rows, error) {
	return x.QueryContext(context.Background(), qr)
}

// QueryRowContext executes a query that returns one row. typically a SELECT
func (x Query) QueryRowContext(ctx context.Context, qr Querier) *sql.Row {
	return qr.QueryRowContext(ctx, x.wr.String(), x.args...)
}

// QueryRowContext executes a query that returns one row. typically a SELECT
func (x Query) QueryRow(qr Querier) *sql.Row {
	return x.QueryRowContext(context.Background(), qr)
}
