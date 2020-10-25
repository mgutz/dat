package main

import (
	"database/sql"
	"fmt"

	"github.com/mgutz/dat/dat"
	runner "github.com/mgutz/dat/sqlx-runner"
)

const dropAllConnectionsSQL = `
select pg_terminate_backend(pid)
from pg_stat_activity
where datname=$1
	and pid <> pg_backend_pid()
`

const allMigrationsSQL = `
select name, up_script, down_script, no_tx_script, created_at
from dat__migrations
order by created_at
`

const lastMigrationSQL = `
select name, up_script, down_script, no_tx_script, created_at
from dat__migrations
order by created_at desc
limit 1
`

const lastNMigrationSQL = `
select name, up_script, down_script, no_tx_script, created_at
from dat__migrations
order by created_at desc
limit $1
`

// NewPostgresAdapter instantiates Postgres adapter
func NewPostgresAdapter() *PostgresAdapter {
	return &PostgresAdapter{dbs: map[string]*runner.DB{}}
}

// PostgresAdapter is the adapter for PostgreSQL
type PostgresAdapter struct {
	dbs          map[string]*runner.DB
	bootstrapped bool
}

// AcquireDB creates a new database connection which implements runner.Connection
func (pg *PostgresAdapter) AcquireDB(connection *Connection) (*runner.DB, error) {
	connectionStr := pg.ConnectionString(connection)
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
	//dat.EnableInterpolation = true

	// set to check things like sessions closing.
	// Should be disabled in production/release builds.
	dat.Strict = false

	conn := runner.NewDB(db, "postgres")
	pg.dbs[connectionStr] = conn
	return conn, nil
}

// ConnectionString returns a connection string built from options.
func (pg *PostgresAdapter) ConnectionString(options *Connection) string {
	return fmt.Sprintf("dbname=%s user=%s password=%s host=%s port=%s %s", options.Database, options.User, options.Password, options.Host, options.Port, options.ExtraParams)
}

// Bootstrap bootstraps dat metadata
func (pg *PostgresAdapter) Bootstrap(ctx *AppContext, conn runner.Connection) error {
	if pg.bootstrapped {
		return nil
	}

	// check to see if there is an init sub directory which executes before
	// any dat scripts. The init/up.sql should be an idempotent
	// script. It was created to migrate data from existing migration tool
	// to dat.
	initScript := readInitScript(ctx.Args)
	if initScript != "" {
		err := execScript(conn, initScript, false)
		if err != nil {
			return err
		}
	}

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
               SELECT string_agg(format(
					  'DROP FUNCTION %s(%s);',
                      oid::regproc,
					  pg_catalog.pg_get_function_identity_arguments(oid)
				   ), E'\n')
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
	if err != nil {
		return err
	}

	pg.bootstrapped = true
	return nil
}

// Create creates a blank database. If the user and database exists, both are
// dropped.
func (pg *PostgresAdapter) Create(ctx *AppContext, superConn runner.Connection) error {
	connection := ctx.Args.Connection

	expressions := []*dat.Expression{
		// drop any existing connections which is helpful
		dat.Prep(dropAllConnectionsSQL, connection.Database),

		dat.Interp(
			`drop database if exists $1;`, dat.UnsafeString(connection.Database),
		),

		dat.Interp(
			`
			do $$begin
				if exists(select 1 from pg_roles where rolname = $1) then
					drop owned by $2 cascade;
					drop user $2;
				end if;
			end$$;
			`,
			connection.User,
			dat.UnsafeString(connection.User),
		),

		dat.Interp(
			`create user $1 password $2 SUPERUSER CREATEROLE;`,
			dat.UnsafeString(connection.User),
			connection.Password,
		),

		dat.Interp(
			`create database $1 owner $2;`,
			dat.UnsafeString(connection.Database),
			dat.UnsafeString(connection.User),
		),
	}

	_, err := superConn.ExecMulti(expressions...)
	return err
}

// ResetRole resets a role droppping anything it owns.
func (pg *PostgresAdapter) ResetRole(ctx *AppContext, superConn runner.Connection) error {
	connection := ctx.Args.Connection

	expressions := []*dat.Expression{
		// drop any existing connections
		dat.Prep(dropAllConnectionsSQL, connection.Database),

		dat.Interp(
			`drop database if exists $1;`, dat.UnsafeString(connection.Database),
		),

		dat.Interp(
			`
			do $$begin
				if exists(select 1 from pg_roles where rolname = $1) then
					drop owned by $2 cascade;
					drop user $2;
				end if;
			end$$;
			`,
			connection.User,
			dat.UnsafeString(connection.User),
		),

		dat.Interp(
			`create user $1 password $2 SUPERUSER CREATEROLE;`,
			dat.UnsafeString(connection.User),
			connection.Password,
		),
	}

	_, err := superConn.ExecMulti(expressions...)
	return err
}

// Exec executes ad-hoc SQL.
func (pg *PostgresAdapter) Exec(conn runner.Connection, sql string, args ...interface{}) error {
	_, err := conn.SQL(sql, args...).Exec()
	return err
}

// GetAllMigrations gets migrations from database.
func (pg *PostgresAdapter) GetAllMigrations(conn runner.Connection) ([]*Migration, error) {
	var migrations []*Migration
	err := conn.SQL(allMigrationsSQL).QueryStructs(&migrations)
	return migrations, err
}

// GetLastMigration gets migrations from database.
func (pg *PostgresAdapter) GetLastMigration(conn runner.Connection) (*Migration, error) {
	var migration Migration
	err := conn.SQL(lastMigrationSQL).QueryStruct(&migration)
	return &migration, err
}

// DeleteMigration deletes a migration from DB
func (pg *PostgresAdapter) DeleteMigration(conn runner.Connection, name string) error {
	sql := `delete from dat__migrations where name=$1`
	_, err := conn.SQL(sql, name).Exec()
	return err
}

// CleanDatabase removes all tables, etc form database.
func (pg *PostgresAdapter) CleanDatabase(conn runner.Connection, name string) error {
	sql := fmt.Sprintf(`
	drop schema public cascade;
	create schema public;
	grant all on schema public to %s;
	grant all on schema public to public;
	comment on schema public is 'standard public schema';
	`, name)
	_, err := conn.SQL(sql).Exec()
	return err
}
