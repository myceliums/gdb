package dialect

type Dialect interface {
	Type(name string, size int) string
	AddTable(name string) string
	DropTable(name string) string
	AddColumn(table, column, typename string, size int) string
	DropColumn(table, column string) string
	AddPrimaryKey(table string, column ...string) string
	UpdatePrimaryKey(table string, column ...string) string
	DropPrimaryKey(table string, column ...string) string
	AddForeignKey(table, columnName, referenceTable, referenceColumn string) string
	UpdateForeignKey(table, columnName, referenceTable, refrenceColumn string) string
	DropForeignKey(table, columnName string) string
	AddUnique(id, table string, column ...string) string
	UpdateUnique(id, table string, column ...string) string
	DeleteUnique(id, table string) string
	AddNotNull(table, column string) string
	DeleteNotNull(table, column string) string
	AddCheck(table, column, check string) string
	UpdateCheck(table, column, check string) string
	DeleteCheck(table, column string) string
	AddEnum(name string, values []string) string
	UpdateEnum(name string, values []string) string
	DeleteEnum(name string) string

	//Select(table string, columns ...string) string
	//And(stmt1, stmt2 string) string
	//Or(stmt1, stmt2 string) string
}
