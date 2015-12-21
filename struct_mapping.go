package dat

import (
	"fmt"
	"reflect"
	"unicode"

	"github.com/mgutz/str"

	"gopkg.in/mgutz/dat.v1/reflectx"
)

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

var fieldMapper = reflectx.NewMapperTagFunc("db", nil, nil)

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return v.IsNil()
	}
	return false
}

// reflectFields gets a cached field information about record
func reflectFields(rec interface{}) *reflectx.StructMap {
	val := reflect.Indirect(reflect.ValueOf(rec))
	vtype := val.Type()
	return fieldMapper.TypeMap(vtype)
}

// ValuesFor ...
func valuesFor(recordType reflect.Type, record reflect.Value, columns []string) ([]interface{}, error) {
	vals := fieldMapper.FieldsByName(record, columns)
	values := make([]interface{}, len(columns))
	for i, val := range vals {
		if !val.IsValid() {
			return nil, fmt.Errorf("Could not find struct tag in type %s: `db:\"%s\"`", recordType.Name(), columns[i])
		}
		values[i] = val.Interface()
	}
	return values, nil
}

func reflectColumns(v interface{}) []string {
	cols := []string{}
	for _, tag := range reflectFields(v).Names {
		cols = append(cols, tag.Name)
	}
	return cols
}

func reflectExcludeColumns(v interface{}, blacklist []string) []string {
	cols := []string{}
	for _, tag := range reflectFields(v).Names {
		if str.SliceContains(blacklist, tag.Name) {
			continue
		}
		cols = append(cols, tag.Name)
	}

	return cols
}
