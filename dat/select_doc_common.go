/*
	NOTE: This is a copy of select_common.go with s/SelectDocBuilder/SelectDocBuilder/g
	DO NOT EDIT BELOW
*/

package dat

import "errors"

// Columns adds additional select columns to the builder.
func (b *SelectDocBuilder) Columns(columns ...string) *SelectDocBuilder {
	if len(columns) == 0 || columns[0] == "" {
		b.err = errors.New("Select requires 1 or more columns")
		return nil
	}

	b.columns = append(b.columns, columns...)
	return b
}

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

// From sets the table to SELECT FROM. JOINs may also be defined here.
func (b *SelectDocBuilder) From(from string) *SelectDocBuilder {
	b.table = from
	return b
}

// For adds FOR clause to SELECT.
func (b *SelectDocBuilder) For(options ...string) *SelectDocBuilder {
	b.fors = options
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
		return sql, args
	})
	return b
}

// Where appends a WHERE clause to the statement for the given string and args
// or map of column/value pairs
func (b *SelectDocBuilder) Where(whereSQLOrMap interface{}, args ...interface{}) *SelectDocBuilder {
	fragment, err := newWhereFragment(whereSQLOrMap, args)
	if err != nil {
		b.err = err
		return b
	}
	b.whereFragments = append(b.whereFragments, fragment)
	return b
}

// GroupBy appends a column to group the statement
func (b *SelectDocBuilder) GroupBy(group string) *SelectDocBuilder {
	b.groupBys = append(b.groupBys, group)
	return b
}

// Having appends a HAVING clause to the statement
func (b *SelectDocBuilder) Having(whereSQLOrMap interface{}, args ...interface{}) *SelectDocBuilder {
	fragment, err := newWhereFragment(whereSQLOrMap, args)
	if err != nil {
		b.err = err
	} else {
		b.havingFragments = append(b.havingFragments, fragment)
	}
	return b
}

// OrderBy appends a column to ORDER the statement by
func (b *SelectDocBuilder) OrderBy(whereSQLOrMap interface{}, args ...interface{}) *SelectDocBuilder {
	fragment, err := newWhereFragment(whereSQLOrMap, args)
	if err != nil {
		b.err = err
	} else {
		b.orderBys = append(b.orderBys, fragment)
	}
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

// Join ...
func (b *SelectDocBuilder) Join(sql string, alias string, joinClause string) *SelectDocBuilder {
	b.joins = append(b.joins, &joinExpression{
		subQuery: sql,
		alias:    alias,
		kind:     InnerJoin,
		expr:     joinClause,
	})
	return b
}

// LeftJoin is LEFT OUTER JOIN
func (b *SelectDocBuilder) LeftJoin(sql string, alias string, joinClause string) *SelectDocBuilder {
	b.joins = append(b.joins, &joinExpression{
		subQuery: sql,
		alias:    alias,
		kind:     LeftOuterJoin,
		expr:     joinClause,
	})
	return b
}

// RightJoin is RIGHT  OUTER JOIN
func (b *SelectDocBuilder) RightJoin(sql string, alias string, joinClause string) *SelectDocBuilder {
	b.joins = append(b.joins, &joinExpression{
		subQuery: sql,
		alias:    alias,
		kind:     RightOuterJoin,
		expr:     joinClause,
	})
	return b
}

// FullJoin is FULL OUTER JOIN
func (b *SelectDocBuilder) FullJoin(sql string, alias string, joinClause string) *SelectDocBuilder {
	b.joins = append(b.joins, &joinExpression{
		subQuery: sql,
		alias:    alias,
		kind:     FullOuterJoin,
		expr:     joinClause,
	})
	return b
}

// With loads a sub query that will be inserted as a "with" table
func (b *SelectDocBuilder) With(column string, sqlOrBuilder interface{}, a ...interface{}) *SelectDocBuilder {
	b.err = storeExpr(&b.subQueriesWith, "SelectDocBuilder.With", column, sqlOrBuilder, a...)
	return b
}

// Union unions with another query or builder.
func (b *SelectDocBuilder) Union(sqlOrBuilder interface{}, a ...interface{}) *SelectDocBuilder {
	switch t := sqlOrBuilder.(type) {
	default:
		b.err = NewError("SelectDocBuilder.Union: sqlOrbuilder accepts only {string, Builder, *SelectDocBuilder} type")
	case *JSQLBuilder:
		t.isParent = false
		sql, args, err := t.ToSQL()
		if err != nil {
			b.err = err
			return b
		}
		b.union = Expr(sql, args...)
	case *SelectDocBuilder:
		t.isParent = false
		sql, args, err := t.ToSQL()
		if err != nil {
			b.err = err
			return b
		}
		b.union = Expr(sql, args...)
	case Builder:
		sql, args, err := t.ToSQL()
		if err != nil {
			b.err = err
			return b
		}
		b.union = Expr(sql, args...)
	case string:
		b.union = Expr(t, a...)
	}
	return b
}
