package dat

import (
	"bytes"
	"reflect"
	"strconv"
)

// InsertBuilder contains the clauses for an INSERT statement
type InsertBuilder struct {
	Table      string
	Cols       []string
	Vals       [][]interface{}
	Records    []interface{}
	Returnings []string
}

// Columns appends columns to insert in the statement
func (b *InsertBuilder) Columns(columns ...string) *InsertBuilder {
	b.Cols = columns
	return b
}

// Values appends a set of values to the statement
func (b *InsertBuilder) Values(vals ...interface{}) *InsertBuilder {
	b.Vals = append(b.Vals, vals)
	return b
}

// Record pulls in values to match Columns from the record
func (b *InsertBuilder) Record(record interface{}) *InsertBuilder {
	b.Records = append(b.Records, record)
	return b
}

// Returning sets the columns for the RETURNING clause
func (b *InsertBuilder) Returning(columns ...string) *InsertBuilder {
	b.Returnings = columns
	return b
}

// Pair adds a key/value pair to the statement
func (b *InsertBuilder) Pair(column string, value interface{}) *InsertBuilder {
	b.Cols = append(b.Cols, column)
	lenVals := len(b.Vals)
	if lenVals == 0 {
		args := []interface{}{value}
		b.Vals = [][]interface{}{args}
	} else if lenVals == 1 {
		b.Vals[0] = append(b.Vals[0], value)
	} else {
		panic("pair only allows you to specify 1 record to insret")
	}
	return b
}

func buildPlaceholders(start, length int) string {
	var buf bytes.Buffer

	// Build the placeholder like "($1,$2,$3)"
	buf.WriteRune('(')
	for i := start; i < start+length; i++ {
		if i > start {
			buf.WriteRune(',')
		}
		buf.WriteRune('$')
		buf.WriteString(strconv.Itoa(i))
	}
	buf.WriteRune(')')
	return buf.String()
}

// ToSql serialized the InsertBuilder to a SQL string
// It returns the string with placeholders and a slice of query arguments
func (b *InsertBuilder) ToSQL() (string, []interface{}) {
	if len(b.Table) == 0 {
		panic("no table specified")
	}
	if len(b.Cols) == 0 {
		panic("no columns specified")
	}
	if len(b.Vals) == 0 && len(b.Records) == 0 {
		panic("no values or records specified")
	}

	var sql bytes.Buffer
	var args []interface{}

	sql.WriteString("INSERT INTO ")
	sql.WriteString(b.Table)
	sql.WriteString(" (")

	for i, c := range b.Cols {
		if i > 0 {
			sql.WriteRune(',')
		}
		Quoter.WriteQuotedColumn(c, &sql)
	}
	sql.WriteString(") VALUES ")

	start := 1
	// Go thru each value we want to insert. Write the placeholders, and collect args
	for i, row := range b.Vals {
		if i > 0 {
			sql.WriteRune(',')
		}
		sql.WriteString(buildPlaceholders(start, len(row)))

		for _, v := range row {
			args = append(args, v)
			start++
		}
	}
	anyVals := len(b.Vals) > 0

	// Go thru the records. Write the placeholders, and do reflection on the records to extract args
	for i, rec := range b.Records {
		if i > 0 || anyVals {
			sql.WriteRune(',')
		}

		ind := reflect.Indirect(reflect.ValueOf(rec))
		vals, err := ValuesFor(ind.Type(), ind, b.Cols)
		if err != nil {
			panic(err.Error())
		}
		sql.WriteString(buildPlaceholders(start, len(vals)))
		for _, v := range vals {
			args = append(args, v)
			start++
		}
	}

	// Go thru the returning clauses
	for i, c := range b.Returnings {
		if i == 0 {
			sql.WriteString(" RETURNING ")
		} else {
			sql.WriteRune(',')
		}
		Quoter.WriteQuotedColumn(c, &sql)
	}

	return sql.String(), args
}

// Interpolate interpolates this builders sql.
func (b *InsertBuilder) Interpolate() (string, error) {
	return interpolate(b)
}

// MustInterpolate must interpolate or panic.
func (b *InsertBuilder) MustInterpolate() string {
	return mustInterpolate(b)
}
