package dialect

import "io"

// Dialect is a parser that transforms the given arguments
// of its functions into an SQL statement of the given dialect
type Dialect interface {
	AddVersionTable() string
	CheckVersion() string
	InsertVersion() string

	AddTable(wr io.Writer, name string, ifnotexists bool)
	DropTable(wr io.Writer, name string)
	AddColumn(wr io.Writer, table, column, typename string, size int)
	UpdateColumn(wr io.Writer, table, colum, typename string, size int)
	DropColumn(wr io.Writer, table, column string)
	AddPrimaryKey(wr io.Writer, table string, column []string)
	UpdatePrimaryKey(wr io.Writer, table string, column []string)
	DropPrimaryKey(wr io.Writer, table string)
	AddForeignKey(wr io.Writer, table, columnName, referenceTable, referenceColumn string)
	UpdateForeignKey(wr io.Writer, table, columnName, referenceTable, refrenceColumn string)
	DropForeignKey(wr io.Writer, table, columnName string)
	AddUnique(wr io.Writer, id, table string, column []string)
	UpdateUnique(wr io.Writer, id, table string, column []string)
	DropUnique(wr io.Writer, id, table string)
	SetNotNull(wr io.Writer, table, column string)
	DeleteNotNull(wr io.Writer, table, column string)
	AddCheck(wr io.Writer, table, column, check string)
	UpdateCheck(wr io.Writer, table, column, check string)
	DropCheck(wr io.Writer, table, column string)
	AddEnum(wr io.Writer, name string, values []string)
	AppendEnum(wr io.Writer, name, values string)
	DropEnum(wr io.Writer, name string)
	SetDefault(wr io.Writer, table, column, value string)
	DropDefault(wr io.Writer, table, column string)
	SetAutoIncrement(wr io.Writer, table, column string)
	UnsetAutoIncrement(wr io.Writer, table, column string)

	Select(wr io.Writer, table string, columns []string)
	Insert(wr io.Writer, table string, columns []string)
	Values(wr io.Writer, vals string)
	Update(wr io.Writer, table string)
	Set(wr io.Writer, column string, sq io.Reader)
	Delete(wr io.Writer, table string)
	Where(wr io.Writer, qc io.Writer)
	InnerJoin(wr io.Writer, table, column, refTable, refColumn string)
	OuterJoin(wr io.Writer, table, column, refTable, refColumn string)
	LeftJoin(wr io.Writer, table, column, refTable, refColumn string)
	RightJoin(wr io.Writer, table, column, refTable, refColumn string)
	Returning(wr io.Writer, columns []string)
	//And(wr io.Writer, stmt1, stmt2 string)
	//Or(wr io.Writer, stmt1, stmt2 string)
	Limit(wr io.Writer, l int)
	//Offset(wr io.Writer, l int)

	Cursor(query io.Reader, key string) string
	CursorSelect(key string) string
	CursorClose(key string) string
}

// GetByDriver returns a dialect of the given driver
func GetByDriver(driver string) Dialect {
	switch driver {
	case "postgres":
		return new(Postgres)
	}

	return nil
}
