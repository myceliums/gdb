# gdb
A golang generated database manager (WIP and a query builder).
Configure your database in a YAML configuration making your documentation code and your code documentation.

The generated code will give you an "Open(string, string) (\*sql.DB, error)" function just like sql.Open but 
returns a fully initialized database with the models you defined in your configuration.
Each update you make in the configuration will be safely implemented in your database model the next time the Open function is called.
See [usage](#usage) for a clear example.

```sh
gdb <options> [configfile]

Options:
  -o          specifies the output file (default is "model.gen.go")
  -pkg        speficies the packagename (default is "model")
```


A configuration example:
```yaml
# ./db.yml

accounts:
  id: serial primary
  username: varchar(50) unique not null
  password: varchar not null
  email: varchar not null unique
  email_verified_at: timestamp
  created_at: timestamp default(NOW())

roles:
  id: serial primary
  name: varchar unique not null

account_roles:
  id: serial primary
  account_id: accounts.id not null
  role_id: roles.id not null

posts:
  id: serial primary
  type: post_type default('general')
  created_by: accounts.id not null
  created_at: timestamp default(NOW())
  context: text not null

posts_type:
- general
- blog
- article
```

## Configuration
tables are made as following:
```yaml
tablename:
  column1: datatype addatives
  column2: datatype addatives
```
A datatype can be either one of the native datatypes or can be one of the defined enums.

The native datatypes:
|Definition|Go type|Postgres|
|-|-|-|
|int, integer, serial|int|INT|
|varchar, string, char, character|string|VARCHAR|
|text|string|TEXT|
|bool,boolean|bool|BOOLEAN|
|datetime, timestamp, time, date|time.Time|TIMESTAMP|
|bigint|int64|BIGINT|
|smallint|int32|SMALLINT|
|float, float32, real|float32|FLOAT|
|double, float64|float64|DOUBLE|

The constraints and properties:
|Definition|Postgres|Description|
|-|-|-|
|primary key, primary|PRIMARY KEY|Adds primary key to column|
|default(\<expression\>)|DEFAULT \<expression\>|Adds default to column|
|unique, unique(\<expression\>)|UNIQUE|Adds unique to column if a arugment is given it will be a grouped unique constraint to the table|
|not null, notnull|NOT NULL|Adds a not null constraint to the column|
|check(\<expression\>)|CHECK(\<expression\>)|Adds a check constraint to the column|
|serial (as type), autoincrement, auto increment|SERIAL (as type)|Auto increments the value with each added table entry|

Enums are defined as following:
```yaml
enum_name:
- value1
- value2
```
Please note that values can only be added to enums and not taken away.
Meaning the configuration reader will only add the new values and cannot remove old/unwanted ones.

## Usage
You can add the command to generate the code for you using "go generate ./..."
```go
// ./dbc/dbc.go

// Package dbc holds the database context
package dbc

//go:generate gdb -pkg dbc -o ./dbc.gen.go ../db.yml
```

Using go generate you can then apply the code to your code.

```go
package main

import (
	"fmt"
	"database/sql"

	"<your_package_name>/dbc"
)

func main() {
	db, err := dbc.Open(`postgres`, `postgres://username:password@localhost:5432/dbname?sslmode=disable`)
	if err != nil {
		panic(err)
	}

	q := `SELECT id, username
		FROM users;`

	rows, err := db.Query(q)
	if err != nil && err != sql.ErrNoRows {
		panic(err)
	}
	defer rows.Close() // nolint: errcheck

	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			panic(err)
		}
		
		fmt.Printf("userID: %d, Name: %s\n", id, name)
	}
}

```

## Todo
- [x] Create initial SQL and differential SQL
- [ ] Create query builder, taking inspiration from "git.ultraware.nl/Nisevoid/qb"
- [ ] Store the configuration hashed
- [ ] Create database read to configuration
