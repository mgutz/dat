package runner

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/mgutz/dat/v1"
)

// Execer implements dat.Execer
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
		return nil, traceError("Exec", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return nil, traceError("Exec", err)
	}
	return &dat.Result{RowsAffected: rowsAffected}, nil
}

// Queryx executes builder's query and returns rows.
func (ex *Execer) Queryx() (*sqlx.Rows, error) {
	n, err := query(ex.runner, ex.builder)
	return n, traceError("Queryx", err)
}

// QueryScalar executes builder's query and scans returned row into destinations.
func (ex *Execer) QueryScalar(destinations ...interface{}) error {
	err := queryScalar(ex.runner, ex.builder, destinations...)
	return traceError("QueryScalar", err)
}

// QuerySlice executes builder's query and builds a slice of values from each row, where
// each row only has one column.
func (ex *Execer) QuerySlice(dest interface{}) error {
	err := querySlice(ex.runner, ex.builder, dest)
	return traceError("QuerySlice", err)
}

// QueryStruct executes builders' query and scans the result row into dest.
func (ex *Execer) QueryStruct(dest interface{}) error {
	// TODO this is a hack. All of this runner, execer nested structs is getting messy.
	// Use a godo task to copy methods instead of this mess.
	if _, ok := ex.builder.(*dat.SelectDocBuilder); ok {
		err := queryJSONStruct(ex.runner, ex.builder, dest)
		return traceError("QueryJSONStruct", err)
	}
	err := queryStruct(ex.runner, ex.builder, dest)
	return traceError("QueryStruct", err)
}

// QueryStructs executes builders' query and scans each row as an item in a slice of structs.
func (ex *Execer) QueryStructs(dest interface{}) error {
	// TODO this is a hack. All of this runner, execer nested structs is getting messy.
	// Use a godo task to copy methods instead of this mess.
	if _, ok := ex.builder.(*dat.SelectDocBuilder); ok {
		err := queryJSONStructs(ex.runner, ex.builder, dest)
		return traceError("QueryJSONStructs", err)
	}
	err := queryStructs(ex.runner, ex.builder, dest)
	return traceError("QueryStructs", err)
}

// QueryObject wraps the builder's query within a `to_json` then executes and unmarshals
// the result into dest.
func (ex *Execer) QueryObject(dest interface{}) error {
	err := queryObject(ex.runner, ex.builder, dest)
	return traceError("QueryObject", err)
}

// QueryJSON wraps the builder's query within a `to_json` then executes and returns
// the JSON []byte representation.
func (ex *Execer) QueryJSON() ([]byte, error) {
	b, err := queryJSON(ex.runner, ex.builder)
	return b, traceError("QueryObject", err)
}

func traceError(name string, err error) error {
	if dat.Strict && err != nil && err != sql.ErrNoRows && err != dat.ErrNotFound {
		logger.Error(name, "err", err)
	}
	return err
}
