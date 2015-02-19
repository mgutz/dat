package dat

type expression struct {
	Sql    string
	Values []interface{}
}

// Expr is a SQL sqlExpressoin with placeholders, and a slice of args to replace them with
func Expr(sql string, values ...interface{}) *expression {
	return &expression{Sql: sql, Values: values}
}
