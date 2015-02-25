--@create_table
CREATE TABLE IF NOT EXISTS TABLE benches (
	id SERIAL PRIMARY KEY,
	amount money,
	store hstore,
	image bytea,
	name text,
	is_ok boolean,
	created_at timestamptz default now()
);

--@clear_table
DELETE FROM benches;
