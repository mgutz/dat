package dat

import "errors"

// DeleteBuilder contains the clauses for a DELETE statement
type DeleteBuilder struct {
	Execer

	table          string
	whereFragments []*whereFragment
	isInterpolated bool
	scope          Scope
	err            error
}

// NewDeleteBuilder creates a new DeleteBuilder for the given table.
func NewDeleteBuilder(table string) *DeleteBuilder {
	var err error
	if table == "" {
		err = errors.New("DeleteFrom requires a table name")
	}
	return &DeleteBuilder{table: table, isInterpolated: EnableInterpolation, err: err}
}

// ScopeMap uses a predefined scope in place of WHERE.
func (b *DeleteBuilder) ScopeMap(mapScope *MapScope, m M) *DeleteBuilder {
	if b.err != nil {
		return b
	}
	b.scope = mapScope.mergeClone(m)
	return b
}

// Scope uses a predefined scope in place of WHERE.
func (b *DeleteBuilder) Scope(sql string, args ...interface{}) *DeleteBuilder {
	if b.err != nil {
		return b
	}
	b.scope = ScopeFunc(func(table string) (string, []interface{}) {
		return escapeScopeTable(sql, table), args
	})
	return b
}

// Where appends a WHERE clause to the statement whereSQLOrMap can be a
// string or map. If it's a string, args wil replaces any places holders
func (b *DeleteBuilder) Where(whereSQLOrMap interface{}, args ...interface{}) *DeleteBuilder {
	if b.err != nil {
		return b
	}
	fragment, err := newWhereFragment(whereSQLOrMap, args)
	if err != nil {
		b.err = err
	} else {
		b.whereFragments = append(b.whereFragments, fragment)
	}
	return b
}

// ToSQL serialized the DeleteBuilder to a SQL string
// It returns the string with placeholders and a slice of query arguments
func (b *DeleteBuilder) ToSQL() (string, []interface{}, error) {
	if b.err != nil {
		return NewDatSQLErr(b.err)
	}

	if len(b.table) == 0 {
		return NewDatSQLError("no table specified")
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
			writeAndFragmentsToSQL(buf, b.whereFragments, &args, &placeholderStartPos)
		}
	} else {
		whereFragment, err := newWhereFragment(b.scope.ToSQL(b.table))
		if err != nil {
			return NewDatSQLErr(err)
		}
		writeScopeCondition(buf, whereFragment, &args, &placeholderStartPos)
	}

	return buf.String(), args, nil
}
