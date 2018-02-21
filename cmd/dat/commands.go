package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"
)

func createDB(options *AppOptions) error {
	superOptions, err := buildSuperOptions(options)
	if err != nil {
		return err
	}

	adapter := NewPostgresAdapter(options)
	superDB, err := adapter.AcquireDB(superOptions)
	if err != nil {
		return err
	}
	defer superDB.Close()

	err = adapter.Create(superDB)
	if err != nil {
		return err
	}

	fmt.Println("Created", adapter.ConnectionString(&options.Connection))
	return nil
}

func list(options *AppOptions) error {
	adapter, db, err := getAdapterAndDB(options)
	if err != nil {
		return err
	}

	migrations, err := adapter.GetAllMigrations(db)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("No migrations found.")
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
func newScript(options *AppOptions) error {
	if len(options.UnparsedArgs) != 2 {
		return errors.New("Usage: dat new TITLE")
	}

	// command is "new title"
	title := options.UnparsedArgs[1]
	upFilename := migrationFile(options.MigrationsDir, title, "up.sql")
	downFilename := migrationFile(options.MigrationsDir, title, "down.sql")

	err := writeFileAll(upFilename, []byte("-- up.sql"))
	if err != nil {
		return err
	}

	err = writeFileAll(downFilename, []byte("-- down.sql"))
	if err != nil {
		return err
	}

	return nil
}

func up(options *AppOptions) error {
	adapter, db, err := getAdapterAndDB(options)
	if err != nil {
		return err
	}

	localMigrations, err := getPartialLocalMigrations(options)
	if err != nil {
		return err
	}

	if len(localMigrations) == 0 {
		// nothing to run is informational not an error
		fmt.Println("Nothing to run. Try 'dat new  some-migration'")
		return nil
	}

	dbMigrations, err := adapter.GetAllMigrations(db)
	if err != nil {
		return err
	}

	// return error if sequence is off between local directories and migrations
	// applied in DB
	err = verifyMigrationsHistory(localMigrations, dbMigrations)
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
			fmt.Printf("Nothing to run.")
		} else {
			// nothing to run is informational not an error
			fmt.Printf("Nothing to run. Last migration was %s\n", dbMigrations[dbLen-1].Name)
		}
		return nil
	}

	for _, migration := range list {
		fmt.Println("loop")
		err = runUpScripts(options, db, migration)
		fmt.Println("AFTER AFTER AFTER")
		if err != nil {
			return err
		}
	}
	fmt.Println("exiting")

	//err = upsertSprocs(options, )
	return nil
}
