package dat

import (
	"bytes"

	"github.com/mgutz/dat/common"
	"github.com/mgutz/str"
)

// Expression holds a sub expression.
type Expression struct {
	SQL         string
	Args        []interface{}
	Interpolate bool
}

// Expr is a SQL expression with placeholders, and a slice of args to replace them with
func Expr(sql string, values ...interface{}) *Expression {
	logger.Warn("DEPRECATION Expr will be deprecated. Use dat.Exp or dat.Expi")
	return &Expression{SQL: sql, Args: values, Interpolate: true}
}

// WriteRelativeArgs writes the args to buf adjusting the placeholder to start at pos.
func (exp *Expression) WriteRelativeArgs(buf common.BufferWriter, args *[]interface{}, pos *int64) {
	remapPlaceholders(buf, exp.SQL, *pos)
	*args = append(*args, exp.Args...)
	*pos += int64(len(exp.Args))
}

// Expression implements Expressioner interface (used in Interpolate).
func (exp *Expression) Expression() (string, []interface{}, error) {
	if exp.Interpolate {
		return Interpolate(exp.SQL, exp.Args)
	}
	return exp.SQL, exp.Args, nil
}

// Prep builds a new non-interpolated expression
func Prep(sql string, values ...interface{}) *Expression {
	return &Expression{SQL: sql, Args: values, Interpolate: false}
}

// Interp is an interpolated expression
func Interp(sql string, values ...interface{}) *Expression {
	return &Expression{SQL: sql, Args: values, Interpolate: true}
}

// AppliedExpression is an expression that is applied to a table of data.
type AppliedExpression struct {
	cmd     string
	arrArgs [][]interface{}
}

// AppliedExpr builds a new AppliedExpression.
func AppliedExpr(cmd string, arrArgs [][]interface{}) *AppliedExpression {
	return &AppliedExpression{cmd, arrArgs}
}

// Expression implements expressioner interface
func (ae *AppliedExpression) Expression() (string, interface{}, error) {
	var buf bytes.Buffer

	for _, args := range ae.arrArgs {
		s, a, err := Interpolate(ae.cmd, args)
		if err != nil {
			return "", nil, err
		}

		if len(a) > 0 {
			return "", nil, logger.Error("Did not expect AppliedExpression to return passthrough arguments")
		}

		buf.WriteString(str.EnsureSuffix(s, ";"))
	}
	return buf.String(), nil, nil
}
