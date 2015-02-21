package runner

import "github.com/mgutz/dat"

// Queryable is any object that can be queried.
type Queryable struct {
	runner runner
}

// DeleteFrom creates a new DeleteBuilder for the given table.
func (q *Queryable) DeleteFrom(table string) *dat.DeleteBuilder {
	b := dat.NewDeleteBuilder(table)
	b.Executable = NewExecer(q.runner, b)
	return b
}

// InsertInto creates a new InsertBuilder for the given table.
func (q *Queryable) InsertInto(table string) *dat.InsertBuilder {
	b := dat.NewInsertBuilder(table)
	b.Executable = NewExecer(q.runner, b)
	return b
}

// Select creates a new SelectBuilder for the given columns.
func (q *Queryable) Select(columns ...string) *dat.SelectBuilder {
	b := dat.NewSelectBuilder(columns...)
	b.Executable = NewExecer(q.runner, b)
	return b
}

// SQL creates a new raw SQL builder.
func (q *Queryable) SQL(sql string, args ...interface{}) *dat.RawBuilder {
	b := dat.NewRawBuilder(sql, args...)
	b.Executable = NewExecer(q.runner, b)
	return b
}

// Update creates a new UpdateBuilder for the given table.
func (q *Queryable) Update(table string) *dat.UpdateBuilder {
	b := dat.NewUpdateBuilder(table)
	b.Executable = NewExecer(q.runner, b)
	return b
}
