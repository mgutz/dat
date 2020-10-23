/**
 * dat is a migration tool
 */
package main

import (
	"fmt"
	"os"

	"github.com/mgutz/logxi"
)

const version = "4.0.0-alpha.1"

func main() {
	// disable logxi logger
	logxi.Suppress(true)

	args, err := parseArgs()
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}

	err = run(args)
	if err != nil {
		logger.Error(fmt.Sprintf("\nERR %s\n", err))
		os.Exit(1)
	}
}

// func usage() string {
// 	const text = `
// dat %s - migration tool

// Usage: dat [command]

// Commands:
//   cli       Runs psql on database
//   createdb  Recreates database
//   dropdb    Drops database
//   down      Migrate down
//   dump      Dumps the database to a file
//   exec      Executes sql string
//   file      Executes sql file
//   init      Initializes migrations dir
//   new       Creates a new migration
//   query     Queries database and prints JSON
//   redo      Redoes the last migration
//   restore   Restores a dump file
//   up        Runs all migrations
// `

// 	return fmt.Sprintf(text, version)
// }

func run(args *CLIArgs) error {

	ctx := &AppContext{
		Options: args,
	}

	switch {
	default:
		rootParser.WriteUsage(os.Stdout)
		return nil
	case args.CLICmd != nil:
		return commandCLI(ctx)
	case args.CreateDBCmd != nil:
		return commandCreateDB(ctx)
	case args.DownCmd != nil:
		return commandDown(ctx)
	case args.DropDBCmd != nil:
		return commandDropDB(ctx)
	case args.DumpCmd != nil:
		return commandDump(ctx)
	case args.ExecCmd != nil:
		return commandExec(ctx)
	case args.FileCmd != nil:
		return commandFile(ctx)
	case args.InitCmd != nil:
		return commandInit(ctx)
	case args.HistoryCmd != nil:
		return commandHistory(ctx)
	case args.NewCmd != nil:
		return commandNew(ctx)
	case args.RedoCmd != nil:
		return commandRedo(ctx)
	case args.RestoreCmd != nil:
		return commandRestore(ctx)
	case args.UpCmd != nil:
		return commandUp(ctx)
	}
}
