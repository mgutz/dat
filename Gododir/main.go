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

	p.Task("createdb", nil, createdb).Description("Creates test database")

	p.Task("test", nil, func(c *Context) {
		c.Run(`go test`, M{"$in": "v1"})
		c.Run(`go test`, M{"$in": "v1/sqlx-runner"})
	}).Src("**/*.go")

	p.Task("test-dir", nil, func(c *Context) {
		dir := c.Args.Leftover()[0]
		Run(`go test`, M{"$in": dir})
	})

	p.Task("test-one", nil, func(c *Context) {
		c.Run(`go test -run TestInsertBytes`, M{"$in": "v1/sqlx-runner"})
	}).Src("*.go")

	p.Task("allocs", nil, func(c *Context) {
		c.Bash(`
		go test -c
		GODEBUG=allocfreetrace=1 ./dat.test -test.bench=BenchmarkSelectBasicSql -test.run=none -test.benchtime=10ms 2>trace.log
		`)
	})

	p.Task("bench", nil, func(c *Context) {
		// Bash("go test -bench . -benchmem 2>/dev/null | column -t")
		// Bash("go test -bench . -benchmem 2>/dev/null | column -t", In{"sqlx-runner"})
		c.Bash("go test -bench . -benchmem", M{"$in": "v1"})
		c.Bash("go test -bench . -benchmem", M{"$in": "v1/sqlx-runner"})
	})
	p.Task("bench2", nil, func(*Context) {
		Bash("go test -bench . -benchmem", M{"$in": "v1/sqlx-runner"})
	})

	p.Task("bench-builder", nil, func(*Context) {
		Bash(`
		go test -c
		#GODEBUG=allocfreetrace=1 ./sqlx-runner.test -test.bench=BenchmarkInsertTransactionDat100 -test.run=none -test.benchtime=1s -test.benchmem 2>trace.log
		GODEBUG=allocfreetrace=1 ./sqlx-runner.test -test.run=none -test.bench . -test.benchmem 2>trace.log
		`, M{"$in": "v1/sqlx-runner"})
	})

	p.Task("install", nil, func(*Context) {
		Run("go install -a", M{"$in": "v1/sqlx-runner"})
	})

	p.Task("default", S{"builder-boilerplate"}, nil)
}

func main() {
	Godo(tasks)
}
