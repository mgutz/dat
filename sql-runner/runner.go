package runner

import (
	"database/sql"
	"reflect"
	"time"

	"github.com/mgutz/dat"
)

// TODO eventreceiver should just be a log interface with a verbose flag

// Unvetted thots:
// Given a query and given a structure (field list), there's 2 sets of fields.
// Take the intersection. We can fill those in. great.
// For fields in the structure that aren't in the query, we'll let that slide if db:"-"
// For fields in the structure that aren't in the query but without db:"-", return error
// For fields in the query that aren't in the structure, we'll ignore them.

type runner interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}

// Runner executes SQL.
type Runner struct {
	runner
}

// M is generic string map.
type M map[string]string

// Exec executes the query built by builder.
func (r Runner) Exec(builder dat.Builder) (sql.Result, error) {
	fullSql, err := builder.Interpolate()
	if err != nil {
		return nil, events.EventErrKv("exec.interpolate", err, M{"sql": fullSql})
	}

	// Start the timer:
	startTime := time.Now()
	defer func() { events.TimingKv("exec", time.Since(startTime).Nanoseconds(), M{"sql": fullSql}) }()

	result, err := r.runner.Exec(fullSql)
	if err != nil {
		return nil, events.EventErrKv("exec.exec", err, M{"sql": fullSql})
	}

	return result, nil
}

// QueryStruct executes the query in builder and loads the resulting data into
// a struct dest must be a pointer to a struct
//
// Returns ErrNotFound if nothing was found
func (r Runner) QueryStruct(dest interface{}, builder dat.Builder) error {
	//
	// Validate the dest, and extract the reflection values we need.
	//
	valueOfDest := reflect.ValueOf(dest)
	indirectOfDest := reflect.Indirect(valueOfDest)
	kindOfDest := valueOfDest.Kind()

	if kindOfDest != reflect.Ptr || indirectOfDest.Kind() != reflect.Struct {
		panic("you need to pass in the address of a struct")
	}

	recordType := indirectOfDest.Type()

	//
	// Get full SQL
	//
	fullSql, err := builder.Interpolate()
	if err != nil {
		return err
	}

	// Start the timer:
	startTime := time.Now()
	defer func() { events.TimingKv("queryStruct", time.Since(startTime).Nanoseconds(), M{"sql": fullSql}) }()

	// Run the query:
	rows, err := r.runner.Query(fullSql)
	if err != nil {
		return events.EventErrKv("queryStruct.query", err, M{"sql": fullSql})
	}
	defer rows.Close()

	// Get the columns of this result set
	columns, err := rows.Columns()
	if err != nil {
		return events.EventErrKv("queryStruct.rows.Columns", err, M{"sql": fullSql})
	}

	// Create a map of this result set to the struct columns
	fieldMap, err := dat.CalculateFieldMap(recordType, columns, false)
	if err != nil {
		return events.EventErrKv("queryStruct.calculateFieldMap", err, M{"sql": fullSql})
	}

	// Build a 'holder', which is an []interface{}. Each value will be the set to address of the field corresponding to our newly made records:
	holder := make([]interface{}, len(fieldMap))

	if rows.Next() {
		// Build a 'holder', which is an []interface{}. Each value will be the address of the field corresponding to our newly made record:
		scannable, err := dat.PrepareHolderFor(indirectOfDest, fieldMap, holder)
		if err != nil {
			return events.EventErrKv("queryStruct.holderFor", err, M{"sql": fullSql})
		}

		// Load up our new structure with the row's values
		err = rows.Scan(scannable...)
		if err != nil {
			return events.EventErrKv("queryStruct.scan", err, M{"sql": fullSql})
		}
		return nil
	}

	if err := rows.Err(); err != nil {
		return events.EventErrKv("queryStruct.rows_err", err, M{"sql": fullSql})
	}

	return dat.ErrNotFound
}

// QueryStructs executes the query in builderand loads the resulting data into
// a slice of structs. dest must be a pointer to a slice of pointers to structs
//
// Returns the number of items found (which is not necessarily the # of items
// set)
func (r Runner) QueryStructs(dest interface{}, builder dat.Builder) (int64, error) {
	//
	// Validate the dest, and extract the reflection values we need.
	//

	// This must be a pointer to a slice
	valueOfDest := reflect.ValueOf(dest)
	kindOfDest := valueOfDest.Kind()

	if kindOfDest != reflect.Ptr {
		panic("invalid type passed to LoadStructs. Need a pointer to a slice")
	}

	// This must a slice
	valueOfDest = reflect.Indirect(valueOfDest)
	kindOfDest = valueOfDest.Kind()

	if kindOfDest != reflect.Slice {
		panic("invalid type passed to LoadStructs. Need a pointer to a slice")
	}

	// The slice elements must be pointers to structures
	recordType := valueOfDest.Type().Elem()
	if recordType.Kind() != reflect.Ptr {
		panic("Elements need to be pointers to structures")
	}

	recordType = recordType.Elem()
	if recordType.Kind() != reflect.Struct {
		panic("Elements need to be pointers to structures")
	}

	//
	// Get full SQL
	//
	fullSql, err := builder.Interpolate()
	if err != nil {
		return 0, events.EventErr("queryStructs.interpolate", err)
	}

	var numberOfRowsReturned int64

	// Start the timer:
	startTime := time.Now()
	defer func() { events.TimingKv("queryStructs", time.Since(startTime).Nanoseconds(), M{"sql": fullSql}) }()

	// Run the query:
	rows, err := r.runner.Query(fullSql)
	if err != nil {
		return 0, events.EventErrKv("queryStructs.query", err, M{"sql": fullSql})
	}
	defer rows.Close()

	// Get the columns returned
	columns, err := rows.Columns()
	if err != nil {
		return numberOfRowsReturned, events.EventErrKv("queryStruct.rows.Columns", err, M{"sql": fullSql})
	}

	// Create a map of this result set to the struct fields
	fieldMap, err := dat.CalculateFieldMap(recordType, columns, false)
	if err != nil {
		return numberOfRowsReturned, events.EventErrKv("queryStructs.calculateFieldMap", err, M{"sql": fullSql})
	}

	// Build a 'holder', which is an []interface{}. Each value will be the set to address of the field corresponding to our newly made records:
	holder := make([]interface{}, len(fieldMap))

	// Iterate over rows and scan their data into the structs
	sliceValue := valueOfDest
	for rows.Next() {
		// Create a new record to store our row:
		pointerToNewRecord := reflect.New(recordType)
		newRecord := reflect.Indirect(pointerToNewRecord)

		// Prepare the holder for this record
		scannable, err := dat.PrepareHolderFor(newRecord, fieldMap, holder)
		if err != nil {
			return numberOfRowsReturned, events.EventErrKv("queryStructs.holderFor", err, M{"sql": fullSql})
		}

		// Load up our new structure with the row's values
		err = rows.Scan(scannable...)
		if err != nil {
			return numberOfRowsReturned, events.EventErrKv("queryStructs.scan", err, M{"sql": fullSql})
		}

		// Append our new record to the slice:
		sliceValue = reflect.Append(sliceValue, pointerToNewRecord)

		numberOfRowsReturned++
	}
	valueOfDest.Set(sliceValue)

	// Check for errors at the end. Supposedly these are error that can happen during iteration.
	if err = rows.Err(); err != nil {
		return numberOfRowsReturned, events.EventErrKv("queryStructs.rows_err", err, M{"sql": fullSql})
	}

	return numberOfRowsReturned, nil
}

// QueryScalar executes the query in builder and loads the resulting data into
// a primitive value
//
// Returns ErrNotFound if no value was found, and it was therefore not set.
func (r Runner) QueryScalar(dest interface{}, builder dat.Builder) error {
	// Validate the dest
	valueOfDest := reflect.ValueOf(dest)
	kindOfDest := valueOfDest.Kind()

	if kindOfDest != reflect.Ptr {
		panic("Destination must be a pointer")
	}

	//
	// Get full SQL
	//
	fullSql, err := builder.Interpolate()
	if err != nil {
		return err
	}

	// Start the timer:
	startTime := time.Now()
	defer func() { events.TimingKv("queryScalar", time.Since(startTime).Nanoseconds(), M{"sql": fullSql}) }()

	// Run the query:
	rows, err := r.runner.Query(fullSql)
	if err != nil {
		return events.EventErrKv("queryScalar.load_value.query", err, M{"sql": fullSql})
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(dest)
		if err != nil {
			return events.EventErrKv("queryScalar.load_value.scan", err, M{"sql": fullSql})
		}
		return nil
	}

	if err := rows.Err(); err != nil {
		return events.EventErrKv("queryScalar.load_value.rows_err", err, M{"sql": fullSql})
	}

	return dat.ErrNotFound
}

// QuerySlice executes the query in builder and loads the resulting data into a
// slice of primitive values
//
// Returns ErrNotFound if no value was found, and it was therefore not set.
func (r Runner) QuerySlice(dest interface{}, builder dat.Builder) (int64, error) {
	// Validate the dest and reflection values we need

	// This must be a pointer to a slice
	valueOfDest := reflect.ValueOf(dest)
	kindOfDest := valueOfDest.Kind()

	if kindOfDest != reflect.Ptr {
		panic("invalid type passed to LoadValues. Need a pointer to a slice")
	}

	// This must a slice
	valueOfDest = reflect.Indirect(valueOfDest)
	kindOfDest = valueOfDest.Kind()

	if kindOfDest != reflect.Slice {
		panic("invalid type passed to LoadValues. Need a pointer to a slice")
	}

	recordType := valueOfDest.Type().Elem()

	recordTypeIsPtr := recordType.Kind() == reflect.Ptr
	if recordTypeIsPtr {
		reflect.ValueOf(dest)
	}

	//
	// Get full SQL
	//
	fullSql, err := builder.Interpolate()
	if err != nil {
		return 0, err
	}

	var numberOfRowsReturned int64

	// Start the timer:
	startTime := time.Now()
	defer func() { events.TimingKv("querySlice", time.Since(startTime).Nanoseconds(), M{"sql": fullSql}) }()

	// Run the query:
	rows, err := r.runner.Query(fullSql)
	if err != nil {
		return numberOfRowsReturned, events.EventErrKv("querySlice.load_all_values.query", err, M{"sql": fullSql})
	}
	defer rows.Close()

	sliceValue := valueOfDest
	for rows.Next() {
		// Create a new value to store our row:
		pointerToNewValue := reflect.New(recordType)
		newValue := reflect.Indirect(pointerToNewValue)

		err = rows.Scan(pointerToNewValue.Interface())
		if err != nil {
			return numberOfRowsReturned, events.EventErrKv("querySlice.load_all_values.scan", err, M{"sql": fullSql})
		}

		// Append our new value to the slice:
		sliceValue = reflect.Append(sliceValue, newValue)

		numberOfRowsReturned++
	}
	valueOfDest.Set(sliceValue)

	if err := rows.Err(); err != nil {
		return numberOfRowsReturned, events.EventErrKv("querySlice.load_all_values.rows_err", err, M{"sql": fullSql})
	}

	return numberOfRowsReturned, nil
}
