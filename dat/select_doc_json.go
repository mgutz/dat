/*
DO NOT EDIT

Copied from jsql_json.go, then s/SelectDocBuilder/SelectDocBuilder/g
*/

package dat

// Many loads a sub query resulting in an array of rows as an alias.
func (b *SelectDocBuilder) Many(column string, sqlOrBuilder interface{}, a ...interface{}) *SelectDocBuilder {
	b.err = storeExpr(&b.subQueriesMany, "SelectDocBuilder.Many", column, sqlOrBuilder, a...)
	return b
}

// Vector loads a sub query resulting in an array of homogeneous scalars as an alias.
func (b *SelectDocBuilder) Vector(column string, sqlOrBuilder interface{}, a ...interface{}) *SelectDocBuilder {
	b.err = storeExpr(&b.subQueriesVector, "SelectDocBuilder.Vector", column, sqlOrBuilder, a...)
	return b
}

// One loads a query resulting in a single row as an alias.
func (b *SelectDocBuilder) One(column string, sqlOrBuilder interface{}, a ...interface{}) *SelectDocBuilder {
	b.err = storeExpr(&b.subQueriesOne, "SelectDocBuilder.One", column, sqlOrBuilder, a...)
	return b
}

// Scalar loads a query resulting in a single scalar as an alias and embeds the scalar in the parent object, rather than as a child object
func (b *SelectDocBuilder) Scalar(column string, sqlOrBuilder interface{}, a ...interface{}) *SelectDocBuilder {
	b.err = storeExpr(&b.subQueriesScalar, "SelectDocBuilder.Scalar", column, sqlOrBuilder, a...)
	return b
}
