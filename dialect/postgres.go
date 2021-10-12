package dialect

import (
	"fmt"
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

func (x Postgres) AddTable(name string, ifnotexists bool) string {
	i := "CREATE TABLE "
	if ifnotexists {
		i += "IF NOT EXISTS "
	}
	i += "%s();\n"
	return fmt.Sprintf(i, name)
}

func (x Postgres) DropTable(name string) string {
	return fmt.Sprintf("DROP TABLE %s CASCADE;\n", name)
}

func (x Postgres) AddColumn(table, column, typename string, size int) string {
	return fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;\n", table, column, x.Type(typename, size))
}

func (x Postgres) UpdateColumn(table, column, typename string, size int) string {
	return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s;\n", table, column, x.Type(typename, size))
}

func (x Postgres) DropColumn(table, column string) string {
	return fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;\n", table, column)
}

func (x Postgres) AddPrimaryKey(table string, columns []string) string {
	return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT pk_%s PRIMARY KEY(%s);\n", table, table, strings.Join(columns, `, `))
}

func (x Postgres) UpdatePrimaryKey(table string, columns []string) string {
	return x.DropPrimaryKey(table) + x.AddPrimaryKey(table, columns)
}

func (x Postgres) DropPrimaryKey(table string) string {
	return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT pk_%s;\n", table, table)
}

func (x Postgres) AddForeignKey(table, column, referenceTable, referenceColumn string) string {
	return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT fk_%s_%s FOREIGN KEY (%s) REFERENCES %s(%s);\n", table, table, column, column, referenceTable, referenceColumn)
}

func (x Postgres) UpdateForeignKey(table, column, refrerenceTable, referenceColumn string) string {
	return x.DropForeignKey(table, column) + x.AddForeignKey(table, column, refrerenceTable, referenceColumn)
}

func (x Postgres) DropForeignKey(table, column string) string {
	return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT fk_%s_%s;\n", table, table, column)
}

func (x Postgres) AddUnique(id, table string, columns []string) string {
	return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT uq_%s UNIQUE(%s);\n", table, id, strings.Join(columns, `, `))
}

func (x Postgres) UpdateUnique(id, table string, columns []string) string {
	return x.DropUnique(id, table) + x.AddUnique(id, table, columns)
}

func (x Postgres) DropUnique(id, table string) string {
	return fmt.Sprintf("ALTER TABLE %s DROP uq_%s;\n", table, id)
}

func (x Postgres) SetNotNull(table, column string) string {
	return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET NOT NULL;\n", table, column)
}

func (x Postgres) DeleteNotNull(table, column string) string {
	return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP NOT NULL;\n", table, column)
}

func (x Postgres) AddCheck(table, column, check string) string {
	return fmt.Sprintf("ALTER TABLE %s ADD CONSTRAINT ch_%s_%s CHECK(%s);\n", table, table, column, check)
}

func (x Postgres) UpdateCheck(table, column, check string) string {
	return x.DropCheck(table, column) + x.AddCheck(table, column, check)
}

func (x Postgres) DropCheck(table, column string) string {
	return fmt.Sprintf("ALTER TABLE %s DROP CONSTRAINT ch_%s_%s;\n", table, table, column)
}

func (x Postgres) AddEnum(name string, values []string) string {
	return fmt.Sprintf("CREATE TYPE %s AS ENUM ('%s');\n", name, strings.Join(values, `', '`))
}

func (x Postgres) AppendEnum(name, value string) string {
	return fmt.Sprintf("ALTER TYPE %s ADD VALUE '%s';\n", name, value)
}

func (x Postgres) DropEnum(name string) string {
	return fmt.Sprintf("DROP TYPE %s;\n", name)
}

func (x Postgres) SetDefault(table, column, value string) string {
	return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT %s;\n", table, column, value)
}

func (x Postgres) DropDefault(table, column string) string {
	return fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP DEFAULT;\n", table, column)
}

func (x Postgres) SetAutoIncrement(table, column string) (q string) {
	q += fmt.Sprintf("CREATE SEQUENCE seq_%s_%s;\n", table, column)
	q += fmt.Sprintf("SELECT setval('seq_%s_%s', (SELECT max(%s) FROM %s));\n", table, column, column, table)
	q += fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s SET DEFAULT nextval('seq_%s_%s'::regclass);\n", table, column, table, column)

	return
}

func (x Postgres) UnsetAutoIncrement(table, column string) (q string) {
	q = fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s DROP DEFAULT;\n", table, column)
	q += fmt.Sprintf("DROP SEQUENCE seq_%s_%s CASCADE;", table, column)

	return
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
