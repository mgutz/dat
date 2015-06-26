package dat

import "time"

// Result serves the same purpose as sql.Result. Defining
// it for the package avoids tight coupling with database/sql.
type Result struct {
	LastInsertID int64
	RowsAffected int64
}

// Execer is any object that executes and queries SQL.
type Execer interface {
	Cache(id string, ttl time.Duration, invalidate bool) Execer
	Exec() (*Result, error)
	QueryScalar(destinations ...interface{}) error
	QuerySlice(dest interface{}) error
	QueryStruct(dest interface{}) error
	QueryStructs(dest interface{}) error
	QueryObject(dest interface{}) error
	QueryJSON() ([]byte, error)
}

const panicExecerMsg = "dat builders are disconnected, use sqlx-runner package"

var nullExecer = &panicExecer{}

// panicExecer is the execer assigned when a builder is first created.
// panicExecer raises a panic if any of the Execer methods are called
// directly from dat. Runners override the execer to work with a live
// database.
type panicExecer struct{}

func (nop *panicExecer) Cache(id string, ttl time.Duration, invalidate bool) Execer {
	panic(panicExecerMsg)
}

// Exec panics when Exec is called.
func (nop *panicExecer) Exec() (*Result, error) {
	panic(panicExecerMsg)
}

// QueryScalar panics when QueryScalar is called.
func (nop *panicExecer) QueryScalar(destinations ...interface{}) error {
	panic(panicExecerMsg)
}

// QuerySlice panics when QuerySlice is called.
func (nop *panicExecer) QuerySlice(dest interface{}) error {
	panic(panicExecerMsg)
}

// QueryStruct panics when QueryStruct is called.
func (nop *panicExecer) QueryStruct(dest interface{}) error {
	panic(panicExecerMsg)
}

// QueryStructs panics when QueryStructs is called.
func (nop *panicExecer) QueryStructs(dest interface{}) error {
	panic(panicExecerMsg)
}

// QueryObject panics when QueryObject is called.
func (nop *panicExecer) QueryObject(dest interface{}) error {
	panic(panicExecerMsg)
}

// QueryJSON panics when QueryJSON is called.
func (nop *panicExecer) QueryJSON() ([]byte, error) {
	panic(panicExecerMsg)
}
