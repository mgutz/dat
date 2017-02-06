package main

import (
	"database/sql"
	"fmt"

	"github.com/mgutz/ansi"
	"github.com/mgutz/str"
	do "gopkg.in/godo.v2"
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
