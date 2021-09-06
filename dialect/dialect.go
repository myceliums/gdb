package dialect

type Dialect interface {
	Type(name string, size int) string
	AddColumn(table, column, typename string, size int) string
	DropColumn(table, column string) string
	AddPrimaryKey(table string, column ...string) string
	UpdatePrimaryKey(table string, column ...string) string
	DropPrimaryKey(table, column ...string) string
	AddForeignKey(table, columnName, referenceTable, referenceColumn string) string
	UpdateForeignKey(table, columnName, referenceTable, refrenceColumn string) string
	DropForeignKey(table, columnName string) string
	AddUnique(table, column string) string
	DeleteUnique(table, column string) string
	AddMultiUnique(id, table string, columns []string) string
	UpdateMultiUnique(id string, colums []string) string
	DeleteMultiUnique(id string) string
}
