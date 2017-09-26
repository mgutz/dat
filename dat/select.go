package dat

type joinExpression struct {
	subQuery string
	alias    string
	kind     int // JoinOn, LeftJoinOn, RightJoinOn, FullJoinOn
	expr     string
}

type subInfo struct {
	*Expression
	alias string
}

// SelectBuilder contains the clauses for a SELECT statement
type SelectBuilder struct {
	Execer

	columns         []string
	distinctColumns []string
	err             error
	fors            []string
	groupBys        []string
	havingFragments []*whereFragment
	isDistinct      bool
	isInterpolated  bool
	joins           []*joinExpression
	limitCount      uint64
	limitValid      bool
	offsetCount     uint64
	offsetValid     bool
	orderBys        []*whereFragment
	scope           Scope
	table           string
	union           *Expression
	subQueriesWith  []*subInfo
	whereFragments  []*whereFragment
}

// NewSelectBuilder creates a new SelectBuilder for the given columns
func NewSelectBuilder(columns ...string) *SelectBuilder {
	b := &SelectBuilder{isInterpolated: EnableInterpolation}
	b.Columns(columns...)
	return b
}

// ToSQL serialized the SelectBuilder to a SQL string
// It returns the string with placeholders and a slice of query arguments
func (b *SelectBuilder) ToSQL() (string, []interface{}, error) {
	if b.err != nil {
		return NewDatSQLErr(b.err)
	}

	if len(b.columns) == 0 {
		return NewDatSQLError("no columns specified")
	}
	if len(b.table) == 0 {
		return NewDatSQLError("no table specified")
	}

	buf := bufPool.Get()
	defer bufPool.Put(buf)
	var args []interface{}
	var placeholderStartPos int64 = 1

	writeWith(buf, b.subQueriesWith, &args, &placeholderStartPos)

	buf.WriteString("SELECT ")

	writeDistinct(buf, b.isDistinct, b.distinctColumns)
	writeColumns(buf, b.columns)

	writeFrom(buf, b.table)

	err := writeJoins(buf, b.joins)
	if err != nil {
		return NewDatSQLErr(err)
	}

	// scope becomes where clause
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
	writeUnion(buf, b.union, &args, &placeholderStartPos)

	return buf.String(), args, nil
}
