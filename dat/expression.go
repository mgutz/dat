package dat

import "gopkg.in/mgutz/dat.v2/common"

// Expression holds a sub expression.
type Expression struct {
	SQL  string
	Args []interface{}
}

// Expr is a SQL expression with placeholders, and a slice of args to replace them with
func Expr(sql string, values ...interface{}) *Expression {
	return &Expression{SQL: sql, Args: values}
}

// WriteRelativeArgs writes the args to buf adjusting the placeholder to start at pos.
func (exp *Expression) WriteRelativeArgs(buf common.BufferWriter, args *[]interface{}, pos *int64) {
	remapPlaceholders(buf, exp.SQL, *pos)
	*args = append(*args, exp.Args...)
	*pos += int64(len(exp.Args))
}
