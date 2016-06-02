package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/mgutz/ansi"
	"github.com/mgutz/str"
	do "gopkg.in/godo.v2"
	"gopkg.in/godo.v2/util"
	"gopkg.in/mgutz/dat.v1"
	"gopkg.in/mgutz/dat.v1/sqlx-runner"
)

func mapBytesToString(m map[string]interface{}) {
	for k, v := range m {
		if b, ok := v.([]byte); ok {
			m[k] = string(b)
		}
	}
}

var lightGreen = ansi.ColorFunc("green+h")
var cyan = ansi.ColorFunc("cyan")

func printMap(m map[string]interface{}) {
	first := true

	fmt.Print(lightGreen("{"))
	for k, v := range m {
		if !first {
			fmt.Print(" ")
		}
		fmt.Printf("%s=%v", cyan(k), v)
		first = false
	}
	fmt.Print(lightGreen("}"))
	fmt.Print("\n")
}

func querySQL(sql string, args []interface{}) {
	fmt.Printf("%s: %s\n", cyan("sql"), sql)
	if len(args) > 0 {
		fmt.Printf("%s: %#v\n", cyan("args"), args)
	}

	conn := getConnection()
	if conn != nil {
		rows, err := conn.DB.Queryx(sql, args...)
		if err != nil {
			util.Error("pg", "Error executing SQL: %s", err.Error())
		}
		for rows.Next() {
			m := map[string]interface{}{}
			rows.MapScan(m)
			mapBytesToString(m)
			b, _ := json.Marshal(m)
			json.Unmarshal(b, &m)
			printMap(m)
		}
	}
	util.Info("pg", "OK\n")
}

func createdb(c *do.Context) {
	user := do.Prompt("superuser: ")
	password := do.PromptPassword("password: ")

	dsn := str.Template("user={{user}} password={{password}} dbname=postgres host=localhost sslmode=disable", do.M{
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
		sql2 := str.Template(cmd, do.M{
			"dbname":   "dbr_test",
			"user":     "dbr",
			"password": "!test",
		})
		_, err = db.Exec(sql2)
		if err != nil {
			panic(err)
		}
	}

	// close superser db connection
	err = db.Close()
	if err != nil {
		panic(err)
	}

	dsn = str.Template("user={{user}} password={{password}} dbname=dbr_test host=localhost sslmode=disable", do.M{
		"user":     user,
		"password": password,
	})
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	_, err = db.Exec("create extension if not exists hstore")
	if err != nil {
		panic(err)
	}

	fmt.Println("OK")
}

var db *runner.DB

func getConnection() *runner.DB {

	if db == nil {
		connectionString := do.Getenv("DAT_DSN")
		db = runner.NewDBFromString("postgres", connectionString)
	}
	return db
}

func pgTasks(p *do.Project) {
	do.Env = `
	DAT_DRIVER=postgres
	DAT_DSN="dbname=dbr_test user=dbr password=!test host=localhost sslmode=disable"
	`

	p.Task("file", nil, func(c *do.Context) {
		filename := c.Args.Leftover()[0]
		if !util.FileExists(filename) {
			util.Error("ERR", "file not found %s", filename)
			return
		}
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			panic(err)
		}
		parts := strings.Split(string(b), "---\n")
		if len(parts) != 2 {
			panic("sql file must have frontmatter")
		}
		var args []interface{}
		err = json.Unmarshal([]byte(parts[0]), &args)
		if err != nil {
			panic(err)
		}
		sql := parts[1]
		sql, args, _ = dat.Interpolate(sql, args)
		querySQL(sql, args)
	}).Desc("Executes a SQL file with placeholders")

	p.Task("query", nil, func(c *do.Context) {
		if len(c.Args.Leftover()) != 1 {
			fmt.Println(`usage: godo query -- "SELECT * ..." `)
			return
		}
		sql := c.Args.Leftover()[0]
		querySQL(sql, nil)

	}).Desc("Executes a query against the database")
}
