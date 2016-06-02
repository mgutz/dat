package runner

import (
	"testing"

	"gopkg.in/mgutz/dat.v1"
	"gopkg.in/stretchr/testify.v1/assert"
)

func TestRealJSON(t *testing.T) {
	j, _ := dat.NewJSON([]int{1, 2, 3})
	var num int
	var err error
	if testDB.Version > 90400 {
		err = testDB.SQL("select $1::json->1", j).QueryScalar(&num)
	} else {
		err = testDB.SQL("select $1->1", j).QueryScalar(&num)
	}

	assert.NoError(t, err)
	assert.Equal(t, 2, num)
}

func TestRealJSONInterpolated(t *testing.T) {
	j, _ := dat.NewJSON([]int{1, 2, 3})
	var num int
	var err error
	if testDB.Version > 90400 {
		err = testDB.SQL("select $1::json->1", j).SetIsInterpolated(true).QueryScalar(&num)
	} else {
		err = testDB.SQL("select $1->1", j).SetIsInterpolated(true).QueryScalar(&num)
	}
	assert.NoError(t, err)
	assert.Equal(t, 2, num)
}
