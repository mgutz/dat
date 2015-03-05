package dat

import (
	"bytes"
	"regexp"
)

// M is a generic map from string to interface{}
type M map[string]interface{}

// Scope defines scope for query.
//
//
type Scope struct {
	SQL    string
	Fields M
}

var reField = regexp.MustCompile(`\B:[A-Za-z_]\w*`)

// NewScope creates a new scope.
func NewScope(sql string, fields M) *Scope {
	return &Scope{SQL: sql, Fields: fields}
}

// CloneMerge creates a clone of scope and merges fields.
func (scope *Scope) cloneMerge(fields M) *Scope {
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

	clone := &Scope{SQL: scope.SQL, Fields: newm}
	return clone
}

// ToSQL converts this scope's SQL to SQL and args.
func (scope *Scope) toSQL(table string) (string, []interface{}) {
	var buf bytes.Buffer
	var n = 1
	var args []interface{}
	sql := reField.ReplaceAllStringFunc(scope.SQL, func(found string) string {
		buf.Reset()
		if found == ":TABLE" {
			Dialect.WriteIdentifier(&buf, table)
			return buf.String()
		}
		if args == nil {
			args = []interface{}{}
		}
		field := found[1:]
		args = append(args, scope.Fields[field])
		writePlaceholder(&buf, n)
		n++
		return buf.String()
	})

	return sql, args
}
