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
