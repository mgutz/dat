package dat

// SelectDocBuilder builds SQL that returns a JSON row.
type SelectDocBuilder struct {
	*SelectBuilder
	jsonBuilder
	innerSQL *Expression
	err      error
}

// NewSelectDocBuilder creates an instance of SelectDocBuilder.
func NewSelectDocBuilder(columns ...string) *SelectDocBuilder {
	sb := NewSelectBuilder(columns...)
	b := &SelectDocBuilder{SelectBuilder: sb}
	b.isParent = true
	return b
}

// InnerSQL sets the SQL after the SELECT (columns...) statement
//
// DEPRECATE this
func (b *SelectDocBuilder) InnerSQL(sql string, a ...interface{}) *SelectDocBuilder {
	b.innerSQL = Expr(sql, a...)
	return b
}

// ToSQL serialized the SelectBuilder to a SQL string
// It returns the string with placeholders and a slice of query arguments
func (b *SelectDocBuilder) ToSQL() (string, []interface{}, error) {
	if b.err != nil {
		return NewDatSQLErr(b.err)
	}

	if len(b.columns) == 0 {
		return NewDatSQLError("no columns specified")
	}
	if len(b.table) == 0 && b.innerSQL == nil {
		return NewDatSQLError("no table specified")
	}

	buf := bufPool.Get()
	defer bufPool.Put(buf)
	var args []interface{}
	var placeholderStartPos int64 = 1

	writeWith(buf, b.subQueriesWith, &args, &placeholderStartPos)

	if b.isParent {
		buf.WriteString("SELECT row_to_json(dat__item.*) FROM ( ")
	}
	buf.WriteString("SELECT ")

	writeDistinct(buf, b.isDistinct, b.distinctColumns)
	writeColumns(buf, b.columns)

	writeMany(buf, b.subQueriesMany, true, &args, &placeholderStartPos)
	writeVector(buf, b.subQueriesVector, true, &args, &placeholderStartPos)
	writeOne(buf, b.subQueriesOne, true, &args, &placeholderStartPos)
	writeScalar(buf, b.subQueriesScalar, true, &args, &placeholderStartPos)

	if b.innerSQL != nil {
		b.innerSQL.WriteRelativeArgs(buf, &args, &placeholderStartPos)
	} else {

		writeFrom(buf, b.table)

		whereFragments := b.whereFragments
		moreWheres, err := writeScope(buf, b.scope, b.table)
		if err != nil {
			return NewDatSQLErr(err)
		}
		whereFragments = append(whereFragments, moreWheres...)
		writeWhere(buf, whereFragments, &args, &placeholderStartPos)

		writeGroupBy(buf, b.groupBys)
		writeHaving(buf, b.havingFragments, &args, &placeholderStartPos)
		writeOrderBy(buf, b.orderBys, &args, &placeholderStartPos)
		writeLimit(buf, b.limitValid, b.limitCount)
		writeOffset(buf, b.offsetValid, b.offsetCount)
		writeFor(buf, b.fors)
	}

	writeUnion(buf, b.union, &args, &placeholderStartPos)

	if b.isParent {
		buf.WriteString(`) as dat__item`)
	}
	return buf.String(), args, nil
}
