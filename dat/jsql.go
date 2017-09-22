package dat

import (
	"strings"

	"github.com/mgutz/str"
)

// JSQLBuilder builds SQL that returns a JSON row.
type JSQLBuilder struct {
	Execer

	isInterpolated   bool
	args             []interface{}
	query            string
	subQueriesMany   []*subInfo
	subQueriesOne    []*subInfo
	subQueriesVector []*subInfo
	subQueriesScalar []*subInfo
	union            *Expression
	isParent         bool
	err              error
}

// NewJSQLBuilder creates an instance of JSQLBuilder.
func NewJSQLBuilder(q string, args ...interface{}) *JSQLBuilder {
	// remove "select" from start of string
	index := str.IndexOf(strings.ToLower(q), "select", 0)
	if index < 0 {
		return &JSQLBuilder{err: logger.Error("Expected query to start with 'select'")}
	}

	// 7 == len("select") + 1
	return &JSQLBuilder{query: q[index+7:], args: args, isParent: true, isInterpolated: EnableInterpolation}
}

// Many loads a sub query resulting in an array of rows as an alias.
func (b *JSQLBuilder) Many(column string, sqlOrBuilder interface{}, a ...interface{}) *JSQLBuilder {
	b.err = storeExpr(&b.subQueriesMany, "JSQLBuilder.Many", column, sqlOrBuilder, a...)
	return b
}

// Vector loads a sub query resulting in an array of homogeneous scalars as an alias.
func (b *JSQLBuilder) Vector(column string, sqlOrBuilder interface{}, a ...interface{}) *JSQLBuilder {
	b.err = storeExpr(&b.subQueriesVector, "JSQLBuilder.Vector", column, sqlOrBuilder, a...)
	return b
}

// One loads a query resulting in a single row as an alias.
func (b *JSQLBuilder) One(column string, sqlOrBuilder interface{}, a ...interface{}) *JSQLBuilder {
	b.err = storeExpr(&b.subQueriesOne, "JSQLBuilder.One", column, sqlOrBuilder, a...)
	return b
}

// Scalar loads a query resulting in a single scalar as an alias and embeds the scalar in the parent object, rather than as a child object
func (b *JSQLBuilder) Scalar(column string, sqlOrBuilder interface{}, a ...interface{}) *JSQLBuilder {
	b.err = storeExpr(&b.subQueriesScalar, "JSQLBuilder.Scalar", column, sqlOrBuilder, a...)
	return b
}

func (b *JSQLBuilder) Union(sqlOrBuilder interface{}, a ...interface{}) *JSQLBuilder {
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

// ToSQL serialized the SelectBuilder to a SQL string
// It returns the string with placeholders and a slice of query arguments
func (b *JSQLBuilder) ToSQL() (string, []interface{}, error) {
	if b.err != nil {
		return NewDatSQLErr(b.err)
	}

	buf := bufPool.Get()
	defer bufPool.Put(buf)
	args := b.args
	placeholderStartPos := int64(len(args)) + 1

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
		buf.WriteString("SELECT row_to_json(dat__item.*) FROM ( SELECT ")
	} else {
		buf.WriteString("SELECT ")
	}

	for _, sub := range b.subQueriesMany {
		buf.WriteString("(SELECT array_agg(dat__")
		buf.WriteString(sub.alias)
		buf.WriteString(".*) FROM (")
		sub.WriteRelativeArgs(buf, &args, &placeholderStartPos)
		buf.WriteString(") AS dat__")
		buf.WriteString(sub.alias)
		buf.WriteString(") AS ")
		writeQuotedIdentifier(buf, sub.alias)
		buf.WriteString(", ")
	}

	for _, sub := range b.subQueriesVector {
		buf.WriteString("(SELECT array_agg(dat__")
		buf.WriteString(sub.alias)
		buf.WriteString(".dat__scalar) FROM (")
		sub.WriteRelativeArgs(buf, &args, &placeholderStartPos)
		buf.WriteString(") AS dat__")
		buf.WriteString(sub.alias)
		buf.WriteString("(dat__scalar)) AS ")
		writeQuotedIdentifier(buf, sub.alias)
		buf.WriteString(", ")
	}

	for _, sub := range b.subQueriesOne {
		buf.WriteString("(SELECT row_to_json(dat__")
		buf.WriteString(sub.alias)
		buf.WriteString(".*) FROM (")
		sub.WriteRelativeArgs(buf, &args, &placeholderStartPos)
		buf.WriteString(") AS dat__")
		buf.WriteString(sub.alias)
		buf.WriteString(") AS ")
		writeQuotedIdentifier(buf, sub.alias)
		buf.WriteString(", ")
	}

	for _, sub := range b.subQueriesScalar {
		buf.WriteString("(SELECT dat__")
		buf.WriteString(sub.alias)
		buf.WriteString(".dat__scalar FROM (")
		sub.WriteRelativeArgs(buf, &args, &placeholderStartPos)
		buf.WriteString(") AS dat__")
		buf.WriteString(sub.alias)
		buf.WriteString("(dat__scalar) limit 1) AS ")
		writeQuotedIdentifier(buf, sub.alias)
		buf.WriteString(", ")
	}

	buf.WriteString(b.query)

	if b.union != nil {
		buf.WriteString(" UNION ")
		b.union.WriteRelativeArgs(buf, &args, &placeholderStartPos)
	}

	if b.isParent {
		buf.WriteString(`) as dat__item`)
	}
	return buf.String(), args, nil
}
