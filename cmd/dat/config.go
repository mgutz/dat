/**
 *
 * dat is a migration tool
 */
package main

import (
	"errors"
	"os"
	"path/filepath"

	conf "github.com/mgutz/configpipe"
)

// ErrConfigNotFound indicates configuration file not found.
var ErrConfigNotFound = errors.New("dat.yaml not found")

// Connection are the options for building connections string.
type Connection struct {
	Database    string
	ExtraParams string
	Host        string
	User        string
	Password    string
	Port        string
}

// AppOptions are the options to connect to a database
type AppOptions struct {
	Connection      Connection
	BatchSeparator  string
	DockerContainer string
	DumpsDir        string
	MigrationsDir   string
	SprocsDir       string
	TablePrefix     string
	Vendor          string
	UnparsedArgs    []string
}

func parseOptions(config *conf.Configuration) (*AppOptions, error) {
	options := &AppOptions{
		Connection: Connection{
			Database:    config.AsString("connection.database"),
			User:        config.AsString("connection.user"),
			Password:    config.AsString("connection.password"),
			Port:        config.OrString("connection.port", "5432"),
			Host:        config.OrString("connection.host", "localhost"),
			ExtraParams: config.AsString("connection.extraParams"),
		},
		BatchSeparator:  config.OrString("batchSeparator", "GO"),
		DockerContainer: config.AsString("dockerContainer"),
		DumpsDir:        config.AsString("dumpsDir"),
		MigrationsDir:   config.OrString("dir", "migrations"),
		SprocsDir:       config.AsString("sprocsDir"),
		TablePrefix:     config.OrString("tablePrefix", "dat"),
		Vendor:          config.OrString("vendor", "postgres"),
	}

	if options.DumpsDir == "" {
		options.DumpsDir = filepath.Join(options.MigrationsDir, "_dumps")
	}

	if options.SprocsDir == "" {
		options.SprocsDir = filepath.Join(options.MigrationsDir, "sprocs")
	}

	// on an error, keep it at zero value, it is checked outside
	unparsed, err := config.StringArray("_unparsed")
	if err == nil {
		options.UnparsedArgs = unparsed
	}

	return options, nil
}

func decryptor(input map[string]interface{}) (map[string]interface{}, error) {
	// ... decrypt some values, add or remove keys
	return input, nil
}

func loadConfig() (*conf.Configuration, error) {
	envmode := os.Getenv("run_env")
	dir := os.Getenv("dir")
	// TODO need to parse "--dir dirname" and "--dir=dirname"
	if dir == "" {
		dir = "migrations"
	}

	var prodConfig conf.Filter
	if envmode == "production" {
		prodConfig = conf.YAML(&conf.File{Path: filepath.Join(dir, "dat-production.yaml")})
	}

	var datYAML conf.Filter
	filename := filepath.Join(dir, "dat.yaml")
	if _, err := os.Stat(filename); err == nil {
		datYAML = conf.YAML(&conf.File{Path: filename})
	}

	// later filters merge over earlier filters
	return conf.Process(
		// read from config.json file (if present)
		datYAML,

		// Any nil filter is noop, so this WILL NOT be processed in development mode.
		prodConfig,

		// read from environment variables that have no prefix replace "_" with "."  for JSON paths
		conf.Env("", "_"),

		// read from argv, storing any non flags in _unparsed slice
		conf.ArgvKeys("_unparsed", "_passthrough"),

		// use custom filter to decrypt encrypted values
		conf.FilterFunc(decryptor),
	)
}
