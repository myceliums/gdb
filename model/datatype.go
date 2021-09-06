package model

// DataType is a model data structure
type DataType interface {
	Type() string
	Ref() *Column
}

// Model contains the database structure
type Model struct {
	Pkg     string
	Columns []*Column
	Enums   []*Enum
}

// Column is a database table column
type Column struct {
	Table    *string
	Name     string
	Datatype DataType
	Size     int
	Default  string
	NotNull  bool
	Unique   string // if value contains '_default' it is a column
	Primary  bool
	Check    string
	rawtype  string
}

// Type is an implementation of Datatype
func (x *Column) Type() string {
	return x.Datatype.Type()
}

// Ref is an implementation of DataType
func (x *Column) Ref() *Column {
	return x
}

// Enum is a numeric object that con be defined in the
// database
type Enum struct {
	Table  *string
	Values []string
}

func (x Enum) Type() string {
	return x.Table
}

func (x Enum) Ref() *Column {
	return nil
}
