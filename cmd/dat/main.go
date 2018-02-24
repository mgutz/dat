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
	// disable dat's logxi logger
	//logxi.Suppress(true)

	config, err := loadConfig()
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}

	options, err := parseOptions(config)
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
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
  con   	Runs psql on database
  createdb  Recreates database
  dropdb    Drops database
  down      Migrate down
  dump      Dumps the database to a file
  exec      Executes sql string
  file      Executes sql file
  init 		Initializes migrations dir
  new       Creates a new migration
  query		Queries database and prints JSON
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
	case "create", "createdb":
		return commandCreateDB(ctx)
	case "con", "console":
		return commandConsole(ctx)
	case "down":
		return commandDown(ctx)
	case "drop", "dropdb":
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
	case "query":
		return commandQuery(ctx)
	case "redo":
		return commandRedo(ctx)
	case "restore":
		return commandRestore(ctx)
	case "up":
		return commandUp(ctx)
	}
}
