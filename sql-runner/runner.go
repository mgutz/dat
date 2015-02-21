package runner

import (
	"database/sql"
	"reflect"
	"time"

	"github.com/mgutz/dat"
)

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

// M is generic string map.
type M map[string]string

// Exec executes the query built by builder.
func exec(runner runner, builder dat.Builder) (sql.Result, error) {
	fullSql, err := builder.Interpolate()
	if err != nil {
		return nil, dat.Events.EventErrKv("exec.interpolate", err, M{"sql": fullSql})
	}

	// Start the timer:
	startTime := time.Now()
	defer func() { dat.Events.TimingKv("exec", time.Since(startTime).Nanoseconds(), M{"sql": fullSql}) }()

	result, err := runner.Exec(fullSql)
	if err != nil {
		return nil, dat.Events.EventErrKv("exec.exec", err, M{"sql": fullSql})
	}

	return result, nil
}

// Query delegates to the internal runner's Query.
func query(runner runner, builder dat.Builder) (*sql.Rows, error) {
	fullSql, err := builder.Interpolate()
	if err != nil {
		return nil, err
	}
	return runner.Query(fullSql)
}

// QueryScan executes the query in builder and loads the resulting data into
// one or more destinations.
//
// Returns ErrNotFound if no value was found, and it was therefore not set.
func queryScan(runner runner, builder dat.Builder, destinations ...interface{}) error {
	fullSql, err := builder.Interpolate()
	if err != nil {
		return err
	}

	// Start the timer:
	startTime := time.Now()
	defer func() { dat.Events.TimingKv("QueryScan", time.Since(startTime).Nanoseconds(), M{"sql": fullSql}) }()

	// Run the query:
	rows, err := runner.Query(fullSql)
	if err != nil {
		return dat.Events.EventErrKv("QueryScan.load_value.query", err, M{"sql": fullSql})
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.Scan(destinations...)

		if err != nil {
			return dat.Events.EventErrKv("QueryScan.load_value.scan", err, M{"sql": fullSql})
		}
		return nil
	}

	if err := rows.Err(); err != nil {
		return dat.Events.EventErrKv("QueryScan.load_value.rows_err", err, M{"sql": fullSql})
	}

	return dat.ErrNotFound
}

// QuerySlice executes the query in builder and loads the resulting data into a
// slice of primitive values
//
// Returns ErrNotFound if no value was found, and it was therefore not set.
func querySlice(runner runner, builder dat.Builder, dest interface{}) (int64, error) {
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
	defer func() { dat.Events.TimingKv("querySlice", time.Since(startTime).Nanoseconds(), M{"sql": fullSql}) }()

	// Run the query:
	rows, err := runner.Query(fullSql)
	if err != nil {
		return numberOfRowsReturned, dat.Events.EventErrKv("querySlice.load_all_values.query", err, M{"sql": fullSql})
	}
	defer rows.Close()

	sliceValue := valueOfDest
	for rows.Next() {
		// Create a new value to store our row:
		pointerToNewValue := reflect.New(recordType)
		newValue := reflect.Indirect(pointerToNewValue)

		err = rows.Scan(pointerToNewValue.Interface())
		if err != nil {
			return numberOfRowsReturned, dat.Events.EventErrKv("querySlice.load_all_values.scan", err, M{"sql": fullSql})
		}

		// Append our new value to the slice:
		sliceValue = reflect.Append(sliceValue, newValue)

		numberOfRowsReturned++
	}
	valueOfDest.Set(sliceValue)

	if err := rows.Err(); err != nil {
		return numberOfRowsReturned, dat.Events.EventErrKv("querySlice.load_all_values.rows_err", err, M{"sql": fullSql})
	}

	return numberOfRowsReturned, nil
}

// QueryStruct executes the query in builder and loads the resulting data into
// a struct dest must be a pointer to a struct
//
// Returns ErrNotFound if nothing was found
func queryStruct(runner runner, builder dat.Builder, dest interface{}) error {
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
	defer func() { dat.Events.TimingKv("QueryStruct", time.Since(startTime).Nanoseconds(), M{"sql": fullSql}) }()

	// Run the query:
	rows, err := runner.Query(fullSql)
	if err != nil {
		return dat.Events.EventErrKv("QueryStruct.query", err, M{"sql": fullSql})
	}
	defer rows.Close()

	// Get the columns of this result set
	columns, err := rows.Columns()
	if err != nil {
		return dat.Events.EventErrKv("QueryStruct.rows.Columns", err, M{"sql": fullSql})
	}

	// Create a map of this result set to the struct columns
	fieldMap, err := dat.CalculateFieldMap(recordType, columns, Strict)
	if err != nil {
		return dat.Events.EventErrKv("QueryStruct.calculateFieldMap", err, M{"sql": fullSql})
	}

	// Build a 'holder', which is an []interface{}. Each value will be the set to address of the field corresponding to our newly made records:
	holder := make([]interface{}, len(fieldMap))

	if rows.Next() {
		// Build a 'holder', which is an []interface{}. Each value will be the address of the field corresponding to our newly made record:
		scannable, err := dat.PrepareHolderFor(indirectOfDest, fieldMap, holder)
		if err != nil {
			return dat.Events.EventErrKv("QueryStruct.holderFor", err, M{"sql": fullSql})
		}

		// Load up our new structure with the row's values
		err = rows.Scan(scannable...)
		if err != nil {
			return dat.Events.EventErrKv("QueryStruct.scan", err, M{"sql": fullSql})
		}
		return nil
	}

	if err := rows.Err(); err != nil {
		return dat.Events.EventErrKv("QueryStruct.rows_err", err, M{"sql": fullSql})
	}

	return dat.ErrNotFound
}

// QueryStructs executes the query in builderand loads the resulting data into
// a slice of structs. dest must be a pointer to a slice of pointers to structs
//
// Returns the number of items found (which is not necessarily the # of items
// set)
func queryStructs(runner runner, builder dat.Builder, dest interface{}) (int64, error) {
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
		return 0, dat.Events.EventErr("QueryStructs.interpolate", err)
	}

	var numberOfRowsReturned int64

	// Start the timer:
	startTime := time.Now()
	defer func() { dat.Events.TimingKv("QueryStructs", time.Since(startTime).Nanoseconds(), M{"sql": fullSql}) }()

	// Run the query:
	rows, err := runner.Query(fullSql)
	if err != nil {
		return 0, dat.Events.EventErrKv("QueryStructs.query", err, M{"sql": fullSql})
	}
	defer rows.Close()

	// Get the columns returned
	columns, err := rows.Columns()
	if err != nil {
		return numberOfRowsReturned, dat.Events.EventErrKv("QueryStruct.rows.Columns", err, M{"sql": fullSql})
	}

	// Create a map of this result set to the struct fields
	fieldMap, err := dat.CalculateFieldMap(recordType, columns, Strict)
	if err != nil {
		return numberOfRowsReturned, dat.Events.EventErrKv("QueryStructs.calculateFieldMap", err, M{"sql": fullSql})
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
			return numberOfRowsReturned, dat.Events.EventErrKv("QueryStructs.holderFor", err, M{"sql": fullSql})
		}

		// Load up our new structure with the row's values
		err = rows.Scan(scannable...)
		if err != nil {
			return numberOfRowsReturned, dat.Events.EventErrKv("QueryStructs.scan", err, M{"sql": fullSql})
		}

		// Append our new record to the slice:
		sliceValue = reflect.Append(sliceValue, pointerToNewRecord)

		numberOfRowsReturned++
	}
	valueOfDest.Set(sliceValue)

	// Check for errors at the end. Supposedly these are error that can happen during iteration.
	if err = rows.Err(); err != nil {
		return numberOfRowsReturned, dat.Events.EventErrKv("QueryStructs.rows_err", err, M{"sql": fullSql})
	}

	return numberOfRowsReturned, nil
}
