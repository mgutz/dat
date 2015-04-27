package dat

import (
	"fmt"
	"log"
	"reflect"
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
		if dbName == "" {
			log.Fatalf("%s must have db struct tags for all fields: `db:\"\"`", vname)
		}
		r.fields = append(r.fields, &field{goName: name, dbName: dbName})
	}
	structCache[vname] = r
	return r
}

// CalculateFieldMap recordType is the type of a structure
func CalculateFieldMap(recordType reflect.Type, columns []string,
	requireAllColumns bool) ([][]int, error) {
	// each value is either the slice to get to the field via
	// FieldByIndex(index []int) in the record, or nil if we don't want to map
	// it to the structure.
	lenColumns := len(columns)
	fieldMap := make([][]int, lenColumns)

	for i, col := range columns {
		fieldMap[i] = nil

		queue := []fieldMapQueueElement{{Type: recordType, Idxs: nil}}

	QueueLoop:
		for len(queue) > 0 {
			curEntry := queue[0]
			queue = queue[1:]

			curType := curEntry.Type
			curIdxs := curEntry.Idxs
			lenFields := curType.NumField()

			for j := 0; j < lenFields; j++ {
				fieldStruct := curType.Field(j)

				// Skip unexported field
				if len(fieldStruct.PkgPath) != 0 {
					continue
				}

				name := fieldStruct.Tag.Get("db")
				if name != "-" {
					if name == "" {
						name = NameMapping(fieldStruct.Name)
					}
					if name == col {
						fieldMap[i] = append(curIdxs, j)
						break QueueLoop
					}
				}

				if fieldStruct.Type.Kind() == reflect.Struct {
					var idxs2 []int
					copy(idxs2, curIdxs)
					idxs2 = append(idxs2, j)
					queue = append(queue, fieldMapQueueElement{Type: fieldStruct.Type, Idxs: idxs2})
				}
			}
		}

		if requireAllColumns && fieldMap[i] == nil {
			return nil, fmt.Errorf(`could not map db column "%s" to struct field (use struct tags)`, col)
		}
	}

	return fieldMap, nil
}

// PrepareHolderFor creates holders for a record.
//
// TODO: fill this in
func PrepareHolderFor(record reflect.Value, fieldMap [][]int, holder []interface{}) ([]interface{}, error) {
	// Given a query and given a structure (field list), there's 2 sets of fields.
	// Take the intersection. We can fill those in. great.
	// For fields in the structure that aren't in the query, we'll let that slide if db:"-"
	// For fields in the structure that aren't in the query but without db:"-", return error
	// For fields in the query that aren't in the structure, we'll ignore them.

	for i, fieldIndex := range fieldMap {
		if fieldIndex == nil {
			holder[i] = &destDummy
		} else {
			field := record.FieldByIndex(fieldIndex)
			holder[i] = field.Addr().Interface()
		}
	}

	return holder, nil
}

// ValuesFor does soemthing
//
// TODO:
func ValuesFor(recordType reflect.Type, record reflect.Value, columns []string) ([]interface{}, error) {
	fieldMap, err := CalculateFieldMap(recordType, columns, true)
	if err != nil {
		fmt.Println("err: calc field map")
		return nil, err
	}

	values := make([]interface{}, len(columns))
	for i, fieldIndex := range fieldMap {
		if fieldIndex == nil {
			panic("wtf bro")
		} else {
			field := record.FieldByIndex(fieldIndex)
			values[i] = field.Interface()
		}
	}

	return values, nil
}
