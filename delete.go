package dat

import (
	"bytes"
	"strconv"
)

// DeleteBuilder contains the clauses for a DELETE statement
type DeleteBuilder struct {
	Executable

	table          string
	whereFragments []*whereFragment
	orderBys       []string
	limitCount     uint
	limitValid     bool
	offsetCount    uint
	offsetValid    bool
}

// NewDeleteBuilder creates a new DeleteBuilder for the given table.
func NewDeleteBuilder(table string) *DeleteBuilder {
	return &DeleteBuilder{table: table}
}

// Where appends a WHERE clause to the statement whereSqlOrMap can be a
// string or map. If it's a string, args wil replaces any places holders
func (b *DeleteBuilder) Where(whereSqlOrMap interface{}, args ...interface{}) *DeleteBuilder {
	b.whereFragments = append(b.whereFragments, newWhereFragment(whereSqlOrMap, args))
	return b
}

// OrderBy appends an ORDER BY clause to the statement
func (b *DeleteBuilder) OrderBy(ord string) *DeleteBuilder {
	b.orderBys = append(b.orderBys, ord)
	return b
}

// Limit sets a LIMIT clause for the statement; overrides any existing LIMIT
func (b *DeleteBuilder) Limit(limit uint) *DeleteBuilder {
	b.limitCount = limit
	b.limitValid = true
	return b
}

// Offset sets an OFFSET clause for the statement; overrides any existing OFFSET
func (b *DeleteBuilder) Offset(offset uint) *DeleteBuilder {
	b.offsetCount = offset
	b.offsetValid = true
	return b
}

// ToSQL serialized the DeleteBuilder to a SQL string
// It returns the string with placeholders and a slice of query arguments
func (b *DeleteBuilder) ToSQL() (string, []interface{}) {
	if len(b.table) == 0 {
		panic("no table specified")
	}

	var sql bytes.Buffer
	var args []interface{}

	sql.WriteString("DELETE FROM ")
	sql.WriteString(b.table)

	var placeholderStartPos int64 = 1

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
		if b.limitCount < maxLookup {
			sql.WriteString(itoaTab[int(b.limitCount)])
		} else {
			sql.WriteString(strconv.FormatUint(uint64(b.limitCount), 10))
		}
	}

	if b.offsetValid {
		sql.WriteString(" OFFSET ")
		if b.offsetCount < maxLookup {
			sql.WriteString(itoaTab[int(b.offsetCount)])
		} else {
			sql.WriteString(strconv.FormatUint(uint64(b.offsetCount), 10))
		}
	}

	return sql.String(), args
}

// Interpolate interpolates this builders sql.
func (b *DeleteBuilder) Interpolate() (string, error) {
	return interpolate(b)
}

// MustInterpolate interpolates this builders sql or panics.
func (b *DeleteBuilder) MustInterpolate() string {
	return mustInterpolate(b)
}
