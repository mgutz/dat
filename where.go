package dat

import (
	"bytes"
	"reflect"
	"strings"
)

// Eq is a map column -> value pairs which must be matched in a query
type Eq map[string]interface{}

type whereFragment struct {
	Condition   string
	Values      []interface{}
	EqualityMap map[string]interface{}
}

func newWhereFragment(whereSqlOrMap interface{}, args []interface{}) *whereFragment {
	switch pred := whereSqlOrMap.(type) {
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

func remapPlaceholders(buf *bytes.Buffer, statement string, pos int64) int64 {
	if !strings.Contains(statement, "$") {
		buf.WriteString(statement)
		return 0
	}

	//var buf bytes.Buffer
	var discardDigits bool
	var replaced int64
	for _, r := range statement {
		if discardDigits {
			if '0' <= r && r <= '9' {
				continue
			}
			discardDigits = false
		}
		if r != '$' {
			buf.WriteRune(r)
		} else if r == '$' {
			// replace relative $1 with absolute like $4
			writePlaceholder64(buf, pos+replaced)
			replaced++
			discardDigits = true
		}
	}
	return replaced
}

// Invariant: only called when len(fragments) > 0
func writeWhereFragmentsToSql(fragments []*whereFragment, sql *bytes.Buffer, args *[]interface{}, pos *int64) {
	hasConditions := false
	for _, f := range fragments {
		if f.Condition != "" {
			if hasConditions {
				sql.WriteString(" AND (")
			} else {
				sql.WriteRune('(')
				hasConditions = true
			}

			if len(f.Values) > 0 {
				// map relative $1, $2 placeholders to absolute
				replaced := remapPlaceholders(sql, f.Condition, *pos)
				*pos += replaced
				*args = append(*args, f.Values...)
			} else {
				sql.WriteString(f.Condition)
			}
			sql.WriteRune(')')
		} else if f.EqualityMap != nil {
			hasConditions = writeEqualityMapToSql(f.EqualityMap, sql, args, hasConditions, pos)
		} else {
			panic("invalid equality map")
		}
	}
}

func writeEqualityMapToSql(eq map[string]interface{}, sql *bytes.Buffer, args *[]interface{}, anyConditions bool, pos *int64) bool {
	for k, v := range eq {
		if v == nil {
			anyConditions = writeWhereCondition(sql, k, " IS NULL", anyConditions)
		} else {
			vVal := reflect.ValueOf(v)

			if vVal.Kind() == reflect.Array || vVal.Kind() == reflect.Slice {
				vValLen := vVal.Len()
				if vValLen == 0 {
					if vVal.IsNil() {
						anyConditions = writeWhereCondition(sql, k, " IS NULL", anyConditions)
					} else {
						if anyConditions {
							sql.WriteString(" AND (1=0)")
						} else {
							sql.WriteString("(1=0)")
						}
					}
				} else if vValLen == 1 {
					anyConditions = writeWhereCondition(sql, k, equalsPlaceholderTab[*pos], anyConditions)
					*args = append(*args, vVal.Index(0).Interface())
					*pos++
				} else {
					// " IN $n"
					anyConditions = writeWhereCondition(sql, k, inPlaceholderTab[*pos], anyConditions)
					*args = append(*args, v)
					*pos++
				}
			} else {
				anyConditions = writeWhereCondition(sql, k, equalsPlaceholderTab[*pos], anyConditions)
				*args = append(*args, v)
				*pos++
			}
		}
	}

	return anyConditions
}

func writeWhereCondition(sql *bytes.Buffer, k string, pred string, anyConditions bool) bool {
	if anyConditions {
		sql.WriteString(" AND (")
	} else {
		sql.WriteRune('(')
		anyConditions = true
	}
	Dialect.WriteIdentifier(sql, k)
	sql.WriteString(pred)
	sql.WriteRune(')')

	return anyConditions
}
