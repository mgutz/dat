package dat

// Builder interface is used to tie SQL generators to executors.
type Builder interface {
	ToSQL() (string, []interface{})
	Interpolate() (string, []interface{}, error)
	MustInterpolate() (string, []interface{})
}

// DeleteFrom creates a new DeleteBuilder for the given table.
func DeleteFrom(table string) *DeleteBuilder {
	return NewDeleteBuilder(table)
}

// InsertInto creates a new InsertBuilder for the given table.
func InsertInto(table string) *InsertBuilder {
	return NewInsertBuilder(table)
}

// Select creates a new SelectBuilder for the given columns.
func Select(columns ...string) *SelectBuilder {
	return NewSelectBuilder(columns...)
}

// SQL creates a new raw SQL builder.
func SQL(sql string, args ...interface{}) *RawBuilder {
	return NewRawBuilder(sql, args...)
}

// Update creates a new UpdateBuilder for the given table.
func Update(table string) *UpdateBuilder {
	return NewUpdateBuilder(table)
}
