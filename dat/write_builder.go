package dat

import (
	"bytes"
	"errors"
	"strings"
)

func writeWith(buf *bytes.Buffer, withs []*subInfo, args *[]interface{}, placeholderStartPos *int64) {
	for i, sub := range withs {
		if i == 0 {
			buf.WriteString("WITH ")
		} else {
			buf.WriteString(", ")
		}
		buf.WriteString(sub.alias)
		buf.WriteString(" AS (")
		sub.WriteRelativeArgs(buf, args, placeholderStartPos)
		buf.WriteString(") ")
	}
}

func writeFor(buf *bytes.Buffer, fors []string) {
	if len(fors) == 0 {
		return
	}
	// add FOR clause
	buf.WriteString(" FOR")
	for _, s := range fors {
		buf.WriteString(" ")
		buf.WriteString(s)
	}

}

func writeFrom(buf *bytes.Buffer, table string) {
	if table != "" {
		buf.WriteString(" FROM ")
		buf.WriteString(table)
	}
}

func writeUnion(buf *bytes.Buffer, union *Expression, args *[]interface{}, placeholderStartPos *int64) {
	if union == nil {
		return
	}
	buf.WriteString(" UNION ")
	union.WriteRelativeArgs(buf, args, placeholderStartPos)
}

func writeLimit(buf *bytes.Buffer, isLimit bool, limit uint64) {
	if isLimit {
		buf.WriteString(" LIMIT ")
		writeUint64(buf, limit)
	}
}

func writeOffset(buf *bytes.Buffer, isOffset bool, offset uint64) {
	if isOffset {
		buf.WriteString(" OFFSET ")
		writeUint64(buf, offset)
	}
}

func writeOrderBy(buf *bytes.Buffer, orderBys []*whereFragment, args *[]interface{}, placeholderStartPos *int64) {
	if len(orderBys) > 0 {
		buf.WriteString(" ORDER BY ")
		writeCommaFragmentsToSQL(buf, orderBys, args, placeholderStartPos)
	}
}

func writeHaving(buf *bytes.Buffer, fragments []*whereFragment, args *[]interface{}, placeholderStartPos *int64) {
	if len(fragments) > 0 {
		buf.WriteString(" HAVING ")
		writeAndFragmentsToSQL(buf, fragments, args, placeholderStartPos)
	}
}

func writeGroupBy(buf *bytes.Buffer, groupBys []string) {
	if len(groupBys) > 0 {
		buf.WriteString(" GROUP BY ")
		for i, s := range groupBys {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(s)
		}
	}
}

func writeWhere(buf *bytes.Buffer, fragments []*whereFragment, args *[]interface{}, placeholderStartPos *int64) {
	if len(fragments) > 0 {
		buf.WriteString(" WHERE ")
		writeAndFragmentsToSQL(buf, fragments, args, placeholderStartPos)
	}
}

func writeScope(buf *bytes.Buffer, scope Scope, table string) ([]*whereFragment, error) {
	var result []*whereFragment

	if scope != nil {
		var where string
		sql, args2 := scope.ToSQL(table)
		sql, where = splitWhere(sql)
		buf.WriteString(sql)
		if where != "" {
			fragment, err := newWhereFragment(where, args2)
			if err != nil {
				return nil, err
			}
			result = append(result, fragment)
		}

		return result, nil
	}

	return result, nil
}

func writeColumns(buf *bytes.Buffer, columns []string) {
	for i, s := range columns {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(s)
	}
}

func writeDistinct(buf *bytes.Buffer, isDistinct bool, distinctColumns []string) {
	if isDistinct {
		if len(distinctColumns) == 0 {
			buf.WriteString("DISTINCT ")
		} else {
			buf.WriteString("DISTINCT ON (")
			for i, s := range distinctColumns {
				if i > 0 {
					buf.WriteString(", ")
				}
				buf.WriteString(s)
			}
			buf.WriteString(") ")
		}
	}
}

func writeJoins(buf *bytes.Buffer, joins []*joinExpression) error {
	for _, join := range joins {
		kind := JoinType[join.kind]
		if kind == "" {
			return errors.New("Invalid join kind")
		}

		buf.WriteString(" ")
		buf.WriteString(kind)
		buf.WriteString(" (")

		if strings.Index(join.subQuery, " ") < 0 {
			// postgres does not allow the table name in parens, expand it to full subquery
			buf.WriteString("select * from ")
		}
		buf.WriteString(join.subQuery)
		buf.WriteString(") AS ")
		buf.WriteString(join.alias)
		buf.WriteString(" ON ")
		buf.WriteString(join.expr)
	}

	return nil

}

func writeMany(buf *bytes.Buffer, subQueriesMany []*subInfo, isPreComma bool, args *[]interface{}, placeholderStartPos *int64) {
	for _, sub := range subQueriesMany {
		if isPreComma {
			buf.WriteString(", ")
		}
		buf.WriteString("(SELECT array_agg(dat__")
		buf.WriteString(sub.alias)
		buf.WriteString(".*) FROM (")
		sub.WriteRelativeArgs(buf, args, placeholderStartPos)
		buf.WriteString(") AS dat__")
		buf.WriteString(sub.alias)
		buf.WriteString(") AS ")
		writeQuotedIdentifier(buf, sub.alias)
		if !isPreComma {
			buf.WriteString(", ")
		}

	}
}

func writeVector(buf *bytes.Buffer, subQueriesVector []*subInfo, isPreComma bool, args *[]interface{}, placeholderStartPos *int64) {
	for _, sub := range subQueriesVector {
		if isPreComma {
			buf.WriteString(", ")
		}
		buf.WriteString("(SELECT array_agg(dat__")
		buf.WriteString(sub.alias)
		buf.WriteString(".dat__scalar) FROM (")
		sub.WriteRelativeArgs(buf, args, placeholderStartPos)
		buf.WriteString(") AS dat__")
		buf.WriteString(sub.alias)
		buf.WriteString("(dat__scalar)) AS ")
		writeQuotedIdentifier(buf, sub.alias)
		if !isPreComma {
			buf.WriteString(", ")
		}

	}
}

func writeOne(buf *bytes.Buffer, subQueriesOne []*subInfo, isPreComma bool, args *[]interface{}, placeholderStartPos *int64) {
	for _, sub := range subQueriesOne {
		if isPreComma {
			buf.WriteString(", ")
		}

		buf.WriteString("(SELECT row_to_json(dat__")
		buf.WriteString(sub.alias)
		buf.WriteString(".*) FROM (")
		sub.WriteRelativeArgs(buf, args, placeholderStartPos)
		buf.WriteString(") AS dat__")
		buf.WriteString(sub.alias)
		buf.WriteString(") AS ")
		writeQuotedIdentifier(buf, sub.alias)
		if !isPreComma {
			buf.WriteString(", ")
		}
	}
}

func writeScalar(buf *bytes.Buffer, subQueriesScalar []*subInfo, isPreComma bool, args *[]interface{}, placeholderStartPos *int64) {
	for _, sub := range subQueriesScalar {
		if isPreComma {
			buf.WriteString(", ")
		}

		buf.WriteString("(SELECT dat__")
		buf.WriteString(sub.alias)
		buf.WriteString(".dat__scalar FROM (")
		sub.WriteRelativeArgs(buf, args, placeholderStartPos)
		buf.WriteString(") AS dat__")
		buf.WriteString(sub.alias)
		buf.WriteString("(dat__scalar) limit 1) AS ")
		writeQuotedIdentifier(buf, sub.alias)
		if !isPreComma {
			buf.WriteString(", ")
		}

	}

}
