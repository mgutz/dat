package dat

import (
	"bytes"
	"fmt"
	"strconv"
)

// UpdateBuilder contains the clauses for an UPDATE statement
type UpdateBuilder struct {
	Executable

	table          string
	setClauses     []*setClause
	whereFragments []*whereFragment
	orderBys       []string
	limitCount     uint64
	limitValid     bool
	offsetCount    uint64
	offsetValid    bool
	returnings     []string
}

type setClause struct {
	column string
	value  interface{}
}

// NewUpdateBuilder creates a new UpdateBuilder for the given table
func NewUpdateBuilder(table string) *UpdateBuilder {
	return &UpdateBuilder{table: table}
}

// Set appends a column/value pair for the statement
func (b *UpdateBuilder) Set(column string, value interface{}) *UpdateBuilder {
	b.setClauses = append(b.setClauses, &setClause{column: column, value: value})
	return b
}

// SetMap appends the elements of the map as column/value pairs for the statement
func (b *UpdateBuilder) SetMap(clauses map[string]interface{}) *UpdateBuilder {
	for col, val := range clauses {
		b = b.Set(col, val)
	}
	return b
}

// Where appends a WHERE clause to the statement
func (b *UpdateBuilder) Where(whereSqlOrMap interface{}, args ...interface{}) *UpdateBuilder {
	b.whereFragments = append(b.whereFragments, newWhereFragment(whereSqlOrMap, args))
	return b
}

// OrderBy appends a column to ORDER the statement by
func (b *UpdateBuilder) OrderBy(ord string) *UpdateBuilder {
	b.orderBys = append(b.orderBys, ord)
	return b
}

// Limit sets a limit for the statement; overrides any existing LIMIT
func (b *UpdateBuilder) Limit(limit uint64) *UpdateBuilder {
	b.limitCount = limit
	b.limitValid = true
	return b
}

// Offset sets an offset for the statement; overrides any existing OFFSET
func (b *UpdateBuilder) Offset(offset uint64) *UpdateBuilder {
	b.offsetCount = offset
	b.offsetValid = true
	return b
}

// Returning sets the columns for the RETURNING clause
func (b *UpdateBuilder) Returning(columns ...string) *UpdateBuilder {
	b.returnings = columns
	return b
}

// ToSQL serialized the UpdateBuilder to a SQL string
// It returns the string with placeholders and a slice of query arguments
func (b *UpdateBuilder) ToSQL() (string, []interface{}) {
	if len(b.table) == 0 {
		panic("no table specified")
	}
	if len(b.setClauses) == 0 {
		panic("no set clauses specified")
	}

	var sql bytes.Buffer
	var args []interface{}

	sql.WriteString("UPDATE ")
	sql.WriteString(b.table)
	sql.WriteString(" SET ")

	var placeholderStartPos int64 = 1

	// Build SET clause SQL with placeholders and add values to args
	for i, c := range b.setClauses {
		if i > 0 {
			sql.WriteString(", ")
		}
		Quoter.WriteQuotedColumn(c.column, &sql)
		if e, ok := c.value.(*expression); ok {
			start := placeholderStartPos
			sql.WriteString(" = ")
			// map relative $1, $2 placeholders to absolute
			remapPlaceholders(&sql, e.Sql, start)
			args = append(args, e.Values...)
			placeholderStartPos += int64(len(e.Values))
		} else {
			if i < maxLookup {
				sql.WriteString(equalsPlaceholderTab[placeholderStartPos])
			} else {
				sql.WriteString(" = $")
				sql.WriteString(strconv.FormatInt(placeholderStartPos, 10))
			}
			placeholderStartPos++
			args = append(args, c.value)
		}
	}

	// Write WHERE clause if we have any fragments
	if len(b.whereFragments) > 0 {
		sql.WriteString(" WHERE ")
		writeWhereFragmentsToSql(b.whereFragments, &sql, &args, &placeholderStartPos)
	}

	// Ordering and limiting
	if len(b.orderBys) > 0 {
		sql.WriteString(" ORDER BY ")
		for i, s := range b.orderBys {
			if i > 0 {
				sql.WriteString(", ")
			}
			sql.WriteString(s)
		}
	}

	if b.limitValid {
		sql.WriteString(" LIMIT ")
		fmt.Fprint(&sql, b.limitCount)
	}

	if b.offsetValid {
		sql.WriteString(" OFFSET ")
		fmt.Fprint(&sql, b.offsetCount)
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
func (b *UpdateBuilder) Interpolate() (string, error) {
	return interpolate(b)
}

// MustInterpolate interpolates this builders sql or panics.
func (b *UpdateBuilder) MustInterpolate() string {
	return mustInterpolate(b)
}
