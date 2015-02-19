package main

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/mgutz/str"
	"github.com/peterh/liner"
	. "gopkg.in/godo.v1"
)

func createdb() {
	line := liner.NewLiner()

	user, _ := line.Prompt("superuser: ")
	password, _ := line.PasswordPrompt("password: ")

	dsn := str.Template("user={{user}} password={{password}} dbname=postgres host=localhost sslmode=disable", M{
		"user":     user,
		"password": password,
	})
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}

	commands := []string{
		"drop database if exists {{dbname}}",
		"drop user if exists {{user}}",
		"create user {{user}} password '{{password}}' CREATEROLE",
		"create database {{dbname}} owner {{user}}",
	}
	for _, cmd := range commands {
		sql := str.Template(cmd, M{
			"dbname":   "dbr_test",
			"user":     "dbr",
			"password": "!test",
		})
		_, err = db.Exec(sql)
		if err != nil {
			panic(err)
		}
	}
	fmt.Println("OK")
}

func tasks(p *Project) {
	Env = `
	DBR_DRIVER=postgres
	DBR_DSN="dbname=dbr_test user=dbr password=!test host=localhost sslmode=disable"
	`
	p.Task("createdb", createdb).Description("Creates test database")

	p.Task("test", func() {
		Run(`go test`)
		Run(`go test`, In{"sql-runner"})
	}).Watch("*.go")

	p.Task("test-some", func() {
		Run(`go test -run InsertReal`, In{"sql-runner"})
	}).Watch("*.go")

	p.Task("bench", func() {
		Bash("go test -bench . -benchmem 2>/dev/null | column -t")
	})
}

// func dum() {
//	b := sqb.Select("column, foobar").Returning("at")
//  b := sqb.Interpolate("asdfasdfadsf", args)
// }

func main() {
	Godo(tasks)
}
