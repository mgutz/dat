package runner

import (
	"github.com/jmoiron/sqlx"
	"github.com/mgutz/dat"
)

// Execer implements Executable
type Execer struct {
	runner
	builder dat.Builder
}

// NewExecer creates a new instance of Execer.
func NewExecer(runner runner, builder dat.Builder) *Execer {
	return &Execer{runner, builder}
}

// Exec executes a builder's query.
func (ex *Execer) Exec() (*dat.Result, error) {
	res, err := exec(ex.runner, ex.builder)
	if err != nil {
		return nil, err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	return &dat.Result{RowsAffected: rowsAffected}, nil
}

// Queryx executes builder's query and returns rows.
func (ex *Execer) Queryx() (*sqlx.Rows, error) {
	return query(ex.runner, ex.builder)
}

// QueryScalar executes builder's query and scans returned row into destinations.
func (ex *Execer) QueryScalar(destinations ...interface{}) error {
	return queryScalar(ex.runner, ex.builder, destinations...)
}

// QuerySlice executes builder's query and builds a slice of values from each row, where
// each row only has one column.
func (ex *Execer) QuerySlice(dest interface{}) error {
	return querySlice(ex.runner, ex.builder, dest)
}

// QueryStruct executes builders' query and scans the result row into dest.
func (ex *Execer) QueryStruct(dest interface{}) error {
	return queryStruct(ex.runner, ex.builder, dest)
}

// QueryStructs executes builders' query and scans each row as an item in a slice of structs.
func (ex *Execer) QueryStructs(dest interface{}) error {
	return queryStructs(ex.runner, ex.builder, dest)
}
