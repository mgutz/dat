package dat

// Union ...
func (b *JSQLBuilder) Union(sqlOrBuilder interface{}, a ...interface{}) *JSQLBuilder {
	switch t := sqlOrBuilder.(type) {
	default:
		b.err = NewError("SelectDocBuilder.Union: sqlOrbuilder accepts only {string, Builder, *SelectDocBuilder} type")
	case JSONBuilder:
		t.setIsParent(false)
		sql, args, err := t.ToSQL()
		if err != nil {
			b.err = err
			return b
		}
		b.union = Expr(sql, args...)
	case Builder:
		sql, args, err := t.ToSQL()
		if err != nil {
			b.err = err
			return b
		}
		b.union = Expr(sql, args...)
	case string:
		b.union = Expr(t, a...)
	}
	return b
}
