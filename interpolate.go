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
func Interpolate(sql string, vals []interface{}) (string, []interface{}, error) {
	// Get the number of arguments to add to this query
	lenVals := len(vals)

	// if there are any []byte types, just pass it through to save memory allocations
	for i := 0; i < lenVals; i++ {
		if _, ok := vals[i].([]byte); ok {
			return sql, vals, nil
		} else if _, ok := vals[i].(*[]byte); ok {
			return sql, vals, nil
		}
	}

	// If our query is blank and has no args return early
	// Args with a blank query is an error
	if sql == "" {
		if lenVals != 0 {
			return "", nil, logger.Error("Interpolation error", "err", ErrArgumentMismatch, "sql", sql, "args", vals)
		}
		return "", nil, nil
	}

	if Strict {
		hasPlaceholders := strings.Contains(sql, "$")

		// If we have no args and the query has no place holders return early
		// No args for a query with place holders is an error
		if lenVals == 0 {
			if hasPlaceholders {
				return "", nil, logger.Error("Interpolation error", "err", ErrArgumentMismatch, "sql", sql, "args", vals)
			}
			return sql, nil, nil
		}

		if lenVals > 0 && !hasPlaceholders {
			return "", nil, logger.Error("Interpolation error", "err", ErrArgumentMismatch, "sql", sql, "args", vals)
		}

		if !hasPlaceholders {
			return sql, nil, nil
		}
	}

	buf := bufPool.Get()
	defer bufPool.Put(buf)
	var accumulateDigits bool
	var digits bytes.Buffer

	newPlaceholderIndex := 0
	var newArgs []interface{}

	var writeValue = func(pos int) error {
		if pos < 0 || pos >= lenVals {
			return ErrArgumentMismatch
		}

		v := vals[pos]

		// mark any arguments not handled with a new placeholder
		// and the arg to the new arguments slice
		var passthroughArg = func(values ...interface{}) {
			newPlaceholderIndex++
			newArgs = append(newArgs, values...)
			writePlaceholder(buf, newPlaceholderIndex)
		}

		if val, ok := v.(UnsafeString); ok {
			buf.WriteString(string(val))
			return nil
		} else if _, ok := v.(JSON); ok {
			valueOfV := reflect.ValueOf(v)
			if valueOfV.IsNil() {
				buf.WriteString("NULL")
				return nil
			}

			passthroughArg(v)
			return nil
		} else if valuer, ok := v.(Expressioner); ok {
			valueOfV := reflect.ValueOf(v)
			if valueOfV.IsNil() {
				buf.WriteString("NULL")
				return nil
			}

			s, args, err := valuer.Expression()
			if err != nil {
				return err
			}

			buf.WriteString(s)
			if len(args) > 0 {
				passthroughArg(args...)
			}
			return nil
		} else if valuer, ok := v.(Interpolator); ok {
			valueOfV := reflect.ValueOf(v)
			if valueOfV.IsNil() {
				buf.WriteString("NULL")
				return nil
			}

			s, err := valuer.Interpolate()
			if err != nil {
				return err
			}
			Dialect.WriteStringLiteral(buf, s)
			return nil
		} else if valuer, ok := v.(driver.Valuer); ok {
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
			return nil
		}

		// Dereference pointer values
		if kindOfV == reflect.Ptr {
			if valueOfV.IsNil() {
				buf.WriteString("NULL")
				return nil
			}
			valueOfV = valueOfV.Elem()
			kindOfV = valueOfV.Kind()
		}

		if kindOfV == reflect.String {
			var str = valueOfV.String()
			if !utf8.ValidString(str) {
				return ErrNotUTF8
			}
			Dialect.WriteStringLiteral(buf, str)
		} else if isInt(kindOfV) {
			var ival = valueOfV.Int()
			writeInt64(buf, ival)
		} else if isUint(kindOfV) {
			var uival = valueOfV.Uint()
			writeUint64(buf, uival)
		} else if isFloat(kindOfV) {
			var fval = valueOfV.Float()
			buf.WriteString(strconv.FormatFloat(fval, 'f', -1, 64))
		} else if kindOfV == reflect.Bool {
			var bval = valueOfV.Bool()
			if bval {
				buf.WriteString(`'t'`)
			} else {
				buf.WriteString(`'f'`)
			}
		} else if kindOfV == reflect.Struct {
			if typeOfV := valueOfV.Type(); typeOfV == typeOfTime {
				t := valueOfV.Interface().(time.Time)
				Dialect.WriteFormattedTime(buf, t)
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
					writeInt64(buf, ival)
				}
			} else if isUint(kindOfSubtype) {
				for i := 0; i < sliceLen; i++ {
					if i > 0 {
						buf.WriteRune(',')
					}
					var uival = valueOfV.Index(i).Uint()
					writeUint64(buf, uival)
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
					Dialect.WriteStringLiteral(buf, str)
				}
			} else {
				return ErrInvalidSliceValue
			}
			buf.WriteRune(')')
		} else {
			passthroughArg(v)
		}

		return nil
	}

	lenSQL := len(sql)
	done := false
	for i, r := range sql {
		if accumulateDigits {
			if '0' <= r && r <= '9' {
				digits.WriteRune(r)
				if i < lenSQL-1 {
					continue
				}
				// last rune is part of a placeholder, fallthrough
				done = true
			}

			digitsStr := digits.String()
			// can be empty $ is followed by a non-digit
			if digitsStr == "" {
				buf.WriteRune('$')
				buf.WriteRune(r)
				accumulateDigits = false
				continue
			}

			pos := 0
			if len(digitsStr) > 2 {
				pos, _ = strconv.Atoi(digitsStr)
			} else {
				pos, _ = atoiTab[digitsStr]
			}
			err := writeValue(pos - 1)
			if err != nil {
				return "", nil, err
			}

			if done {
				break
			}
			accumulateDigits = false
		}

		if r == '$' && i < lenSQL-1 {
			digits.Reset()
			accumulateDigits = true
			continue
		}

		buf.WriteRune(r)
	}

	return buf.String(), newArgs, nil
}

func interpolate(builder Builder) (string, []interface{}, error) {
	sql, args := builder.ToSQL()
	if builder.IsInterpolated() {
		return Interpolate(sql, args)
	}
	return sql, args, nil
}
