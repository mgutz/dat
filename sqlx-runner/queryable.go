package runner

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/mgutz/dat"
)

// Queryable is an object that can be queried.
type Queryable struct {
	runner runner
}

// WrapSqlxExt converts a sqlx.Ext to a *Queryable
func WrapSqlxExt(e sqlx.Ext) *Queryable {
	switch e := e.(type) {
	default:
		panic(fmt.Sprintf("unexpected type %T", e))
	case runner:
		return &Queryable{e}
	}
}

// DeleteFrom creates a new DeleteBuilder for the given table.
func (q *Queryable) DeleteFrom(table string) *dat.DeleteBuilder {
	b := dat.NewDeleteBuilder(table)
	b.Execer = NewExecer(q.runner, b)
	return b
}

// Exec executes a SQL query with optional arguments.
func (q *Queryable) Exec(cmd string, args ...interface{}) (*dat.Result, error) {
	var result sql.Result
	var err error

	if len(args) == 0 {
		result, err = q.runner.Exec(cmd)
	} else {
		result, err = q.runner.Exec(cmd, args...)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}
	return &dat.Result{RowsAffected: rowsAffected}, nil
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
