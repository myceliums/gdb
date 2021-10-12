# gdb
A golang generated database manager.
Configure your database in a YAML configuration making your documentation code and your code documentation.
```yaml
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

## Datatypes
Good to note is that if a type is defined as int and has the primary key constraint then the type
will automatically add auto increment (serial in postgres).

Varchar's and integers can have lengths set after them like: varchar(50) or string(50)
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

The constraints and autoincrement:
|Definition|Postgres|Description|
|-|-|-|
|primary key, primary|PRIMARY KEY|Adds primary key to column|
|default(\<expression\>)|DEFAULT \<expression\>|Adds default to column|
|unique, unique(\<expression\>)|UNIQUE|Adds unique to column if a arugment is given it will be a grouped unique constraint to the table|
|not null, notnull|NOT NULL|Adds a not null constraint to the column|
|check(\<expression\>)|CHECK(\<expression\>)|Adds a check constraint to the column|
|serial (as type), autoincrement, auto increment|SERIAL (as type)|Auto increments the value with each added table entry|
