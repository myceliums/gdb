package model

import (
	"log"
	"regexp"
	"strconv"

	"gopkg.in/yaml.v2"
)

var (
	// typeReg is a regular expression for matching the types
	// [0] varchar(100)
	// [1] varchar
	// [2] (100)
	// [3] 100
	typeReg = regexp.MustCompile(`^([\w\-\.]+)(\(([\d]{1,3})\))?`)

	// defaultReg
	// [0] default(NOW())
	// [1] NOW()
	defaultReg = regexp.MustCompile(`default\(([\w\(\)'"]+)\)`)

	// checkReg
	// [0] check(x > y)
	// [1] x > y
	checkReg = regexp.MustCompile(`check\(([\w\(\)\-\'\"\=\>\<\!\%\/\+]+)\)`)

	// notnullReg is used to match the context to see if it contains a not null statement
	notnullReg = regexp.MustCompile(`notnull|not null`)

	// primaryReg
	// [0] primary(id_name)
	// [1] (id_name)
	// [2] id_name
	primaryReg = regexp.MustCompile(`primary key|primarykey|primary(\((\w+)\))?`)

	// uniqueReg
	// [0] unique(id_name)
	// [1] (id_name)
	// [2] id_name
	uniqueReg = regexp.MustCompile(`unique(\((\w+)\))?`)
)

// DataType is a model data structure
type DataType interface {
	Type() string
	Ref() *Column
}

type config struct {
	Pkg    string                       `yaml:"pkg"`
	Tables map[string]map[string]string `yaml:"tables"`
	Enums  map[string][]string          `yaml:"enums"`
}

// New returns a new initialized model
func New(in []byte) (*Model, error) {
	x := new(Model)
	x.Tables = map[string][]*Column{}
	x.Primaries = map[string][]*Column{}
	x.Uniques = map[string][]*Column{}

	conf := new(config)
	if err := yaml.Unmarshal(in, conf); err != nil {
		return nil, err
	}

	for table, columns := range conf.Tables {
		for name, context := range columns {
			col := new(Column)
			col.Table = &table
			col.Name = name

			col.rawtype, col.Size = rawtype(context)
			if col.rawtype == `` {
				log.Fatalf("error in type for table %s column %s", table, name)
			}

			primary := getSecondSubmatchOrColumn(primaryReg, name, context)
			if primary != `` {
				if x.Primaries[primary] == nil {
					x.Primaries[primary] = []*Column{}
				}
				x.Primaries[primary] = append(x.Primaries[primary], col)
			}

			unique := getSecondSubmatchOrColumn(uniqueReg, name, context)
			if unique != `` {
				if x.Uniques[unique] == nil {
					x.Uniques[unique] = []*Column{}
				}
				x.Uniques[unique] = append(x.Uniques[unique], col)
			}

			col.Default = getFirstSubmatch(defaultReg, context)
			col.Check = getFirstSubmatch(checkReg, context)
			col.NotNull = notnullReg.MatchString(context)

			x.Tables[table] = append(x.Tables[table], col)
		}
	}

	x.Enums = getEnums(conf.Enums)

	//TODO get datatype

	return x, nil
}

// Model contains the database structure
type Model struct {
	Pkg       string
	Tables    map[string][]*Column
	Enums     []*Enum
	Uniques   map[string][]*Column
	Primaries map[string][]*Column
}

// Column is a database table column
type Column struct {
	Table    *string
	Name     string
	Datatype DataType
	Size     int
	Default  string
	NotNull  bool
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

// Enum is a numeric object that con be defined in the database
type Enum struct {
	Name   string
	Values []string
}

// Type is an implementation of Datatype
func (x Enum) Type() string {
	return x.Name
}

// Enum is a numeric object that con be defined in the database
func (x Enum) Ref() *Column {
	return nil
}

func rawtype(context string) (rawtype string, size int) {
	typeMatch := typeReg.FindStringSubmatch(context)
	if i := len(typeMatch); i < 1 || i == 2 {
		return
	}

	rawtype = typeMatch[1]

	if len(typeMatch) == 4 && typeMatch[3] != `` {
		s, err := strconv.Atoi(typeMatch[3])
		if err != nil {
			return
		}

		size = s
	}

	return
}

func getFirstSubmatch(reg *regexp.Regexp, context string) string {
	matches := reg.FindStringSubmatch(context)
	if len(matches) < 2 {
		return ``
	}

	return matches[1]
}

func getSecondSubmatchOrColumn(reg *regexp.Regexp, columnName, context string) string {
	matches := reg.FindStringSubmatch(context)
	length := len(matches)
	switch {
	case length == 3 && matches[2] != ``:
		return matches[2]
	case length == 1 && matches[0] != ``:
		return columnName
	}

	return ``
}

func getEnums(enumConfig map[string][]string) (enums []*Enum) {
	for name, vals := range enumConfig {
		enum := new(Enum)
		enum.Name = name
		enum.Values = vals

		enums = append(enums, enum)
	}
	return
}
