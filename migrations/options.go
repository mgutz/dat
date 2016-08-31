package migrations

import (
	"bytes"
	"strconv"
)

// DBOptions are options to connect to the database.
type DBOptions struct {
	BatchSeparator string
	DBName         string
	Host           string
	Password       string
	Port           int
	SSLMode        bool
	User           string
}

// String returns a connection string.
func (opts *DBOptions) String() string {
	// dbname=dat_test user=dat password=!test host=localhost
	var buffer bytes.Buffer

	if opts.DBName != "" {
		buffer.WriteString("dbname=")
		buffer.WriteString(opts.DBName)
		buffer.WriteString(" ")
	}

	if opts.Host != "" {
		buffer.WriteString("host=")
		buffer.WriteString(opts.Host)
		buffer.WriteString(" ")
	}

	if opts.Password != "" {
		buffer.WriteString("password=")
		buffer.WriteString(opts.Password)
		buffer.WriteString(" ")
	}

	if opts.Port != 0 {
		buffer.WriteString("port=")
		buffer.WriteString(strconv.Itoa(opts.Port))
		buffer.WriteString(" ")
	}

	buffer.WriteString("sslmode=")
	if opts.SSLMode {
		buffer.WriteString("true")
	} else {
		buffer.WriteString("false")
	}
	buffer.WriteString(" ")

	return buffer.String()
}

// NewDBOptions creates a new instance of DBOptions with default settings. This should
// be used instead of creating DBOptions directly.
func NewDBOptions() *DBOptions {
	return &DBOptions{
		Host:           "localhost",
		DBName:         "postgres",
		User:           "postgres",
		BatchSeparator: "GO",
		SSLMode:        false,
		Port:           5432,
	}
}
