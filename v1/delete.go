package dat

// DeleteBuilder contains the clauses for a DELETE statement
type DeleteBuilder struct {
	Execer

	table          string
	whereFragments []*whereFragment
	orderBys       []string
	limitCount     uint64
	limitValid     bool
	offsetCount    uint64
	offsetValid    bool
	id             int
	isInterpolated bool
	scope          Scope
}

// NewDeleteBuilder creates a new DeleteBuilder for the given table.
func NewDeleteBuilder(table string) *DeleteBuilder {
	if table == "" {
		logger.Error("DeleteFrom requires a table name.")
		return nil
	}
	return &DeleteBuilder{table: table, isInterpolated: EnableInterpolation}
}

// ScopeMap uses a predefined scope in place of WHERE.
func (b *DeleteBuilder) ScopeMap(mapScope *MapScope, m M) *DeleteBuilder {
	b.scope = mapScope.mergeClone(m)
	return b
}

// Scope uses a predefined scope in place of WHERE.
func (b *DeleteBuilder) Scope(sql string, args ...interface{}) *DeleteBuilder {
	b.scope = ScopeFunc(func(table string) (string, []interface{}) {
		return escapeScopeTable(sql, table), args
	})
	return b
}

// Where appends a WHERE clause to the statement whereSqlOrMap can be a
// string or map. If it's a string, args wil replaces any places holders
func (b *DeleteBuilder) Where(whereSqlOrMap interface{}, args ...interface{}) *DeleteBuilder {
	b.whereFragments = append(b.whereFragments, newWhereFragment(whereSqlOrMap, args))
	return b
}

// OrderBy appends an ORDER BY clause to the statement
func (b *DeleteBuilder) OrderBy(ord string) *DeleteBuilder {
	b.orderBys = append(b.orderBys, ord)
	return b
}

// Limit sets a LIMIT clause for the statement; overrides any existing LIMIT
func (b *DeleteBuilder) Limit(limit uint64) *DeleteBuilder {
	b.limitCount = limit
	b.limitValid = true
	return b
}

// Offset sets an OFFSET clause for the statement; overrides any existing OFFSET
func (b *DeleteBuilder) Offset(offset uint64) *DeleteBuilder {
	b.offsetCount = offset
	b.offsetValid = true
	return b
}

// ToSQL serialized the DeleteBuilder to a SQL string
// It returns the string with placeholders and a slice of query arguments
func (b *DeleteBuilder) ToSQL() (string, []interface{}) {
	if len(b.table) == 0 {
		panic("no table specified")
	}

	buf := bufPool.Get()
	defer bufPool.Put(buf)

	var args []interface{}

	buf.WriteString("DELETE FROM ")
	buf.WriteString(b.table)

	var placeholderStartPos int64 = 1

	// Write WHERE clause if we have any fragments
	if b.scope == nil {
		if len(b.whereFragments) > 0 {
			buf.WriteString(" WHERE ")
			writeWhereFragmentsToSql(buf, b.whereFragments, &args, &placeholderStartPos)
		}
	} else {
		whereFragment := newWhereFragment(b.scope.ToSQL(b.table))
		writeScopeCondition(buf, whereFragment, &args, &placeholderStartPos)
	}

	// Ordering and limiting
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

	return buf.String(), args
}
