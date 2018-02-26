-- up.sql

create table accounts (
	id serial primary key,
	balance numeric(12, 8) default 0.0,
	created_at timestamptz default now()
);

-- create test account 700
insert into accounts(id) values(700);

-- reserve first 1000 for special use
alter sequence accounts_id_seq restart with 1000;
