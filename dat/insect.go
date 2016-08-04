package dat

import (
	"reflect"

	"github.com/pkg/errors"
)

// InsectBuilder inserts or selects an existing row when executed.
//
//	// Inserts new row unless there exists a record where
//	// `name='mario' AND email='mario@acme.com'`
//	conn.Insect("people").
//		Columns("name", "email").
//		Values("mario", "mario@acme.com").
//		Returning("id", "name", "email")
//
//	// Inserts unless there exists a record with ID of 1.
//	// Insect WILL NOT update the row if it exists.
//	conn.Insect("people").
//		Columns("name", "email").
//		Values("mario", "mario@acme.com").
//		Where("id=$1", 1).
//		Returning("id", "name", "email")
type InsectBuilder struct {
	Execer

	cols           []string
	err            error
	isBlacklist    bool
	isInterpolated bool
	record         interface{}
	returnings     []string
	table          string
	vals           []interface{}
	whereFragments []*whereFragment
}

// NewInsectBuilder creates a new InsectBuilder for the given table.
func NewInsectBuilder(table string) *InsectBuilder {
	if table == "" {
		logger.Error("Insect requires a table name.")
		return nil
	}
	return &InsectBuilder{table: table, isInterpolated: EnableInterpolation}
}

// Columns appends columns to insert in the statement
func (b *InsectBuilder) Columns(columns ...string) *InsectBuilder {
	return b.Whitelist(columns...)
}

// Blacklist defines a blacklist of columns and should only be used
// in conjunction with Record.
func (b *InsectBuilder) Blacklist(columns ...string) *InsectBuilder {
	b.isBlacklist = true
	b.cols = columns
	return b
}

// Whitelist defines a whitelist of columns to be inserted. To
// specify all columsn of a record use "*".
func (b *InsectBuilder) Whitelist(columns ...string) *InsectBuilder {
	b.cols = columns
	return b
}

// Values appends a set of values to the statement
func (b *InsectBuilder) Values(vals ...interface{}) *InsectBuilder {
	b.vals = vals
	return b
}

// Record pulls in values to match Columns from the record
func (b *InsectBuilder) Record(record interface{}) *InsectBuilder {
	b.record = record
	return b
}

// Returning sets the columns for the RETURNING clause
func (b *InsectBuilder) Returning(columns ...string) *InsectBuilder {
	b.returnings = columns
	return b
}

// ToSQL serialized the InsectBuilder to a SQL string
// It returns the string with placeholders and a slice of query arguments
func (b *InsectBuilder) ToSQL() (string, []interface{}, error) {
	if b.err != nil {
		return NewDatSQLErr(b.err)
	}
	var err error

	if len(b.table) == 0 {
		return NewDatSQLErr(errors.New("no table specified"))
	}
	lenCols := len(b.cols)
	if lenCols == 0 {
		return NewDatSQLErr(errors.New("no columns specified"))
	}
	if len(b.vals) == 0 && b.record == nil {
		return NewDatSQLErr(errors.New("no values or records specified"))
	}

	if b.record == nil && b.cols[0] == "*" {
		return NewDatSQLErr(errors.New(`"*" can only be used in conjunction with Record`))
	}
	if b.record == nil && b.isBlacklist {
		return NewDatSQLErr(errors.New(`Blacklist can only be used in conjunction with Record`))
	}

	cols := b.cols
	vals := b.vals
	returnings := b.returnings
	whereFragments := b.whereFragments

	// reflect fields removing blacklisted columns
	if b.record != nil && b.isBlacklist {
		cols = reflectExcludeColumns(b.record, cols)
	}
	// reflect all fields
	if b.record != nil && cols[0] == "*" {
		cols = reflectColumns(b.record)
	}

	whereAdded := false

	// build where clause from columns and values
	if len(whereFragments) == 0 && b.record == nil {
		whereAdded = true
		for i, column := range cols {
			fragment, err := newWhereFragment(column+"=$1", vals[i:i+1])
			if err != nil {
				return NewDatSQLErr(err)
			}
			whereFragments = append(whereFragments, fragment)
		}
	}

	if len(returnings) == 0 {
		returnings = cols
	}

	/*
	   END GOAL:

	   	   WITH sel AS (
	   	       SELECT id, user_name, auth_id, auth_provider
	   	       FROM users
	   	       WHERE user_name = $1 and auth_id = $2 and auth_provider = $3
	   	   ), ins AS (
	   	       INSERT INTO users (user_name, auth_id, auth_provider)
	   	       SELECT $1, $2, $3
	   	       WHERE NOT EXISTS (SELECT 1 FROM sel)
	   	       RETURNING id, user_name, auth_id, auth_provider
	   	   )
	   	   SELECT * FROM ins
	   	   UNION ALL
	   	   SELECT * FROM sel
	*/
	if b.record != nil {
		ind := reflect.Indirect(reflect.ValueOf(b.record))
		vals, err = valuesFor(ind.Type(), ind, cols)
		if err != nil {
			return NewDatSQLErr(err)
		}
	}

	buf := bufPool.Get()
	defer bufPool.Put(buf)
	var args []interface{}
	var selectSQL string

	buf.WriteString("WITH sel AS (")

	sb := NewSelectBuilder(returnings...).
		From(b.table)
	sb.whereFragments = whereFragments
	selectSQL, args, err = sb.ToSQL()
	if err != nil {
		return NewDatSQLErr(err)
	}
	buf.WriteString(selectSQL)

	buf.WriteString("), ins AS (")

	buf.WriteString(" INSERT INTO ")
	writeIdentifier(buf, b.table)
	buf.WriteString("(")
	writeIdentifiers(buf, cols, ",")
	buf.WriteString(") SELECT ")

	if whereAdded {
		writePlaceholders(buf, len(args), ",", 1)
	} else {
		writePlaceholders(buf, len(vals), ",", len(args)+1)
		args = append(args, vals...)
	}

	buf.WriteString(" WHERE NOT EXISTS (SELECT 1 FROM sel) RETURNING ")
	writeIdentifiers(buf, returnings, ",")

	buf.WriteString(") SELECT * FROM ins UNION ALL SELECT * FROM sel")

	return buf.String(), args, nil
}

// Where appends a WHERE clause to the statement for the given string and args
// or map of column/value pairs
func (b *InsectBuilder) Where(whereSQLOrMap interface{}, args ...interface{}) *InsectBuilder {
	fragment, err := newWhereFragment(whereSQLOrMap, args)
	if err != nil {
		b.err = err
		return b
	}
	b.whereFragments = append(b.whereFragments, fragment)
	return b
}
