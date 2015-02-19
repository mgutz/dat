package dat

import (
	"bytes"
	"reflect"
	"strconv"
)

// InsertBuilder contains the clauses for an INSERT statement
type InsertBuilder struct {
	Executable

	table      string
	cols       []string
	vals       [][]interface{}
	records    []interface{}
	returnings []string
}

// NewInsertBuilder creates a new InsertBuilder for the given table.
func NewInsertBuilder(table string) *InsertBuilder {
	return &InsertBuilder{table: table}
}

// Columns appends columns to insert in the statement
func (b *InsertBuilder) Columns(columns ...string) *InsertBuilder {
	b.cols = columns
	return b
}

// Values appends a set of values to the statement
func (b *InsertBuilder) Values(vals ...interface{}) *InsertBuilder {
	b.vals = append(b.vals, vals)
	return b
}

// Record pulls in values to match Columns from the record
func (b *InsertBuilder) Record(record interface{}) *InsertBuilder {
	b.records = append(b.records, record)
	return b
}

// Returning sets the columns for the RETURNING clause
func (b *InsertBuilder) Returning(columns ...string) *InsertBuilder {
	b.returnings = columns
	return b
}

// Pair adds a key/value pair to the statement
func (b *InsertBuilder) Pair(column string, value interface{}) *InsertBuilder {
	b.cols = append(b.cols, column)
	lenVals := len(b.vals)
	if lenVals == 0 {
		args := []interface{}{value}
		b.vals = [][]interface{}{args}
	} else if lenVals == 1 {
		b.vals[0] = append(b.vals[0], value)
	} else {
		panic("pair only allows you to specify 1 record to insret")
	}
	return b
}

func buildPlaceholders(buf *bytes.Buffer, start, length int) {
	// Build the placeholder like "($1,$2,$3)"
	buf.WriteRune('(')
	for i := start; i < start+length; i++ {
		if i > start {
			buf.WriteRune(',')
		}
		if i < maxLookup {
			buf.WriteString(placeholderTab[i])
		} else {
			buf.WriteRune('$')
			buf.WriteString(strconv.Itoa(i))
		}
	}
	buf.WriteRune(')')
}

// ToSQL serialized the InsertBuilder to a SQL string
// It returns the string with placeholders and a slice of query arguments
func (b *InsertBuilder) ToSQL() (string, []interface{}) {
	if len(b.table) == 0 {
		panic("no table specified")
	}
	if len(b.cols) == 0 {
		panic("no columns specified")
	}
	if len(b.vals) == 0 && len(b.records) == 0 {
		panic("no values or records specified")
	}

	var sql bytes.Buffer
	var args []interface{}

	sql.WriteString("INSERT INTO ")
	sql.WriteString(b.table)
	sql.WriteString(" (")

	for i, c := range b.cols {
		if i > 0 {
			sql.WriteRune(',')
		}
		Quoter.WriteQuotedColumn(c, &sql)
	}
	sql.WriteString(") VALUES ")

	start := 1
	// Go thru each value we want to insert. Write the placeholders, and collect args
	for i, row := range b.vals {
		if i > 0 {
			sql.WriteRune(',')
		}
		buildPlaceholders(&sql, start, len(row))

		for _, v := range row {
			args = append(args, v)
			start++
		}
	}
	anyVals := len(b.vals) > 0

	// Go thru the records. Write the placeholders, and do reflection on the records to extract args
	for i, rec := range b.records {
		if i > 0 || anyVals {
			sql.WriteRune(',')
		}

		ind := reflect.Indirect(reflect.ValueOf(rec))
		vals, err := ValuesFor(ind.Type(), ind, b.cols)
		if err != nil {
			panic(err.Error())
		}
		buildPlaceholders(&sql, start, len(vals))
		for _, v := range vals {
			args = append(args, v)
			start++
		}
	}

	// Go thru the returning clauses
	for i, c := range b.returnings {
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
