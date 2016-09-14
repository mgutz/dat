package migrations

import (
	"fmt"
	"os/exec"
	"strconv"

	"gopkg.in/mgutz/dat.v3/dat"
	runner "gopkg.in/mgutz/dat.v3/sqlx-runner"
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
	CreatedAt dat.NullTime `db:"created_at"`
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

// Add adds a migration to the database.
func Add(migration *Migration) error {
	q := `
INSERT INTO $1__migrations(version, up, down)
VALUES($2, $3, $4)
	`
	mustUserDB().SQL(q, _namespace, migration.Version, migration.Up, migration.Down)
	return nil
}

// AddMigration adds a migration to the database.
func AddMigration(conn runner.Connection, version string, upScript string, downScript string) error {
	q := `
INSERT INTO $1__migrations (version, up, down)
VALUES ($2, $3, $4)
	`
	_, err := conn.SQL(q, _namespace, version, upScript, downScript).Exec()
	return err
}

// Bootstrap bootstraps user's database with metadata tables
func (pg *Postgres) Bootstrap(conn runner.Connection, namespace string) error {
	q := `
CREATE TABLE IF NOT EXISTS $1__migrations (
  version varchar(256) not null primary key,
  up text,
  down text,
  created_at timestamp default current_timestamp
);

CREATE TABLE IF NOT EXISTS $1__sprocs (
  name text primary key,
  crc text not null,
  created_at timestamp default current_timestamp
);

CREATE OR REPLACE FUNCTION $1__delfunc(_name text) RETURNS VOID AS $$
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
	_, err := conn.SQL(q, _namespace).Exec()
	return err
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

func unquoted(name string) dat.UnsafeString {
	return dat.UnsafeString(name)
}

// CreateUserDB creates user database. Super user database options must be valid.
func CreateUserDB(conn runner.Connection) error {
	script := `
DROP USER IF EXISTS $2;
GO
CREATE USER $2 PASSWORD $3 SUPERUSER CREATEROLE;
GO
CREATE DATABASE $1 OWNER $2;
`
	err := DropUserDB(conn)
	if err != nil {
		return err
	}
	return conn.ExecScript(
		script,
		unquoted(_userOptions.DBName),
		unquoted(_userOptions.User),
		_userOptions.Password,
	)
}

// DropUserDB drops user database.
func DropUserDB(conn runner.Connection) error {
	script := `
# kill all connections first
# NOTE: pid is procpid in PostgreSQL < 9.2
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE datname=$1
	and pid <> pg_backend_pid()
GO
DROP DATABASE IF EXISTS $2;
GO
	`
	return conn.ExecScript(
		script,
		_userOptions.DBName,
		unquoted(_userOptions.DBName),
	)
}

// FetchMigrations fetches all migrations from the database sorted by created_at desc
func FetchMigrations(conn runner.Connection) ([]*Migration, error) {
	var migrations []*Migration
	q := `
SELECT *
FROM $1__migrations
ORDER BY VERSION DESC;
	`
	err := conn.SQL(q, _namespace).QueryStructs(&migrations)
	return migrations, err
}

// GetSproc gets registered stored procedure metadata.
func GetSproc(conn runner.Connection, name string) ([]*Sproc, error) {
	q := `
SELECT name, crc
FROM $1__sprocs
WHERE name = $2

UNION ALL

SELECT  proname AS name, '0' AS crc
FROM    pg_catalog.pg_namespace n
JOIN    pg_catalog.pg_proc p ON pronamespace = n.oid
WHERE	nspname = 'public'
	AND proname NOT IN (
		SELECT name
		FROM $1__sprocs
		WHERE name = $2
	)
`

	var sprocs []*Sproc
	err := conn.SQL(q, _namespace, name).QueryStructs(&sprocs)
	return sprocs, err
}

// Last retrieves the last migration
func Last(conn runner.Connection, version string) (*Migration, error) {
	q := `
SELECT *
FROM $1__migrations
ORDER BY VERSION DESC
LIMIT 1;
	`
	var migration *Migration
	err := conn.SQL(q, _namespace).QueryStruct(&migration)
	return migration, err
}

// RegisterSproc registers a stored procedure into dat__sprocs.
func RegisterSproc(conn runner.Connection, name string, crc string) error {
	q := `
DELETE FROM $1__sprocs WHERE name = $2;
INSERT INTO $1__sprocs WHERE (name, crc) VALUES ($2, $3);
	`
	_, err := conn.SQL(q, _namespace, name, crc).Exec()
	return err
}

// Remove removes a migration from the database
func Remove(conn runner.Connection, version string) error {
	q := `
DELETE
FROM $1__migrations
WHERE version = $2
	`
	_, err := conn.SQL(q, _namespace, version).Exec()
	return err
}
