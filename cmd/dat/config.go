/**
 *
 * dat is a migration tool
 */
package main

import (
	"os"

	conf "github.com/mgutz/configpipe"
)

// Connection are the options for building connections string.
type Connection struct {
	Database    string
	ExtraParams string
	Host        string
	User        string
	Password    string
}

// AppOptions are the options to connect to a database
type AppOptions struct {
	Connection     Connection
	BatchSeparator string
	MigrationsDir  string
	TablePrefix    string
	Vendor         string
	UnparsedArgs   []string
}

func parseOptions(config *conf.Configuration) (*AppOptions, error) {
	options := &AppOptions{
		Connection: Connection{
			Database:    config.MustString("connection.database"),
			User:        config.MustString("connection.user"),
			Password:    config.AsString("connection.password"),
			Host:        config.OrString("connection.host", "localhost"),
			ExtraParams: config.AsString("connection.extraParams"),
		},
		BatchSeparator: config.OrString("batchSeparator", "GO"),
		MigrationsDir:  config.OrString("migrationsDir", "./migrations"),
		TablePrefix:    config.OrString("tablePrefix", "dat"),
		Vendor:         config.OrString("vendor", "postgres"),
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

	var prodConfig conf.Filter
	if envmode == "production" {
		prodConfig = conf.YAML(&conf.File{Path: "dat-production.yaml"})
	}

	// later filters merge over earlier filters
	return conf.Process(
		// read from config.json file (if present)
		conf.YAML(&conf.File{Path: "dat.yaml"}),

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
