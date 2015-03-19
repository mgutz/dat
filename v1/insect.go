package dat

import (
	"reflect"

	"github.com/mgutz/str"
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

	isInterpolated bool
	table          string
	cols           []string
	isBlacklist    bool
	vals           []interface{}
	record         interface{}
	returnings     []string
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
func (b *InsectBuilder) ToSQL() (string, []interface{}) {
	if len(b.table) == 0 {
		panic("no table specified")
	}
	lenCols := len(b.cols)
	if lenCols == 0 {
		panic("no columns specified")
	}
	if len(b.vals) == 0 && b.record == nil {
		panic("no values or records specified")
	}

	if b.record == nil && b.cols[0] == "*" {
		panic(`"*" can only be used in conjunction with Record`)
	}
	if b.record == nil && b.isBlacklist {
		panic(`Blacklist can only be used in conjunction with Record`)
	}

	// reflect fields removing blacklisted columns
	if b.record != nil && b.isBlacklist {
		info := reflectFields(b.record)
		lenFields := len(info.fields)
		cols := []string{}
		for i := 0; i < lenFields; i++ {
			f := info.fields[i]
			if str.SliceContains(b.cols, f.dbName) {
				continue
			}
			cols = append(cols, f.dbName)
		}
		b.cols = cols
	}
	// reflect all fields
	if b.record != nil && b.cols[0] == "*" {
		info := reflectFields(b.record)
		b.cols = info.Columns()
	}

	whereAdded := false

	// build where clause from columns and values
	if len(b.whereFragments) == 0 && b.record == nil {
		whereAdded = true
		for i, column := range b.cols {
			b.whereFragments = append(b.whereFragments, newWhereFragment(column+"=$1", b.vals[i:i+1]))
		}
	}

	if len(b.returnings) == 0 {
		b.returnings = b.cols
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
		var err error
		b.vals, err = ValuesFor(ind.Type(), ind, b.cols)
		if err != nil {
			panic(err.Error())
		}
	}

	buf := bufPool.Get()
	defer bufPool.Put(buf)
	var args []interface{}
	var selectSQL string

	buf.WriteString("WITH sel AS (")

	sb := NewSelectBuilder(b.returnings...).
		From(b.table)
	sb.whereFragments = b.whereFragments
	selectSQL, args = sb.ToSQL()
	buf.WriteString(selectSQL)

	buf.WriteString("), ins AS (")

	buf.WriteString(" INSERT INTO ")
	writeIdentifier(buf, b.table)
	buf.WriteString("(")
	writeIdentifiers(buf, b.cols, ",")
	buf.WriteString(") SELECT ")

	if whereAdded {
		writePlaceholders(buf, len(args), ",", 1)
	} else {
		writePlaceholders(buf, len(b.vals), ",", len(args)+1)
		args = append(args, b.vals...)
	}

	buf.WriteString(" WHERE NOT EXISTS (SELECT 1 FROM sel) RETURNING ")
	writeIdentifiers(buf, b.returnings, ",")

	buf.WriteString(") SELECT * FROM ins UNION ALL SELECT * FROM sel")

	return buf.String(), args
}

// Where appends a WHERE clause to the statement for the given string and args
// or map of column/value pairs
func (b *InsectBuilder) Where(whereSqlOrMap interface{}, args ...interface{}) *InsectBuilder {
	b.whereFragments = append(b.whereFragments, newWhereFragment(whereSqlOrMap, args))
	return b
}
