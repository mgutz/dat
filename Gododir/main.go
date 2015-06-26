package main

import (
	"fmt"

	_ "github.com/lib/pq"
	do "gopkg.in/godo.v2"
)

func tasks(p *do.Project) {
	do.Env = `
	DAT_DRIVER=postgres
	DAT_DSN="dbname=dbr_test user=dbr password=!test host=localhost sslmode=disable"
	`
	generateTasks(p)
	p.Use("pg", pgTasks)

	p.Task("createdb", nil, createdb).Description("Creates test database")

	p.Task("test", nil, func(c *do.Context) {
		c.Run(`go test -race`)
		c.Run(`go test -race`, do.M{"$in": "sqlx-runner"})
	}).Src("**/*.go").
		Desc("test with -race flag")

	p.Task("test-fast", nil, func(c *do.Context) {
		c.Run(`go test`)
		c.Run(`go test`, do.M{"$in": "sqlx-runner"})
	}).Src("**/*.go").
		Desc("fater test without -race flag")

	p.Task("test-dir", nil, func(c *do.Context) {
		dir := c.Args.NonFlags()[0]
		c.Run(`go test`, do.M{"$in": dir})
	})

	p.Task("test-one", nil, func(c *do.Context) {
		c.Run(`LOGXI=* go test -run TestSelectDocDate`, do.M{"$in": "sqlx-runner"})
	}).Src("*.go")

	p.Task("allocs", nil, func(c *do.Context) {
		c.Bash(`
		go test -c
		GODEBUG=allocfreetrace=1 ./dat.test -test.bench=BenchmarkSelectBasicSql -test.run=none -test.benchtime=10ms 2>trace.log
		`)
	})

	p.Task("hello", nil, func(*do.Context) {
		fmt.Println("hello?")
	})

	p.Task("bench", nil, func(c *do.Context) {
		// Bash("go test -bench . -benchmem 2>/dev/null | column -t")
		// Bash("go test -bench . -benchmem 2>/dev/null | column -t", In{"sqlx-runner"})
		c.Bash("go test -bench . -benchmem")
		c.Bash("go test -bench . -benchmem", do.M{"$in": "sqlx-runner"})
	})
	p.Task("bench2", nil, func(c *do.Context) {
		c.Bash("go test -bench . -benchmem", do.M{"$in": "sqlx-runner"})
	})

	p.Task("bench-builder", nil, func(c *do.Context) {
		c.Bash(`
		go test -c
		#GODEBUG=allocfreetrace=1 ./sqlx-runner.test -test.bench=BenchmarkInsertTransactionDat100 -test.run=none -test.benchtime=1s -test.benchmem 2>trace.log
		GODEBUG=allocfreetrace=1 ./sqlx-runner.test -test.run=none -test.bench . -test.benchmem 2>trace.log
		`, do.M{"$in": "sqlx-runner"})
	})

	p.Task("install", nil, func(c *do.Context) {
		c.Run("go install -a", do.M{"$in": "sqlx-runner"})
	})

	p.Task("default", do.S{"builder-boilerplate"}, nil)

	p.Task("example", nil, func(c *do.Context) {
	})

	p.Task("lint", nil, func(c *do.Context) {
		c.Bash(`
		echo Directory=.
		golint

		cd sqlx-runner
		echo
		echo Directory=sqlx-runner
		golint

		cd ../kvs
		echo
		echo Directory=kvs
		golint

		cd ../postgres
		echo
		echo Directory=postgres
		golint
		`)
	})

	p.Task("mocks", nil, func(c *do.Context) {
		// go get github.com/vektra/mockery
		c.Run("mockery --dir=kvs --all")
	})
}

func main() {
	do.Godo(tasks)
}
