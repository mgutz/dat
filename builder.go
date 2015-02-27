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
	b := NewDeleteBuilder(table)
	b.Execer = &panicExecer{}
	return b
}

// InsertInto creates a new InsertBuilder for the given table.
func InsertInto(table string) *InsertBuilder {
	b := NewInsertBuilder(table)
	b.Execer = &panicExecer{}
	return b
}

// Select creates a new SelectBuilder for the given columns.
func Select(columns ...string) *SelectBuilder {
	b := NewSelectBuilder(columns...)
	b.Execer = &panicExecer{}
	return b
}

// SQL creates a new raw SQL builder.
func SQL(sql string, args ...interface{}) *RawBuilder {
	b := NewRawBuilder(sql, args...)
	b.Execer = &panicExecer{}
	return b
}

// Update creates a new UpdateBuilder for the given table.
func Update(table string) *UpdateBuilder {
	b := NewUpdateBuilder(table)
	b.Execer = &panicExecer{}
	return b
}
