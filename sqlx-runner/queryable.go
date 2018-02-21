package runner

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/mgutz/dat/dat"
)

// Queryable is an object that can be queried.
type Queryable struct {
	runner database
}

// WrapSqlxExt converts a sqlx.Ext to a *Queryable
func WrapSqlxExt(e sqlx.Ext) (*Queryable, error) {
	switch e := e.(type) {
	default:
		return nil, dat.NewError(fmt.Sprintf("unexpected type %T", e))
	case database:
		return &Queryable{e}, nil
	}
}

// Call creates a new CallBuilder for the given sproc and args.
func (q *Queryable) Call(sproc string, args ...interface{}) *dat.CallBuilder {
	b := dat.NewCallBuilder(sproc, args...)
	b.Execer = NewExecer(q.runner, b)
	return b
}

// DeleteFrom creates a new DeleteBuilder for the given table.
func (q *Queryable) DeleteFrom(table string) *dat.DeleteBuilder {
	b := dat.NewDeleteBuilder(table)
	b.Execer = NewExecer(q.runner, b)
	return b
}

// Exec executes a SQL query with optional arguments.
func (q *Queryable) Exec(cmd string, args ...interface{}) (*dat.Result, error) {
	return q.SQL(cmd, args...).Exec()
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
	if err != nil {
		return logSQLError(err, "ExecBuilder", sql, args)
	}
	return nil
}

// ExecMulti executes multiple SQL statements returning the number of
// statements executed, or the index at which an error occurred.
func (q *Queryable) ExecMulti(commands ...*dat.Expression) (int, error) {
	for i, cmd := range commands {
		_, err := q.SQL(cmd.SQL, cmd.Args...).Exec()
		if err != nil {
			return i, err
		}
	}
	return len(commands), nil
}

// InsertInto creates a new InsertBuilder for the given table.
func (q *Queryable) InsertInto(table string) *dat.InsertBuilder {
	b := dat.NewInsertBuilder(table)
	b.Execer = NewExecer(q.runner, b)
	return b
}

// Insect inserts or selects.
func (q *Queryable) Insect(table string) *dat.InsectBuilder {
	b := dat.NewInsectBuilder(table)
	b.Execer = NewExecer(q.runner, b)
	return b
}

// JSQL creates a new JSON SQL builder.
func (q *Queryable) JSQL(sql string, args ...interface{}) *dat.JSQLBuilder {
	b := dat.NewJSQLBuilder(sql, args...)
	b.Execer = NewExecer(q.runner, b)
	return b
}

// Select creates a new SelectBuilder for the given columns.
func (q *Queryable) Select(columns ...string) *dat.SelectBuilder {
	b := dat.NewSelectBuilder(columns...)
	b.Execer = NewExecer(q.runner, b)
	return b
}

// SelectDoc creates a new SelectBuilder for the given columns.
func (q *Queryable) SelectDoc(columns ...string) *dat.SelectDocBuilder {
	b := dat.NewSelectDocBuilder(columns...)
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

// Upsert creates a new UpdateBuilder for the given table.
func (q *Queryable) Upsert(table string) *dat.UpsertBuilder {
	b := dat.NewUpsertBuilder(table)
	b.Execer = NewExecer(q.runner, b)
	return b
}
