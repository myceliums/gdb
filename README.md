# gdb
A golang generated database manager

## Datatypes
Good to note is that if a type is defined as int and has the primary key constraint then the type
will automatically add auto increment (serial in postgres).

Varchar's and integers can have lengths set after them like: varchar(50) or string(50)
|Definition|Go type|Postgres|
|-|-|-|
|int, integer|int|INT|
|varchar, string, char, character|string|VARCHAR|
|text|string|TEXT|
|bool,boolean|bool|BOOLEAN|
|datetime, timestamp, time, date|time.Time|TIMESTAMP|
|bigint|int64|BIGINT|
|smallint|int32|SMALLINT|
|float, float32, real|float32|FLOAT|
|double, float64|float64|DOUBLE|

The constraints:
|Definition|Postgres|Description|
|-|-|-|
|primary key, primary|PRIMARY KEY|Adds primary key to collumn|
|default(\<expression\>)|DEFAULT \<expression\>|Adds default to collum|
|unique, unique(\<expression\>)|UNIQUE|Adds unique to collumn if a arugment is given it will be a grouped unique constraint to the table|
|not null, notnull|NOT NULL|Adds a not null constraint to the collumn|
|check(\<expression\>)|CHECK(\<expression\>)|Adds a check constraint to the collumn|
