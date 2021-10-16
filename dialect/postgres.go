package dialect

import (
	"fmt"
	"io"
	"strings"
)

type Postgres string

func (x Postgres) Type(name string, size int) string {
	switch name {
	case `varchar`, `string`, `charactervarying`:
		name = `VARCHAR`
	case `int`, `smallint`, `bigint`, `float`, `timestamp`, `boolean`, `double`, `text`:
		name = strings.ToUpper(name)
	}
	i := name
	if (name == `VARCHAR` || name == `INT`) && size > 0 {
		i = fmt.Sprintf("%s(%d)", i, size)
	}

	return i
}

func (x Postgres) AddTable(wr io.Writer, name string, ifnotexists bool) {
	i := "CREATE TABLE "
	if ifnotexists {
		i += "IF NOT EXISTS "
	}
	i += "%s();\n"
	fmt.Fprintf(wr, i, name) // nolint: errcheck
}

func (x Postgres) DropTable(wr io.Writer, name string) {
	fmt.Fprintf(wr, "DROP TABLE %s CASCADE;\n", name) // nolint: errcheck
}

func (x Postgres) AddColumn(wr io.Writer, table, column, typename string, size int) {
	fmt.Fprintf(wr, "ALTER TABLE %s ADD COLUMN %s %s;\n", table, column, x.Type(typename, size)) // nolint: errcheck
}

func (x Postgres) UpdateColumn(wr io.Writer, table, column, typename string, size int) {
	fmt.Fprintf(wr, "ALTER TABLE %s ALTER COLUMN %s TYPE %s;\n", table, column, x.Type(typename, size)) // nolint: errcheck
}

func (x Postgres) DropColumn(wr io.Writer, table, column string) {
	fmt.Fprintf(wr, "ALTER TABLE %s DROP COLUMN %s;\n", table, column) // nolint: errcheck
}

func (x Postgres) AddPrimaryKey(wr io.Writer, table string, columns []string) {
	fmt.Fprintf(wr, "ALTER TABLE %s ADD CONSTRAINT pk_%s PRIMARY KEY(%s);\n", table, table, strings.Join(columns, `, `)) // nolint: errcheck
}

func (x Postgres) UpdatePrimaryKey(wr io.Writer, table string, columns []string) {
	x.DropPrimaryKey(wr, table)
	x.AddPrimaryKey(wr, table, columns)
}

func (x Postgres) DropPrimaryKey(wr io.Writer, table string) {
	fmt.Fprintf(wr, "ALTER TABLE %s DROP CONSTRAINT pk_%s;\n", table, table) // nolint: errcheck
}

func (x Postgres) AddForeignKey(wr io.Writer, table, column, referenceTable, referenceColumn string) {
	fmt.Fprintf(wr, "ALTER TABLE %s ADD CONSTRAINT fk_%s_%s FOREIGN KEY (%s) REFERENCES %s(%s);\n", table, table, column, column, referenceTable, referenceColumn) // nolint: errcheck
}

func (x Postgres) UpdateForeignKey(wr io.Writer, table, column, refrerenceTable, referenceColumn string) {
	x.DropForeignKey(wr, table, column)
	x.AddForeignKey(wr, table, column, refrerenceTable, referenceColumn)
}

func (x Postgres) DropForeignKey(wr io.Writer, table, column string) {
	fmt.Fprintf(wr, "ALTER TABLE %s DROP CONSTRAINT fk_%s_%s;\n", table, table, column) // nolint: errcheck
}

func (x Postgres) AddUnique(wr io.Writer, id, table string, columns []string) {
	fmt.Fprintf(wr, "ALTER TABLE %s ADD CONSTRAINT uq_%s UNIQUE(%s);\n", table, id, strings.Join(columns, `, `)) // nolint: errcheck
}

func (x Postgres) UpdateUnique(wr io.Writer, id, table string, columns []string) {
	x.DropUnique(wr, id, table)
	x.AddUnique(wr, id, table, columns)
}

func (x Postgres) DropUnique(wr io.Writer, id, table string) {
	fmt.Fprintf(wr, "ALTER TABLE %s DROP uq_%s;\n", table, id) // nolint: errcheck
}

func (x Postgres) SetNotNull(wr io.Writer, table, column string) {
	fmt.Fprintf(wr, "ALTER TABLE %s ALTER COLUMN %s SET NOT NULL;\n", table, column) // nolint: errcheck
}

func (x Postgres) DeleteNotNull(wr io.Writer, table, column string) {
	fmt.Fprintf(wr, "ALTER TABLE %s ALTER COLUMN %s DROP NOT NULL;\n", table, column) // nolint: errhcheck
}

func (x Postgres) AddCheck(wr io.Writer, table, column, check string) {
	fmt.Fprintf(wr, "ALTER TABLE %s ADD CONSTRAINT ch_%s_%s CHECK(%s);\n", table, table, column, check) // nolint: errcheck
}

func (x Postgres) UpdateCheck(wr io.Writer, table, column, check string) {
	x.DropCheck(wr, table, column)
	x.AddCheck(wr, table, column, check)
}

func (x Postgres) DropCheck(wr io.Writer, table, column string) {
	fmt.Fprintf(wr, "ALTER TABLE %s DROP CONSTRAINT ch_%s_%s;\n", table, table, column) // nolint: errcheck
}

func (x Postgres) AddEnum(wr io.Writer, name string, values []string) {
	fmt.Fprintf(wr, "CREATE TYPE %s AS ENUM ('%s');\n", name, strings.Join(values, `', '`)) // nolint: errcheck
}

func (x Postgres) AppendEnum(wr io.Writer, name, value string) {
	fmt.Fprintf(wr, "ALTER TYPE %s ADD VALUE '%s';\n", name, value) // nolint: errcheck
}

func (x Postgres) DropEnum(wr io.Writer, name string) {
	fmt.Fprintf(wr, "DROP TYPE %s;\n", name) // nolint: errcheck
}

func (x Postgres) SetDefault(wr io.Writer, table, column, value string) {
	fmt.Fprintf(wr, "ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s;\n", table, column, value) // nolint: errcheck
}

func (x Postgres) DropDefault(wr io.Writer, table, column string) {
	fmt.Fprintf(wr, "ALTER TABLE %s ALTER COLUMN %s DROP DEFAULT;\n", table, column) // nolint: errcheck
}

func (x Postgres) SetAutoIncrement(wr io.Writer, table, column string) {
	fmt.Fprintf(wr, "CREATE SEQUENCE seq_%s_%s;\n", table, column)                                                                // nolint: errcheck
	fmt.Fprintf(wr, "SELECT setval('seq_%s_%s', (SELECT max(%s) FROM %s));\n", table, column, column, table)                      // nolint: errcheck
	fmt.Fprintf(wr, "ALTER TABLE %s ALTER COLUMN %s SET DEFAULT nextval('seq_%s_%s'::regclass);\n", table, column, table, column) // nolint: errcheck
}

func (x Postgres) UnsetAutoIncrement(wr io.Writer, table, column string) {
	fmt.Fprintf(wr, "ALTER TABLE %s ALTER COLUMN %s DROP DEFAULT;\n", table, column) // nolint: errcheck
	fmt.Fprintf(wr, "DROP SEQUENCE seq_%s_%s CASCADE;", table, column)               // nolint: errcheck
}

func (x Postgres) Select(wr io.Writer, table string, columns []string) {
	fmt.Fprintf(wr, "SELECT %s FROM %s\n", strings.Join(columns, ", "), table)
}

func (x Postgres) Insert(wr io.Writer, table string, columns []string) {
	fmt.Fprintf(wr, "INSERT INTO %s (%s)\n", table, strings.Join(columns, ", "))
}

func (x Postgres) Values(wr io.Writer, vals string) {
	fmt.Fprintf(wr, " VALUES (%s)\n", vals)
}

func (x Postgres) Update(wr io.Writer, table string) {
	fmt.Fprintf(wr, "UPDATE %s\n", table)
}

func (x Postgres) Set(wr io.Writer, column string, sq io.Reader) {
	q := "SET %s ="
	if sq != nil {
		q = fmt.Sprintf("%s (%s)\n", q, sq)
	} else {
		q = "?\n"
	}

	fmt.Fprintf(wr, q, column)
}

func (x Postgres) Delete(wr io.Writer, table string) {
	fmt.Fprintf(wr, "DELETE FROM %s\n", table)
}

func (x Postgres) Where(wr io.Writer, qc io.Writer) {
	fmt.Fprint(wr, "WHERE %s\n", qc)
}

func (x Postgres) InnerJoin(wr io.Writer, table, column, refTable, refColumn string) {
	fmt.Fprintf(wr, "INNER JOIN %s ON %s.%s = %s.%s\n", refTable, refTable, refColumn, table, column)
}

func (x Postgres) OuterJoin(wr io.Writer, table, column, refTable, refColumn string) {
	fmt.Fprintf(wr, "OUTER JOIN %s ON %s.%s = %s.%s\n", refTable, refTable, refColumn, table, column)
}

func (x Postgres) RightJoin(wr io.Writer, table, column, refTable, refColumn string) {
	fmt.Fprintf(wr, "RIGHT JOIN %s ON %s.%s = %s.%s\n", refTable, refTable, refColumn, table, column)
}

func (x Postgres) LeftJoin(wr io.Writer, table, column, refTable, refColumn string) {
	fmt.Fprintf(wr, "LEFT JOIN %s ON %s.%s = %s.%s\n", refTable, refTable, refColumn, table, column)
}

func (x Postgres) Returning(wr io.Writer, columns []string) {
	fmt.Fprintf(wr, "RETURNING %s", strings.Join(columns, ", "))
}

func (x Postgres) Limit(wr io.Writer, i int) {
	fmt.Fprintf(wr, "LIMIT %d\n", i)
}

func (x Postgres) AddVersionTable() string {
	return "CREATE TABLE IF NOT EXISTS versions (id INT NOT NULL, config TEXT NOT NULL);\n"
}

func (x Postgres) CheckVersion() string {
	return "SELECT id, config FROM versions ORDER BY id DESC;\n"
}

func (x Postgres) InsertVersion() string {
	return "INSERT INTO versions (id, config) VALUES($1, $2);\n"
}

func (x Postgres) Cursor(query io.Reader, key string) string {
	return fmt.Sprintf("DECLARE %s CURSOR FOR %s;", key, query)
}

func (x Postgres) CursorSelect(key string) string {
	return fmt.Sprintf("FETCH NEXT FROM %s;", key)
}

func (x Postgres) CursorClose(key string) string {
	return fmt.Sprintf("CLOSE %s;", key)
}
