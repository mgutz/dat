package dat

import (
	"bytes"
	"regexp"
	"strings"
)

// M is a generic map from string to interface{}
type M map[string]interface{}

var reField = regexp.MustCompile(`\B:[A-Za-z_]\w*`)

// Scope predefines parameterized JOIN and WHERE conditions.
type Scope interface {
	ToSQL(table string) (string, []interface{})
}

// ScopeFunc is an ad-hoc scope function.
type ScopeFunc func(table string) (string, []interface{})

// ToSQL converts scoped func to sql.
func (sf ScopeFunc) ToSQL(table string) (string, []interface{}) {
	return sf(table)
}

// MapScope defines scope for using a fields map.
type MapScope struct {
	SQL    string
	Fields M
}

// NewScope creates a new scope.
func NewScope(sql string, fields M) *MapScope {
	return &MapScope{SQL: sql, Fields: fields}
}

// Clone creates a clone of scope and merges fields.
func (scope *MapScope) mergeClone(fields M) *MapScope {
	newm := M{}
	for key, val := range scope.Fields {
		if fields != nil {
			if val, ok := fields[key]; ok {
				newm[key] = val
				continue
			}
		}
		newm[key] = val
	}

	clone := &MapScope{SQL: scope.SQL, Fields: newm}
	return clone
}

// ToSQL converts this scope's SQL to SQL and args.
func (scope *MapScope) ToSQL(table string) (string, []interface{}) {
	buf := bufPool.Get()
	defer bufPool.Put(buf)

	var n = 1
	var args []interface{}
	sql := reField.ReplaceAllStringFunc(scope.SQL, func(found string) string {
		buf.Reset()
		if found == ":TABLE" {
			Dialect.WriteIdentifier(buf, table)
			return buf.String()
		}
		if args == nil {
			args = []interface{}{}
		}
		field := found[1:]
		args = append(args, scope.Fields[field])
		writePlaceholder(buf, n)
		n++
		return buf.String()
	})

	return sql, args
}

// escapeScopeTable escapes :TABLE in sql using Dialect.WriteIdentifer.
func escapeScopeTable(sql string, table string) string {
	if !strings.Contains(sql, ":TABLE") {
		return sql
	}

	var buf bytes.Buffer
	Dialect.WriteIdentifier(&buf, table)
	quoted := buf.String()
	return strings.Replace(sql, ":TABLE", quoted, -1)
}

var reWhereClause = regexp.MustCompile(`\s*(WHERE|where)\b`)

// splitWhere splits a query on the word WHERE
func splitWhere(query string) (sql string, where string) {
	indices := reWhereClause.FindStringIndex(query)
	// grab only the first location
	if len(indices) == 0 {
		return query, ""
	}

	// may have leading spaces
	where = query[indices[0]:]
	idx := strings.Index(where, "WHERE")
	if idx == -1 {
		idx = strings.Index(where, "where")
	}
	// 5 == len("WHERE")
	where = where[idx+5:]

	sql = query[0:indices[0]]
	return sql, where
}
