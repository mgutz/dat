package runner

import (
	"database/sql"
	"reflect"
	"time"

	"github.com/jmoiron/sqlx"
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
	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
	QueryRowx(query string, args ...interface{}) *sqlx.Row
	Select(dest interface{}, query string, args ...interface{}) error
	Get(dest interface{}, query string, args ...interface{}) error
}

// M is generic string map.
type M map[string]string

// Exec executes the query built by builder.
func exec(runner runner, builder dat.Builder) (sql.Result, error) {
	fullSQL, args, err := builder.Interpolate()
	if err != nil {
		return nil, dat.Events.EventErrKv("exec.interpolate", err, M{"sql": fullSQL})
	}

	// Start the timer:
	startTime := time.Now()
	defer func() { dat.Events.TimingKv("exec", time.Since(startTime).Nanoseconds(), M{"sql": fullSQL}) }()

	var result sql.Result
	if args == nil {
		result, err = runner.Exec(fullSQL)
	} else {
		result, err = runner.Exec(fullSQL, args...)
	}
	if err != nil {
		return nil, dat.Events.EventErrKv("exec.exec", err, M{"sql": fullSQL})
	}

	return result, nil
}

// Query delegates to the internal runner's Query.
func query(runner runner, builder dat.Builder) (*sqlx.Rows, error) {
	fullSQL, args, err := builder.Interpolate()
	if err != nil {
		return nil, err
	}

	if args == nil {
		return runner.Queryx(fullSQL)
	}
	return runner.Queryx(fullSQL, args...)
}

// QueryScan executes the query in builder and loads the resulting data into
// one or more destinations.
//
// Returns ErrNotFound if no value was found, and it was therefore not set.
func queryScalar(runner runner, builder dat.Builder, destinations ...interface{}) error {
	fullSQL, args, err := builder.Interpolate()
	if err != nil {
		return err
	}

	// Start the timer:
	startTime := time.Now()
	defer func() { dat.Events.TimingKv("QueryScalar", time.Since(startTime).Nanoseconds(), M{"sql": fullSQL}) }()

	// Run the query:
	var rows *sqlx.Rows
	if args == nil {
		rows, err = runner.Queryx(fullSQL)
	} else {
		rows, err = runner.Queryx(fullSQL, args...)
	}
	if err != nil {
		return dat.Events.EventErrKv("QueryScalar.load_value.query", err, M{"sql": fullSQL})
	}
	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(destinations...)
		if err != nil {
			return dat.Events.EventErrKv("QueryScalar.load_value.scan", err, M{"sql": fullSQL})
		}
		return nil
	}
	if err := rows.Err(); err != nil {
		return dat.Events.EventErrKv("QueryScalar.load_value.rows_err", err, M{"sql": fullSQL})
	}

	return dat.ErrNotFound
}

// QuerySlice executes the query in builder and loads the resulting data into a
// slice of primitive values
//
// Returns ErrNotFound if no value was found, and it was therefore not set.
func querySlice(runner runner, builder dat.Builder, dest interface{}) error {
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

	fullSQL, args, err := builder.Interpolate()
	if err != nil {
		return err
	}

	// Start the timer:
	startTime := time.Now()
	defer func() { dat.Events.TimingKv("querySlice", time.Since(startTime).Nanoseconds(), M{"sql": fullSQL}) }()

	// Run the query:
	var rows *sqlx.Rows
	if args == nil {
		rows, err = runner.Queryx(fullSQL)
	} else {
		rows, err = runner.Queryx(fullSQL, args...)
	}
	if err != nil {
		return dat.Events.EventErrKv("querySlice.load_all_values.query", err, M{"sql": fullSQL})
	}
	defer rows.Close()

	sliceValue := valueOfDest
	for rows.Next() {

		// Create a new value to store our row:
		pointerToNewValue := reflect.New(recordType)
		newValue := reflect.Indirect(pointerToNewValue)

		err = rows.Scan(pointerToNewValue.Interface())
		if err != nil {
			return dat.Events.EventErrKv("querySlice.load_all_values.scan", err, M{"sql": fullSQL})
		}

		// Append our new value to the slice:
		sliceValue = reflect.Append(sliceValue, newValue)
	}
	valueOfDest.Set(sliceValue)

	if err := rows.Err(); err != nil {
		return dat.Events.EventErrKv("querySlice.load_all_values.rows_err", err, M{"sql": fullSQL})
	}

	return nil
}

// QueryStruct executes the query in builder and loads the resulting data into
// a struct dest must be a pointer to a struct
//
// Returns ErrNotFound if nothing was found
func queryStruct(runner runner, builder dat.Builder, dest interface{}) error {
	fullSQL, args, err := builder.Interpolate()
	if err != nil {
		return err
	}

	// Start the timer:
	startTime := time.Now()
	defer func() { dat.Events.TimingKv("QueryStruct", time.Since(startTime).Nanoseconds(), M{"sql": fullSQL}) }()

	// Run the query:

	if args == nil {
		return runner.Get(dest, fullSQL)
	}
	return runner.Get(dest, fullSQL, args...)
}

// QueryStructs executes the query in builderand loads the resulting data into
// a slice of structs. dest must be a pointer to a slice of pointers to structs
//
// Returns the number of items found (which is not necessarily the # of items
// set)
func queryStructs(runner runner, builder dat.Builder, dest interface{}) error {
	fullSQL, args, err := builder.Interpolate()
	if err != nil {
		return dat.Events.EventErr("QueryStructs.interpolate", err)
	}

	// Start the timer:
	startTime := time.Now()
	defer func() { dat.Events.TimingKv("QueryStructs", time.Since(startTime).Nanoseconds(), M{"sql": fullSQL}) }()

	if args == nil {
		return runner.Select(dest, fullSQL)
	}
	return runner.Select(dest, fullSQL, args...)
}
