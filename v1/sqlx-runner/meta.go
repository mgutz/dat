package runner

import (
	"database/sql"
	"fmt"
	"hash/fnv"

	"github.com/mgutz/dat/v1"
)

// MustCreateMetaTable creates the dat__meta table or panics.
func (con *Connection) MustCreateMetaTable() {
	// pg function to delete a function without having to worry about
	// the arguments changing.
	delfunc := `
CREATE OR REPLACE FUNCTION dat__delfunc(_name text) returns void AS $$
BEGIN
    EXECUTE (
       SELECT string_agg(format('DROP FUNCTION %s(%s);'
                         ,oid::regproc
                         ,pg_catalog.pg_get_function_identity_arguments(oid))
              ,E'\n')
       FROM   pg_proc
       WHERE  proname = _name
       AND    pg_function_is_visible(oid)
    );
exception when others then
    -- do nothing, EXEC above returns an exception if it does not
	-- find existing function
END $$ LANGUAGE plpgsql;`

	// The table used to track versions of functions and eventually
	// migrations
	createMeta := `
CREATE TABLE IF NOT EXISTS dat__meta (
	id serial primary key,
	kind text,
	name text,
	version text,
	meta1 text,
	meta2 text,
	meta3 text,
	created_at timestamptz default now()
)
	`

	sess, err := con.NewSession()
	if err != nil {
		logger.Fatal("Could not create session")
	}
	defer sess.AutoRollback()

	_, err = sess.ExecMulti(
		dat.Expr(delfunc),
		dat.Expr(createMeta),
	)
	if err != nil {
		logger.Fatal("Could not execute Multi SQL")
		panic(err)
	}
	sess.Commit()
}

// MustRegisterFunction registers a user defined function but will not recreate it
// unles the hash has changed with version. This is useful for keeping user defined
// functions defined in source code.
func (con *Connection) MustRegisterFunction(name string, version string, body string) {
	sess, err := con.NewSession()
	if err != nil {
		logger.Fatal("Could not register function", "err", err, "name", name)
	}
	defer sess.AutoRollback()

	h := fnv.New64a()
	h.Write([]byte(body))
	crc := fmt.Sprintf("%s-%X", version, h.Sum64())

	// check if the function already exists
	var metaID int
	err = con.
		SQL(`SELECT id FROM dat__meta WHERE kind = 'function' AND version = $1 AND name = $2`, crc, name).
		QueryScalar(&metaID)
	if err != nil && err != sql.ErrNoRows && err != dat.ErrNotFound {
		logger.Fatal("Could not get metadata for function", "err", err)
	}

	if metaID == 0 {
		logger.Debug("Adding function", "name", name)
		commands := []*dat.Expression{
			dat.Expr(`
				INSERT INTO dat__meta (kind, version, name)
				VALUES ('function', $1, $2)
			`, crc, name),

			// have to use a special delete function since Postgres will
			// not delete a function if arguments change
			dat.Expr(`
				SELECT dat__delfunc($1)
			`, name),

			dat.Expr(body),
		}

		_, err := sess.ExecMulti(commands...)
		if err != nil {
			logger.Fatal("Could not insert function", "err", err)
		}
	}
	sess.Commit()
}
