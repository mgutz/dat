package main

import "fmt"

func printUsage(exe string, version string) {
	usage := `
%s %s

Usage: %s [options] [command] [filename.sql]

Commands:
  console         Runs database CLI
  createdb        (Re)Create database from config.js and $NODE_ENV
  down            Undo {COUNT|VERSION|all} migrations
  dropdb          Drops the database
  file            Execute SQL script in file
  history         Show migrations in database. (default)
  init            Creates migration directory with config
  redo            Undo last dir if applied and migrate up
  new             Generate new migration directory
  ping            Pings the database (0 exit code means OK)
  up              Execute new migrations
  exec            Execute an expression

Options:
      --directory use different directory than migrations
      --me        use current user as superuser
  -?, --help      output usage information
  -U, --username  superuser username
  -W, --password  superuser password
  -V, --version   output the version number
	`

	fmt.Println(fmt.Sprintf(usage, exe, version, exe))
}

// main is THE entry point
func main() {
}
