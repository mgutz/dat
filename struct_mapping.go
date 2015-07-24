package dat

import (
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

var fieldMapper = reflectx.NewMapperFunc("db", snakeCase)

// reflectFields gets a cached field information about record
func reflectFields(rec interface{}) *reflectx.FieldMeta {
	val := reflect.Indirect(reflect.ValueOf(rec))
	vtype := val.Type()
	return fieldMapper.TypeMap(vtype)
}

// ValuesFor ...
func valuesFor(recordType reflect.Type, record reflect.Value, columns []string) ([]interface{}, error) {
	vals := fieldMapper.FieldsByName(record, columns)
	values := make([]interface{}, len(columns))
	for i, v := range vals {
		values[i] = v.Interface()
	}
	return values, nil
}

func reflectColumns(v interface{}) []string {
	cols := []string{}
	for _, tag := range reflectFields(v).Names {
		cols = append(cols, tag)
	}
	return cols
}

func reflectBlacklistedColumns(v interface{}, blacklist []string) []string {
	cols := []string{}
	for _, tag := range reflectFields(v).Names {
		if str.SliceContains(blacklist, tag) {
			continue
		}
		cols = append(cols, tag)
	}

	return cols
}
