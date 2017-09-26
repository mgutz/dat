package dat

import (
	"strings"

	"github.com/mgutz/str"
)

type jsonBuilder struct {
	subQueriesMany   []*subInfo
	subQueriesOne    []*subInfo
	subQueriesVector []*subInfo
	subQueriesScalar []*subInfo
	isParent         bool
}

// JSQLBuilder builds SQL that returns a JSON row. JSQLBuilder accepts
// a raw SQL Query unlike SelectDoc.
type JSQLBuilder struct {
	Execer
	jsonBuilder

	args           []interface{}
	isInterpolated bool
	query          string
	subQueriesWith []*subInfo
	union          *Expression
	err            error
}

// NewJSQLBuilder creates an instance of JSQLBuilder.
func NewJSQLBuilder(q string, args ...interface{}) *JSQLBuilder {
	// remove "select" from start of string
	index := str.IndexOf(strings.ToLower(q), "select", 0)
	if index < 0 {
		return &JSQLBuilder{err: logger.Error("Expected query to start with 'select'")}
	}

	// 7 == len("select ")
	b := &JSQLBuilder{query: q[index+7:], args: args, isInterpolated: EnableInterpolation}
	b.isParent = true
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

	writeWith(buf, b.subQueriesWith, &args, &placeholderStartPos)

	if b.isParent {
		buf.WriteString("SELECT row_to_json(dat__item.*) FROM ( ")
	}
	buf.WriteString("SELECT ")

	// track if any new columns were inserted to add comma below
	writeMany(buf, b.subQueriesMany, false, &args, &placeholderStartPos)
	writeVector(buf, b.subQueriesVector, false, &args, &placeholderStartPos)
	writeOne(buf, b.subQueriesOne, false, &args, &placeholderStartPos)
	writeScalar(buf, b.subQueriesScalar, false, &args, &placeholderStartPos)

	buf.WriteString(b.query)

	writeUnion(buf, b.union, &args, &placeholderStartPos)

	if b.isParent {
		buf.WriteString(`) as dat__item`)
	}
	return buf.String(), args, nil
}
