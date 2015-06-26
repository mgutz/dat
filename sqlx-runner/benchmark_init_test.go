package runner

import "gopkg.in/mgutz/dat.v1"

func benchReset() {
	var sql = `
	DROP TABLE IF EXISTS benches;
	CREATE TABLE benches (
		id SERIAL PRIMARY KEY,
		amount decimal,
		image bytea,
		is_ok boolean,
		name text,
		n integer,
		created_at timestamptz default now()
	);
	DELETE FROM benches;
	`
	err := execMulti(sql)
	if err != nil {
		panic(err)
	}
	return
}

// execMulti executes grouped SQL statements in a string delimited by a marker.
// The marker is "^GO$" which means GO on a line by itself.
func execMulti(sql string) error {
	statements, err := dat.SQLSliceFromString(sql)
	if err != nil {
		return err
	}
	// TODO this should be in transaction
	for _, sq := range statements {
		_, err := testDB.SQL(sq).Exec()
		if err != nil {
			return err
		}
	}
	return nil
}
