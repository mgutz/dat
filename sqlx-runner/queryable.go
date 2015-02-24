package runner

import "github.com/mgutz/dat"

// Queryable is an object that can be queried.
type Queryable struct {
	runner runner
}

// DeleteFrom creates a new DeleteBuilder for the given table.
func (q *Queryable) DeleteFrom(table string) *dat.DeleteBuilder {
	b := dat.NewDeleteBuilder(table)
	b.Execer = NewExecer(q.runner, b)
	return b
}

// ExecBuilder executes the SQL in builder.
func (q *Queryable) ExecBuilder(b dat.Builder) error {
	sql, args, err := b.Interpolate()
	if err != nil {
		return err
	}

	if len(args) == 0 {
		_, err = q.runner.Exec(sql)
	} else {
		_, err = q.runner.Exec(sql, args...)
	}
	return err
}

// InsertInto creates a new InsertBuilder for the given table.
func (q *Queryable) InsertInto(table string) *dat.InsertBuilder {
	b := dat.NewInsertBuilder(table)
	b.Execer = NewExecer(q.runner, b)
	return b
}

// Select creates a new SelectBuilder for the given columns.
func (q *Queryable) Select(columns ...string) *dat.SelectBuilder {
	b := dat.NewSelectBuilder(columns...)
	b.Execer = NewExecer(q.runner, b)
	return b
}

// SQL creates a new raw SQL builder.
func (q *Queryable) SQL(sql string, args ...interface{}) *dat.RawBuilder {
	b := dat.NewRawBuilder(sql, args...)
	b.Execer = NewExecer(q.runner, b)
	return b
}

// Update creates a new UpdateBuilder for the given table.
func (q *Queryable) Update(table string) *dat.UpdateBuilder {
	b := dat.NewUpdateBuilder(table)
	b.Execer = NewExecer(q.runner, b)
	return b
}
