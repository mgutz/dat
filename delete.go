package dat

import (
	"bytes"
	"strconv"
)

// DeleteBuilder contains the clauses for a DELETE statement
type DeleteBuilder struct {
	Table          string
	WhereFragments []*whereFragment
	OrderBys       []string
	LimitCount     uint64
	LimitValid     bool
	OffsetCount    uint64
	OffsetValid    bool
}

// Where appends a WHERE clause to the statement whereSqlOrMap can be a
// string or map. If it's a string, args wil replaces any places holders
func (b *DeleteBuilder) Where(whereSqlOrMap interface{}, args ...interface{}) *DeleteBuilder {
	b.WhereFragments = append(b.WhereFragments, newWhereFragment(whereSqlOrMap, args))
	return b
}

// OrderBy appends an ORDER BY clause to the statement
func (b *DeleteBuilder) OrderBy(ord string) *DeleteBuilder {
	b.OrderBys = append(b.OrderBys, ord)
	return b
}

// OrderDir appends an ORDER BY clause with a direction to the statement
func (b *DeleteBuilder) OrderDir(ord string, isAsc bool) *DeleteBuilder {
	if isAsc {
		b.OrderBys = append(b.OrderBys, ord+" ASC")
	} else {
		b.OrderBys = append(b.OrderBys, ord+" DESC")
	}
	return b
}

// Limit sets a LIMIT clause for the statement; overrides any existing LIMIT
func (b *DeleteBuilder) Limit(limit uint64) *DeleteBuilder {
	b.LimitCount = limit
	b.LimitValid = true
	return b
}

// Offset sets an OFFSET clause for the statement; overrides any existing OFFSET
func (b *DeleteBuilder) Offset(offset uint64) *DeleteBuilder {
	b.OffsetCount = offset
	b.OffsetValid = true
	return b
}

// ToSQL serialized the DeleteBuilder to a SQL string
// It returns the string with placeholders and a slice of query arguments
func (b *DeleteBuilder) ToSQL() (string, []interface{}) {
	if len(b.Table) == 0 {
		panic("no table specified")
	}

	var sql bytes.Buffer
	var args []interface{}

	sql.WriteString("DELETE FROM ")
	sql.WriteString(b.Table)

	var placeholderStartPos int64 = 1

	// Write WHERE clause if we have any fragments
	if len(b.WhereFragments) > 0 {
		sql.WriteString(" WHERE ")
		writeWhereFragmentsToSql(b.WhereFragments, &sql, &args, &placeholderStartPos)
	}

	// Ordering and limiting
	if len(b.OrderBys) > 0 {
		sql.WriteString(" ORDER BY ")
		for i, s := range b.OrderBys {
			if i > 0 {
				sql.WriteString(", ")
			}
			sql.WriteString(s)
		}
	}

	if b.LimitValid {
		sql.WriteString(" LIMIT ")
		sql.WriteString(strconv.FormatUint(b.LimitCount, 10))
	}

	if b.OffsetValid {
		sql.WriteString(" OFFSET ")
		sql.WriteString(strconv.FormatUint(b.OffsetCount, 10))
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
