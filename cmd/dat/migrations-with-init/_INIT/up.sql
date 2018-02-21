/**
 * This idempotent script is run each time dat starts to convert legacy
 * migrations to `dat`. In our case, it migrates `mygrate` tables to `dat`.
 *
 * dat runs `_INIT/up.sql` if it exists.
 */
do $adhoc$ begin
	if not exists (
		select 1
		from information_schema.tables
		where table_schema = 'public'
			and table_name = 'dat__migrations'
	) then
		create table dat__migrations (
			name text primary key,
			up_script text not null,
			down_script text default '',
			no_tx_script text default '',
			created_at timestamptz default now()
		);

		insert into dat__migrations (name, up_script, down_script, created_at)
		select version, up, down, created_at
		from  schema_migrations
		order by created_at;
	end if;

	if not exists (
		select 1
		from information_schema.tables
		where table_schema = 'public'
			and table_name = 'dat__sprocs'
	) then
		create table dat__sprocs (
			name text primary key,
			script text not null,
			crc text not null,
			updated_at timestamptz default now(),
			created_at timestamptz default now()
		);

		insert into dat__sprocs (name, script, crc, created_at, updated_at)
		select name, '', crc, created_at, created_at
		from  mygrate__sprocs
		order by name;
	end if;
end $adhoc$;
