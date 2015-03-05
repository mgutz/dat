package dat

import "bytes"

// SelectBuilder contains the clauses for a SELECT statement
type SelectBuilder struct {
	Execer

	isDistinct      bool
	isInterpolated  bool
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
	scope           *Scope
}

// NewSelectBuilder creates a new SelectBuilder for the given columns
func NewSelectBuilder(columns ...string) *SelectBuilder {
	if len(columns) == 0 || columns[0] == "" {
		logger.Error("Select requires 1 or more columns")
		return nil
	}
	return &SelectBuilder{columns: columns, isInterpolated: EnableInterpolation}
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

// Scope uses a predefined scope in place of WHERE.
func (b *SelectBuilder) Scope(sc *Scope, override M) *SelectBuilder {
	b.scope = sc.cloneMerge(override)
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

	var buf bytes.Buffer
	var args []interface{}

	buf.WriteString("SELECT ")

	if b.isDistinct {
		buf.WriteString("DISTINCT ")
	}

	for i, s := range b.columns {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(s)
	}

	buf.WriteString(" FROM ")
	buf.WriteString(b.table)

	var placeholderStartPos int64 = 1
	if b.scope == nil {
		if len(b.whereFragments) > 0 {
			buf.WriteString(" WHERE ")
			writeWhereFragmentsToSql(b.whereFragments, &buf, &args, &placeholderStartPos)
		}
	} else {
		whereFragment := newWhereFragment(b.scope.toSQL(b.table))
		writeScopeCondition(whereFragment, &buf, &args, &placeholderStartPos)
	}

	if len(b.groupBys) > 0 {
		buf.WriteString(" GROUP BY ")
		for i, s := range b.groupBys {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(s)
		}
	}

	if len(b.havingFragments) > 0 {
		buf.WriteString(" HAVING ")
		writeWhereFragmentsToSql(b.havingFragments, &buf, &args, &placeholderStartPos)
	}

	if len(b.orderBys) > 0 {
		buf.WriteString(" ORDER BY ")
		for i, s := range b.orderBys {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(s)
		}
	}

	if b.limitValid {
		buf.WriteString(" LIMIT ")
		writeUint64(&buf, b.limitCount)
	}

	if b.offsetValid {
		buf.WriteString(" OFFSET ")
		writeUint64(&buf, b.offsetCount)
	}

	return buf.String(), args
}

// Interpolate interpolates this builders sql.
func (b *SelectBuilder) Interpolate() (string, []interface{}, error) {
	return interpolate(b)
}

// IsInterpolated determines if this builder will interpolate when
// Interpolate() is called.
func (b *SelectBuilder) IsInterpolated() bool {
	return b.isInterpolated
}

// SetIsInterpolated sets whether this builder should interpolate.
func (b *SelectBuilder) SetIsInterpolated(enable bool) *SelectBuilder {
	b.isInterpolated = enable
	return b
}
