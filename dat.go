package dat

// Builder interface is used to tie SQL generators to executors.
type Builder interface {
	ToSQL() (string, []interface{})
	Interpolate() (string, error)
	MustInterpolate() string
}

// DeleteFrom creates a new DeleteBuilder for the given table.
func DeleteFrom(table string) *DeleteBuilder {
	return &DeleteBuilder{Table: table}
}

// InsertInto creates a new InsertBuilder for the given table.
func InsertInto(table string) *InsertBuilder {
	return &InsertBuilder{Table: table}
}

// Select creates a new SelectBuilder for the given columns
func Select(columns ...string) *SelectBuilder {
	return &SelectBuilder{Columns: columns}
}

// SQL creates a new RawBuilder for the given SQL string and arguments
func SQL(sql string, args ...interface{}) *RawBuilder {
	return &RawBuilder{sql: sql, args: args}
}

// Update creates a new UpdateBuilder for the given table
func Update(table string) *UpdateBuilder {
	return &UpdateBuilder{Table: table}
}
