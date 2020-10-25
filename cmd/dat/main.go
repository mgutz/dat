/**
 * dat is a migration tool
 */
package main

import (
	"fmt"
	"os"
)

const version = "4.0.0-alpha.1"

func main() {
	// disable logxi logger
	//logxi.Suppress(true)

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

func run(args *CLIArgs) error {

	ctx := &AppContext{
		Args: args,
	}

	switch {
	default:
		rootParser.WriteUsage(os.Stdout)
		return nil
	case args.CleanCmd != nil:
		return commandClean(ctx)
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
