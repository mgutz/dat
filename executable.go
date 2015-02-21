package dat

import "database/sql"

// Executable is an object that can be queried.
type Executable interface {
	Exec() (sql.Result, error)
	//Query() (*sql.Rows, error)
	QueryScalar(destinations ...interface{}) error
	QuerySlice(dest interface{}) error
	QueryStruct(dest interface{}) error
	QueryStructs(dest interface{}) error
}
