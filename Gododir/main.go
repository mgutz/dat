package main

import (
	_ "github.com/lib/pq"
	. "github.com/mgutz/godo/v2"
)

func tasks(p *Project) {
	Env = `
	DAT_DRIVER=postgres
	DAT_DSN="dbname=dbr_test user=dbr password=!test host=localhost sslmode=disable"
	`

	generateTasks(p)
	p.Use("pg", pgTasks)

	p.Task("createdb", createdb).Description("Creates test database")

	p.Task("test", func(c *Context) {
		Run(`go test`, M{"$in": "v1"})
		Run(`go test`, M{"$in": "v1/sqlx-runner"})
	}).Src("**/*.go")

	p.Task("test-dir", func(c *Context) {
		dir := c.Args.Leftover()[0]
		Run(`go test`, M{"$in": dir})
	})

	p.Task("test-one", func() {
		Run(`go test -run TestInsertBytes`, M{"$in": "v1/sqlx-runner"})
	}).Src("*.go")

	p.Task("allocs", func() {
		Bash(`
		go test -c
		GODEBUG=allocfreetrace=1 ./dat.test -test.bench=BenchmarkSelectBasicSql -test.run=none -test.benchtime=10ms 2>trace.log
		`)
	})

	p.Task("bench", func() {
		// Bash("go test -bench . -benchmem 2>/dev/null | column -t")
		// Bash("go test -bench . -benchmem 2>/dev/null | column -t", In{"sqlx-runner"})
		Bash("go test -bench . -benchmem", M{"$in": "v1"})
		Bash("go test -bench . -benchmem", M{"$in": "v1/sqlx-runner"})
	})
	p.Task("bench2", func() {
		Bash("go test -bench . -benchmem", M{"$in": "v1/sqlx-runner"})
	})

	p.Task("bench-builder", func() {
		Bash(`
		go test -c
		#GODEBUG=allocfreetrace=1 ./sqlx-runner.test -test.bench=BenchmarkInsertTransactionDat100 -test.run=none -test.benchtime=1s -test.benchmem 2>trace.log
		GODEBUG=allocfreetrace=1 ./sqlx-runner.test -test.run=none -test.bench . -test.benchmem 2>trace.log
		`, M{"$in": "v1/sqlx-runner"})
	})

	p.Task("install", func() {
		Run("go install -a", M{"$in": "v1/sqlx-runner"})
	})

	p.Task("default", S{"builder-boilerplate"})
}

func main() {
	Godo(tasks)
}
