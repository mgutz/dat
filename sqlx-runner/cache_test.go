package runner

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/syreclabs/dat"
	"gopkg.in/stretchr/testify.v1/assert"
)

func TestCachePre(t *testing.T) {
	installFixtures()
}

func TestCacheSelectDocBytes(t *testing.T) {
	Cache.FlushDB()
	for i := 0; i < 2; i++ {
		b, err := testDB.SelectDoc("id", "comment").
			From("comments").
			OrderBy("id").
			Cache("selectdoc.1", 1*time.Second, false).
			QueryJSON()

		assert.NoError(t, err)

		var comments jo.Object
		err = json.Unmarshal(b, &comments)
		assert.NoError(t, err)

		assert.Equal(t, "A very good day", comments.MustString("[0].comment"))
		assert.Equal(t, "Yum. Apple pie.", comments.MustString("[1].comment"))
	}
}

func TestCacheSelectDocBytesByHash(t *testing.T) {
	Cache.FlushDB()
	for i := 0; i < 2; i++ {
		b, err := testDB.SelectDoc("id", "comment").
			From("comments").
			OrderBy("id").
			Cache("", 1*time.Second, false).
			QueryJSON()

		assert.NoError(t, err)

		var comments jo.Object
		err = json.Unmarshal(b, &comments)
		assert.NoError(t, err)

		assert.Equal(t, "A very good day", comments.MustString("[0].comment"))
		assert.Equal(t, "Yum. Apple pie.", comments.MustString("[1].comment"))
	}
}

func TestCacheSelectDocObject(t *testing.T) {
	Cache.FlushDB()
	for i := 0; i < 2; i++ {
		var comments jo.Object
		err := testDB.SelectDoc("id", "comment").
			From("comments").
			OrderBy("id").
			Cache("selectdoc.2", 1*time.Second, false).
			QueryObject(&comments)

		assert.NoError(t, err)
		assert.Equal(t, "A very good day", comments.MustString("[0].comment"))
		assert.Equal(t, "Yum. Apple pie.", comments.MustString("[1].comment"))
	}
}

func TestCacheSelectDocNested(t *testing.T) {
	Cache.FlushDB()
	assert := assert.New(t)
	for i := 0; i < 2; i++ {

		var obj jo.Object

		posts := dat.SelectDoc("id", "title").
			Many("comments", `SELECT * FROM comments WHERE comments.id = posts.id`).
			From("posts").
			Where("user_id = people.id")

		err := testDB.
			SelectDoc("id", "name").
			Many("posts", posts).
			From("people").
			Where("id = $1", 1).
			SetIsInterpolated(true).
			Cache("selectdoc.3", 1*time.Second, false).
			QueryStruct(&obj)

		assert.NoError(err)
		assert.Equal("Mario", obj.AsString("name"))
		assert.Equal(1, obj.AsInt("id"))

		assert.Equal("A very good day", obj.AsString("posts[0].comments[0].comment"))
		assert.Equal("Yum. Apple pie.", obj.AsString("posts[1].comments[0].comment"))
	}
}

func TestCacheSelectQueryStructs(t *testing.T) {
	Cache.FlushDB()
	for i := 0; i < 2; i++ {
		var people []Person
		err := testDB.
			Select("id", "name", "email").
			From("people").
			OrderBy("id ASC").
			Cache("selectdoc.4", 1*time.Second, false).
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
	}
}

func TestCacheSelectQueryStruct(t *testing.T) {
	Cache.FlushDB()
	for i := 0; i < 2; i++ {

		// Found:
		var person Person
		err := testDB.
			Select("id", "name", "email").
			From("people").
			Where("email = $1", "john@acme.com").
			Cache("selectdoc.5", 1*time.Second, false).
			QueryStruct(&person)
		assert.NoError(t, err)
		assert.True(t, person.ID > 0)
		assert.Equal(t, person.Name, "John")
		assert.True(t, person.Email.Valid)
		assert.Equal(t, person.Email.String, "john@acme.com")

		// Not found:
		var person2 Person
		err = testDB.
			Select("id", "name", "email").
			From("people").Where("email = $1", "dontexist@acme.com").
			Cache("selectdoc.6", 1*time.Second, false).
			QueryStruct(&person2)
		assert.Contains(t, err.Error(), "no rows")
	}
}

func TestCacheSelectBySqlQueryStructs(t *testing.T) {
	Cache.FlushDB()
	for i := 0; i < 2; i++ {
		var people []*Person
		dat.EnableInterpolation = true
		err := testDB.
			SQL("SELECT name FROM people WHERE email IN $1", []string{"john@acme.com"}).
			Cache("selectdoc.7", 1*time.Second, false).
			QueryStructs(&people)
		dat.EnableInterpolation = false

		assert.NoError(t, err)
		assert.Equal(t, len(people), 1)
		assert.Equal(t, people[0].Name, "John")
		assert.Equal(t, people[0].ID, int64(0))       // not set
		assert.Equal(t, people[0].Email.Valid, false) // not set
		assert.Equal(t, people[0].Email.String, "")   // not set
	}
}

func TestCacheSelectQueryScalar(t *testing.T) {
	Cache.FlushDB()
	for i := 0; i < 2; i++ {

		var name string
		err := testDB.
			Select("name").
			From("people").
			Where("email = 'john@acme.com'").
			Cache("selectdoc.8", 1*time.Second, false).
			QueryScalar(&name)

		assert.NoError(t, err)
		assert.Equal(t, name, "John")

		var id int64
		err = testDB.
			Select("id").
			From("people").
			Limit(1).
			Cache("selectdoc.9", 1*time.Second, false).
			QueryScalar(&id)

		assert.NoError(t, err)
		assert.True(t, id > 0)
	}
}

func TestCacheSelectQuerySlice(t *testing.T) {
	Cache.FlushDB()
	for i := 0; i < 2; i++ {
		var names []string
		err := testDB.
			Select("name").
			From("people").
			Cache("selectdoc.10", 1*time.Second, false).
			QuerySlice(&names)

		assert.NoError(t, err)
		assert.Equal(t, len(names), 6)
		assert.Equal(t, names, []string{"Mario", "John", "Grant", "Tony", "Ester", "Reggie"})

		var ids []int64
		err = testDB.
			Select("id").
			From("people").
			Limit(1).
			Cache("selectdoc.11", 1*time.Second, false).
			QuerySlice(&ids)

		assert.NoError(t, err)
		assert.Equal(t, len(ids), 1)
		assert.Equal(t, ids, []int64{1})
	}
}

func TestCacheSelectQuerySliceByHash(t *testing.T) {
	Cache.FlushDB()
	for i := 0; i < 2; i++ {
		var names []string
		err := testDB.
			Select("name").
			From("people").
			Cache("", 1*time.Second, false).
			QuerySlice(&names)

		assert.NoError(t, err)
		assert.Equal(t, len(names), 6)
		assert.Equal(t, names, []string{"Mario", "John", "Grant", "Tony", "Ester", "Reggie"})

		var ids []int64
		err = testDB.
			Select("id").
			From("people").
			Limit(1).
			Cache("", 1*time.Second, false).
			QuerySlice(&ids)

		assert.NoError(t, err)
		assert.Equal(t, len(ids), 1)
		assert.Equal(t, ids, []int64{1})
	}
}
