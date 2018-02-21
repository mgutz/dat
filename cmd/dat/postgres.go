package main

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"

	"github.com/mgutz/dat/dat"
	runner "github.com/mgutz/dat/sqlx-runner"
)

// NewPostgresAdapter instantiates Postgres adapter
func NewPostgresAdapter(options *AppOptions) *PostgresAdapter {
	return &PostgresAdapter{options: options, dbs: map[string]*runner.DB{}}
}

// PostgresAdapter is the adapter for PostgreSQL
type PostgresAdapter struct {
	options *AppOptions
	dbs     map[string]*runner.DB
}

// AcquireDB creates a new database connection which implements runner.Connection
func (pg *PostgresAdapter) AcquireDB(options *AppOptions) (*runner.DB, error) {
	connectionStr := pg.ConnectionString(&options.Connection)

	if conn, ok := pg.dbs[connectionStr]; ok {
		return conn, nil
	}

	// create a normal database connection through database/sql
	db, err := sql.Open("postgres", connectionStr)
	if err != nil {
		panic(err)
	}

	// ensures the database can be pinged with an exponential backoff (15 min)
	// set this to enable interpolation
	dat.EnableInterpolation = true

	// set to check things like sessions closing.
	// Should be disabled in production/release builds.
	dat.Strict = false

	conn := runner.NewDB(db, "postgres")
	pg.dbs[connectionStr] = conn
	return conn, nil
}

// ConnectionString returns a connection string built from options.
func (pg *PostgresAdapter) ConnectionString(options *Connection) string {
	return fmt.Sprintf("dbname=%s user=%s password=%s host=%s %s", options.Database, options.User, options.Password, options.Host, options.ExtraParams)
}

// Bootstrap bootstraps dat metadata
func (pg *PostgresAdapter) Bootstrap(conn runner.Connection) error {
	sql := `
		create table if not exists dat__migrations (
			name text primary key,
			up_script text not null,
			down_script text default '',
			no_tx_script text default '',
			created_at timestamptz default now()
		);

		create table if not exists dat__sprocs (
			name text primary key,
			script text not null,
			crc text not null,
			updated_at timestamptz default now(),
			created_at timestamptz default now()
		);

        CREATE OR REPLACE FUNCTION dat__delfunc(_name text) returns void AS $$
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
        exception when others then
            -- do nothing, EXEC above returns an exception if it does not
          -- find existing function
        END $$ LANGUAGE plpgsql;
      `
	_, err := conn.SQL(sql).Exec()
	return err
}

// Create creates database.
func (pg *PostgresAdapter) Create(superConn runner.Connection) error {
	options := pg.options
	connection := pg.options.Connection

	expressions := []*dat.Expression{
		// drop any existing connections which is helpful
		dat.Expr(`
			select pg_terminate_backend(pid)
			from pg_stat_activity
			where datname=$1
				and pid <> pg_backend_pid();
			`,
			options.Connection.Database,
		),

		dat.Expr(
			`drop database if exists $1;`, dat.UnsafeString(connection.Database),
		),

		dat.Expr(
			`drop user if exists $1;`,
			dat.UnsafeString(connection.User),
		),

		dat.Expr(
			`create user $1 password $2 SUPERUSER CREATEROLE;`,
			dat.UnsafeString(connection.User),
			connection.Password,
		),

		dat.Expr(
			`create database $1 owner $2;`,
			dat.UnsafeString(connection.Database),
			dat.UnsafeString(connection.User),
		),
	}

	_, err := superConn.ExecMulti(expressions...)
	return err
}

// Drop drops database with option to force which means drop all
// connections.
func (pg *PostgresAdapter) Drop(superConn runner.Connection) error {
	connection := pg.options.Connection

	expressions := []*dat.Expression{
		// drop any existing connections which is helpful
		dat.Expr(
			`drop database if exists $1;`,
			dat.UnsafeString(connection.Database),
		),

		dat.Expr(
			`drop user if exists $1;`,
			dat.UnsafeString(connection.User),
		),
	}

	_, err := superConn.ExecMulti(expressions...)
	return err
}

// Dump dumps a database into directory.
func (PostgresAdapter) Dump(dir string) error {
	return nil
}

// Exec executes ad-hoc SQL.
func (pg *PostgresAdapter) Exec(conn runner.Connection, sql string) error {
	_, err := conn.SQL(sql).Exec()
	return err
}

// GetAllMigrations gets migrations from database.
func (pg *PostgresAdapter) GetAllMigrations(conn runner.Connection) ([]*Migration, error) {
	sql := `
      select name, up_script, down_script, no_tx_script, created_at
      from dat__migrations
	  order by created_at;
	`

	var migrations []*Migration
	err := conn.SQL(sql).QueryStructs(&migrations)
	return migrations, err
}

// GetLastMigration gets migrations from database.
func (pg *PostgresAdapter) GetLastMigration(conn runner.Connection) (*Migration, error) {
	sql := `
      select name, up_script, down_script, no_tx_script, created_at
      from dat__migrations
	  order by created_at desc
	  limit 1;
	`

	var migration Migration
	err := conn.SQL(sql).QueryStruct(&migration)
	return &migration, err
}

// AddMigration adds a migration.
func (pg *PostgresAdapter) AddMigration(conn runner.Connection, migration *Migration) error {
	sql := `
		insert into dat__migrations (name, up_script, down_script, no_tx_script)
		values ($1, $2, $3, $4)
	`
	_, err := conn.
		SQL(sql, migration.Name, migration.UpScript, migration.DownScript, migration.NoTransactionScript).
		Exec()
	return err
}

// DeleteMigration deletes a migration from DB
func (pg *PostgresAdapter) DeleteMigration(conn runner.Connection, name string) error {
	sql := `delete from dat__migrations where name=$1`
	_, err := conn.SQL(sql, name).Exec()
	return err
}

// Redo redoes the last migration.
func (pg *PostgresAdapter) Redo(conn runner.Connection, migration *Migration) error {
	lastMigration, err := pg.GetLastMigration(conn)
	if err != nil {
		return err
	}
	// execute down migration
	// TODO what if you're redoing something that is an error
	_, err = conn.SQL(lastMigration.DownScript).Exec()
	if err != nil {
		return err
	}

	// execute new migration
	return pg.AddMigration(conn, migration)
}

// Restore restores a database from directory.
func (PostgresAdapter) Restore(dir string) error {
	return nil
}

// UpsertSproc implements Adapter method.
func (PostgresAdapter) UpsertSproc(body string) error {
	return nil
}

// ParseSprocName parses the sproc name from a body.
func (PostgresAdapter) parseSprocName(body string) (string, error) {
	return "", nil
}

var reBatchSeparator = regexp.MustCompile(`^GO\n`)

// Executes a script which may have a batch separator (default is GO)
func execScript(conn runner.Connection, script string) error {
	statements := reBatchSeparator.Split(script, -1)
	if len(statements) == 0 {
		return nil
	}

	for _, statement := range statements {
		if statement == "" {
			continue
		}

		_, err := conn.SQL(statement).Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

// runUpScripts run a migration's notx and up scripts
func runUpScripts(options *AppOptions, conn runner.Connection, migration *Migration) error {
	noTxFilename := scriptFilename(options, migration, "notx.sql")
	if _, err := os.Stat(noTxFilename); err == nil {
		// notx is an optional script
		script, err := readFileText(noTxFilename)
		if err != nil {
			return err
		}

		err = execScript(conn, script)
		if err != nil {
			return err
		}

		migration.NoTransactionScript = script
		// path/to/whatever does not exist
	}

	upScript, err := readFileText(scriptFilename(options, migration, "up.sql"))
	if err != nil {
		return err
	}
	migration.UpScript = upScript

	downScript, err := readFileText(scriptFilename(options, migration, "down.sql"))
	if err != nil {
		return err
	}
	migration.DownScript = downScript

	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.AutoRollback()

	err = execScript(tx, upScript)
	if err != nil {
		return err
	}

	q := `
		insert into dat__migrations (name, up_script, down_script, no_tx_script)
		values ($1, $2, $3, $4);
	`

	_, err = tx.SQL(
		q,
		migration.Name,
		migration.UpScript,
		migration.DownScript,
		migration.NoTransactionScript,
	).Exec()
	if err != nil {
		return err
	}

	tx.Commit()
	fmt.Println("returning")
	return nil
}
