# dat Migration Utility

* SQL files only

* Stored procedures are updated as needed and run outside of migrations.

* Dump and restore database. Will use docker container's pg_dump and pg_restore
  if `dockerContainer` is set in`dat.yaml`

* Batch separator `GO`

## Workflow

Initialize migrations

```sh
dat init
```

Create database

```sh
# edit configuration file
vim migrations/dat.yaml
dat createdb
```

Create migration with description "add-tables"

```sh
dat new add-tables

# edit `down.sql`, `up.sql`, `notx.sql` script in migrations/TIMESTAMP-add-tables
```

Run migrations

```sh
dat up
```

Redo last migration if you made a mistake. `down.sql` must have been valid

```sh
dat redo
```

Start from fresh DB

```sh
# same as create
dat createdb
```

Dump database to send to colleague or to creat a snapshot

```sh
# set dockerContainer if running Postgres in docker to use pg_dump in container
# creates migrations/_dumps/ISSUE-31 file
dat dump ISSUE-31 --dockerContainer=postgres-svc
```

Restore dump

```sh
dat dump ISSUE-31 --dockerContainer=postgres-svc
```

Create/edit a stored procedure

```sh
# add or edit file to any file with .sql extension under migrations/sprocs
# use 'create function' to define (no need for create or replace)
vim migrations/sprocs/calc_tax.sql

# then run migrations
dat up
```

## Directory Structure and Files

```
migrations/
    TIMESTAMP-description1/
        up.sql      # up script
        down.sql    # down script
        notx.sql    # up script which runs outside of transaction
    TIMESTAMP-description2/
    ...
    TIMESTAMP-descriptionn/

    _init/
        up.sql      # idempotent script to run before dat bootstraps

    sprocs/
        any.sql     # 1 or more files sproc files

    dat.yaml        # configuration file
```

## FAQ

Q. How to add multiple sprocs in file?

A. Use `GO` separator.

```
    # file=migrations/sprocs/foobar.sql

    create function f_foo()
    returns void as $$
    begin
    end;
    $$ language plpgsql;

    GO

    create function f_bar()
    returns void as $$
    begin
    end;
    $$ language plpgsql;
```
