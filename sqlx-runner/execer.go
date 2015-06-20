package runner

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/jmoiron/sqlx"
	"gopkg.in/mgutz/dat.v1"
)

// Execer executes queries against a database.
type Execer struct {
	database
	builder dat.Builder

	cacheID         string
	cacheTTL        time.Duration
	cacheInvalidate bool
}

// NewExecer creates a new instance of Execer.
func NewExecer(database database, builder dat.Builder) *Execer {
	return &Execer{
		database: database,
		builder:  builder,
	}
}

// Cache caches the results of queries for Select and SelectDoc.
func (ex *Execer) Cache(id string, ttl time.Duration, invalidate bool) dat.Execer {
	ex.cacheID = id
	ex.cacheTTL = ttl
	ex.cacheInvalidate = invalidate
	return ex
}

// Exec executes a builder's query.
func (ex *Execer) Exec() (*dat.Result, error) {
	res, err := exec(ex)
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
	n, err := query(ex)
	return n, traceError("Queryx", err)
}

// QueryScalar executes builder's query and scans returned row into destinations.
func (ex *Execer) QueryScalar(destinations ...interface{}) error {
	err := queryScalar(ex, destinations...)
	return traceError("QueryScalar", err)
}

// QuerySlice executes builder's query and builds a slice of values from each row, where
// each row only has one column.
func (ex *Execer) QuerySlice(dest interface{}) error {
	err := querySlice(ex, dest)
	return traceError("QuerySlice", err)
}

// QueryStruct executes builders' query and scans the result row into dest.
func (ex *Execer) QueryStruct(dest interface{}) error {
	// TODO this is a hack. All of this runner, execer nested structs is getting messy.
	// Use a godo task to copy methods instead of this mess.
	if _, ok := ex.builder.(*dat.SelectDocBuilder); ok {
		err := queryJSONStruct(ex, dest)
		return traceError("QueryJSONStruct", err)
	}
	err := queryStruct(ex, dest)
	return traceError("QueryStruct", err)
}

// QueryStructs executes builders' query and scans each row as an item in a slice of structs.
func (ex *Execer) QueryStructs(dest interface{}) error {
	// TODO this is a hack. All of this runner, execer nested structs is getting messy.
	// Use a godo task to copy methods instead of this mess.
	if _, ok := ex.builder.(*dat.SelectDocBuilder); ok {
		err := queryJSONStructs(ex, dest)
		return traceError("QueryJSONStructs", err)
	}

	err := queryStructs(ex, dest)
	return traceError("QueryStructs", err)
}

// QueryObject wraps the builder's query within a `to_json` then executes and unmarshals
// the result into dest.
func (ex *Execer) QueryObject(dest interface{}) error {
	// TODO this is a hack. All of this runner, execer nested structs is messy.
	// Use a godo task to copy methods instead of this mess.
	if _, ok := ex.builder.(*dat.SelectDocBuilder); ok {
		b, err := queryJSONBlob(ex, false)
		if err != nil {
			return err
		}
		return json.Unmarshal(b, dest)
	}

	err := queryObject(ex, dest)
	return traceError("QueryObject", err)
}

// QueryJSON wraps the builder's query within a `to_json` then executes and returns
// the JSON []byte representation.
func (ex *Execer) QueryJSON() ([]byte, error) {
	// TODO this is a hack. All of this runner, execer nested structs is messy.
	// Use a godo task to copy methods instead of this mess.
	if _, ok := ex.builder.(*dat.SelectDocBuilder); ok {
		return queryJSONBlob(ex, false)
	}

	b, err := queryJSON(ex)
	return b, traceError("QueryObject", err)
}

func traceError(name string, err error) error {
	if dat.Strict && err != nil && err != sql.ErrNoRows && err != dat.ErrNotFound {
		logger.Error(name, "err", err)
	}
	return err
}
