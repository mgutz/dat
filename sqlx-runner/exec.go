package runner

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/syreclabs/dat"
	"github.com/syreclabs/dat/kvs"
)

// database is the interface for sqlx's DB or Tx against which
// queries can be executed
type database interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Queryx(query string, args ...interface{}) (*sqlx.Rows, error)
	QueryRowx(query string, args ...interface{}) *sqlx.Row
	Select(dest interface{}, query string, args ...interface{}) error
	Get(dest interface{}, query string, args ...interface{}) error
}

func toOutputStr(args []interface{}) string {
	if args == nil {
		return "nil"
	}
	var buf bytes.Buffer
	for i, arg := range args {
		if i > 0 {
			buf.WriteString(" ")
		}
		buf.WriteString("$")
		buf.WriteString(strconv.Itoa(i + 1))
		buf.WriteString("=")
		switch t := arg.(type) {
		default:
			buf.WriteString(fmt.Sprintf("%v", t))
		case []byte:
			buf.WriteString("<binary>")
		}
	}
	return buf.String()
}

func logSQLError(err error, msg string, statement string, args []interface{}) error {
	if err != nil && err != sql.ErrNoRows {
		logger.Error(msg, "err", err, "sql", statement, "args", toOutputStr(args))
	}
	return err
}

func logExecutionTime(start time.Time, sql string, args []interface{}) {
	logged := false
	if logger.IsWarn() {
		elapsed := time.Since(start)
		if LogQueriesThreshold > 0 && elapsed.Nanoseconds() > LogQueriesThreshold.Nanoseconds() {
			if len(args) > 0 {
				logger.Warn("SLOW query", "elapsed", fmt.Sprintf("%s", elapsed), "sql", sql, "args", toOutputStr(args))
			} else {
				logger.Warn("SLOW query", "elapsed", fmt.Sprintf("%s", elapsed), "sql", sql)
			}
			logged = true
		}
	}

	if logger.IsInfo() && !logged {
		elapsed := time.Since(start)
		logger.Info("Query time", "elapsed", fmt.Sprintf("%s", elapsed), "sql", sql)
	}
}

// Exec executes the query built by builder.
func exec(execer *Execer) (sql.Result, error) {
	fullSQL, args, err := execer.builder.Interpolate()
	if err != nil {
		logger.Error("exec.10", "err", err, "sql", fullSQL)
		return nil, err
	}
	defer logExecutionTime(time.Now(), fullSQL, args)

	var result sql.Result
	if args == nil {
		result, err = execer.database.Exec(fullSQL)
	} else {
		result, err = execer.database.Exec(fullSQL, args...)
	}
	if err != nil {
		return nil, logSQLError(err, "exec.30", fullSQL, args)
	}

	return result, nil
}

// Query delegates to the internal runner's Query.
func query(execer *Execer) (*sqlx.Rows, error) {
	fullSQL, args, err := execer.builder.Interpolate()
	if err != nil {
		return nil, err
	}

	defer logExecutionTime(time.Now(), fullSQL, args)
	var rows *sqlx.Rows
	if args == nil {
		rows, err = execer.database.Queryx(fullSQL)
	} else {
		rows, err = execer.database.Queryx(fullSQL, args...)
	}
	if err != nil {
		return nil, logSQLError(err, "query", fullSQL, args)
	}

	return rows, nil
}

// QueryScan executes the query in builder and loads the resulting data into
// one or more destinations.
//
// Returns ErrNotFound if no value was found, and it was therefore not set.
func queryScalar(execer *Execer, destinations ...interface{}) error {
	fullSQL, args, blob, err := cacheOrSQL(execer)
	if err != nil {
		return err
	}
	if blob != nil {
		err = json.Unmarshal(blob, &destinations)
		if err == nil {
			return nil
		}
		// log it and fallthrough to let the query continue
		logger.Warn("queryScalar.2: Could not unmarshal cache data. Continuing with query")
	}

	defer logExecutionTime(time.Now(), fullSQL, args)
	// Run the query:
	var rows *sqlx.Rows
	if args == nil {
		rows, err = execer.database.Queryx(fullSQL)
	} else {
		rows, err = execer.database.Queryx(fullSQL, args...)
	}
	if err != nil {
		return logSQLError(err, "QueryScalar.load_value.query", fullSQL, args)
	}

	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(destinations...)
		if err != nil {
			return logSQLError(err, "QueryScalar.load_value.scan", fullSQL, args)
		}

		setCache(execer, destinations, dtStruct)

		return nil
	}
	if err := rows.Err(); err != nil {
		return logSQLError(err, "QueryScalar.load_value.rows_err", fullSQL, args)
	}

	return dat.ErrNotFound
}

// QuerySlice executes the query in builder and loads the resulting data into a
// slice of primitive values
//
// Returns ErrNotFound if no value was found, and it was therefore not set.
func querySlice(execer *Execer, dest interface{}) error {
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

	fullSQL, args, blob, err := cacheOrSQL(execer)
	if err != nil {
		return err
	}
	if blob != nil {
		err = json.Unmarshal(blob, &dest)
		if err == nil {
			return nil
		}
		// log it and fallthrough to let the query continue
		logger.Warn("querySlice.2: Could not unmarshal cache data. Continuing with query")
	}

	defer logExecutionTime(time.Now(), fullSQL, args)
	var rows *sqlx.Rows
	if args == nil {
		rows, err = execer.database.Queryx(fullSQL)
	} else {
		rows, err = execer.database.Queryx(fullSQL, args...)
	}
	if err != nil {
		return logSQLError(err, "querySlice.load_all_values.query", fullSQL, args)
	}

	sliceValue := valueOfDest
	defer rows.Close()
	for rows.Next() {
		// Create a new value to store our row:
		pointerToNewValue := reflect.New(recordType)
		newValue := reflect.Indirect(pointerToNewValue)

		err = rows.Scan(pointerToNewValue.Interface())
		if err != nil {
			return logSQLError(err, "querySlice.load_all_values.scan", fullSQL, args)
		}

		// Append our new value to the slice:
		sliceValue = reflect.Append(sliceValue, newValue)
	}
	valueOfDest.Set(sliceValue)

	if err := rows.Err(); err != nil {
		return logSQLError(err, "querySlice.load_all_values.rows_err", fullSQL, args)
	}

	setCache(execer, dest, dtStruct)

	return nil
}

// QueryStruct executes the query in builder and loads the resulting data into
// a struct dest must be a pointer to a struct
//
// Returns ErrNotFound if nothing was found
func queryStruct(execer *Execer, dest interface{}) error {
	fullSQL, args, blob, err := cacheOrSQL(execer)
	if err != nil {
		return err
	}
	if blob != nil {
		err = json.Unmarshal(blob, &dest)
		if err == nil {
			return nil
		}
		// log it and fallthrough to let the query continue
		logger.Warn("queryStruct.2: Could not unmarshal queryStruct cache data. Continuing with query")
	}

	defer logExecutionTime(time.Now(), fullSQL, args)
	if args == nil {
		err = execer.database.Get(dest, fullSQL)
	} else {
		err = execer.database.Get(dest, fullSQL, args...)
	}
	if err != nil {
		logSQLError(err, "queryStruct.3", fullSQL, args)
		return err
	}

	setCache(execer, dest, dtStruct)

	return nil
}

// QueryStructs executes the query in builderand loads the resulting data into
// a slice of structs. dest must be a pointer to a slice of pointers to structs
//
// Returns the number of items found (which is not necessarily the # of items
// set)
func queryStructs(execer *Execer, dest interface{}) error {
	fullSQL, args, blob, err := cacheOrSQL(execer)
	if err != nil {
		logger.Error("queryStructs.1: Could not convert to SQL", "err", err)
		return err
	}
	if blob != nil {
		err = json.Unmarshal(blob, dest)
		if err == nil {
			return nil
		}
		// log it and let the query continue
		logger.Warn("queryStructs.2: Could not unmarshal queryStruct cache data. Continuing with query", "err", err)
	}

	defer logExecutionTime(time.Now(), fullSQL, args)
	if args == nil {
		err = execer.database.Select(dest, fullSQL)
	} else {
		err = execer.database.Select(dest, fullSQL, args...)
	}
	if err != nil {
		logSQLError(err, "queryStructs", fullSQL, args)
	}

	setCache(execer, dest, dtStruct)
	return err
}

// queryJSONStruct executes the query in builder and loads the resulting data into
// a struct, using json.Unmarshal().
//
// Returns ErrNotFound if nothing was found
func queryJSONStruct(execer *Execer, dest interface{}) error {
	blob, err := queryJSONBlob(execer, true)
	if err != nil {
		return err
	}
	return json.Unmarshal(blob, dest)
}

// queryJSONBlob executes the query in builder and loads the resulting data
// into a blob. If a single item is to be returned, set single to true.
//
// Returns ErrNotFound if nothing was found
func queryJSONBlob(execer *Execer, single bool) ([]byte, error) {
	fullSQL, args, blob, err := cacheOrSQL(execer)
	if err != nil {
		return nil, err
	}
	if blob != nil {
		return blob, nil
	}

	defer logExecutionTime(time.Now(), fullSQL, args)
	var rows *sqlx.Rows
	// Run the query:
	if args == nil {
		rows, err = execer.database.Queryx(fullSQL)
	} else {
		rows, err = execer.database.Queryx(fullSQL, args...)
	}
	if err != nil {
		return nil, logSQLError(err, "queryJSONStructs", fullSQL, args)
	}

	// TODO optimize this later, may be better to
	var buf bytes.Buffer
	i := 0
	if single {
		defer rows.Close()
		for rows.Next() {
			if i == 1 {
				if dat.Strict {
					logSQLError(errors.New("Multiple results returned"), "Expected single result", fullSQL, args)
					logger.Fatal("Expected single result, got many")
				} else {
					break
				}
			}
			i++

			err = rows.Scan(&blob)
			if err != nil {
				return nil, err
			}
			buf.Write(blob)
		}
	} else {
		defer rows.Close()
		for rows.Next() {
			if i == 0 {
				buf.WriteRune('[')
			} else {
				buf.WriteRune(',')
			}
			i++

			err = rows.Scan(&blob)
			if err != nil {
				return nil, err
			}
			buf.Write(blob)
		}
		if i > 0 {
			buf.WriteRune(']')
		}
	}

	if i == 0 {
		return nil, sql.ErrNoRows
	}

	blob = buf.Bytes()
	setCache(execer, blob, dtBytes)
	return blob, nil
}

// queryJSON executes the query in builder and loads the resulting data into
// a struct, using json.Unmarshal().
//
// Returns ErrNotFound if nothing was found
func queryJSONStructs(execer *Execer, dest interface{}) error {
	blob, err := queryJSONBlob(execer, false)
	if err != nil {
		return err
	}
	return json.Unmarshal(blob, dest)
}

// cacheOrSQL attempts to get a valeu from cache, otherwise it builds
// the SQL and args to be executed. If value = "" then the SQL is built.
func cacheOrSQL(execer *Execer) (sql string, args []interface{}, value []byte, err error) {
	// if a cacheID exists, return the value ASAP
	if Cache != nil && execer.cacheTTL > 0 && execer.cacheID != "" && !execer.cacheInvalidate {
		v, err := Cache.Get(execer.cacheID)
		//logger.Warn("DBG cacheOrSQL.1 getting by id", "id", execer.cacheID, "v", v, "err", err)
		if err != nil && err != kvs.ErrNotFound {
			logger.Error("Unable to read cache key. Continuing with query", "key", execer.cacheID, "err", err)
		} else if v != "" {
			//logger.Warn("DBG cacheOrSQL.11 HIT", "v", v)
			return "", nil, []byte(v), nil
		}
	}

	fullSQL, args, err := execer.builder.Interpolate()
	if err != nil {
		return "", nil, nil, err
	}

	// if there is no cacheID, use the checksum of SQL as the ID
	if Cache != nil && execer.cacheTTL > 0 && execer.cacheID == "" {
		// this must be set for setCache() to work below
		execer.cacheID = kvs.Hash(fullSQL)

		if !execer.cacheInvalidate {
			v, err := Cache.Get(execer.cacheID)
			//logger.Warn("DBG cacheOrSQL.2 getting by hash", "hash", execer.cacheID, "v", v, "err", err)
			if v != "" && (err == nil || err != kvs.ErrNotFound) {
				//logger.Warn("DBG cacheOrSQL.22 HIT")
				return "", nil, []byte(v), nil
			}
		}
	}

	return fullSQL, args, nil, nil
}

const (
	dtStruct = iota
	dtString
	dtBytes
)

// Sets the cache value using the execer.ID key. Note that execer.ID
// is set as a side-effect of calling cacheOrSQL function above if
// execer.cacheID is not set. data must be a string or a value that
// can be json.Marshal'ed to string.
func setCache(execer *Execer, data interface{}, dataType int) {
	if Cache == nil || execer.cacheTTL < 1 {
		return
	}

	var s string
	switch dataType {
	case dtStruct:
		b, err := json.Marshal(data)
		if err != nil {
			logger.Warn("Could not marshal data, clearing", "key", execer.cacheID, "err", err)
			err = Cache.Del(execer.cacheID)
			if err != nil {
				logger.Error("Could not delete cache key", "key", execer.cacheID, "err", err)
			}
			return
		}
		s = string(b)
	case dtString:
		s = data.(string)
	case dtBytes:
		s = string(data.([]byte))
	}

	//logger.Warn("DBG setting cache", "key", execer.cacheID, "data", string(b), "ttl", execer.cacheTTL)
	err := Cache.Set(execer.cacheID, s, execer.cacheTTL)
	if err != nil {
		logger.Warn("Could not set cache. Query will proceed without caching", "err", err)
	}
}

// queryJSON executes the query in builder and loads the resulting JSON into
// a bytes slice compatible.
//
// Returns ErrNotFound if nothing was found
func queryJSON(execer *Execer) ([]byte, error) {
	fullSQL, args, blob, err := cacheOrSQL(execer)
	if err != nil {
		return nil, err
	}
	if blob != nil {
		return blob, nil
	}

	defer logExecutionTime(time.Now(), fullSQL, args)
	jsonSQL := fmt.Sprintf("SELECT TO_JSON(ARRAY_AGG(__datq.*)) FROM (%s) AS __datq", fullSQL)

	if args == nil {
		err = execer.database.Get(&blob, jsonSQL)
	} else {
		err = execer.database.Get(&blob, jsonSQL, args...)
	}
	if err != nil {
		logSQLError(err, "queryJSON", jsonSQL, args)
	}

	setCache(execer, blob, dtBytes)

	return blob, err
}

// queryObject executes the query in builder and loads the resulting data into
// an object agreeable with json.Unmarshal.
//
// Returns ErrNotFound if nothing was found
func queryObject(execer *Execer, dest interface{}) error {
	blob, err := queryJSON(execer)
	if err != nil {
		return err
	}
	return json.Unmarshal(blob, dest)
}
