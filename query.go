package query

import (
	"bytes"
	"strconv"
	"strings"
)

// Option provided query option.
type Option func(q Query) Query

// Query creates a statement query.
type Query struct {
	stmt    statement
	clauses []clause
	args    []interface{}
}

type statement uint8

const (
	noneStmt statement = iota
	selectStmt
	insertStmt
	updateStmt
	deleteStmt
)

var parentheses map[clauseKind]struct{} = map[clauseKind]struct{}{
	whereKind: {},
	countKind: {},
}

// Delete creates a DELETE statement query.
func Delete(opts ...Option) Query {
	q := Query{
		stmt: deleteStmt,
	}

	for _, opt := range opts {
		q = opt(q)
	}

	return q
}

// Insert creates an INSERT statement query.
func Insert(opts ...Option) Query {
	q := Query{
		stmt: insertStmt,
	}

	for _, opt := range opts {
		q = opt(q)
	}

	return q
}

// Select creates a SELECT statement query.
func Select(opts ...Option) Query {
	q := Query{
		stmt: selectStmt,
	}

	for _, opt := range opts {
		q = opt(q)
	}

	return q
}

// Union adds the UNION clause to each of the queries given.
func Union(queries ...Query) Query {
	q := Query{
		stmt: noneStmt,
	}

	for _, qry := range queries {
		u := union{
			query: qry,
		}

		q.args = append(q.args, qry.args...)
		q.clauses = append(q.clauses, u)
	}

	return q
}

// Update creates an UPDATE statement query.
func Update(opts ...Option) Query {
	q := Query{
		stmt: updateStmt,
	}

	for _, opt := range opts {
		q = opt(q)
	}

	return q
}

// Args returns the arguments set for the given query.
func (q Query) Args() []interface{} {
	return q.args
}

func (q Query) buildInitial() string {
	buf := &bytes.Buffer{}

	switch q.stmt {
	case selectStmt:
		buf.WriteString("SELECT ")
	case insertStmt:
		buf.WriteString("INSERT ")
	case updateStmt:
		buf.WriteString("UPDATE ")
	case deleteStmt:
		buf.WriteString("DELETE ")
	}

	clauses := make(map[clauseKind]struct{})
	end := len(q.clauses) - 1

	for i, c := range q.clauses {
		var (
			prev clause
			next clause
		)

		if i > 0 {
			prev = q.clauses[i-1]
		}

		if i < end {
			next = q.clauses[i+1]
		}

		kind := c.kind()

		// Build clauses in the query once.
		if _, ok := clauses[kind]; !ok {
			if kind != countKind {
				clauses[kind] = struct{}{}
			}

			kind.build(buf)

			if _, ok := parentheses[kind]; ok {
				buf.WriteString("(")
			}
		}

		c.build(buf)

		if next != nil {
			cat := next.cat()

			// Determine if the clause needs wrapping in parentheses. We wrap
			// clauses under these conditions:
			//
			//   - If the previous clause is the same, but the concatenation
			//     string is different.
			//   - If the next clause is a different kind from the current one.
			if next.kind() == kind {
				wrap := prev != nil && prev.kind() == kind && cat != c.cat()

				if wrap {
					buf.WriteString(")")
				}

				buf.WriteString(cat)

				if wrap {
					buf.WriteString("(")
				}
			} else {
				if _, ok := parentheses[kind]; ok {
					buf.WriteString(")")
				}

				buf.WriteString(" ")
			}
		}

		if i == end {
			if _, ok := parentheses[kind]; ok {
				buf.WriteString(")")
			}
		}
	}

	return buf.String()
}

// BuildMySQL the final query string and return it.
func (q Query) BuildMySQL() string {
	built := q.buildInitial()
	return built
}

// Build the final query string and return it. This will replace all
// placeholder values in the query '?', with the respective PostgreSQL bind
// param.
func (q Query) Build() string {
	built := q.buildInitial()

	query := make([]byte, 0, len(built))
	param := int64(0)

	for i := strings.Index(built, "?"); i != -1; i = strings.Index(built, "?") {
		param++

		query = append(query, built[:i]...)
		query = append(query, '$')
		query = strconv.AppendInt(query, param, 10)

		built = built[i+1:]
	}

	query = append(query, []byte(strings.TrimPrefix(built, " "))...)

	return string(query)
}
