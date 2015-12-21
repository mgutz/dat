package dat

import (
	"testing"
	"time"

	"gopkg.in/stretchr/testify.v1/assert"
)

func TestIssue26(t *testing.T) {

	type Model struct {
		ID        int64     `json:"id" db:"id"`
		CreatedAt time.Time `json:"createdAt" db:"created_at"`
		UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
	}

	type Customer struct {
		Model
		First string `json:"first" db:"first"`
		Last  string `json:"last" db:"last"`
	}

	customer := Customer{}
	sql, args :=
		Update("customers").
			SetBlacklist(customer, "id", "created_at", "updated_at").
			Where("id = $1", customer.ID).
			Returning("updated_at").ToSQL()

	assert.Equal(t, sql, `UPDATE "customers" SET "first" = $1, "last" = $2 WHERE (id = $3) RETURNING "updated_at"`)
	assert.Exactly(t, args, []interface{}{"", "", int64(0)})
}

func TestIssue29(t *testing.T) {
	sql, args := Select("a").From("people").Where("email = $1", "foo@acme.com").OrderBy("people.name <-> $1", "foo").ToSQL()
	assert.Equal(t, sql, `SELECT a FROM people WHERE (email = $1) ORDER BY people.name <-> $2`)
	assert.Exactly(t, args, []interface{}{"foo@acme.com", "foo"})

	sql2, _, err := Interpolate(sql, args)
	assert.NoError(t, err)
	assert.Equal(t, stripWS(`SELECT a FROM people WHERE (email = 'foo@acme.com') ORDER BY people.name <-> 'foo'`), stripWS(sql2))
}
