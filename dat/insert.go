package dat

import (
	"bytes"
	"reflect"

	"github.com/pkg/errors"
	"github.com/mgutz/str"
	"fmt"
)

// InsertBuilder contains the clauses for an INSERT statement
type InsertBuilder struct {
	Execer

	isInterpolated bool
	table          string
	cols           []string
	conflict       *ConflictAction
	isBlacklist    bool
	vals           [][]interface{}
	records        []interface{}
	returnings     []string
	err            error
}

type ConflictAction struct {
	Target         string
	Action         string
	UpdateExcludes []string
	UpdateColumns  []string
}

// NewInsertBuilder creates a new InsertBuilder for the given table.
func NewInsertBuilder(table string) *InsertBuilder {
	if table == "" {
		logger.Error("InsertInto requires a table name.")
		return nil
	}
	return &InsertBuilder{table: table, isInterpolated: EnableInterpolation}
}

// Columns appends columns to insert in the statement
func (b *InsertBuilder) Columns(columns ...string) *InsertBuilder {
	return b.Whitelist(columns...)
}

// Blacklist defines a blacklist of columns and should only be used
// in conjunction with Record.
func (b *InsertBuilder) Blacklist(columns ...string) *InsertBuilder {
	b.isBlacklist = true
	b.cols = columns
	return b
}

// Whitelist defines a whitelist of columns to be inserted. To
// specify all columns of a record use "*".
func (b *InsertBuilder) Whitelist(columns ...string) *InsertBuilder {
	b.cols = columns
	return b
}

// Values appends a set of values to the statement
func (b *InsertBuilder) Values(vals ...interface{}) *InsertBuilder {
	b.vals = append(b.vals, vals)
	return b
}

// Record pulls in values to match Columns from the record
func (b *InsertBuilder) Record(record interface{}) *InsertBuilder {
	b.records = append(b.records, record)
	return b
}

// OnConflict sets ON CONFLICT clause. Only supported on Postgres 9.5+. Unfortunately,
// this dat package is just a builder and version cannot be checked.
func (b *InsertBuilder) OnConflict(target string, action string) *InsertBuilder {
	b.conflict = &ConflictAction{
		Target: target,
		Action: action,
	}
	return b
}

func (b *InsertBuilder) OnConflictUpdate(target string, columns ...string) *InsertBuilder {
	b.conflict = &ConflictAction{
		Target:        target,
		UpdateColumns: columns,
	}
	return b
}

func (b *InsertBuilder) OnConflictUpdateExclude(target string, excludeColumns ...string) *InsertBuilder {
	b.conflict = &ConflictAction{
		Target:         target,
		UpdateExcludes: excludeColumns,
	}
	return b
}

// Returning sets the columns for the RETURNING clause
func (b *InsertBuilder) Returning(columns ...string) *InsertBuilder {
	b.returnings = columns
	return b
}

// Pair adds a key/value pair to the statement
func (b *InsertBuilder) Pair(column string, value interface{}) *InsertBuilder {
	b.cols = append(b.cols, column)
	lenVals := len(b.vals)
	if lenVals == 0 {
		args := []interface{}{value}
		b.vals = [][]interface{}{args}
	} else if lenVals == 1 {
		b.vals[0] = append(b.vals[0], value)
	} else {
		b.err = errors.New("pair only allows you to specify 1 record to insert")
	}
	return b
}

// ToSQL serialized the InsertBuilder to a SQL string
// It returns the string with placeholders and a slice of query arguments
func (b *InsertBuilder) ToSQL() (string, []interface{}, error) {
	if b.err != nil {
		return "", nil, b.err
	}

	if len(b.table) == 0 {
		return "", nil, NewError("no table specified")
	}
	lenCols := len(b.cols)
	lenRecords := len(b.records)
	if lenCols == 0 {
		return "", nil, NewError("no columns specified")
	}
	if len(b.vals) == 0 && lenRecords == 0 {
		return "", nil, NewError("no values or records specified")
	}

	if lenRecords == 0 && b.cols[0] == "*" {
		return "", nil, NewError(`"*" can only be used in conjunction with Record`)
	}

	if lenRecords == 0 && b.isBlacklist {
		return "", nil, NewError("Blacklist can only be used in conjunction with Record")
	}

	cols := b.cols

	// reflect fields removing blacklisted columns
	if lenRecords > 0 && b.isBlacklist {
		cols = reflectExcludeColumns(b.records[0], cols)
	}
	// reflect all fields
	if lenRecords > 0 && cols[0] == "*" {
		cols = reflectColumns(b.records[0])
	}

	var sql bytes.Buffer
	var args []interface{}

	sql.WriteString("INSERT INTO ")
	sql.WriteString(b.table)
	sql.WriteString(" (")

	for i, c := range cols {
		if i > 0 {
			sql.WriteRune(',')
		}
		writeIdentifier(&sql, c)
	}
	sql.WriteString(") VALUES ")

	start := 1
	// Go thru each value we want to insert. Write the placeholders, and collect args
	for i, row := range b.vals {
		if i > 0 {
			sql.WriteRune(',')
		}
		buildPlaceholders(&sql, start, len(row))

		for _, v := range row {
			args = append(args, v)
			start++
		}
	}
	anyVals := len(b.vals) > 0

	// Go thru the records. Write the placeholders, and do reflection on the records to extract args
	for i, rec := range b.records {
		if i > 0 || anyVals {
			sql.WriteRune(',')
		}

		ind := reflect.Indirect(reflect.ValueOf(rec))
		vals, err := valuesFor(ind.Type(), ind, cols)
		if err != nil {
			return "", nil, err
		}
		buildPlaceholders(&sql, start, len(vals))
		for _, v := range vals {
			args = append(args, v)
			start++
		}
	}

	if b.conflict != nil {
		sql.WriteString(" ON CONFLICT ")
		sql.WriteString(b.conflict.Target)
		sql.WriteString(" DO ")
		var updateColumns []string
		if b.conflict.Action != "" {
			sql.WriteString(b.conflict.Action)
		} else if len(b.conflict.UpdateColumns) > 0 {
			if b.conflict.UpdateColumns[0] == "*" {
				updateColumns = cols
			} else {
				for _, c := range b.conflict.UpdateColumns {
					if !str.SliceContains(cols, c) {
						return "", nil, fmt.Errorf(`column "%s" is not in insert columns`, c)
					}
					updateColumns = append(updateColumns, c)
				}
			}
		} else if len(b.conflict.UpdateExcludes) > 0 {
			for _, c := range cols {
				if str.SliceContains(b.conflict.UpdateExcludes, c) {
					continue
				}
				updateColumns = append(updateColumns, c)
			}
		} else {
			sql.WriteString("NOTHING")
		}
		if len(updateColumns) > 0 {
			sql.WriteString("UPDATE SET ")
			first := true
			for _, c := range updateColumns {
				if !first {
					sql.WriteRune(',')
				}
				sql.WriteString(c)
				sql.WriteString(" = EXCLUDED.")
				sql.WriteString(c)
				first = false
			}
		}
	}
	// Go thru the returning clauses
	for i, c := range b.returnings {
		if i == 0 {
			sql.WriteString(" RETURNING ")
		} else {
			sql.WriteRune(',')
		}
		writeIdentifier(&sql, c)
	}

	return sql.String(), args, nil
}
