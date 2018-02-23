package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/mgutz/str"

	runner "github.com/mgutz/dat/sqlx-runner"
	survey "gopkg.in/AlecAivazis/survey.v1"
)

func buildSuperOptions(options *AppOptions) (*AppOptions, error) {
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
	superOptions := AppOptions(*options)
	superOptions.Connection.User = answers.SuperUser
	superOptions.Connection.Password = answers.SuperPassword
	superOptions.Connection.Database = "postgres"

	return &superOptions, nil
}

func getAdapterAndDB(options *AppOptions) (*PostgresAdapter, *runner.DB, error) {
	adapter := NewPostgresAdapter(options)
	db, err := adapter.AcquireDB(options)
	if err != nil {
		return nil, nil, err
	}

	err = adapter.Bootstrap(db)
	return adapter, db, err
}

var reMigrationDir = regexp.MustCompile(`[0-9]+-[\w\-]+$`)

func getMigrationSubDirectories(options *AppOptions) ([]string, error) {

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
func getPartialLocalMigrations(options *AppOptions) ([]*Migration, error) {
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
func scriptFilename(options *AppOptions, migration *Migration, subFile string) string {
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
func readInitScript(options *AppOptions) string {
	path := filepath.Join(options.MigrationsDir, "_init", "up.sql")
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
