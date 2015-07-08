package dat

type subInfo struct {
	*Expression
	alias string
}

// SelectDocBuilder builds SQL that returns a JSON row.
type SelectDocBuilder struct {
	*SelectBuilder
	subQueries    []*subInfo
	subQueriesOne []*subInfo
	innerSQL      *Expression
	isParent      bool
}

// NewSelectDocBuilder creates an instance of SelectDocBuilder.
func NewSelectDocBuilder(columns ...string) *SelectDocBuilder {
	sb := NewSelectBuilder(columns...)
	return &SelectDocBuilder{SelectBuilder: sb, isParent: true}
}

// InnerSQL sets the SQL after the SELECT (columns...) statement
func (b *SelectDocBuilder) InnerSQL(sql string, a ...interface{}) *SelectDocBuilder {
	b.innerSQL = Expr(sql, a...)
	return b
}

// Many loads a sub query resulting in an array of rows as an alias.
func (b *SelectDocBuilder) Many(column string, sqlOrBuilder interface{}, a ...interface{}) *SelectDocBuilder {
	switch t := sqlOrBuilder.(type) {
	default:
		panic("sqlOrbuilder accepts only {string, Builder, *SelectDocBuilder} type")
	case *SelectDocBuilder:
		t.isParent = false
		b.subQueries = append(b.subQueries, &subInfo{Expr(t.ToSQL()), column})
	case Builder:
		b.subQueries = append(b.subQueries, &subInfo{Expr(t.ToSQL()), column})
	case string:
		b.subQueries = append(b.subQueries, &subInfo{Expr(t, a...), column})
	}
	return b
}

// One loads a query resulting in a single row as an alias.
func (b *SelectDocBuilder) One(column string, sqlOrBuilder interface{}, a ...interface{}) *SelectDocBuilder {
	switch t := sqlOrBuilder.(type) {
	default:
		panic("sqlOrbuilder accepts only {string, Builder, *SelectDocBuilder} type")
	case *SelectDocBuilder:
		t.isParent = false
		b.subQueriesOne = append(b.subQueriesOne, &subInfo{Expr(t.ToSQL()), column})
	case Builder:
		b.subQueriesOne = append(b.subQueriesOne, &subInfo{Expr(t.ToSQL()), column})
	case string:
		b.subQueriesOne = append(b.subQueriesOne, &subInfo{Expr(t, a...), column})
	}
	return b
}

// ToSQL serialized the SelectBuilder to a SQL string
// It returns the string with placeholders and a slice of query arguments
func (b *SelectDocBuilder) ToSQL() (string, []interface{}) {
	if len(b.columns) == 0 {
		panic("no columns specified")
	}
	if len(b.table) == 0 && b.innerSQL == nil {
		panic("no table specified")
	}

	buf := bufPool.Get()
	defer bufPool.Put(buf)
	var args []interface{}
	var placeholderStartPos int64 = 1

	/*
		SELECT
			row_to_json(item.*)
		FROM (
			SELECT 	ID,
				NAME,
				(
					select ARRAY_AGG(dat__1.*)
					from (
						SELECT ID, user_id, title FROM posts WHERE posts.user_id = people.id
					) as dat__1
				) as posts
			FROM
				people
			WHERE
				ID in (1, 2)
		) as item
	*/

	if b.isParent {
		//buf.WriteString("SELECT convert_to(row_to_json(dat__item.*)::text, 'UTF8') FROM ( SELECT ")
		buf.WriteString("SELECT row_to_json(dat__item.*) FROM ( SELECT ")
	} else {
		buf.WriteString("SELECT ")
	}

	if b.isDistinct {
		if len(b.distinctColumns) == 0 {
			buf.WriteString("DISTINCT ")
		} else {
			buf.WriteString("DISTINCT ON (")
			for i, s := range b.distinctColumns {
				if i > 0 {
					buf.WriteString(", ")
				}
				buf.WriteString(s)
			}
			buf.WriteString(") ")
		}
	}

	for i, s := range b.columns {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(s)
	}

	/*
		(
			select ARRAY_AGG(dat__1.*)
			from (
				SELECT ID, user_id, title FROM posts WHERE posts.user_id = people.id
			) as dat__1
		) as posts
	*/

	for _, sub := range b.subQueries {
		buf.WriteString(", (SELECT array_agg(dat__")
		buf.WriteString(sub.alias)
		buf.WriteString(".*) FROM (")
		sub.WriteRelativeArgs(buf, &args, &placeholderStartPos)
		buf.WriteString(") AS dat__")
		buf.WriteString(sub.alias)
		buf.WriteString(") AS ")
		Dialect.WriteIdentifier(buf, sub.alias)
	}

	for _, sub := range b.subQueriesOne {
		buf.WriteString(", (SELECT row_to_json(dat__")
		buf.WriteString(sub.alias)
		buf.WriteString(".*) FROM (")
		sub.WriteRelativeArgs(buf, &args, &placeholderStartPos)
		buf.WriteString(") AS dat__")
		buf.WriteString(sub.alias)
		buf.WriteString(") AS ")
		Dialect.WriteIdentifier(buf, sub.alias)
	}

	if b.innerSQL != nil {
		b.innerSQL.WriteRelativeArgs(buf, &args, &placeholderStartPos)
	} else {
		buf.WriteString(" FROM ")
		buf.WriteString(b.table)

		if b.scope != nil {
			var where string
			sql, args := b.scope.ToSQL(b.table)
			sql, where = splitWhere(sql)
			buf.WriteString(sql)
			if where != "" {
				fragment := newWhereFragment(where, args)
				b.whereFragments = append(b.whereFragments, fragment)
			}
		}

		if len(b.whereFragments) > 0 {
			buf.WriteString(" WHERE ")
			writeWhereFragmentsToSql(buf, b.whereFragments, &args, &placeholderStartPos)
		}

		// if b.scope == nil {
		// 	if len(b.whereFragments) > 0 {
		// 		buf.WriteString(" WHERE ")
		// 		writeWhereFragmentsToSql(buf, b.whereFragments, &args, &placeholderStartPos)
		// 	}
		// } else {
		// 	whereFragment := newWhereFragment(b.scope.ToSQL(b.table))
		// 	writeScopeCondition(buf, whereFragment, &args, &placeholderStartPos)
		// }

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
			writeWhereFragmentsToSql(buf, b.havingFragments, &args, &placeholderStartPos)
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
			writeUint64(buf, b.limitCount)
		}

		if b.offsetValid {
			buf.WriteString(" OFFSET ")
			writeUint64(buf, b.offsetCount)
		}
	}

	if b.isParent {
		buf.WriteString(`) as dat__item`)
	}

	return buf.String(), args
}

//// Copied form SelectBuilder

// Distinct marks the statement as a DISTINCT SELECT
func (b *SelectDocBuilder) Distinct() *SelectDocBuilder {
	b.isDistinct = true
	return b
}

// DistinctOn sets the columns for DISTINCT ON
func (b *SelectDocBuilder) DistinctOn(columns ...string) *SelectDocBuilder {
	b.isDistinct = true
	b.distinctColumns = columns
	return b
}

// From sets the table to SELECT FROM
func (b *SelectDocBuilder) From(from string) *SelectDocBuilder {
	b.table = from
	return b
}

// ScopeMap uses a predefined scope in place of WHERE.
func (b *SelectDocBuilder) ScopeMap(mapScope *MapScope, m M) *SelectDocBuilder {
	b.scope = mapScope.mergeClone(m)
	return b
}

// Scope uses a predefined scope in place of WHERE.
func (b *SelectDocBuilder) Scope(sql string, args ...interface{}) *SelectDocBuilder {
	b.scope = ScopeFunc(func(table string) (string, []interface{}) {
		return escapeScopeTable(sql, table), args
	})
	return b
}

// Where appends a WHERE clause to the statement for the given string and args
// or map of column/value pairs
func (b *SelectDocBuilder) Where(whereSqlOrMap interface{}, args ...interface{}) *SelectDocBuilder {
	b.whereFragments = append(b.whereFragments, newWhereFragment(whereSqlOrMap, args))
	return b
}

// GroupBy appends a column to group the statement
func (b *SelectDocBuilder) GroupBy(group string) *SelectDocBuilder {
	b.groupBys = append(b.groupBys, group)
	return b
}

// Having appends a HAVING clause to the statement
func (b *SelectDocBuilder) Having(whereSqlOrMap interface{}, args ...interface{}) *SelectDocBuilder {
	b.havingFragments = append(b.havingFragments, newWhereFragment(whereSqlOrMap, args))
	return b
}

// OrderBy appends a column to ORDER the statement by
func (b *SelectDocBuilder) OrderBy(ord string) *SelectDocBuilder {
	b.orderBys = append(b.orderBys, ord)
	return b
}

// Limit sets a limit for the statement; overrides any existing LIMIT
func (b *SelectDocBuilder) Limit(limit uint64) *SelectDocBuilder {
	b.limitCount = limit
	b.limitValid = true
	return b
}

// Offset sets an offset for the statement; overrides any existing OFFSET
func (b *SelectDocBuilder) Offset(offset uint64) *SelectDocBuilder {
	b.offsetCount = offset
	b.offsetValid = true
	return b
}

// Paginate sets LIMIT/OFFSET for the statement based on the given page/perPage
// Assumes page/perPage are valid. Page and perPage must be >= 1
func (b *SelectDocBuilder) Paginate(page, perPage uint64) *SelectDocBuilder {
	b.Limit(perPage)
	b.Offset((page - 1) * perPage)
	return b
}
