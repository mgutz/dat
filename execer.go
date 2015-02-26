package dat

// Result serves the same purpose as sql.Result. Defining
// it for the package avoids tight coupling with database/sql.
type Result struct {
	LastInsertID int64
	RowsAffected int64
}

// Execer is any object that executes and queries SQL.
type Execer interface {
	Exec() (*Result, error)
	QueryScalar(destinations ...interface{}) error
	QuerySlice(dest interface{}) error
	QueryStruct(dest interface{}) error
	QueryStructs(dest interface{}) error
}
