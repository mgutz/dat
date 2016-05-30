package dat

import (
	"bytes"
	"reflect"
	"strconv"

	"github.com/syreclabs/dat/common"
)

var bufPool = common.NewBufferPool()

// listIdentifiers returns a CSV with quoted column identifiers
//
// return "col1", "col2", ... "coln"
func joinIdentifiers(columns []string, join string) string {
	buf := bufPool.Get()
	defer bufPool.Put(buf)

	for i, column := range columns {
		if i > 0 {
			buf.WriteString(join)
		}
		Dialect.WriteIdentifier(buf, column)
	}
	return buf.String()
}

func writeIdentifiers(buf *bytes.Buffer, columns []string, join string) {
	for i, column := range columns {
		if i > 0 {
			buf.WriteString(join)
		}
		Dialect.WriteIdentifier(buf, column)
	}
}

func writeIdentifier(buf *bytes.Buffer, name string) {
	Dialect.WriteIdentifier(buf, name)
}

// joinKVPlaceholders returns "col1" = $1 and "col2" = $2  ... and "coln" = $n
func joinKVPlaceholders(columns []string, join string, offset int) string {
	buf := bufPool.Get()
	defer bufPool.Put(buf)

	for i, column := range columns {
		if i > 0 {
			buf.WriteString(join)
		}
		Dialect.WriteIdentifier(buf, column)
		buf.WriteRune('=')
		writePlaceholder(buf, i+offset)
	}
	return buf.String()
}

func buildPlaceholders(buf *bytes.Buffer, start, length int) {
	// Build the placeholder like "($1,$2,$3)"
	buf.WriteRune('(')
	for i := start; i < start+length; i++ {
		if i > start {
			buf.WriteRune(',')
		}
		writePlaceholder(buf, i)
	}
	buf.WriteRune(')')
}

// joinPlaceholders returns $1, $2 ... , $n
func joinPlaceholders(length int, join string, offset int) string {
	buf := bufPool.Get()
	defer bufPool.Put(buf)

	for i := 0; i < length; i++ {
		if i > 0 {
			buf.WriteString(join)
		}
		writePlaceholder(buf, i+offset)
	}
	return buf.String()
}

// joinPlaceholders returns $1, $2 ... , $n
func writePlaceholders(buf *bytes.Buffer, length int, join string, offset int) {
	for i := 0; i < length; i++ {
		if i > 0 {
			buf.WriteString(join)
		}
		writePlaceholder(buf, i+offset)
	}
}

// writeKVAssign wrties "col1" = $1, "col2" = $2, ... "coln" = $n
func writeKVAssign(buf *bytes.Buffer, columns []string, values []interface{}, args *[]interface{}) {
	var placeholderStartPos int64 = 1

	// Build SET clause SQL with placeholders and add values to args
	for i, column := range columns {
		if i > 0 {
			buf.WriteString(", ")
		}
		value := values[i]
		writeIdentifier(buf, column)
		if e, ok := value.(*Expression); ok {
			start := placeholderStartPos
			buf.WriteString(" = ")
			// map relative $1, $2 placeholders to absolute
			remapPlaceholders(buf, e.Sql, start)
			*args = append(*args, e.Args...)
			placeholderStartPos += int64(len(e.Args))
		} else {
			if i < maxLookup {
				buf.WriteString(equalsPlaceholderTab[placeholderStartPos])
			} else {
				if placeholderStartPos < maxLookup {
					buf.WriteString(equalsPlaceholderTab[placeholderStartPos])
				} else {
					buf.WriteString(" = $")
					buf.WriteString(strconv.FormatInt(placeholderStartPos, 10))
				}
			}
			placeholderStartPos++
			*args = append(*args, value)
		}
	}

}

// writeKVWhere writes "col1" = $1 AND "col2" = $2, ... "coln" = $n
func writeKVWhere(buf common.BufferWriter, columns []string, values []interface{}, args *[]interface{}, anyConditions bool, pos *int64) bool {
	if len(columns) != len(values) {
		panic("Mismatch of column and values")
	}

	for i, k := range columns {
		v := values[i]

		if v == nil {
			anyConditions = writeWhereCondition(buf, k, " IS NULL", anyConditions)
		} else if e, ok := v.(*Expression); ok {
			start := pos
			buf.WriteString(" = ")
			// map relative $1, $2 placeholders to absolute
			remapPlaceholders(buf, e.Sql, *start)
			*args = append(*args, e.Args...)
			*pos += int64(len(e.Args))
		} else {
			vVal := reflect.ValueOf(v)

			if vVal.Kind() == reflect.Array || vVal.Kind() == reflect.Slice {
				vValLen := vVal.Len()
				if vValLen == 0 {
					if vVal.IsNil() {
						anyConditions = writeWhereCondition(buf, k, " IS NULL", anyConditions)
					} else {
						if anyConditions {
							buf.WriteString(" AND (1=0)")
						} else {
							buf.WriteString("(1=0)")
						}
					}
				} else if vValLen == 1 {
					anyConditions = writeWhereCondition(buf, k, equalsPlaceholderTab[*pos], anyConditions)
					*args = append(*args, vVal.Index(0).Interface())
					*pos++
				} else {
					// " IN $n"
					anyConditions = writeWhereCondition(buf, k, inPlaceholderTab[*pos], anyConditions)
					*args = append(*args, v)
					*pos++
				}
			} else {
				anyConditions = writeWhereCondition(buf, k, equalsPlaceholderTab[*pos], anyConditions)
				*args = append(*args, v)
				*pos++
			}
		}
	}

	return anyConditions
}
