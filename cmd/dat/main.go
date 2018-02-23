/**
 *
 * dat is a migration tool
 */
package main

import (
	"fmt"
	"os"

	"github.com/mgutz/logxi"
)

func main() {
	// disable dat's logxi logger
	logxi.Suppress(true)

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
dat v1.0.0.0-alpha.1 - simple migration tool

Usage: dat [command]

Commands:
  createdb  Recreates database
  dropdb    Drops database
  down      Migrate down
  dump      Dumps the database to a file
  exec      Executes sql from command line
  file      Executes sql file
  init 		Initializes migrations dir w/ dat.yaml
  new       Creates a new migration
  redo      Redoes the last migration
  restore   Restores a dump file
  up        Runs all migrations
`
}

func run(ctx *AppContext, command string) error {
	switch command {
	default:
		logger.Info(usage())
		return nil
	case "createdb":
		return commandCreateDB(ctx)
	case "down":
		return commandDown(ctx)
	case "dropdb":
		return commandDropDB(ctx)
	case "dump":
		return commandDump(ctx)
	case "exec":
		return commandExec(ctx)
	case "file":
		return commandFile(ctx)
	case "init":
		return commandInit(ctx)
	case "list":
		return commandList(ctx)
	case "new":
		return commandNew(ctx)
	case "redo":
		return commandRedo(ctx)
	case "restore":
		return commandRestore(ctx)
	case "up":
		return commandUp(ctx)
	}
}
