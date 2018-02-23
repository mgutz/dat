/**
 *
 * dat is a migration tool
 */
package main

import (
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
		logger.Error(usage())
		os.Exit(1)
	}

	ctx := &AppContext{
		Options: options,
	}

	command := options.UnparsedArgs[0]
	err = run(ctx, command)
	if err != nil {
		logger.Error(fmt.Sprintf("\nERR %s\n", err))
		os.Exit(1)
	}
}

func usage() string {
	return `
dat v0.0.0 - simple migration tool

Usage: dat [command]

Commands:
  createdb  Recreates database and runs migrations
  dropdb    Drops database
  //down      Migrate down
  dump      Dumps the database
  exec      Executes sql from command line
  file      Executes sql file
  new       Creates a new migration
  redo      Redoes the last migration
  restore   Restores the database
  up        Runs all migrations
`
}

func run(ctx *AppContext, command string) error {
	switch command {
	default:
		logger.Info(usage())
		return nil
	case "createdb":
		return createDB(ctx)
	case "dropdb":
		return dropDB(ctx)
	case "dump":
		return dump(ctx)
	case "exec":
		return execUserString(ctx)
	case "list":
		return list(ctx)
	case "new":
		return newScript(ctx)
	case "restore":
		return restore(ctx)
	case "up":
		return up(ctx)
	}
}
