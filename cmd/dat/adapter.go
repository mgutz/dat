package main

import (
	"github.com/mgutz/dat/dat"
	runner "github.com/mgutz/dat/sqlx-runner"
)

// AdapterContext is the context to pass to adapters
type AdapterContext struct {
	ConnectionOptions *AppOptions
}

// Migration is meta for migrations.
type Migration struct {
	CreatedAt           dat.NullTime `db:"created_at"`
	DownScript          string       `db:"down_script"`
	Name                string       `db:"name"`
	NoTransactionScript string       `db:"no_tx_script"`
	UpScript            string       `db:"up_script"`
}

// Sproc is short for stored procedure
type Sproc struct {
	CRC       string       `db:"crc"`
	CreatedAt dat.NullTime `db:"created_at"`
	Name      string       `db:"name"`
	Script    string       `db:"script"`
	UpdatedAt dat.NullTime `db:"updated_at"`
}

// Adapter is the interface for applying migrations
type Adapter interface {
	// Bootstrap boostraps database with dat metadata (idempotent)
	Bootstrap(conn runner.Connection) error

	// ConnectionString creates a connection string from options.
	ConnectionString(options *AppOptions) string

	// Creates database which requires a super user.
	Create(conn runner.Connection) error

	// Drops database
	Drop(conn runner.Connection) error

	// Dumps database to specified directory
	Dump(dir string) error

	// Execute a statement
	Exec(conn runner.Connection, sql string) error

	// GetMigrations gets list of migrations from DB
	GetMigrations(conn runner.Connection) ([]*Migration, error)

	// Redo redoes the last migration by undoing it then rerunning it.
	Redo(conn runner.Connection, migration *Migration) error

	// Restore restsors DB from specified directory
	Restore(dir string) error

	// Updates or inserts sproc
	UpsertSproc(body string) error
}

/**
dat.Migrate(&Jobs{
	dat.MigrateDirectory("./migrations/111203-asdfasdf-as"),
	MoreComplicated(),
	dat.MigrateSprocsDir("./sprocs")
})
*/
