package main

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/mgutz/str"

	runner "github.com/mgutz/dat/sqlx-runner"
	"github.com/mgutz/jo"
	"github.com/olekukonko/tablewriter"
)

func commandCreateDB(ctx *AppContext) error {
	superOptions, err := buildSuperOptions(ctx.Options)
	if err != nil {
		return err
	}

	adapter := NewPostgresAdapter()
	superDB, err := adapter.AcquireDB(&superOptions.Connection)
	if err != nil {
		return err
	}
	defer superDB.Close()

	err = adapter.Create(ctx, superDB)
	if err != nil {
		return err
	}

	logger.Info("Created %s", adapter.ConnectionString(&ctx.Options.Connection))
	return nil
}

// commandConsole runs psql console
func commandConsole(ctx *AppContext) error {
	connection := ctx.Options.Connection
	var args []string
	var exe string

	container := ctx.Options.DockerContainer
	passwordEnv := "PGPASSWORD=" + connection.Password
	if container == "" {
		exe = "psql"
	} else {
		exe = "docker"
		args = []string{
			"exec",
			"-it",
			"-e",
			passwordEnv,
			container,
			"psql",
		}
	}

	args = append(args,
		"--dbname="+connection.Database,
		"--host="+connection.Host,
		"--username="+connection.User,
		"--port="+connection.Port,
	)

	logger.Info("exe=%s args=%v\n", exe, args)
	cmd := exec.Command(exe, args...)
	cmd.Env = append(os.Environ(), passwordEnv)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Down undoes 1 or more migrations based on argument. Default is 1 down.
func commandDown(ctx *AppContext) error {
	var count int
	if len(ctx.Options.UnparsedArgs) == 1 {
		count = 1
	} else {
		arg := ctx.Options.UnparsedArgs[1]
		n, err := strconv.Atoi(arg)
		if err != nil {
			return errors.New("Usage: dat down [n]")
		}
		count = n
		if count == 0 {
			count = 1
		}
	}

	_, db, err := getAdapterAndDB(ctx)
	if err != nil {
		return err
	}
	defer db.Close()

	return migrateDown(db, count)
}

// dropDB drops the database
func commandDropDB(ctx *AppContext) error {
	adapter := NewPostgresAdapter()

	// drop the database

	superOptions, err := buildSuperOptions(ctx.Options)
	if err != nil {
		return err
	}

	db, err := adapter.AcquireDB(&superOptions.Connection)
	if err != nil {
		return err
	}
	defer db.Close()

	err = adapter.Drop(ctx, db)
	if err != nil {
		return err
	}
	return nil
}

// dump dumps a database to a file for use by restore
func commandDump(ctx *AppContext) error {
	superOptions, err := buildSuperOptions(ctx.Options)
	if err != nil {
		return err
	}

	newFilename := getCommandArg1(ctx)
	if newFilename == "" {
		newFilename = timestampedName("dump")
	}
	dumpsDir := ctx.Options.DumpsDir
	destination := filepath.Join(dumpsDir, newFilename)
	err = os.MkdirAll(dumpsDir, os.ModePerm)
	if err != nil {
		return err
	}

	var args []string
	var exe string

	passwordEnv := "PGPASSWORD=" + superOptions.Connection.Password
	if ctx.Options.DockerContainer == "" {
		exe = "pg_dump"
	} else {
		exe = "docker"
		args = []string{
			"exec",
			"-e",
			passwordEnv,
			ctx.Options.DockerContainer,
			"pg_dump",
		}
	}

	args = append(args,
		"--dbname="+ctx.Options.Connection.Database,
		"--host="+ctx.Options.Connection.Host,
		"--username="+superOptions.Connection.User,
		"--port="+ctx.Options.Connection.Port,
		"-Fc",
	)

	var buf bytes.Buffer

	logger.Info("exe=%s args=%v\n", exe, args)
	cmd := exec.Command(exe, args...)
	cmd.Env = append(os.Environ(), passwordEnv)
	cmd.Stdout = &buf
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		logger.Info(buf.String())
		return err
	}

	err = writeFileAll(destination, buf.Bytes())
	if err != nil {
		logger.Error("Error writing dump file %s", destination)
		return err
	}
	logger.Info("dumped %s", destination)
	return err
}

func commandExec(ctx *AppContext) error {
	q := getCommandArg1(ctx)
	if q == "" {
		return errors.New(`Usage: dat exec [sql_string]`)
	}

	adapter := NewPostgresAdapter()
	db, err := adapter.AcquireDB(&ctx.Options.Connection)
	if err != nil {
		return err
	}
	defer db.Close()

	return execScript(db, q, false)
}

func commandQuery(ctx *AppContext) error {
	q := getCommandArg1(ctx)
	if q == "" {
		return errors.New(`Usage: dat exec [sql_string]`)
	}

	adapter := NewPostgresAdapter()
	db, err := adapter.AcquireDB(&ctx.Options.Connection)
	if err != nil {
		return err
	}
	defer db.Close()

	var obj jo.Object
	err = db.SQL(q).QueryObject(&obj)
	if err != nil {
		return err
	}

	logger.Info(obj.Prettify())
	return nil
}

func commandFile(ctx *AppContext) error {
	filename := getCommandArg1(ctx)
	if filename == "" {
		return errors.New(`Usage: dat file [filename]`)
	}

	adapter := NewPostgresAdapter()
	db, err := adapter.AcquireDB(&ctx.Options.Connection)
	if err != nil {
		return err
	}
	defer db.Close()

	sql, err := readFileText(filename)
	if err != nil {
		return err
	}

	_, err = db.SQL(sql).Exec()
	if err != nil {
		logger.Error(sprintPQError(sql, err))
		return err
	}

	logger.Info("OK %s\n", filename)
	return nil
}

func commandInit(ctx *AppContext) error {
	filename := filepath.Join(ctx.Options.MigrationsDir, "dat.yaml")
	if _, err := os.Stat(filename); err == nil {
		return errors.New("File exists " + filename)
	}

	err := writeFileAll(filename, []byte(datYAMLExample))
	if err != nil {
		return err
	}

	logger.Info("Edit %s", filename)

	err = os.MkdirAll(filepath.Join(ctx.Options.MigrationsDir, "sprocs"), os.ModePerm)
	return err
}

func commandList(ctx *AppContext) error {
	adapter, db, err := getAdapterAndDB(ctx)
	if err != nil {
		return err
	}

	migrations, err := adapter.GetAllMigrations(db)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Info("A No migrations found.")
			return nil
		}
		return err
	}

	if len(migrations) == 0 {
		logger.Info("No migrations found")
		return nil
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Name", "Created"})
	table.SetBorder(false)

	for _, migration := range migrations {
		dateStr := fmt.Sprintf("%s", migration.CreatedAt.Time.Format("2006-01-02 15:04 PM Mon"))
		row := []string{migration.Name, dateStr}
		table.Append(row)
	}

	table.Render()
	return nil
}

// newScripts creates a new script directory with blank `up.sql` and `down.sql`
func commandNew(ctx *AppContext) error {
	title := getCommandArg1(ctx)
	if title == "" {
		return errors.New("Usage: dat new TITLE")
	}

	upFilename := migrationFile(ctx.Options.MigrationsDir, title, "up.sql")
	downFilename := migrationFile(ctx.Options.MigrationsDir, title, "down.sql")

	err := writeFileAll(upFilename, []byte("-- up.sql"))
	if err != nil {
		return err
	}

	err = writeFileAll(downFilename, []byte("-- down.sql"))
	if err != nil {
		return err
	}

	logger.Info("Created %s\n", filepath.Dir(upFilename))

	return nil
}

func commandRedo(ctx *AppContext) error {
	adapter, db, err := getAdapterAndDB(ctx)
	if err != nil {
		return err
	}
	defer db.Close()

	last, err := adapter.GetLastMigration(db)
	if err != nil {
		return err
	}

	err = migrateDown(db, 1)

	return runUpScripts(ctx, db, last)
}

func migrateDown(conn runner.Connection, count int) error {
	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.AutoRollback()

	var migrations []*Migration
	err = tx.SQL(lastNMigrationSQL, count).QueryStructs(&migrations)
	if err != nil {
		return err
	}

	// run the down script then delete it from migrations table
	for _, migration := range migrations {
		logger.Info("DB down script %s ... ", migration.Name)
		if !str.IsEmpty(migration.DownScript) {
			err = execScript(tx, migration.DownScript, true)
			if err != nil {
				logger.Info("\n")
				return err
			}
		}

		_, err = tx.SQL(`delete from dat__migrations where name=$1`, migration.Name).Exec()
		if err != nil {
			logger.Info("\n")
			return err
		}
		logger.Info("OK\n")
	}

	tx.Commit()
	return nil
}

// restore restores `dump` file
func commandRestore(ctx *AppContext) error {
	adapter := NewPostgresAdapter()

	superOptions, err := buildSuperOptions(ctx.Options)
	if err != nil {
		return err
	}

	db, err := adapter.AcquireDB(&superOptions.Connection)
	if err != nil {
		return err
	}

	err = adapter.ResetRole(ctx, db)
	db.Close()
	if err != nil {
		return err
	}

	filename := getCommandArg1(ctx)
	if filename == "" {
		dumpFiles, err := getDumpFiles(ctx)
		if err != nil {
			return err
		}

		filename, err = askOption("Choose dump file", dumpFiles)
		if err != nil {
			return err
		}
	} else {
		// TODO assumes filename is within _dumps
		filename = filepath.Join(ctx.Options.DumpsDir, filename)
	}

	var exe string
	var args []string
	passwordEnv := "PGPASSWORD=" + superOptions.Connection.Password
	dockerContainer := ctx.Options.DockerContainer
	if dockerContainer == "" {
		exe = "pg_restore"
	} else {
		exe = "docker"
		args = []string{
			"exec",
			"-i",
			"-e",
			passwordEnv,
			dockerContainer,
			"pg_restore",
		}
	}

	args = append(
		args,
		"--dbname="+superOptions.Connection.Database,
		"--host="+ctx.Options.Connection.Host,
		"--username="+superOptions.Connection.User,
		"--create",
	)

	cmd := exec.Command(exe, args...)
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	buf := bytes.NewBuffer(b)

	cmd.Env = append(os.Environ(), passwordEnv)
	cmd.Stdout = os.Stdout
	cmd.Stdin = buf
	cmd.Stderr = os.Stderr

	logger.Info("%s %v\n", exe, args)
	err = cmd.Run()
	if err == nil {
		logger.Info("restored %s", filename)
	}
	return err
}

func commandUp(ctx *AppContext) error {
	adapter, db, err := getAdapterAndDB(ctx)
	if err != nil {
		return err
	}

	localMigrations, err := getPartialLocalMigrations(ctx.Options)
	if err != nil {
		return err
	}

	if len(localMigrations) == 0 {
		logger.Info("Nothing to run. Try 'dat new some-migration'")
		return nil
	}

	dbMigrations, err := adapter.GetAllMigrations(db)
	if err != nil {
		return err
	}

	// return error if sequence is off between local directories and migrations
	// applied in DB
	err = verifyMigrationsHistory(ctx, localMigrations, dbMigrations)
	if err != nil {
		return err
	}

	dbLen := len(dbMigrations)
	localLen := len(localMigrations)
	var list []*Migration
	if dbLen == 0 {
		// run all local
		list = localMigrations
	} else {
		lastDBMigration := dbMigrations[dbLen-1]
		// find pending migrations relative to last one run in DB
		start := -1
		for i := 0; i < localLen; i++ {
			current := localMigrations[i]
			if lastDBMigration.Name == current.Name {
				start = i + 1
				break
			}
		}
		if start > -1 {
			list = localMigrations[start:]
		} else {
			list = localMigrations
		}
	}

	if len(list) == 0 {
		if dbLen == 0 {
			logger.Info("Nothing to run.")
		} else {
			// nothing to run is informational not an error
			logger.Info("Nothing to run, last migration was %s\n", dbMigrations[dbLen-1].Name)
		}
	}

	for _, migration := range list {
		err = runUpScripts(ctx, db, migration)
		if err != nil {
			return err
		}
	}

	err = upsertSprocs(db, ctx.Options)
	return err
}
