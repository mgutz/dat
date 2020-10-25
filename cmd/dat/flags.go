/**
 *
 * dat is a migration tool
 */
package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alexflint/go-arg"
	"github.com/joho/godotenv"
)

// ErrConfigNotFound indicates configuration file not found.
var ErrConfigNotFound = errors.New("dat.yaml not found")

// Connection are the options for building connections string.
type Connection struct {
	Database    string `arg:"-d,--database,required,env:dat_database" placeholder:"DB"`
	ExtraParams string `arg:"-e,--extraParams,env:dat_extraParams" help:"Extra  connection params" placeholder:"QS"`
	Host        string `arg:"-h,--host,env:dat_host" default:"localhost" description:"Host"`
	Password    string `arg:"--password,env:dat_password"`
	Port        string `arg:"-p,--port,env:dat_port" default:"5432"`
	User        string `arg:"-u,--user,env:dat_user" help:"User"`
}

// CLIArgs are the options to connect to a database
type CLIArgs struct {
	CleanCmd *struct {
		Owner string `arg:"positional" help:"Owner"`
	} `arg:"subcommand:clean" help:"Remove data and tables from database"`

	CLICmd *struct {
	} `arg:"subcommand:cli" help:"Run psql"`

	CreateDBCmd *struct {
	} `arg:"subcommand:createdb" help:"(Re)Creates the database"`

	DownCmd *struct {
		Count int `arg:"positional" default:"1"`
	} `arg:"subcommand:down" help:"Migrate down"`

	DropDBCmd *struct {
		Msg string `arg:"positional"`
	} `arg:"subcommand:dropdb" help:"Drop database"`

	DumpCmd *struct {
		Filename string `arg:"positional"`
	} `arg:"subcommand:dump" help:"Dump database to FILENAME"`

	ExecCmd *struct {
		Query        string `arg:"positional" help:"SQL Query"`
		OutputFormat string `arg:"-o,--outputFormat" help:"Output format (json)"`
	} `arg:"subcommand:exec" help:"Exec a query"`

	FileCmd *struct {
		Filename string `arg:"positional"`
	} `arg:"subcommand:file" help:"Execute a SQL file"`

	InitCmd *struct {
		Msg string `arg:"positional"`
	} `arg:"subcommand:init" help:"Initializes dir with migrations subdir"`

	HistoryCmd *struct {
		Msg string `arg:"positional"`
	} `arg:"subcommand:history" help:"Displays history"`

	NewCmd *struct {
		Title string `arg:"positional" help:"kebab-case-title"`
	} `arg:"subcommand:new" help:"Creates a timestamped migration"`

	RedoCmd *struct {
		Msg string `arg:"positional"`
	} `arg:"subcommand:redo" help:"Reruns last migration"`

	RestoreCmd *struct {
		Filename string `arg:"positional"`
	} `arg:"subcommand:restore" help:"Restore database dmp"`

	UpCmd *struct {
		Msg string `arg:"positional"`
	} `arg:"subcommand:up" help:"migrate up"`

	Connection
	BatchSeparator  string `arg:"--batchSeparator" default:"GO" help:"Statement separator" env:"dat_batchSeparator" placeholder:"SEP"`
	DockerContainer string `arg:"-"`
	DumpsDir        string `arg:"--dumpsDir,env:dat_dumpsDir" placeholder:"DIR"`
	InitDir         string `arg:"--initDir,env:dat_initDir" placeholder:"DIR"`
	MigrationsDir   string `arg:"--dir,env:dat_dir" default:"migrations" help:"Migrations directory"  placeholder:"DIR"`
	SprocsDir       string `arg:"--sprocsDir,env:dat_sprocsDir" help:"Stored procedures"  placeholder:"DIR"`
}

func loadEnvFiles() error {
	err := godotenv.Load()
	if err != nil {
		if os.IsNotExist(err) {
			// do nothing, it's not error if .env file does not exist
			return nil
		}

		return fmt.Errorf("Cannot load .env file: %w", err)
	}
	return nil
}

var (
	rootParser *arg.Parser
)

func parseArgs() (*CLIArgs, error) {
	err := loadEnvFiles()
	if err != nil {
		return nil, err
	}

	var args CLIArgs
	rootParser = arg.MustParse(&args)

	if args.DumpsDir == "" {
		args.DumpsDir = filepath.Join(args.MigrationsDir, "dumps")
	}

	if args.InitDir == "" {
		args.InitDir = filepath.Join(args.MigrationsDir, "init")
	}

	if args.SprocsDir == "" {
		args.SprocsDir = filepath.Join(args.MigrationsDir, "sprocs")
	}

	// b, _ := json.MarshalIndent(args, "", "    ")
	// fmt.Println("args=", string(b))

	return &args, nil
}
