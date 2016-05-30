package runner

import (
	"testing"

	"github.com/syreclabs/dat"
	"gopkg.in/stretchr/testify.v1/assert"
)

func TestSelectQueryEmbedded(t *testing.T) {
	s := beginTxWithFixtures()
	defer s.AutoRollback()

	type PostEmbedded struct {
		ID    int    `db:"id"`
		State string `db:"state"`
		User  struct {
			ID int64 `db:"author_id"`
		}
	}

	type User struct {
		ID int64 `db:"user_id"`
	}

	type PostEmbedded2 struct {
		ID    int    `db:"id"`
		State string `db:"state"`
		User
	}

	var post2 PostEmbedded2

	// THIS RESULTS IN ERROR
	// var post PostEmbedded
	// err := s.Select("id", "state", "42 as user_id").
	// 	From("posts").
	// 	Where("id = $1", 1).
	// 	QueryStruct(&post)

	// assert.Error(t, err)

	err := s.Select("id", "state", "42 as user_id").
		From("posts").
		Where("id = $1", 1).
		QueryStruct(&post2)

	assert.NoError(t, err)
	assert.Equal(t, 1, post2.ID)
	assert.EqualValues(t, 42, post2.User.ID)
}

func TestSelectQueryStructs(t *testing.T) {
	s := beginTxWithFixtures()
	defer s.AutoRollback()

	var people []Person
	err := s.
		Select("id", "name", "email").
		From("people").
		OrderBy("id ASC").
		QueryStructs(&people)

	assert.NoError(t, err)
	assert.Equal(t, len(people), 6)

	// Make sure that the Ids are set. It's possible (maybe?) that different DBs set ids differently so
	// don't assume they're 1 and 2.
	assert.True(t, people[0].ID > 0)
	assert.True(t, people[1].ID > people[0].ID)

	assert.Equal(t, people[0].Name, "Mario")
	assert.True(t, people[0].Email.Valid)
	assert.Equal(t, people[0].Email.String, "mario@acme.com")
	assert.Equal(t, people[1].Name, "John")
	assert.True(t, people[1].Email.Valid)
	assert.Equal(t, people[1].Email.String, "john@acme.com")

	// TODO: test map
}

func TestSelectQueryStruct(t *testing.T) {
	s := beginTxWithFixtures()
	defer s.AutoRollback()

	// Found:
	var person Person
	err := s.
		Select("id", "name", "email").
		From("people").
		Where("email = $1", "john@acme.com").
		QueryStruct(&person)
	assert.NoError(t, err)
	assert.True(t, person.ID > 0)
	assert.Equal(t, person.Name, "John")
	assert.True(t, person.Email.Valid)
	assert.Equal(t, person.Email.String, "john@acme.com")

	// Not found:
	var person2 Person
	err = s.
		Select("id", "name", "email").
		From("people").Where("email = $1", "dontexist@acme.com").
		QueryStruct(&person2)
	assert.Contains(t, err.Error(), "no rows")
}

func TestSelectQueryDistinctOn(t *testing.T) {
	s := beginTxWithFixtures()
	defer s.AutoRollback()

	// Found:
	var person Person
	err := s.
		Select("id", "name", "email").
		DistinctOn("id").
		From("people").
		Where("email = $1", "john@acme.com").
		QueryStruct(&person)
	assert.NoError(t, err)
	assert.True(t, person.ID > 0)
	assert.Equal(t, person.Name, "John")
	assert.True(t, person.Email.Valid)
	assert.Equal(t, person.Email.String, "john@acme.com")

	// Not found:
	var person2 Person
	err = s.
		Select("id", "name", "email").
		From("people").Where("email = $1", "dontexist@acme.com").
		QueryStruct(&person2)
	assert.Contains(t, err.Error(), "no rows")
}

func TestSelectBySqlQueryStructs(t *testing.T) {
	s := beginTxWithFixtures()
	defer s.AutoRollback()

	var people []*Person
	dat.EnableInterpolation = true
	err := s.
		SQL("SELECT name FROM people WHERE email IN $1", []string{"john@acme.com"}).
		QueryStructs(&people)
	dat.EnableInterpolation = false

	assert.NoError(t, err)
	assert.Equal(t, len(people), 1)
	assert.Equal(t, people[0].Name, "John")
	assert.EqualValues(t, people[0].ID, 0)        // not set
	assert.Equal(t, people[0].Email.Valid, false) // not set
	assert.Equal(t, people[0].Email.String, "")   // not set
}

func TestSelectQueryScalar(t *testing.T) {
	s := beginTxWithFixtures()
	defer s.AutoRollback()

	var name string
	err := s.
		Select("name").
		From("people").
		Where("email = 'john@acme.com'").
		QueryScalar(&name)

	assert.NoError(t, err)
	assert.Equal(t, name, "John")

	var id int64
	err = s.Select("id").From("people").Limit(1).QueryScalar(&id)

	assert.NoError(t, err)
	assert.True(t, id > 0)
}

func TestSelectQuerySlice(t *testing.T) {
	s := beginTxWithFixtures()
	defer s.AutoRollback()

	var names []string
	err := s.Select("name").From("people").QuerySlice(&names)

	assert.NoError(t, err)
	assert.Equal(t, len(names), 6)
	assert.Equal(t, names, []string{"Mario", "John", "Grant", "Tony", "Ester", "Reggie"})

	var ids []int64
	err = s.Select("id").From("people").Limit(1).QuerySlice(&ids)

	assert.NoError(t, err)
	assert.Equal(t, len(ids), 1)
	assert.Equal(t, ids, []int64{1})
}

func TestScalar(t *testing.T) {
	s := beginTxWithFixtures()
	defer s.AutoRollback()

	var name string
	err := s.Select("name").From("people").Where("email = 'john@acme.com'").QueryScalar(&name)
	assert.NoError(t, err)
	assert.Equal(t, name, "John")

	var count int64
	err = s.Select("COUNT(*)").From("people").QueryScalar(&count)
	assert.NoError(t, err)
	assert.EqualValues(t, count, 6)
}

func TestSelectExpr(t *testing.T) {
	s := beginTxWithFixtures()
	defer s.AutoRollback()

	var name string
	scope := dat.Expr("email = $1", "john@acme.com")
	err := s.Select("name").From("people").Where(scope).QueryScalar(&name)
	assert.NoError(t, err)
	assert.Equal(t, name, "John")

	var count int64
	err = s.Select("COUNT(*)").From("people").QueryScalar(&count)
	assert.NoError(t, err)
	assert.EqualValues(t, count, 6)
}

func TestSelectScope(t *testing.T) {
	s := beginTxWithFixtures()
	defer s.AutoRollback()

	var id int
	err := s.
		InsertInto("people").
		Columns("name").
		Values("mgutz").
		Returning("id").
		QueryScalar(&id)
	assert.NoError(t, err)
	assert.True(t, id > 0)

	var postID int
	err = s.
		InsertInto("posts").
		Columns("title", "state", "user_id").
		Values("my post", "published", id).
		Returning("id").
		QueryScalar(&postID)
	assert.NoError(t, err)
	assert.True(t, postID > 0)

	publishedByUser := `
		INNER JOIN people P on (P.id = posts.user_id)
		WHERE
			P.name = $1 AND
			posts.state = 'published' AND
			posts.deleted_at IS NULL
		`
	var posts []*Post
	err = s.
		Select("posts.*").
		From("posts").
		Scope(publishedByUser, "mgutz").
		QueryStructs(&posts)
	assert.NoError(t, err)
	assert.Equal(t, posts[0].Title, "my post")
}

func TestSelectScoped(t *testing.T) {
	s := beginTxWithFixtures()
	defer s.AutoRollback()

	var id int
	err := s.
		InsertInto("people").
		Columns("name").
		Values("mgutz").
		Returning("id").
		QueryScalar(&id)
	assert.NoError(t, err)
	assert.True(t, id > 0)

	var postID int
	err = s.
		InsertInto("posts").
		Columns("title", "state", "user_id").
		Values("my post", "published", id).
		Returning("id").
		QueryScalar(&postID)
	assert.NoError(t, err)
	assert.True(t, postID > 0)

	publishedByUser := `
		INNER JOIN people on (people.id = p.user_id)
		WHERE
			people.name = $1 AND
			p.state = 'published' AND
			p.deleted_at IS NULL
	`

	var posts []*Post
	err = s.
		Select("p.*").
		From("posts p").
		Scope(publishedByUser, "mgutz").
		QueryStructs(&posts)
	assert.NoError(t, err)
	assert.Equal(t, posts[0].Title, "my post")
}

// Series of tests that test mapping struct fields to columns
