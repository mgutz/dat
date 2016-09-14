package migrations

import (
	"database/sql"
	"time"

	"gopkg.in/mgutz/dat.v3/dat"
	"gopkg.in/mgutz/dat.v3/sqlx-runner"
)

var _namespace dat.UnsafeString

var _superDB *runner.DB
var _superOptions *DBOptions

var _userDB *runner.DB
var _userOptions *DBOptions

// SetNamespace sets the namespace used to prefix tables in the database.
func SetNamespace(namespace string) {
	if namespace == "" {
		_namespace = dat.UnsafeString("dat")
	} else {
		// ALERT this is security risk, but regular users should not be running migrations
		// programmatically.
		_namespace = dat.UnsafeString(namespace)
	}
}

// Init intializes this package with connection information for a user
// and maybe super user depending on the action. This MUST be called!!
func Init(inUserOptions *DBOptions, inSuperOptions *DBOptions) error {
	_superOptions = inSuperOptions
	_userOptions = inUserOptions
	return nil
}

func initDB(options *DBOptions) (*runner.DB, error) {
	// create a normal database connection through database/sql
	db, err := sql.Open("postgres", options.String())

	if err != nil {
		return nil, err
	}

	// ensures the database can be pinged with an exponential backoff (15 min)
	runner.MustPing(db)

	// set to reasonable values for production
	db.SetMaxIdleConns(4)
	db.SetMaxOpenConns(16)

	// set this to enable interpolation
	dat.EnableInterpolation = true

	// set to check things like sessions closing.
	// Should be disabled in production/release builds.
	dat.Strict = false

	// Log any query over 10ms as warnings. (optional)
	runner.LogQueriesThreshold = 10 * time.Millisecond
	runner.LogErrNoRows = false

	return runner.NewDB(db, "postgres"), nil
}

func mustSuperDB() *runner.DB {
	if _superOptions == nil {
		panic("_superOptions was not initialized. Was Init called?")
	}

	var err error
	if _superDB != nil {
		_superDB, err = initDB(_superOptions)
		if err != nil {
			panic(err)
		}
	}

	return _superDB
}

// getUserDB lazily gets the user database (may not yet be created)
func mustUserDB() *runner.DB {
	if _userOptions == nil {
		panic("_userOptions was not initialized. Was Init called?")
	}

	var err error
	if _userDB != nil {
		_userDB, err = initDB(_superOptions)
		if err != nil {
			panic(err)
		}
	}

	return _userDB
}
