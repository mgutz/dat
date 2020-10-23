package main

import (
	"database/sql"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mgutz/str"

	"github.com/mgutz/dat/dat"
	runner "github.com/mgutz/dat/sqlx-runner"
)

func dirExists(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && stat.IsDir()
}

var reSprocName = regexp.MustCompile(`(?mi)^\s*create function\s(\w+(\.(\w+))?)`)

// upsertSprocs parses all sprocs in SQL files under migrations/sprocs and
// insert/updates the database
func upsertSprocs(conn runner.Connection, options *CLIArgs) error {
	sprocsDir := options.SprocsDir
	if !dirExists(sprocsDir) {
		return nil
	}

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
			script := strings.TrimSpace(sproc)
			if !str.IsEmpty(script) {
				name := parseSprocName(script)
				localSprocs = append(localSprocs, &Sproc{
					Name:   name,
					Script: script,
					CRC:    hash([]byte(script)),
				})
			}
		}
	}

	tx, err := conn.Begin()
	if err != nil {
		return err
	}
	defer tx.AutoRollback()

	// need to know if it exists to display inserting or updating feedback
	existsSproc := `
	select name, crc
	from dat__sprocs
	where name = $1

	union all

	select  proname, '0'
	from pg_catalog.pg_namespace n
	join pg_catalog.pg_proc p on pronamespace = n.oid
	where
		nspname = 'public'
		and proname = $1
		and not exists (
			select 1
			from dat__sprocs
			where name = $1
		)
	`

	executed := 0
	for _, localSproc := range localSprocs {
		found := false
		var dbSproc Sproc
		err = tx.SQL(existsSproc, localSproc.Name).QueryStruct(&dbSproc)
		if err != nil {
			if err != sql.ErrNoRows {
				return err
			}
		} else {
			found = true
		}

		addSproc := false
		if !found {
			addSproc = true
			logger.Info("Inserting %s\n", localSproc.Name)
		} else if dbSproc.CRC != localSproc.CRC {
			addSproc = true
			logger.Info("Updating %s\n", localSproc.Name)
		}

		if addSproc {
			trackAndDeleteSQL := `
			delete from dat__sprocs where name = $1;
			insert into dat__sprocs (name, crc, script) values ($1, $2, $3);
			select dat__delfunc($1);
			`

			_, err = tx.ExecExpr(dat.Interp(trackAndDeleteSQL, localSproc.Name, localSproc.CRC, localSproc.Script))
			if err != nil {
				return err
			}
			err = execScript(tx, localSproc.Script, false)
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

func parseSprocName(body string) string {
	matches := reSprocName.FindStringSubmatch(body)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
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
