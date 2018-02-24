package main

import (
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"regexp"

	runner "github.com/mgutz/dat/sqlx-runner"
)

func dirExists(path string) bool {
	stat, err := os.Stat(path)
	return err == nil && stat.IsDir()
}

var reSprocName = regexp.MustCompile(`(?mi)^\s*create function\s(\w+(\.(\w+))?)`)

// upsertSprocs parses all sprocs in SQL files under migrations/sprocs and
// insert/updates the database
func upsertSprocs(conn runner.Connection, options *AppOptions) error {
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

			logger.Info("Updating sproc %s ...\n", localSproc.Name)
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
