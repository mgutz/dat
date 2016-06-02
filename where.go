package dat

import (
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/mgutz/dat.v1/common"
)

// Eq is a map column -> value pairs which must be matched in a query
type Eq map[string]interface{}

type whereFragment struct {
	Condition   string
	Values      []interface{}
	EqualityMap map[string]interface{}
}

func newWhereFragment(whereSQLOrMap interface{}, args []interface{}) *whereFragment {
	switch pred := whereSQLOrMap.(type) {
	case Expression:
		return &whereFragment{Condition: pred.Sql, Values: pred.Args}
	case *Expression:
		return &whereFragment{Condition: pred.Sql, Values: pred.Args}
	case string:
		return &whereFragment{Condition: pred, Values: args}
	case map[string]interface{}:
		return &whereFragment{EqualityMap: pred}
	case Eq:
		return &whereFragment{EqualityMap: map[string]interface{}(pred)}
	default:
		panic("Invalid argument passed to Where. Pass a string or an Eq map.")
	}
}

var rePlaceholder = regexp.MustCompile(`\$\d+`)

func remapPlaceholders(buf common.BufferWriter, statement string, start int64) int64 {
	if !strings.Contains(statement, "$") {
		buf.WriteString(statement)
		return 0
	}

	highest := 0
	pos := int(start) - 1 // 0-based
	statement = rePlaceholder.ReplaceAllStringFunc(statement, func(s string) string {
		i, _ := strconv.Atoi(s[1:])
		if i > highest {
			highest = i
		}

		sum := strconv.Itoa(pos + i)
		return "$" + sum
	})

	buf.WriteString(statement)
	return int64(highest)
}

// Invariant: for scope conditions only
func writeScopeCondition(buf common.BufferWriter, f *whereFragment, args *[]interface{}, pos *int64) {
	buf.WriteRune(' ')
	if len(f.Values) > 0 {
		// map relative $1, $2 placeholders to absolute
		replaced := remapPlaceholders(buf, f.Condition, *pos)
		*pos += replaced
		*args = append(*args, f.Values...)
	} else {
		buf.WriteString(f.Condition)
	}
}

func writeAndFragmentsToSQL(buf common.BufferWriter, fragments []*whereFragment, args *[]interface{}, pos *int64) {
	writeFragmentsToSQL(" AND ", true, buf, fragments, args, pos)
}

func writeCommaFragmentsToSQL(buf common.BufferWriter, fragments []*whereFragment, args *[]interface{}, pos *int64) {
	writeFragmentsToSQL(", ", false, buf, fragments, args, pos)
}

// Invariant: only called when len(fragments) > 0
func writeFragmentsToSQL(delimiter string, addParens bool, buf common.BufferWriter, fragments []*whereFragment, args *[]interface{}, pos *int64) {
	hasConditions := false
	for _, f := range fragments {
		if f.Condition != "" {
			if hasConditions {
				buf.WriteString(delimiter)
			} else {
				hasConditions = true
			}

			if addParens {
				buf.WriteRune('(')
			}

			if len(f.Values) > 0 {
				// map relative $1, $2 placeholders to absolute
				replaced := remapPlaceholders(buf, f.Condition, *pos)
				*pos += replaced
				*args = append(*args, f.Values...)
			} else {
				buf.WriteString(f.Condition)
			}
			if addParens {
				buf.WriteRune(')')
			}
		} else if f.EqualityMap != nil {
			hasConditions = writeEqualityMapToSQL(buf, f.EqualityMap, args, hasConditions, pos)
		} else {
			panic("invalid equality map")
		}
	}
}

func writeEqualityMapToSQL(buf common.BufferWriter, eq map[string]interface{}, args *[]interface{}, anyConditions bool, pos *int64) bool {
	for k, v := range eq {
		if v == nil {
			anyConditions = writeWhereCondition(buf, k, " IS NULL", anyConditions)
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

func writeWhereCondition(buf common.BufferWriter, k string, pred string, anyConditions bool) bool {
	if anyConditions {
		buf.WriteString(" AND (")
	} else {
		buf.WriteRune('(')
		anyConditions = true
	}
	Dialect.WriteIdentifier(buf, k)
	buf.WriteString(pred)
	buf.WriteRune(')')

	return anyConditions
}
