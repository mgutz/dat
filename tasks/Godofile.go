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
	DAT_DRIVER=postgres
	DAT_DSN="dbname=dbr_test user=dbr password=!test host=localhost sslmode=disable"
	`
	p.Task("createdb", createdb).Description("Creates test database")

	p.Task("test", func(c *Context) {
		Run(`go test`)
		Run(`go test`, In{"sqlx-runner"})
	}).Watch("**/*.go")

	p.Task("test-dir", func(c *Context) {
		dir := c.Args.Leftover()[0]
		Run(`go test`, In{dir})
	})

	p.Task("test-one", func() {
		Run(`go test -run TestInsertBytes`, In{"sqlx-runner"})
	}).Watch("*.go")

	p.Task("allocs", func() {
		Bash(`
		go test -c
		GODEBUG=allocfreetrace=1 ./dat.test -test.bench=BenchmarkSelectBasicSql -test.run=none -test.benchtime=10ms 2>trace.log
		`)
	})

	p.Task("bench", func() {
		Bash("go test -bench . -benchmem 2>/dev/null | column -t")
		Bash("go test -bench . -benchmem 2>/dev/null | column -t", In{"sqlx-runner"})
	})

	p.Task("bench-builder", func() {
		Bash(`
		go test -c
		#GODEBUG=allocfreetrace=1 ./sqlx-runner.test -test.bench=BenchmarkInsertTransactionDat100 -test.run=none -test.benchtime=1s -test.benchmem 2>trace.log
		GODEBUG=allocfreetrace=1 ./sqlx-runner.test -test.run=none -test.bench . -test.benchmem 2>trace.log
		`, In{"sqlx-runner"})
	})

	p.Task("install", func() {
		Run("go install -a", In{"sqlx-runner"})
	})
}

func main() {
	Godo(tasks)
}
