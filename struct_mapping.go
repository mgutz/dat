package dat

import (
	"fmt"
	"log"
	"reflect"
	"unicode"

	"github.com/jmoiron/sqlx/reflectx"
)

var destDummy interface{}

type fieldMapQueueElement struct {
	Type reflect.Type
	Idxs []int
}

type field struct {
	// name of go field
	goName string
	// name of datbase column
	dbName string
}
type record struct {
	fields    []*field
	dbColumns []string
}

func (r *record) Columns() []string {
	if r.dbColumns != nil {
		return r.dbColumns
	}

	lenFields := len(r.fields)
	r.dbColumns = make([]string, lenFields)
	for i, f := range r.fields {
		r.dbColumns[i] = f.dbName
	}
	return r.dbColumns
}

func newRecord() *record {
	return &record{}
}

// structCache maps type name -> record
var structCache = map[string]*record{}

// reflectFields gets a cached field information about record
func reflectFields(rec interface{}) *record {
	val := reflect.Indirect(reflect.ValueOf(rec))
	vname := val.String()
	vtype := val.Type()

	if structCache[vname] != nil {
		return structCache[vname]
	}

	r := &record{}
	//fmt.Println(val.Type().String(), val.Type().Name())
	for i := 0; i < vtype.NumField(); i++ {
		f := vtype.Field(i)

		// skip unexported
		if len(f.PkgPath) != 0 {
			continue
		}
		name := f.Name
		dbName := f.Tag.Get("db")
		logger.Debug("reflect field", "field", f)
		if dbName == "" {
			log.Fatalf("%s must have db struct tags for all fields: `db:\"\"`", vname)
		}
		r.fields = append(r.fields, &field{goName: name, dbName: dbName})
	}
	structCache[vname] = r
	return r
}

// ToSnake convert the given string to snake case following the Golang format:
// acronyms are converted to lower-case and preceded by an underscore.
func snakeCase(in string) string {
	runes := []rune(in)
	length := len(runes)

	var out []rune
	for i := 0; i < length; i++ {
		if i > 0 && unicode.IsUpper(runes[i]) && ((i+1 < length && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	return string(out)
}

//var tagMapper = reflectx.NewMapperFunc("db", strings.ToLower)
var fieldMapper = reflectx.NewMapperFunc("db", snakeCase)

// CalculateFieldMap recordType is the type of a structure
func CalculateFieldMap(recordType reflect.Type, columns []string,
	requireAllColumns bool) ([][]int, error) {
	return fieldMapper.TraversalsByName(recordType, columns), nil
}

// ValuesFor ...
func ValuesFor(recordType reflect.Type, record reflect.Value, columns []string) ([]interface{}, error) {
	fieldMap, err := CalculateFieldMap(recordType, columns, true)
	if err != nil {
		fmt.Println("err: calc field map")
		return nil, err
	}

	values := make([]interface{}, len(columns))
	for i, fieldIndex := range fieldMap {
		if fieldIndex == nil {
			panic("No entry for fieldIndex in fieldmap")
		} else {
			field := record.FieldByIndex(fieldIndex)
			values[i] = field.Interface()
		}
	}

	return values, nil
}
