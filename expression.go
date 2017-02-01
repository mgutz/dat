package dat

import "gopkg.in/mgutz/dat.v1/common"

// Expression holds a sub expression.
type Expression struct {
	Sql  string
	Args []interface{}
}

// Expr is a SQL expression with placeholders, and a slice of args to replace them with
func Expr(sql string, values ...interface{}) *Expression {
	return &Expression{Sql: sql, Args: values}
}

// WriteRelativeArgs writes the args to buf adjusting the placeholder to start at pos.
func (exp *Expression) WriteRelativeArgs(buf common.BufferWriter, args *[]interface{}, pos *int64) {
	remapPlaceholders(buf, exp.Sql, *pos)
	*args = append(*args, exp.Args...)
	*pos += int64(len(exp.Args))
}

// Expression implements Expressioner interface (used in Interpolate).
func (exp *Expression) Expression() (string, []interface{}, error) {
	return Interpolate(exp.Sql, exp.Args)
}
