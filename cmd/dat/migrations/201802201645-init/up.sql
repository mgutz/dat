-- migrations/201802201645-init/up.sql
create table foo (
	id serial primary key,
	created_at timestamptz default now()
);
