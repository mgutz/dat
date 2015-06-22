package runner

import (
	"testing"

	"gopkg.in/stretchr/testify.v1/assert"
)

func TestSprocRows(t *testing.T) {
	// returns multiple rows
	sql := `
CREATE OR REPLACE FUNCTION rows_table(x int, y int)
RETURNS TABLE(sum int, prod int) AS $$
BEGIN
	return query
	select x + y as sum, x * y as prod;
END;
$$ LANGUAGE plpgsql;
`
	testDB.DB.MustExec(sql)

	var sum int
	var prod int
	testDB.SQL(`select * from rows_table(1, 2)`).QueryScalar(&sum, &prod)
	assert.Equal(t, sum, 3)
	assert.Equal(t, prod, 2)
}

func TestSprocRowOut(t *testing.T) {
	// returns a single row
	sql := `
CREATE OR REPLACE FUNCTION row_out(x int, y int, OUT prod int, OUT sum int) AS $$
BEGIN
	sum := x + y;
	prod := x * y;
END;
$$ LANGUAGE plpgsql;
`
	testDB.DB.MustExec(sql)

	var sum int
	var prod int
	testDB.SQL(`select * from row_out(1, 2)`).QueryScalar(&prod, &sum)
	assert.Equal(t, sum, 3)
	assert.Equal(t, prod, 2)
}

// func TestSprocRowType(t *testing.T) {
// 	// returns a single row
// 	sql := `
// --create type row_type_result as (prod int, sum int);

// CREATE OR REPLACE FUNCTION row_type(x int, y int) RETURNS row_type_result AS $$
// DECLARE
// 	result row_type_result%rowtype;
// BEGIN
// 	-- using aliases did not work
// 	-- select into result x+y as sum, x * y as prod;
// 	select into result x*y, x+y;
// 	return result;
// END;
// $$ LANGUAGE plpgsql;
// `
// 	conn.DB.MustExec(sql)

// 	var sum int
// 	var prod int
// 	conn.SQL(`select * from row_type(1, 2)`).QueryScalar(&prod, &sum)
// 	assert.Equal(t, sum, 3)
// 	assert.Equal(t, prod, 2)
// }

func TestCallRows(t *testing.T) {
	// returns multiple rows
	sql := `
CREATE OR REPLACE FUNCTION rows_table(x int, y int)
RETURNS TABLE(sum int, prod int) AS $$
BEGIN
	return query
	select x + y as sum, x * y as prod;
END;
$$ LANGUAGE plpgsql;
`
	testDB.DB.MustExec(sql)

	var sum int
	var prod int
	testDB.Call("rows_table", 1, 2).QueryScalar(&sum, &prod)
	assert.Equal(t, sum, 3)
	assert.Equal(t, prod, 2)
}

func TestCallNoArgs(t *testing.T) {
	// returns multiple rows
	sql := `
CREATE OR REPLACE FUNCTION hello(OUT s text) AS $$
BEGIN
	s = 'Hello world!';
END;
$$ LANGUAGE plpgsql;
`
	testDB.DB.MustExec(sql)

	var s string
	err := testDB.Call("hello").QueryScalar(&s)
	assert.NoError(t, err)
	assert.Equal(t, "Hello world!", s)
}

func TestCallNoArgsReturns(t *testing.T) {
	// returns multiple rows
	sql := `
CREATE OR REPLACE FUNCTION hello2() RETURNS text AS $$
BEGIN
	return 'Hello world!';
END;
$$ LANGUAGE plpgsql;
`
	testDB.DB.MustExec(sql)

	var s string
	err := testDB.Call("hello2").QueryScalar(&s)
	assert.NoError(t, err)
	assert.Equal(t, "Hello world!", s)
}
