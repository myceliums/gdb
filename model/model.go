package model

import (
	"fmt"
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

func newConfig(in []byte) (*config, error) {
	x := new(config)

	if err := yaml.Unmarshal(in, x); err != nil {
		return nil, err
	}

	return x, nil
}

type config struct {
	Pkg    string                       `yaml:"pkg"`
	Tables map[string]map[string]string `yaml:"tables"`
	Enums  map[string][]string          `yaml:"enums"`
}

// New returns a new initialized model
func New(in []byte) (*Model, error) {
	var x Model
	x.Tables = map[string][]*Column{}
	x.Primaries = map[string][]*Column{}
	x.Uniques = map[string][]*Column{}
	x.aliases = primitiveTypesAliases()

	conf, err := newConfig(in)
	if err != nil {
		return nil, err
	}

	x, err = appendTablesAndColums(x, conf)
	if err != nil {
		return nil, err
	}

	x = appendEnums(x, conf)

	x, err = getDataTypes(x)
	if err != nil {
		return nil, err
	}

	return &x, nil
}

// Model contains the database structure
type Model struct {
	Pkg       string
	Tables    map[string][]*Column
	Enums     []*Enum
	Uniques   map[string][]*Column
	Primaries map[string][]*Column
	aliases   map[string]DataType
	conf      config
}

//
func (x Model) Compare(old Model) map[string]string {
	// TODO return a list of this added, updated and deleted
	// list will look like '+default(NOW()),-notnull,~check(started_at < ended_at)
	//
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

func appendTablesAndColums(m Model, conf *config) (Model, error) {
	for table, columns := range conf.Tables {
		for name, content := range columns {
			col := new(Column)
			col.Table = &table
			col.Name = name

			col.rawtype, col.Size = rawtype(content)
			if col.rawtype == `` {
				return m, fmt.Errorf("no type found in table %s column %s", table, name)
			}

			primary := getSecondSubmatchOrColumn(primaryReg, name, content)
			if primary != `` {
				if m.Primaries[primary] == nil {
					m.Primaries[primary] = []*Column{}
				}
				m.Primaries[primary] = append(m.Primaries[primary], col)
			}

			unique := getSecondSubmatchOrColumn(uniqueReg, name, content)
			if unique != `` {
				if m.Uniques[unique] == nil {
					m.Uniques[unique] = []*Column{}
				}
				m.Uniques[unique] = append(m.Uniques[unique], col)
			}

			col.Default = getFirstSubmatch(defaultReg, content)
			col.Check = getFirstSubmatch(checkReg, content)
			col.NotNull = notnullReg.MatchString(content)

			m.Tables[table] = append(m.Tables[table], col)
			m.aliases[table+`.`+name] = col
		}
	}

	return m, nil
}

func appendEnums(m Model, conf *config) Model {
	for name, values := range conf.Enums {
		enum := new(Enum)
		enum.Name = name
		enum.Values = values

		m.Enums = append(m.Enums, enum)
		m.aliases[name] = enum
	}

	return m
}

func getDataTypes(m Model) (Model, error) {
	for table, cols := range m.Tables {
		for _, col := range cols {
			if m.aliases[col.rawtype] == nil {
				return m, fmt.Errorf("unrecognized datatype in table %s column %s type: %s", table, col.Name, col.rawtype)
			}
			col.Datatype = m.aliases[col.rawtype]
		}
	}

	return m, nil
}

type primitiveType string

// Type is an implementation of the Datatype
func (x primitiveType) Type() string {
	return string(x)
}

// Ref is an implementation of the Datatype
func (x primitiveType) Ref() *Column {
	return nil
}

func primitiveTypesAliases() map[string]DataType {
	m := map[string]DataType{}
	varchar := primitiveType(`varchar`)
	m[`string`] = &varchar
	m[`varchar`] = &varchar
	m[`char`] = &varchar
	m[`character`] = &varchar

	integer := primitiveType(`int`)
	m[`int`] = &integer
	m[`integer`] = &integer

	float := primitiveType(`float`)
	m[`float`] = &float
	m[`float32`] = &float
	m[`real`] = &float

	timestamp := primitiveType(`timestamp`)
	m[`timestamp`] = &timestamp
	m[`date`] = &timestamp
	m[`time`] = &timestamp
	m[`datetime`] = &timestamp

	boolean := primitiveType(`boolean`)
	m[`boolean`] = &boolean
	m[`bool`] = &boolean

	double := primitiveType(`double`)
	m[`double`] = &double
	m[`float64`] = &double

	text := primitiveType(`text`)
	m[`text`] = &text

	bigint := primitiveType(`bigint`)
	m[`bigint`] = &bigint

	smallint := primitiveType(`smallint`)
	m[`smallint`] = &smallint

	return m
}
