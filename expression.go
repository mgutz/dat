package dat

// Expression is used by query builder to hold a sub expression.
type Expression struct {
	Sql    string
	Values []interface{}
}

// Expr is a SQL expression with placeholders, and a slice of args to replace them with
func Expr(sql string, values ...interface{}) *Expression {
	return &Expression{Sql: sql, Values: values}
}
