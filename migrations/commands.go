package migrations

import (
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
func Console() error {
	cmd := exec.Command("psql",
		"-U", _userOptions.User,
		"-h", _userOptions.Host,
		"-p", strconv.Itoa(_userOptions.Port),
	)
	if _userOptions.Password != "" {
		cmd.Env = append(cmd.Env, "PGPASSWORD=", _userOptions.Password)
	}
	cmd.Run()

	return nil
}

// CreateUserDB creates user database. Super user database options must be valid.
func CreateUserDB() error {
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
