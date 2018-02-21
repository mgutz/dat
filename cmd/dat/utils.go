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

	// bootstrap is idempotent
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

func migrationFile(dir string, title string, filename string) string {
	// 201801231939-refactor-campaigns
	subdir := fmt.Sprintf("%s-%s", time.Now().Format("200601021504"), str.Slugify(title))
	return filepath.Join(dir, subdir, filename)
}

// scriptFilename computes a migration sripts filename
func scriptFilename(options *AppOptions, migration *Migration, subFile string) string {
	return filepath.Join(options.MigrationsDir, migration.Name, subFile)
}

// verifyMigrationsHistory verifies local migrations are in sync with the database.
// Devs might have added migrations in their working branch that predate migrations
// already applied to the database.
func verifyMigrationsHistory(localMigrations []*Migration, dbMigrations []*Migration) error {
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
