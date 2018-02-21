/**
 *
 * dat is a migration tool
 */
package main

import (
	"errors"
	"fmt"
	"os"
)

func main() {
	config, err := loadConfig()
	if err != nil {
		panic(err)
	}

	options, err := parseOptions(config)
	if err != nil {
		panic(err)
	}

	if len(options.UnparsedArgs) == 0 {
		printUsage()
		return
	}

	command := options.UnparsedArgs[0]
	err = run(command, options)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("OK")
}

func printUsage() {
	usage := `
dat v0.0.0 - migration tool

Usage: dat [COMMAND]

Commands:
  createdb  Recreates database and runs migrations
  down      Migrate down
  dump      Dumps the database
  exec      Executes sql from command line
  file      Executes sql file
  new       Creates a new migration
  redo      Redoes the last migration
  restore   Restores the database
  up        Runs all migrations
`
	fmt.Println(usage)
}

// NewAdapter creates adapter.
func NewAdapter(options *AppOptions) (*PostgresAdapter, error) {
	switch options.Vendor {
	default:
		return nil, errors.New("Unknown vendor: " + options.Vendor)
	case "postgres":
		return NewPostgresAdapter(options), nil
	}
}

func run(command string, options *AppOptions) error {
	switch command {
	default:
		printUsage()
		return nil
	case "createdb":
		return createDB(options)
	case "list":
		return list(options)
	case "new":
		return newScript(options)
	case "up":
		return up(options)
	}
}
