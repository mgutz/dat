package dat

// Builder interface is used to tie SQL generators to executors.
type Builder interface {
	// ToSQL builds the SQL and arguments from builder.
	ToSQL() (string, []interface{})

	// Interpolate builds the interpolation SQL and arguments from builder.
	// If interpolation flag is disabled then this is just a passthrough to ToSQL.
	Interpolate() (string, []interface{}, error)

	// IsInterpolated determines if this builder will interpolate when
	// Interpolate() is called.
	IsInterpolated() bool
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
