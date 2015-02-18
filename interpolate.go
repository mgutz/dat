package dat

import (
	"bytes"
	"database/sql/driver"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Need to turn \x00, \n, \r, \, ', " and \x1a
// Returns an escaped, quoted string. eg, "hello 'world'" -> "'hello \'world\''"
func escapeAndQuoteString(val string) string {
	var buf bytes.Buffer

	buf.WriteRune('\'')

	for _, char := range val {
		if char == '\'' { // single quote: ' -> \'
			buf.WriteString("\\'")
		} else if char == '"' { // double quote: " -> \"
			buf.WriteString("\\\"")
		} else if char == '\\' { // slash: \ -> "\\"
			buf.WriteString("\\\\")
		} else if char == '\n' { // control: newline: \n -> "\n"
			buf.WriteString("\\n")
		} else if char == '\r' { // control: return: \r -> "\r"
			buf.WriteString("\\r")
		} else if char == 0 { // control: NUL: 0 -> "\x00"
			buf.WriteString("\\x00")
		} else if char == 0x1a { // control: \x1a -> "\x1a"
			buf.WriteString("\\x1a")
		} else {
			buf.WriteRune(char)
		}
	}

	buf.WriteRune('\'')

	return buf.String()
}

func isUint(k reflect.Kind) bool {
	return k == reflect.Uint ||
		k == reflect.Uint8 ||
		k == reflect.Uint16 ||
		k == reflect.Uint32 ||
		k == reflect.Uint64
}

func isInt(k reflect.Kind) bool {
	return k == reflect.Int ||
		k == reflect.Int8 ||
		k == reflect.Int16 ||
		k == reflect.Int32 ||
		k == reflect.Int64
}

func isFloat(k reflect.Kind) bool {
	return k == reflect.Float32 ||
		k == reflect.Float64
}

// sql is like "id = $1 OR username = $2"
// vals is like []interface{}{4, "bob"}
// NOTE that vals can only have values of certain types:
//   - Integers (signed and unsigned)
//   - floats
//   - strings (that are valid utf-8)
//   - booleans
//   - times
var typeOfTime = reflect.TypeOf(time.Time{})

// Interpolate takes a SQL string with placeholders and a list of arguments to
// replace them with. Returns a blank string and error if the number of placeholders
// does not match the number of arguments.
func Interpolate(sql string, vals []interface{}) (string, error) {
	// Get the number of arguments to add to this query
	maxVals := len(vals)

	// If our query is blank and has no args return early
	// Args with a blank query is an error
	if sql == "" {
		if maxVals != 0 {
			return "", ErrArgumentMismatch
		}
		return "", nil
	}

	lenVals := len(vals)
	hasPlaceholders := strings.Contains(sql, "$1")

	// If we have no args and the query has no place holders return early
	// No args for a query with place holders is an error
	if lenVals == 0 {
		if hasPlaceholders {
			return "", ErrArgumentMismatch
		}
		return sql, nil
	}

	if lenVals > 0 && !hasPlaceholders {
		return "", ErrArgumentMismatch
	}

	var buf bytes.Buffer
	var accumulateDigits bool
	var digits bytes.Buffer

	var writeValue = func(pos int) error {
		if pos < 0 || pos >= lenVals {
			return ErrArgumentMismatch
		}

		v := vals[pos]

		valuer, ok := v.(driver.Valuer)
		if ok {
			val, err := valuer.Value()
			if err != nil {
				return err
			}
			v = val
		}

		valueOfV := reflect.ValueOf(v)
		kindOfV := valueOfV.Kind()

		if v == nil {
			buf.WriteString("NULL")
		} else if _, ok := v.(defaultType); ok {
			buf.WriteString("DEFAULT")
		} else if isInt(kindOfV) {
			var ival = valueOfV.Int()
			buf.WriteString(strconv.FormatInt(ival, 10))
		} else if isUint(kindOfV) {
			var uival = valueOfV.Uint()
			buf.WriteString(strconv.FormatUint(uival, 10))
		} else if kindOfV == reflect.String {
			var str = valueOfV.String()
			if !utf8.ValidString(str) {
				return ErrNotUTF8
			}
			buf.WriteString(escapeAndQuoteString(str))
		} else if isFloat(kindOfV) {
			var fval = valueOfV.Float()
			buf.WriteString(strconv.FormatFloat(fval, 'f', -1, 64))
		} else if kindOfV == reflect.Bool {
			var bval = valueOfV.Bool()
			if bval {
				buf.WriteRune('1')
			} else {
				buf.WriteRune('0')
			}
		} else if kindOfV == reflect.Struct {
			if typeOfV := valueOfV.Type(); typeOfV == typeOfTime {
				t := valueOfV.Interface().(time.Time)
				buf.WriteString(escapeAndQuoteString(t.UTC().Format(timeFormat)))
			} else {
				return ErrInvalidValue
			}
		} else if kindOfV == reflect.Slice {
			typeOfV := reflect.TypeOf(v)
			subtype := typeOfV.Elem()
			kindOfSubtype := subtype.Kind()
			sliceLen := valueOfV.Len()

			if sliceLen == 0 {
				return ErrInvalidSliceLength
			}

			buf.WriteRune('(')
			if isInt(kindOfSubtype) {
				for i := 0; i < sliceLen; i++ {
					if i > 0 {
						buf.WriteRune(',')
					}
					var ival = valueOfV.Index(i).Int()
					buf.WriteString(strconv.FormatInt(ival, 10))
				}
			} else if isUint(kindOfSubtype) {
				for i := 0; i < sliceLen; i++ {
					if i > 0 {
						buf.WriteRune(',')
					}
					var uival = valueOfV.Index(i).Uint()
					buf.WriteString(strconv.FormatUint(uival, 10))
				}
			} else if kindOfSubtype == reflect.String {
				for i := 0; i < sliceLen; i++ {
					if i > 0 {
						buf.WriteRune(',')
					}
					var str = valueOfV.Index(i).String()
					if !utf8.ValidString(str) {
						return ErrNotUTF8
					}
					buf.WriteString(escapeAndQuoteString(str))
				}
			} else {
				return ErrInvalidSliceValue
			}
			buf.WriteRune(')')
		} else {
			return ErrInvalidValue
		}
		return nil
	}

	lenSql := len(sql)
	done := false
	for i, r := range sql {
		if accumulateDigits {
			if '0' <= r && r <= '9' {
				digits.WriteRune(r)
				if i < lenSql-1 {
					continue
				}
				// the last rune is part of a placeholder and its value must be
				// written
				done = true
			}

			pos, _ := strconv.Atoi(digits.String())
			err := writeValue(pos - 1)
			if err != nil {
				return "", err
			}

			if done {
				break
			}
			accumulateDigits = false
		}

		if r == '$' {
			digits.Reset()
			accumulateDigits = true
			continue
		}
		buf.WriteRune(r)
	}

	return buf.String(), nil
}

func interpolate(builder Builder) (string, error) {
	sql, args := builder.ToSQL()
	return Interpolate(sql, args)
}

func mustInterpolate(builder Builder) string {
	sql, args := builder.ToSQL()

	fullSql, err := Interpolate(sql, args)
	if err != nil {
		panic(events.EventErrKv("mustInterpolate", err, kvs{"sql": fullSql}))
	}
	return fullSql
}
