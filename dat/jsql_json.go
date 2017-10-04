package dat

//// JSON related methods

// setIsParent sets whether this builder is used a parent or sub-query builder.
func (b *JSQLBuilder) setIsParent(value bool) {
	b.isParent = value
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
