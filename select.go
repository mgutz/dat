package dat

import (
	"bytes"
	"strconv"
)

// SelectBuilder contains the clauses for a SELECT statement
type SelectBuilder struct {
	Executable

	isDistinct      bool
	columns         []string
	table           string
	whereFragments  []*whereFragment
	groupBys        []string
	havingFragments []*whereFragment
	orderBys        []string
	limitCount      uint64
	limitValid      bool
	offsetCount     uint64
	offsetValid     bool
}

// NewSelectBuilder creates a new SelectBuilder for the given columns
func NewSelectBuilder(columns ...string) *SelectBuilder {
	return &SelectBuilder{columns: columns}
}

// Distinct marks the statement as a DISTINCT SELECT
func (b *SelectBuilder) Distinct() *SelectBuilder {
	b.isDistinct = true
	return b
}

// From sets the table to SELECT FROM
func (b *SelectBuilder) From(from string) *SelectBuilder {
	b.table = from
	return b
}

// Where appends a WHERE clause to the statement for the given string and args
// or map of column/value pairs
func (b *SelectBuilder) Where(whereSqlOrMap interface{}, args ...interface{}) *SelectBuilder {
	b.whereFragments = append(b.whereFragments, newWhereFragment(whereSqlOrMap, args))
	return b
}

// GroupBy appends a column to group the statement
func (b *SelectBuilder) GroupBy(group string) *SelectBuilder {
	b.groupBys = append(b.groupBys, group)
	return b
}

// Having appends a HAVING clause to the statement
func (b *SelectBuilder) Having(whereSqlOrMap interface{}, args ...interface{}) *SelectBuilder {
	b.havingFragments = append(b.havingFragments, newWhereFragment(whereSqlOrMap, args))
	return b
}

// OrderBy appends a column to ORDER the statement by
func (b *SelectBuilder) OrderBy(ord string) *SelectBuilder {
	b.orderBys = append(b.orderBys, ord)
	return b
}

// Limit sets a limit for the statement; overrides any existing LIMIT
func (b *SelectBuilder) Limit(limit uint64) *SelectBuilder {
	b.limitCount = limit
	b.limitValid = true
	return b
}

// Offset sets an offset for the statement; overrides any existing OFFSET
func (b *SelectBuilder) Offset(offset uint64) *SelectBuilder {
	b.offsetCount = offset
	b.offsetValid = true
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
	if len(b.columns) == 0 {
		panic("no columns specified")
	}
	if len(b.table) == 0 {
		panic("no table specified")
	}

	var sql bytes.Buffer
	var args []interface{}

	sql.WriteString("SELECT ")

	if b.isDistinct {
		sql.WriteString("DISTINCT ")
	}

	for i, s := range b.columns {
		if i > 0 {
			sql.WriteString(", ")
		}
		sql.WriteString(s)
	}

	sql.WriteString(" FROM ")
	sql.WriteString(b.table)

	var placeholderStartPos int64 = 1
	if len(b.whereFragments) > 0 {
		sql.WriteString(" WHERE ")
		writeWhereFragmentsToSql(b.whereFragments, &sql, &args, &placeholderStartPos)
	}

	if len(b.groupBys) > 0 {
		sql.WriteString(" GROUP BY ")
		for i, s := range b.groupBys {
			if i > 0 {
				sql.WriteString(", ")
			}
			sql.WriteString(s)
		}
	}

	if len(b.havingFragments) > 0 {
		sql.WriteString(" HAVING ")
		writeWhereFragmentsToSql(b.havingFragments, &sql, &args, &placeholderStartPos)
	}

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
			sql.WriteString(strconv.FormatUint(b.limitCount, 10))
		}
	}

	if b.offsetValid {
		sql.WriteString(" OFFSET ")
		if b.offsetCount < maxLookup {
			sql.WriteString(itoaTab[int(b.offsetCount)])
		} else {
			sql.WriteString(strconv.FormatUint(b.offsetCount, 10))
		}
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
