package main

import (
	"database/sql"
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"os/exec"
	"path/filepath"

	runner "github.com/mgutz/dat/sqlx-runner"
	"github.com/olekukonko/tablewriter"
)

func createDB(ctx *AppContext) error {
	superOptions, err := buildSuperOptions(ctx.Options)
	if err != nil {
		return err
	}

	adapter := NewPostgresAdapter(ctx.Options)
	superDB, err := adapter.AcquireDB(superOptions)
	if err != nil {
		return err
	}
	defer superDB.Close()

	err = adapter.Create(superDB)
	if err != nil {
		return err
	}

	logger.Info("Created %s", adapter.ConnectionString(&ctx.Options.Connection))
	return nil
}

// dump dumps a database to a file for use by restore
func dump(ctx *AppContext) error {
	superOptions, err := buildSuperOptions(ctx.Options)
	if err != nil {
		return err
	}

	var newFilename string
	if len(ctx.Options.UnparsedArgs) > 1 {
		newFilename = ctx.Options.UnparsedArgs[1]
	} else {
		newFilename = timestampedName("dump")
	}
	dumpsDir := ctx.Options.DumpsDir
	destination := filepath.Join(dumpsDir, newFilename)
	err = os.MkdirAll(dumpsDir, os.ModePerm)
	if err != nil {
		return err
	}

	cmd := exec.Command(
		"pg_dump",
		"--dbname="+ctx.Options.Connection.Database,
		"--host="+ctx.Options.Connection.Host,
		"--username="+superOptions.Connection.User,
		"-Fc",
		"-f",
		destination,
	)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+superOptions.Connection.Password)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err == nil {
		logger.Info("dumped %s", destination)
	}
	return err
}

func list(ctx *AppContext) error {
	adapter, db, err := getAdapterAndDB(ctx.Options)
	if err != nil {
		return err
	}

	migrations, err := adapter.GetAllMigrations(db)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Info("No migrations found.")
			return nil
		}
		return err
	}

	if len(migrations) == 0 {
		return errors.New("No migrations found")
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
func newScript(ctx *AppContext) error {
	if len(ctx.Options.UnparsedArgs) != 2 {
		return errors.New("Usage: dat new TITLE")
	}

	// command is "new title"
	title := ctx.Options.UnparsedArgs[1]
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

	logger.Info("migrations created %s\n", filepath.Dir(upFilename))

	return nil
}

// restore restores `dump` file
func restore(ctx *AppContext) error {
	adapter := NewPostgresAdapter(ctx.Options)

	// drop the database

	superOptions, err := buildSuperOptions(ctx.Options)
	if err != nil {
		return err
	}

	db, err := adapter.AcquireDB(superOptions)
	if err != nil {
		return err
	}

	err = adapter.Create(db)
	if err != nil {
		db.Close()
		return err
	}
	db.Close()

	var filename string
	if len(ctx.Options.UnparsedArgs) > 1 {
		// TODO assumes filename is within _dumps
		filename = filepath.Join(ctx.Options.DumpsDir, ctx.Options.UnparsedArgs[1])
	} else {
		dumpFiles, err := getDumpFiles(ctx)
		if err != nil {
			return err
		}

		filename, err = askOption("Choose dump file", dumpFiles)
		if err != nil {
			return err
		}
	}

	cmd := exec.Command(
		"pg_restore",
		"--dbname="+ctx.Options.Connection.Database,
		"--host="+ctx.Options.Connection.Host,
		"--username="+superOptions.Connection.User,
		"--create",
		filename,
	)
	cmd.Env = append(os.Environ(), "PGPASSWORD="+superOptions.Connection.Password)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err == nil {
		logger.Info("restored %s", filename)
	}
	return err
}

// dropDB drops the database
func dropDB(ctx *AppContext) error {
	adapter := NewPostgresAdapter(ctx.Options)

	// drop the database

	superOptions, err := buildSuperOptions(ctx.Options)
	if err != nil {
		return err
	}

	db, err := adapter.AcquireDB(superOptions)
	if err != nil {
		return err
	}

	err = adapter.Drop(db)
	if err != nil {
		return err
	}
	db.Close()
	return nil
}

func up(ctx *AppContext) error {
	adapter, db, err := getAdapterAndDB(ctx.Options)
	if err != nil {
		return err
	}

	localMigrations, err := getPartialLocalMigrations(ctx.Options)
	if err != nil {
		return err
	}

	if len(localMigrations) == 0 {
		logger.Info("Nothing to run. Try 'dat new  some-migration'")
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

// upsertSprocs parses all sprocs in SQL files under migrations/sprocs and
// insert/updates the database
func upsertSprocs(conn runner.Connection, options *AppOptions) error {
	sprocsDir := options.SprocsDir
	files, err := getSprocFiles(sprocsDir)
	if err != nil {
		return err
	}

	var localSprocs []*Sproc
	for _, file := range files {
		filename := filepath.Join(sprocsDir, file)
		script, err := readFileText(filename)
		if err != nil {
			return err
		}

		sprocs := reBatchSeparator.Split(script, -1)
		if len(sprocs) == 0 {
			continue
		}

		for _, sproc := range sprocs {
			name := parseSprocName(sproc)
			localSprocs = append(localSprocs, &Sproc{
				Name:   name,
				Script: sproc,
				CRC:    hash([]byte(sproc)),
			})
		}
	}

	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.AutoRollback()

	var dbSprocs []*Sproc

	existsSproc := `
	SELECT name, crc
	FROM dat__sprocs
	`
	err = tx.SQL(existsSproc).QueryStructs(&dbSprocs)
	if err != nil {
		return err
	}

	executed := 0
	for _, localSproc := range localSprocs {
		found := sprocFind(dbSprocs, localSproc.Name)
		if found == nil {
			insert := `
			insert into dat__sprocs (name, crc, script)
			values ($1, $2, $3)
			`
			logger.Info("Inserting sproc %s ...", localSproc.Name)
			_, err := tx.SQL(insert, localSproc.Name, localSproc.CRC, localSproc.Script).Exec()
			if err != nil {
				return err
			}
			executed++
		} else if found.CRC != localSproc.CRC {
			update := `
			update dat__sprocs
			set
				crc = $1,
				updated_at = now(),
				script = $2
			where name = $3
			`

			logger.Info("Updating sproc %s ...", localSproc.Name)
			_, err := tx.SQL(update, localSproc.CRC, localSproc.Script, localSproc.Name).Exec()

			if err != nil {
				return err
			}
			executed++
		}
	}

	if executed == 0 {
		logger.Info("No sprocs need updating")
	}

	tx.Commit()

	return nil
}

func hash(b []byte) string {
	hash := fnv.New64a()
	hash.Write(b)
	return fmt.Sprintf("%x", hash.Sum64())
}

func sprocFind(sprocs []*Sproc, name string) *Sproc {

	if len(sprocs) > 0 {
		for _, sproc := range sprocs {
			if sproc.Name == name {
				return sproc
			}
		}
	}

	return nil
}
