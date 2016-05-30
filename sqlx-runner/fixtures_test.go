package runner

import "github.com/syreclabs/dat"

type Team struct {
	ID        int64  `db:"id"`
	Name      string `db:"name"`
	CreatedAt dat.NullTime
}

type Person struct {
	ID        int64           `db:"id"`
	Amount    dat.NullFloat64 `db:"amount"`
	Doc       dat.NullString  `db:"doc"`
	Email     dat.NullString  `db:"email"`
	Foo       string          `db:"foo"`
	Image     []byte          `db:"image"`
	Key       dat.NullString  `db:"key"`
	Name      string          `db:"name"`
	CreatedAt dat.NullTime    `db:"created_at"`
}

type Post struct {
	ID        int          `db:"id"`
	UserID    int          `db:"user_id"`
	State     string       `db:"state"`
	Title     string       `db:"title"`
	DeletedAt dat.NullTime `db:"deleted_at"`
	CreatedAt dat.NullTime `db:"created_at"`
}

type Comment struct {
	ID        int          `db:"id"`
	UserID    int          `db:"user_id"`
	PostID    int          `db:"post_id"`
	Comment   string       `db:"comment"`
	CreatedAt dat.NullTime `db:"created_at"`
}

const seedData = `
INSERT INTO people (id, name, email) VALUES
	(1, 'Mario', 'mario@acme.com'),
	(2, 'John', 'john@acme.com'),
	(3, 'Grant', 'grant@acme.com'),
	(4, 'Tony', 'tony@acme.com'),
	(5, 'Ester', 'ester@acme.com'),
	(6, 'Reggie', 'reggie@acme.com');

INSERT INTO posts (id, user_id, title, state) VALUES
	(1, 1, 'Day 1', 'published'),
	(2, 1, 'Day 2', 'draft'),
	(3, 2, 'Apple', 'published'),
	(4, 2, 'Orange', 'draft');

INSERT INTO comments (id, user_id, post_id, comment) VALUES
	(1, 1, 1, 'A very good day'),
	(2, 2, 3, 'Yum. Apple pie.');

alter sequence people_id_seq RESTART with 100;
alter sequence posts_id_seq RESTART with 100;
alter sequence comments_id_seq RESTART with 100;
`

const createTables = `
	DROP TABLE IF EXISTS comments;
	DROP TABLE IF EXISTS posts;
	DROP TABLE IF EXISTS people;

	CREATE TABLE people (
		id SERIAL PRIMARY KEY,
		amount decimal,
		doc hstore,
		email text,
		foo text default 'bar',
		image bytea,
		key text,
		name text NOT NULL,
		created_at timestamptz default now()
	);
	CREATE TABLE posts (
		id SERIAL PRIMARY KEY,
		user_id int references people(id),
		state text,
		title text,
		deleted_at timestamptz,
		created_at timestamptz default now()
	);
	CREATE TABLE comments (
		id SERIAL PRIMARY KEY,
		user_id int references people(id),
		post_id int references posts(id),
		comment text not null,
		created_at timestamptz default now()
	);
`

func installFixtures() {
	sqlToRun := []string{
		createTables,
		seedData,
	}

	for _, v := range sqlToRun {
		_, err := testDB.Exec(v)
		if err != nil {
			logger.Fatal("Failed to execute statement", "sql", v, "err", err)
		}
	}
}
