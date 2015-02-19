package dat

import (
	"bytes"
	"fmt"
)

// SelectBuilder contains the clauses for a SELECT statement
type SelectBuilder struct {
	Executable

	IsDistinct      bool
	Columns         []string
	Table           string
	WhereFragments  []*whereFragment
	GroupBys        []string
	HavingFragments []*whereFragment
	OrderBys        []string
	LimitCount      uint64
	LimitValid      bool
	OffsetCount     uint64
	OffsetValid     bool
}

// NewSelectBuilder creates a new SelectBuilder for the given columns
func NewSelectBuilder(columns ...string) *SelectBuilder {
	return &SelectBuilder{Columns: columns}
}

// Distinct marks the statement as a DISTINCT SELECT
func (b *SelectBuilder) Distinct() *SelectBuilder {
	b.IsDistinct = true
	return b
}

// From sets the table to SELECT FROM
func (b *SelectBuilder) From(from string) *SelectBuilder {
	b.Table = from
	return b
}

// Where appends a WHERE clause to the statement for the given string and args
// or map of column/value pairs
func (b *SelectBuilder) Where(whereSqlOrMap interface{}, args ...interface{}) *SelectBuilder {
	b.WhereFragments = append(b.WhereFragments, newWhereFragment(whereSqlOrMap, args))
	return b
}

// GroupBy appends a column to group the statement
func (b *SelectBuilder) GroupBy(group string) *SelectBuilder {
	b.GroupBys = append(b.GroupBys, group)
	return b
}

// Having appends a HAVING clause to the statement
func (b *SelectBuilder) Having(whereSqlOrMap interface{}, args ...interface{}) *SelectBuilder {
	b.HavingFragments = append(b.HavingFragments, newWhereFragment(whereSqlOrMap, args))
	return b
}

// OrderBy appends a column to ORDER the statement by
func (b *SelectBuilder) OrderBy(ord string) *SelectBuilder {
	b.OrderBys = append(b.OrderBys, ord)
	return b
}

// Limit sets a limit for the statement; overrides any existing LIMIT
func (b *SelectBuilder) Limit(limit uint64) *SelectBuilder {
	b.LimitCount = limit
	b.LimitValid = true
	return b
}

// Offset sets an offset for the statement; overrides any existing OFFSET
func (b *SelectBuilder) Offset(offset uint64) *SelectBuilder {
	b.OffsetCount = offset
	b.OffsetValid = true
	return b
}

// Paginate sets LIMIT/OFFSET for the statement based on the given page/perPage
// Assumes page/perPage are valid. Page and perPage must be >= 1
func (b *SelectBuilder) Paginate(page, perPage uint64) *SelectBuilder {
	b.Limit(perPage)
	b.Offset((page - 1) * perPage)
	return b
}

// ToSQL serialized the SelectBuilder to a SQL string
// It returns the string with placeholders and a slice of query arguments
func (b *SelectBuilder) ToSQL() (string, []interface{}) {
	if len(b.Columns) == 0 {
		panic("no columns specified")
	}
	if len(b.Table) == 0 {
		panic("no table specified")
	}

	var sql bytes.Buffer
	var args []interface{}

	sql.WriteString("SELECT ")

	if b.IsDistinct {
		sql.WriteString("DISTINCT ")
	}

	for i, s := range b.Columns {
		if i > 0 {
			sql.WriteString(", ")
		}
		sql.WriteString(s)
	}

	sql.WriteString(" FROM ")
	sql.WriteString(b.Table)

	var placeholderStartPos int64 = 1
	if len(b.WhereFragments) > 0 {
		sql.WriteString(" WHERE ")
		writeWhereFragmentsToSql(b.WhereFragments, &sql, &args, &placeholderStartPos)
	}

	if len(b.GroupBys) > 0 {
		sql.WriteString(" GROUP BY ")
		for i, s := range b.GroupBys {
			if i > 0 {
				sql.WriteString(", ")
			}
			sql.WriteString(s)
		}
	}

	if len(b.HavingFragments) > 0 {
		sql.WriteString(" HAVING ")
		writeWhereFragmentsToSql(b.HavingFragments, &sql, &args, &placeholderStartPos)
	}

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
		fmt.Fprint(&sql, b.LimitCount)
	}

	if b.OffsetValid {
		sql.WriteString(" OFFSET ")
		fmt.Fprint(&sql, b.OffsetCount)
	}

	return sql.String(), args
}

// Interpolate interpolates this builders sql.
func (b *SelectBuilder) Interpolate() (string, error) {
	return interpolate(b)
}

// MustInterpolate interpolates this builders sql or panics.
func (b *SelectBuilder) MustInterpolate() string {
	return mustInterpolate(b)
}
