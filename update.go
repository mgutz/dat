package dat

import (
	"reflect"
	"strconv"
)

// UpdateBuilder contains the clauses for an UPDATE statement
type UpdateBuilder struct {
	Execer

	isInterpolated bool
	table          string
	setClauses     []*setClause
	whereFragments []*whereFragment
	orderBys       []string
	limitCount     uint64
	limitValid     bool
	offsetCount    uint64
	offsetValid    bool
	returnings     []string
	scope          Scope
}

type setClause struct {
	column string
	value  interface{}
}

// NewUpdateBuilder creates a new UpdateBuilder for the given table
func NewUpdateBuilder(table string) *UpdateBuilder {
	if table == "" {
		logger.Error("Update requires a table name")
		return nil
	}
	return &UpdateBuilder{table: table, isInterpolated: EnableInterpolation}
}

// Set appends a column/value pair for the statement
func (b *UpdateBuilder) Set(column string, value interface{}) *UpdateBuilder {
	b.setClauses = append(b.setClauses, &setClause{column: column, value: value})
	return b
}

// SetMap appends the elements of the map as column/value pairs for the statement
func (b *UpdateBuilder) SetMap(clauses map[string]interface{}) *UpdateBuilder {
	for col, val := range clauses {
		b = b.Set(col, val)
	}
	return b
}

// SetBlacklist creates SET clause(s) using a record and blacklist of columns
func (b *UpdateBuilder) SetBlacklist(rec interface{}, blacklist ...string) *UpdateBuilder {
	if len(blacklist) == 0 {
		panic("SetBlacklist requires a list of columns names")
	}

	columns := reflectExcludeColumns(rec, blacklist)
	ind := reflect.Indirect(reflect.ValueOf(rec))
	vals, err := valuesFor(ind.Type(), ind, columns)
	if err != nil {
		panic(err)
	}

	for i, val := range vals {
		b.Set(columns[i], val)
	}

	return b
}

// SetWhitelist creates SET clause(s) using a record and whitelist of columns.
// To specify all columns, use "*".
func (b *UpdateBuilder) SetWhitelist(rec interface{}, whitelist ...string) *UpdateBuilder {
	var columns []string
	if len(whitelist) == 0 || whitelist[0] == "*" {
		columns = reflectColumns(rec)
	} else {
		columns = whitelist
	}

	ind := reflect.Indirect(reflect.ValueOf(rec))
	vals, err := valuesFor(ind.Type(), ind, columns)
	if err != nil {
		panic(err)
	}

	for i, val := range vals {
		b.Set(columns[i], val)
	}

	return b
}

// ScopeMap uses a predefined scope in place of WHERE.
func (b *UpdateBuilder) ScopeMap(mapScope *MapScope, m M) *UpdateBuilder {
	b.scope = mapScope.mergeClone(m)
	return b
}

// Scope uses a predefined scope in place of WHERE.
func (b *UpdateBuilder) Scope(sql string, args ...interface{}) *UpdateBuilder {
	b.scope = ScopeFunc(func(table string) (string, []interface{}) {
		return escapeScopeTable(sql, table), args
	})
	return b
}

// Where appends a WHERE clause to the statement
func (b *UpdateBuilder) Where(whereSQLOrMap interface{}, args ...interface{}) *UpdateBuilder {
	b.whereFragments = append(b.whereFragments, newWhereFragment(whereSQLOrMap, args))
	return b
}

// OrderBy appends a column to ORDER the statement by
func (b *UpdateBuilder) OrderBy(ord string) *UpdateBuilder {
	b.orderBys = append(b.orderBys, ord)
	return b
}

// Limit sets a limit for the statement; overrides any existing LIMIT
func (b *UpdateBuilder) Limit(limit uint64) *UpdateBuilder {
	b.limitCount = limit
	b.limitValid = true
	return b
}

// Offset sets an offset for the statement; overrides any existing OFFSET
func (b *UpdateBuilder) Offset(offset uint64) *UpdateBuilder {
	b.offsetCount = offset
	b.offsetValid = true
	return b
}

// Returning sets the columns for the RETURNING clause
func (b *UpdateBuilder) Returning(columns ...string) *UpdateBuilder {
	b.returnings = columns
	return b
}

// ToSQL serialized the UpdateBuilder to a SQL string
// It returns the string with placeholders and a slice of query arguments
func (b *UpdateBuilder) ToSQL() (string, []interface{}) {
	if len(b.table) == 0 {
		panic("no table specified")
	}
	if len(b.setClauses) == 0 {
		panic("no set clauses specified")
	}

	buf := bufPool.Get()
	defer bufPool.Put(buf)
	var args []interface{}

	buf.WriteString("UPDATE ")
	writeIdentifier(buf, b.table)
	buf.WriteString(" SET ")

	var placeholderStartPos int64 = 1

	// Build SET clause SQL with placeholders and add values to args
	for i, c := range b.setClauses {
		if i > 0 {
			buf.WriteString(", ")
		}
		Dialect.WriteIdentifier(buf, c.column)
		if e, ok := c.value.(*Expression); ok {
			start := placeholderStartPos
			buf.WriteString(" = ")
			// map relative $1, $2 placeholders to absolute
			remapPlaceholders(buf, e.Sql, start)
			args = append(args, e.Args...)
			placeholderStartPos += int64(len(e.Args))
		} else {
			// TOOD
			if placeholderStartPos < maxLookup {
				buf.WriteString(equalsPlaceholderTab[placeholderStartPos])
			} else {
				buf.WriteString(" = $")
				buf.WriteString(strconv.FormatInt(placeholderStartPos, 10))
			}
			placeholderStartPos++
			args = append(args, c.value)
		}
	}

	if b.scope == nil {
		if len(b.whereFragments) > 0 {
			buf.WriteString(" WHERE ")
			writeAndFragmentsToSQL(buf, b.whereFragments, &args, &placeholderStartPos)
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

	// Go thru the returning clauses
	for i, c := range b.returnings {
		if i == 0 {
			buf.WriteString(" RETURNING ")
		} else {
			buf.WriteRune(',')
		}
		Dialect.WriteIdentifier(buf, c)
	}

	return buf.String(), args
}
