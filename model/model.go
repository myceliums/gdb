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
	primaryReg = regexp.MustCompile(`primary key|primarykey|primary`)

	// uniqueReg
	// [0] unique(id_name)
	// [1] (id_name)
	// [2] id_name
	uniqueReg = regexp.MustCompile(`unique(\((\w+)\))?`)

	// autoIncrementReg
	autoIncrementReg = regexp.MustCompile(`^serial|auto\ ?increment`)
)

// DataType is a model data structure
type DataType interface {
	Type() string
}

func newConfig(in []byte) (*Config, error) {
	var cfile map[string]interface{}

	if err := yaml.Unmarshal(in, &cfile); err != nil {
		return nil, err
	}

	x := new(Config)
	x.Tables = map[string]map[string]string{}
	x.Enums = map[string][]string{}
	x.raw = in

	for k, t := range cfile {
		switch m := t.(type) {
		case []interface{}:
			var vals []string
			for _, mm := range m {
				if sm, ok := mm.(string); ok {
					vals = append(vals, sm)
				}
			}
			x.Enums[k] = vals
		case map[interface{}]interface{}:
			vals := map[string]string{}
			for kk, mm := range m {
				sk, kok := kk.(string)
				smm, vok := mm.(string)
				if kok && vok {
					vals[sk] = smm
				}
			}
			x.Tables[k] = vals
		default:
			return nil, fmt.Errorf("'%s' has an unknown configuration. want either map[string][]string or map[string]map[string]string, but has %T", k, t)
		}
	}

	return x, nil
}

type Config struct {
	Tables map[string]map[string]string `yaml:"tables,flow"`
	Enums  map[string][]string          `yaml:"enums,flow"`
	raw    []byte
}

// New returns a new initialized model
func New(in []byte) (*Model, error) {
	var x Model
	x.Tables = map[string]map[string]*Column{}
	x.Primaries = map[string][]*Column{}
	x.Uniques = map[string][]*Column{}
	x.Foreigns = map[string]*Column{}
	x.Enums = map[string]*Enum{}
	x.aliases = primitiveTypesAliases()

	conf, err := newConfig(in)
	if err != nil {
		return nil, err
	}

	x.conf = conf

	x, err = appendTablesAndColums(x, conf.Tables)
	if err != nil {
		return nil, err
	}

	x = appendEnums(x, conf.Enums)

	x, err = getDataTypes(x)
	if err != nil {
		return nil, err
	}

	return &x, nil
}

// Model contains the database structure
type Model struct {
	Tables    map[string]map[string]*Column
	Enums     map[string]*Enum
	Uniques   map[string][]*Column
	Primaries map[string][]*Column
	Foreigns  map[string]*Column
	aliases   map[string]DataType
	conf      *Config
}

// Config returns the raw config
func (x Model) Config() []byte {
	return x.conf.raw
}

// Column is a database table column
type Column struct {
	Table        string
	Name         string
	Datatype     DataType
	Ref          *Column
	Size         int
	Default      string
	NotNull      bool
	Check        string
	Primary      string
	Unique       string
	AutoIncement bool
	rawtype      string
	raw          string
}

// Type is an implementation of Datatype
func (x *Column) Type() string {
	return x.Datatype.Type()
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

func appendTablesAndColums(m Model, tables map[string]map[string]string) (Model, error) {
	for table, columns := range tables {
		if _, ok := m.Tables[table]; !ok {
			m.Tables[table] = map[string]*Column{}
		}

		for name, content := range columns {
			tname := table

			col := new(Column)
			col.raw = content

			col.Table = tname
			col.Name = name

			col.rawtype, col.Size = rawtype(content)
			if col.rawtype == `` {
				return m, fmt.Errorf("no type found in table %s column %s", table, name)
			}

			if primaryReg.MatchString(content) {
				if m.Primaries[table] == nil {
					m.Primaries[table] = []*Column{}
				}
				m.Primaries[table] = append(m.Primaries[table], col)
				col.Primary = table
			}

			unique := getSecondSubmatchOrColumn(uniqueReg, name, content)
			if unique != `` {
				if m.Uniques[unique] == nil {
					m.Uniques[unique] = []*Column{}
				}
				m.Uniques[unique] = append(m.Uniques[unique], col)
				col.Unique = unique
			}

			col.Default = getFirstSubmatch(defaultReg, content)
			col.Check = getFirstSubmatch(checkReg, content)
			col.NotNull = notnullReg.MatchString(content)
			col.AutoIncement = autoIncrementReg.MatchString(content)

			m.Tables[table][name] = col
			m.aliases[table+`.`+name] = col
		}
	}

	return m, nil
}

func appendEnums(m Model, enums map[string][]string) Model {
	if enums == nil {
		m.Enums = map[string]*Enum{}
	}

	for name, values := range enums {
		enum := new(Enum)
		enum.Name = name
		enum.Values = values

		m.Enums[name] = enum
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
			if ref, ok := col.Datatype.(*Column); ok {
				col.Ref = ref
				m.Foreigns[col.Table+`.`+col.Name] = col
			}
		}
	}

	return m, nil
}

type primitiveType string

// Type is an implementation of the Datatype
func (x primitiveType) Type() string {
	return string(x)
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
	m[`serial`] = &integer

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
