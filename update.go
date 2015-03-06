package dat

import (
	"bytes"
	"log"
	"reflect"
	"strconv"

	"github.com/mgutz/str"
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
func (b *UpdateBuilder) SetBlacklist(rec interface{}, columns ...string) *UpdateBuilder {
	val := reflect.Indirect(reflect.ValueOf(rec))
	vname := val.String()
	vtype := val.Type()

	if len(columns) == 0 {
		panic("SetBlacklist a list of columns names")
	}

	for i := 0; i < vtype.NumField(); i++ {
		f := vtype.Field(i)
		dbName := f.Tag.Get("db")
		if dbName == "" {
			log.Fatalf("%s must have db struct tags for all fields: `db:\"\"`", vname)
		}
		if !str.SliceContains(columns, dbName) {
			value := val.Field(i).Interface()
			b.Set(dbName, value)
		}
	}
	return b
}

// SetWhitelist creates SET clause(s) using a record and whitelist of columns.
// To specify all columns, use "*".
func (b *UpdateBuilder) SetWhitelist(rec interface{}, columns ...string) *UpdateBuilder {
	val := reflect.Indirect(reflect.ValueOf(rec))
	vname := val.String()
	vtype := val.Type()

	isWildcard := len(columns) == 0 || columns[0] == "*"

	for i := 0; i < vtype.NumField(); i++ {
		f := vtype.Field(i)
		dbName := f.Tag.Get("db")
		if dbName == "" {
			log.Fatalf("%s must have db struct tags for all fields: `db:\"\"`", vname)
		}
		value := val.Field(i).Interface()

		if isWildcard {
			b.Set(dbName, value)
		} else if str.SliceContains(columns, dbName) {
			b.Set(dbName, value)
		}

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
func (b *UpdateBuilder) Where(whereSqlOrMap interface{}, args ...interface{}) *UpdateBuilder {
	b.whereFragments = append(b.whereFragments, newWhereFragment(whereSqlOrMap, args))
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

	var sql bytes.Buffer
	var args []interface{}

	sql.WriteString("UPDATE ")
	sql.WriteString(b.table)
	sql.WriteString(" SET ")

	var placeholderStartPos int64 = 1

	// Build SET clause SQL with placeholders and add values to args
	for i, c := range b.setClauses {
		if i > 0 {
			sql.WriteString(", ")
		}
		Dialect.WriteIdentifier(&sql, c.column)
		if e, ok := c.value.(*Expression); ok {
			start := placeholderStartPos
			sql.WriteString(" = ")
			// map relative $1, $2 placeholders to absolute
			remapPlaceholders(&sql, e.Sql, start)
			args = append(args, e.Args...)
			placeholderStartPos += int64(len(e.Args))
		} else {
			if i < maxLookup {
				sql.WriteString(equalsPlaceholderTab[placeholderStartPos])
			} else {
				if placeholderStartPos < maxLookup {
					sql.WriteString(equalsPlaceholderTab[placeholderStartPos])
				} else {
					sql.WriteString(" = $")
					sql.WriteString(strconv.FormatInt(placeholderStartPos, 10))
				}
			}
			placeholderStartPos++
			args = append(args, c.value)
		}
	}

	if b.scope == nil {
		if len(b.whereFragments) > 0 {
			sql.WriteString(" WHERE ")
			writeWhereFragmentsToSql(b.whereFragments, &sql, &args, &placeholderStartPos)
		}
	} else {
		whereFragment := newWhereFragment(b.scope.ToSQL(b.table))
		writeScopeCondition(whereFragment, &sql, &args, &placeholderStartPos)
	}

	// Ordering and limiting
	if len(b.orderBys) > 0 {
		sql.WriteString(" ORDER BY ")
		for i, s := range b.orderBys {
			if i > 0 {
				sql.WriteString(", ")
			}
			sql.WriteString(s)
		}
	}

	if b.limitValid {
		sql.WriteString(" LIMIT ")
		writeUint64(&sql, b.limitCount)
	}

	if b.offsetValid {
		sql.WriteString(" OFFSET ")
		writeUint64(&sql, b.offsetCount)
	}

	// Go thru the returning clauses
	for i, c := range b.returnings {
		if i == 0 {
			sql.WriteString(" RETURNING ")
		} else {
			sql.WriteRune(',')
		}
		Dialect.WriteIdentifier(&sql, c)
	}

	return sql.String(), args
}

// Interpolate interpolates this builders sql.
func (b *UpdateBuilder) Interpolate() (string, []interface{}, error) {
	return interpolate(b)
}

// IsInterpolated determines if this builder will interpolate when
// Interpolate() is called.
func (b *UpdateBuilder) IsInterpolated() bool {
	return b.isInterpolated
}

// SetIsInterpolated sets whether this builder should interpolate.
func (b *UpdateBuilder) SetIsInterpolated(enable bool) *UpdateBuilder {
	b.isInterpolated = enable
	return b
}
