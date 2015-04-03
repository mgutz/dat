package runner

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mgutz/dat/v1"
)

type runner interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
	QueryRowx(query string, args ...interface{}) *sqlx.Row
	Select(dest interface{}, query string, args ...interface{}) error
	Get(dest interface{}, query string, args ...interface{}) error
}

// Exec executes the query built by builder.
func exec(runner runner, builder dat.Builder) (sql.Result, error) {
	fullSQL, args, err := builder.Interpolate()
	if err != nil {
		logger.Error("exec.10", "err", err, "sql", fullSQL)
		return nil, err
	}

	if logger.IsInfo() {
		startTime := time.Now()
		defer func() {
			logger.Info("exec.20", "elapsed", time.Since(startTime).Nanoseconds(), "sql", fullSQL)
		}()
	}

	var result sql.Result
	if args == nil {
		result, err = runner.Exec(fullSQL)
	} else {
		result, err = runner.Exec(fullSQL, args...)
	}
	if err != nil {
		logger.Error("exec.30", "err", err, "sql", fullSQL)
		return nil, err
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

	if logger.IsInfo() {
		// Start the timer:
		startTime := time.Now()
		defer func() {
			logger.Info("QueryScalar",
				"elapsed", time.Since(startTime).Nanoseconds(),
				"sql", fullSQL,
			)
		}()
	}

	// Run the query:
	var rows *sqlx.Rows
	if args == nil {
		rows, err = runner.Queryx(fullSQL)
	} else {
		rows, err = runner.Queryx(fullSQL, args...)
	}
	if err != nil {
		logger.Error("QueryScalar.load_value.query",
			"err", err,
			"sql", fullSQL,
		)
		return err
	}
	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(destinations...)
		if err != nil {
			logger.Error("QueryScalar.load_value.scan",
				"err", err,
				"sql", fullSQL,
			)
			return err
		}
		return nil
	}
	if err := rows.Err(); err != nil {
		logger.Error("QueryScalar.load_value.rows_err",
			"err", err,
			"sql", fullSQL,
		)
		return err
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

	if logger.IsInfo() {
		// Start the timer:
		startTime := time.Now()
		defer func() {
			logger.Info("querySlice",
				"elapsed", time.Since(startTime).Nanoseconds(),
				"sql", fullSQL,
			)
		}()
	}

	// Run the query:
	var rows *sqlx.Rows
	if args == nil {
		rows, err = runner.Queryx(fullSQL)
	} else {
		rows, err = runner.Queryx(fullSQL, args...)
	}
	if err != nil {
		logger.Error("querySlice.load_all_values.query",
			"err", err,
			"sql", fullSQL,
		)
		return err
	}
	defer rows.Close()

	sliceValue := valueOfDest
	for rows.Next() {

		// Create a new value to store our row:
		pointerToNewValue := reflect.New(recordType)
		newValue := reflect.Indirect(pointerToNewValue)

		err = rows.Scan(pointerToNewValue.Interface())
		if err != nil {
			logger.Error("querySlice.load_all_values.scan",
				"err", err,
				"sql", fullSQL,
			)
			return err
		}

		// Append our new value to the slice:
		sliceValue = reflect.Append(sliceValue, newValue)
	}
	valueOfDest.Set(sliceValue)

	if err := rows.Err(); err != nil {
		logger.Error("querySlice.load_all_values.rows_err",
			"err", err, "sql", fullSQL)
		return err
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

	if logger.IsInfo() {
		// Start the timer:
		startTime := time.Now()
		defer func() {
			logger.Info("QueryStruct",
				"elapsed", time.Since(startTime).Nanoseconds(),
				"sql", fullSQL,
			)
		}()
	}

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
		logger.Error("QueryStructs.interpolate", "err", err)
		return err
	}

	if logger.IsInfo() {
		// Start the timer:
		startTime := time.Now()
		defer func() {
			logger.Info("QueryStructs",
				"elapsed", time.Since(startTime).Nanoseconds(),
				"sql", fullSQL,
			)
		}()
	}

	if args == nil {
		return runner.Select(dest, fullSQL)
	}
	return runner.Select(dest, fullSQL, args...)
}

// queryJSONStrut executes the query in builder and loads the resulting data into
// a struct, using json.Unmarshal().
//
// Returns ErrNotFound if nothing was found
func queryJSONStruct(runner runner, builder dat.Builder, dest interface{}) error {
	fullSQL, args, err := builder.Interpolate()
	if err != nil {
		return err
	}

	if logger.IsInfo() {
		// Start the timer:
		startTime := time.Now()
		defer func() {
			logger.Info("QueryJSON",
				"elapsed", time.Since(startTime).Nanoseconds(),
				"sql", fullSQL,
			)
		}()
	}

	var blob []byte

	// Run the query:
	if args == nil {
		err = runner.Get(&blob, fullSQL)
	} else {
		err = runner.Get(&blob, fullSQL, args...)
	}
	if err != nil {
		return err
	}

	return json.Unmarshal(blob, dest)
}

// queryJSON executes the query in builder and loads the resulting data into
// a struct, using json.Unmarshal().
//
// Returns ErrNotFound if nothing was found
func queryJSONStructs(runner runner, builder dat.Builder, dest interface{}) error {
	fullSQL, args, err := builder.Interpolate()
	if err != nil {
		return err
	}

	if logger.IsInfo() {
		// Start the timer:
		startTime := time.Now()
		defer func() {
			logger.Info("QueryJSON",
				"elapsed", time.Since(startTime).Nanoseconds(),
				"sql", fullSQL,
			)
		}()
	}

	var rows *sqlx.Rows
	// Run the query:
	if args == nil {
		rows, err = runner.Queryx(fullSQL)
	} else {
		rows, err = runner.Queryx(fullSQL, args...)
	}
	if err != nil {
		return err
	}

	// TODO optimize this later, may be better to
	var buf bytes.Buffer
	var blob []byte
	i := 0
	for rows.Next() {
		if i == 0 {
			buf.WriteRune('[')
		} else {
			buf.WriteRune(',')
		}
		i++

		err = rows.Scan(&blob)
		if err != nil {
			return err
		}
		buf.Write(blob)
	}

	if i == 0 {
		return sql.ErrNoRows
	}
	buf.WriteRune(']')
	return json.Unmarshal(buf.Bytes(), dest)
}

// queryJSON executes the query in builder and loads the resulting JSON into
// a bytes slice.
//
// Returns ErrNotFound if nothing was found
func queryJSON(runner runner, builder dat.Builder) ([]byte, error) {
	fullSQL, args, err := builder.Interpolate()
	if err != nil {
		return nil, err
	}

	fullSQL = fmt.Sprintf("SELECT TO_JSON(ARRAY_AGG(__datq.*)) FROM (%s) AS __datq", fullSQL)

	if logger.IsInfo() {
		// Start the timer:
		startTime := time.Now()
		defer func() {
			logger.Info("QueryJSON",
				"elapsed", time.Since(startTime).Nanoseconds(),
				"sql", fullSQL,
			)
		}()
	}

	var blob []byte

	// Run the query:
	if args == nil {
		err = runner.Get(&blob, fullSQL)
	} else {
		err = runner.Get(&blob, fullSQL, args...)
	}
	return blob, err
}

// queryObject executes the query in builder and loads the resulting data into
// a simple object.
//
// Returns ErrNotFound if nothing was found
func queryObject(runner runner, builder dat.Builder, dest interface{}) error {
	blob, err := queryJSON(runner, builder)
	if err != nil {
		return err
	}
	return json.Unmarshal(blob, dest)
}
