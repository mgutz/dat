package runner

import "github.com/mgutz/dat"

func benchReset() {
	var sql = `
	DROP TABLE IF EXISTS benches;
	CREATE TABLE benches (
		id SERIAL PRIMARY KEY,
		amount money,
		image bytea,
		is_ok boolean,
		name text,
		n integer,
		created_at timestamptz default now()
	);
	DELETE FROM benches;
	`
	err := conn.ExecMulti(sql)
	if err != nil {
		panic(err)
	}
	return
}

// benchInsertBuilders builds an insert statement with
// many values.
//
// INSERT INTO(benches)
// VALUES (row0), (row1), ... (rown-1)
func benchInsertBuilder(rows int, argc int) (dat.Builder, error) {
	if argc > 4 {
		panic("args must be <= 4")
	}

	columns := []string{"amount", "name", "n", "is_ok"}
	values := []interface{}{42.0, "foo", 42, "true"}
	builder := dat.
		NewInsertBuilder("benches").
		Whitelist(columns[0:argc]...)

	// fill image with random bytes
	maxImage := 256
	image := make([]byte, maxImage)
	for i := 0; i < maxImage; i++ {
		image[i] = byte(i % 256)
	}

	for i := 0; i < rows; i++ {
		builder.Values(values[0:argc]...)
	}
	return builder, nil
}
