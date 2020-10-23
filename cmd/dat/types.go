package main

import "github.com/mgutz/dat/dat"

// AppContext is the context we pass around instead of having globals
type AppContext struct {
	Options *CLIArgs
}

// AdapterContext is the context to pass to adapters
type AdapterContext struct {
	ConnectionOptions *CLIArgs
}

// Migration is meta for migrations.
type Migration struct {
	CreatedAt           dat.NullTime `db:"created_at"`
	DownScript          string       `db:"down_script"`
	Name                string       `db:"name"`
	NoTransactionScript string       `db:"no_tx_script"`
	UpScript            string       `db:"up_script"`
}

// String implements Stringer.
func (m *Migration) String() string {
	return m.Name
}

// Sproc is short for stored procedure
type Sproc struct {
	CRC       string       `db:"crc"`
	CreatedAt dat.NullTime `db:"created_at"`
	Name      string       `db:"name"`
	Script    string       `db:"script"`
	UpdatedAt dat.NullTime `db:"updated_at"`
}

// String implements Stringer.
func (s *Sproc) String() string {
	return s.Name
}
