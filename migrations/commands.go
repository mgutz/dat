package migrations

import (
	"fmt"
	"os/exec"
	"strconv"

	"gopkg.in/mgutz/dat.v3/dat"
)

// Migration is an entry in the migrations table.
type Migration struct {
	Version   string       `db:"version"`
	Up        string       `db:"up"`
	Down      string       `db:"down"`
	CreatedAt dat.NullTime `db:"created_at"`
}

// Sproc is an entry in the sprocs table.
type Sproc struct {
	Name      string       `db:"name"`
	CRC       string       `db:"crc"`
	CreatedAT dat.NullTime `db:"created_at"`
}

// SQLError is a standardized error from SQL Server
type SQLError struct {
	Name     string
	Message  string
	Filename string
	Line     int
	Column   int
}

// ToSQLError converts something to SQLError
func ToSQLError() string {
	// do we get any error from go's database?
	return ""
}

// Postgres is an abstraction for Postgres
type Postgres struct {
	port int
	host string
}

// Bootstrap bootstraps user's database with metadata tables
func (pg *Postgres) Bootstrap(namespace string) error {
	q := `
create table if not exists $1__migrations (
  version varchar(256) not null primary key,
  up text,
  down text,
  created_at timestamp default current_timestamp
);

create table if not exists $1__sprocs (
  name text primary key,
  crc text not null,
  created_at timestamp default current_timestamp
);

CREATE OR REPLACE FUNCTION $1__delfunc(_name text) returns void AS $$
BEGIN
    EXECUTE (
       SELECT string_agg(format('DROP FUNCTION %s(%s);'
                         ,oid::regproc
                         ,pg_catalog.pg_get_function_identity_arguments(oid))
              ,E'\n')
       FROM   pg_proc
       WHERE  proname = _name
       AND    pg_function_is_visible(oid)
    );
EXCEPTION WHEN OTHERS THEN
  -- do nothing, EXEC above returns an exception if existing function not found
END $$ LANGUAGE plpgsql;
`
	// should this be the user database
	_, err := mustSuperDB().SQL(q, _namespace).Exec()
	return err
}

// Add adds a migration to the database.
func Add(migration *Migration) error {
	q := `
		INSERT INTO $1__migrations(version, up, down)
		VALUES($2, $3, $4)
	`
	mustUserDB().SQL(q, _namespace, migration.Version, migration.Up, migration.Down)
	return nil
}

// Console launches psql console.
// TODO this aint working
func Console() error {
	cmd := exec.Command("psql",
		"-U", _userOptions.User,
		"-h", _userOptions.Host,
		"-p", strconv.Itoa(_userOptions.Port),
	)
	if _userOptions.Password != "" {
		cmd.Env = append(cmd.Env, "PGPASSWORD=", _userOptions.Password)
	}
	err := cmd.Run()
	if err != nil {
		fmt.Println("ERR", err)
		return err
	}
	return nil
}

func table(name string) dat.UnsafeString {
	return dat.UnsafeString(name)
}

// CreateUserDB creates user database. Super user database options must be valid.
func CreateUserDB(superOptions *DBOptions, userOptions *DBOptions) error {
	// kills all existing connections then (re)creates the database specified in user options
	statements := `	
	# kill all connections first
	# NOTE: pid is procpid in PostgreSQL < 9.2
	"""
	select pg_terminate_backend(pid)
	from pg_stat_activity
	where datname='#{config.database}'
		and pid <> pg_backend_pid()
	"""
	GO
	"drop database if exists #{config.database};"
	GO
	"drop user if exists #{config.user};"
	GO
	"create user #{config.user} password '#{config.password}' SUPERUSER CREATEROLE;"
	GO
	"create database #{config.database} owner #{config.user};"
`

	/*
		  createDatabase: (defaultUser, argv) ->
		    self = @
		    config = @config
		    using = @using

		    doCreate = (err, result) ->
		      {user, password, host, port} = result
		      password = null if password.trim().length == 0

		      statements = [
		          # kill all connections first
		          # NOTE: pid is procpid in PostgreSQL < 9.2
		          """
		            select pg_terminate_backend(pid)
		            from pg_stat_activity
		            where datname='#{config.database}'
		              and pid <> pg_backend_pid()
		          """
		          "drop database if exists #{config.database};"
		          "drop user if exists #{config.user};"
		          "create user #{config.user} password '#{config.password}' SUPERUSER CREATEROLE;"
		          "create database #{config.database} owner #{config.user};"
		      ]

		      rootConfig =
		        user: user
		        password: password
		        host: config.host
		        port: config.port
		        database: "postgres"
		      RootDb = Postgres.define(rootConfig)
		      store = new RootDb()

		      execRootSql = (sql, cb) ->
		        console.log 'SQL', sql
		        store.sql(sql).exec cb

		      Async.forEachSeries statements, execRootSql, (err) ->
		        if (err)
		          console.error err
		          console.error "Verify migrations/config.js has the correct host and port"
		          process.exit 1
		        else
		          console.log """Created
		\tdatabase: #{config.database}
		\tuser: #{config.user}
		\tpassword: #{config.password}
		\thost: #{config.host}
		\tport: #{config.port}
		"""
		          console.log "OK"
		          process.exit 0

		    if argv.user
		      @argvSuperUser argv, doCreate
		    else
		      @promptSuperUser defaultUser, doCreate


	*/
	return nil
}

// DropUserDB drops user database.
func DropUserDB() error {
	return nil
}

// ExecFileCLI executes a file through `psql` command line utility.
func ExecFileCLI(filename string) error {
	return nil
}

// ExecFileDriver executes a file through `dat` which supports batch separator.
func ExecFileDriver() error {
	return nil
}

// FetchMigrations fetches all migrations from the database sorted by created_at desc
func FetchMigrations() ([]*Migration, error) {
	return nil, nil
}

// RegisterSproc registers a stored procedure and logs it dat__sprocs
func RegisterSproc() error {
	return nil
}

// Remove removes a migration from the database
func Remove(version string) error {
	return nil
}
