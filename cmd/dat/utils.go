package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/mgutz/str"

	"github.com/mgutz/dat/dat"
	runner "github.com/mgutz/dat/sqlx-runner"
	survey "gopkg.in/AlecAivazis/survey.v1"
)

func buildSuperOptions(options *CLIArgs) (*CLIArgs, error) {
	questions := []*survey.Question{
		{
			Name:   "SuperUser",
			Prompt: &survey.Input{Message: "admin (postgres)"},
		},
		{
			Name:     "SuperPassword",
			Prompt:   &survey.Password{Message: "password"},
			Validate: survey.Required,
		},
	}

	var answers struct {
		SuperUser     string
		SuperPassword string
	}

	err := survey.Ask(questions, &answers)
	if err != nil {
		return nil, err
	}

	if answers.SuperUser == "" {
		answers.SuperUser = "postgres"
	}

	// use conversion to clone, then set admin credentials
	superOptions := CLIArgs(*options)
	superOptions.Connection.User = answers.SuperUser
	superOptions.Connection.Password = answers.SuperPassword
	superOptions.Connection.Database = "postgres"

	return &superOptions, nil
}

func getAdapterAndDB(ctx *AppContext) (*PostgresAdapter, *runner.DB, error) {
	adapter := NewPostgresAdapter()
	db, err := adapter.AcquireDB(&ctx.Options.Connection)
	if err != nil {
		return nil, nil, err
	}

	err = adapter.Bootstrap(ctx, db)
	return adapter, db, err
}

var reMigrationDir = regexp.MustCompile(`[0-9]+-[\w\-]+$`)

func getMigrationSubDirectories(options *CLIArgs) ([]string, error) {

	var files []string
	err := filepath.Walk(options.MigrationsDir+"/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && reMigrationDir.MatchString(path) {
			files = append(files, info.Name())
		}
		return nil
	})

	// sort in DESC order
	//sort.Sort(sort.StringSlice(files))
	return files, err
}

var reSQLFile = regexp.MustCompile(`[\w\-]+.sql$`)

func getSprocFiles(sprocsDir string) ([]string, error) {
	var files []string
	err := filepath.Walk(sprocsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && reSQLFile.MatchString(info.Name()) {
			files = append(files, info.Name())
		}
		return nil
	})

	// sort in DESC order
	//sort.Sort(sort.StringSlice(files))
	return files, err
}

func getDumpFiles(ctx *AppContext) ([]string, error) {
	dir := ctx.Options.DumpsDir

	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			files = append(files, path)
		}

		return nil
	})

	// sort in DESC order
	//sort.Sort(sort.StringSlice(files))
	return files, err
}

// gets local migrations names only, it does not fill in DownScript, UpScript and
// NoTransactionScript
func getPartialLocalMigrations(options *CLIArgs) ([]*Migration, error) {
	dirs, err := getMigrationSubDirectories(options)
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
			return []*Migration{}, nil
		}

		return nil, err
	}

	meta := make([]*Migration, len(dirs))
	for i, dir := range dirs {
		meta[i] = &Migration{Name: dir}
	}

	return meta, nil
}

func timestampedName(name string) string {
	return fmt.Sprintf("%s-%s", time.Now().Format("200601021504"), str.Slugify(name))
}

func migrationFile(dir string, title string, filename string) string {
	// 201801231939-refactor-campaigns
	subdir := timestampedName(title)
	return filepath.Join(dir, subdir, filename)
}

// scriptFilename computes a migration sripts filename
func scriptFilename(options *CLIArgs, migration *Migration, subFile string) string {
	return filepath.Join(options.MigrationsDir, migration.Name, subFile)
}

func migrationFindIndexOf(migrations []*Migration, name string) int {
	if len(migrations) > 0 {
		for i, migration := range migrations {
			if migration.Name == name {
				return i
			}
		}
	}

	return -1
}

// verifyMigrationsHistory verifies local migrations are in sync with the database.
// Devs might have added migrations in their working branch that predate migrations
// already applied to the database.
//
// assumes localMigrations and dbMigrations are sorted in ASC order
func verifyMigrationsHistory(ctx *AppContext, localMigrations []*Migration, dbMigrations []*Migration) error {
	if len(dbMigrations) == 0 {
		return nil
	}

	inError := false

	// print any migration in DB that doesn't exist locally
	for _, migration := range dbMigrations {
		idx := migrationFindIndexOf(localMigrations, migration.Name)
		if idx == -1 {
			logger.Info("Migration %s was migrated in database but does not exist in local migrations.\n", migration.Name)
			inError = true
		}
	}

	// log any directory which has not been executed but is younger than last migration in DB
	lastMigration := dbMigrations[len(dbMigrations)-1]
	for _, localMigration := range localMigrations {
		localName := localMigration.Name
		if localName < lastMigration.Name {
			idx := migrationFindIndexOf(dbMigrations, localName)
			if idx == -1 {
				logger.Info("Migration %s will not be migrated. Its timestamp is younger than last migration %s\n", localName, lastMigration.Name)
				inError = true
			}
		}
	}

	if inError {
		return fmt.Errorf("Local migrations are out of sync with %s database, rename as needed", ctx.Options.Connection.Database)
	}

	return nil
}

func readFileText(filename string) (string, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// writeFileAll write text to a file. Subdirectories are created recursively like
// `mkdirp`.
func writeFileAll(filename string, b []byte) error {
	dir := filepath.Dir(filename)
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, b, 0644)
}

// readInitScript reads migrations/_init/up.sql. If any error occurs, it returns
// an empty string.
func readInitScript(options *CLIArgs) string {
	path := filepath.Join(options.InitDir, "up.sql")
	s, _ := readFileText(path)
	return s
}

func askOption(prompt string, options []string) (string, error) {
	questions := []*survey.Question{
		{
			Name: "Option",
			Prompt: &survey.Select{
				Message: prompt,
				Options: options,
			},
		},
	}

	var answers struct {
		Option string
	}
	err := survey.Ask(questions, &answers)
	return answers.Option, err
}

var reBatchSeparator = regexp.MustCompile(`(?m)^GO\n`)

// Executes a script which may have a batch separator (default is GO). Filename
// is used for error reporting
func execScript(conn runner.Connection, script string, warnEmpty bool) error {
	statements := reBatchSeparator.Split(script, -1)
	if len(statements) == 0 {
		return nil
	}

	for _, statement := range statements {
		if statement == "" {
			continue
		}

		if str.IsEmpty(statement) {
			return nil
		}

		_, err := conn.SQL(statement).Exec()
		if err != nil {
			if strings.Contains(err.Error(), "empty statement") {
				if warnEmpty {
					logger.Info("\nEmpty query:\n")
					logger.Info("%s\n", statement)
				}
				continue
			}

			logger.Error("\n%s\n", sprintPQError(statement, err))
			return err
		}
	}

	return nil
}

func execFile(ctx *AppContext, conn runner.Connection, filename string) (string, error) {
	logger.Info("%s ... ", filename)
	script, err := readFileText(filename)
	if err != nil {
		logger.Info("\n")
		return "", err
	}

	err = execScript(conn, script, false)
	if err != nil {
		return "", err
	}
	logger.Info("OK\n")
	return script, nil
}

// runUpScripts run a migration's notx and up scripts
func runUpScripts(ctx *AppContext, conn runner.Connection, migration *Migration) error {
	// notx.sql is not required
	noTxFilename := scriptFilename(ctx.Options, migration, "notx.sql")
	if _, err := os.Stat(noTxFilename); err == nil {
		// notx is an optional script
		script, err := execFile(ctx, conn, noTxFilename)
		if err != nil {
			return err
		}

		migration.NoTransactionScript = script
		// path/to/whatever does not exist
	}

	// down.sql is not required
	downFilename := scriptFilename(ctx.Options, migration, "down.sql")
	if _, err := os.Stat(downFilename); err == nil {
		downScript, err := readFileText(downFilename)
		if err != nil {
			return err
		}
		migration.DownScript = downScript
	}

	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.AutoRollback()

	upScript, err := execFile(ctx, conn, scriptFilename(ctx.Options, migration, "up.sql"))
	if err != nil {
		return err
	}
	migration.UpScript = upScript

	q := `
		insert into dat__migrations (name, up_script, down_script, no_tx_script)
		values ($1, $2, $3, $4);
	`

	_, err = tx.SQL(
		q,
		migration.Name,
		migration.UpScript,
		migration.DownScript,
		migration.NoTransactionScript,
	).Exec()
	if err != nil {
		return err
	}

	tx.Commit()
	return nil
}

// Drop drops database with option to force which means drop all connections.
func (pg *PostgresAdapter) Drop(ctx *AppContext, superConn runner.Connection) error {
	connection := ctx.Options.Connection

	expressions := []*dat.Expression{
		// drop any existing connections which is helpful
		dat.Interp(`
				select pg_terminate_backend(pid)
				from pg_stat_activity
				where datname=$1
					and pid <> pg_backend_pid();
			`,
			connection.Database,
		),

		dat.Interp(
			`drop database if exists $1;`,
			dat.UnsafeString(connection.Database),
		),

		dat.Interp(
			`drop user if exists $1;`,
			dat.UnsafeString(connection.User),
		),
	}
	_, err := superConn.ExecMulti(expressions...)
	return err
}

func sprintPQError(script string, err error) string {
	if err == nil {
		return ""
	}

	if e, ok := err.(*pq.Error); ok {
		//. TODO need to show line number, column on syntax errors
		// fmt.Println("Code", e.Code)
		// fmt.Println("Column", e.Column)
		// fmt.Println("Line", e.Line)
		// fmt.Println("Position", e.Position)
		// fmt.Println("Message", e.Message)
		// fmt.Println("Detail", e.Detail)
		// fmt.Println("Hint", e.Hint)
		// fmt.Println("Severity", e.Severity)

		if e.Position != "" {
			line, col, _ := calcLineCol(script, e.Position)
			return fmt.Sprintf("[%s code=%s]\n%s at line=%d col=%d\n", e.Severity, e.Code, e.Message, line, col)
		}
		return fmt.Sprintf("[%s code=%s]\n%s", e.Severity, e.Code, e.Message)
	}

	return ""
}

func calcLineCol(script string, pos string) (int, int, error) {
	position, err := strconv.Atoi(pos)
	if err != nil {
		return 0, 0, err
	}

	position++ // starts at column 1 in text
	line := 1
	column := 0
	max := len(script)

	i := 0
	for i < max && i < position {
		ch := script[i]
		// Windows
		if ch == '\r' {
			if i+1 < max && script[i+1] == '\n' {
				i++
			}
			if i < max-1 {
				line++
				column = 0
			}
		} else if ch == '\n' {
			if i < max-1 {
				line++
				column = 0
			}
		} else {
			column++
		}
		i++
	}

	return line, column, nil
}

const datYAMLExample = `
connection:
  host: localhost
  database: dev_db
  user: grace
  password: "!development"
  extraParams: "sslmode=disable"

# if using docker set to container name to use pg_dump and pg_restore inside container
# dockerContainer: postgres-svc
`
