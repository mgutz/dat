package dat

import (
	"fmt"
	"reflect"
	"github.com/mgutz/str"

	"gopkg.in/mgutz/dat.v1/reflectx"
)

var fieldMapper = reflectx.NewMapperTagFunc("db", nil, nil)

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
	for _, name := range reflectFields(v).DeclaredNames {
		cols = append(cols, name)
	}
	return cols
}

func reflectExcludeColumns(v interface{}, blacklist []string) []string {
	cols := []string{}
	for _, name := range reflectFields(v).DeclaredNames {
		if str.SliceContains(blacklist, name) {
			continue
		}
		cols = append(cols, name)
	}

	return cols
}
