package query

import (
	"strconv"
	"strings"
)

type statement uint

// Option is the type for the first class functions that should be used for
// modifying a Query as it is being built. This will be passed the latest
// state of the Query, and should return that same Query once any modifications
// have been made.
type Option func(Query) Query

// Query contains the state of a Query that is being built. The only way this
// should be modified is via the use of the Option first class function.
type Query struct {
	stmt    statement
	table   string
	exprs   []Expr
	clauses []Clause
	args    []interface{}
}

//go:generate stringer -type statement -linecomment
const (
	_Stmt statement = iota //
	_Delete                // DELETE
	_Insert                // INSERT
	_Select                // SELECT
	_Update                // UPDATE
)

// Delete builds up a DELETE query on the given table applying the given
// options.
func Delete(table string, opts ...Option) Query {
	q := Query{
		stmt:  _Delete,
		table: table,
	}

	for _, opt := range opts {
		q = opt(q)
	}
	return q
}

// Insert builds up an INSERT query on the given table using the given leading
// expression, and applying the given options.
func Insert(table string, expr Expr, opts ...Option) Query {
	q := Query{
		stmt:  _Insert,
		table: table,
		exprs: []Expr{expr},
	}

	for _, opt := range opts {
		q = opt(q)
	}
	return q
}

// Select will build up a SELECT query using the given leading expression, and
// applying the given options.
func Select(expr Expr, opts ...Option) Query {
	q := Query{
		stmt:  _Select,
		exprs: []Expr{expr},
	}

	for _, opt := range opts {
		q = opt(q)
	}
	return q
}

// Update will build up an UPDATE query on the given table applying the given
// options.
func Update(table string, opts ...Option) Query {
	q := Query{
		stmt:  _Update,
		table: table,
	}

	for _, opt := range opts {
		q = opt(q)
	}
	return q
}

// Union returns a new Query that applies the UNION clause to all fo the given
// queries. This allows for multiple queries to be used within a single query.
func Union(queries ...Query) Query {
	var q0 Query

	for _, q := range queries {
		q0.args = append(q0.args, q.args...)
		q0.clauses = append(q0.clauses, unionClause{
			q: q,
		})
	}
	return q0
}

// Options applies all of the given options to the current query being built.
func Options(opts ...Option) Option {
	return func(q Query) Query {
		for _, opt := range opts {
			q = opt(q)
		}
		return q
	}
}

// conj returns the string that should be used for conjoining multiple clauses
// of the same type.
func (q Query) conj(cl Clause) string {
	if cl == nil {
		return ""
	}

	switch v := cl.(type) {
	case whereClause:
		return " " + v.conjunction + " "
	case unionClause:
		return " " + cl.Kind().String() + " "
	case setClause, valuesClause:
		return ", "
	default:
		return " "
	}
}

// buildInitial builds up the initial query using ? as the placeholder. This
// will correctly wrap the portions of the query in parenthese depending on the
// clauses in the query, and how these clauses are conjoined.
func (q Query) buildInitial() string {
	var buf strings.Builder

	buf.WriteString(q.stmt.String())

	switch q.stmt {
	case _Insert:
		buf.WriteString(" INTO " + q.table)
	case _Update:
		buf.WriteString(" " + q.table + " ")
	case _Delete:
		buf.WriteString(" FROM " + q.table + " ")
	}

	for _, expr := range q.exprs {
		buf.WriteByte(' ')

		if q.stmt == _Insert {
			buf.WriteByte('(')
		}

		buf.WriteString(expr.Build())

		if q.stmt == _Insert {
			buf.WriteByte(')')
		}
		buf.WriteByte(' ')
	}

	clauses := make(map[clauseKind]struct{})
	end := len(q.clauses) - 1

	for i, cl := range q.clauses {
		var (
			prev Clause
			next Clause
		)

		if i > 0 {
			prev = q.clauses[i-1]
		}

		if i < end {
			next = q.clauses[i+1]
		}

		kind := cl.Kind()

		if kind != _UnionClause {
			// Write the string of the clause kind only once, this avoids something
			// like multiple WHERE clauses being built into the query.
			if _, ok := clauses[kind]; !ok {
				clauses[kind] = struct{}{}

				buf.WriteString(kind.String() + " ")

				if kind == _WhereClause {
					buf.WriteByte('(')
				}
			}
		}

		buf.WriteString(cl.Build())

		if next != nil {
			conj := q.conj(next)

			// Determine if the clause needs wrapping in parentheses. We wrap
			// clauses under these conditions:
			//
			// - If the next clause is a different kind from the current one
			if next.Kind() == kind {
				wrap := false

				if prev != nil {
					// Wrap the clause in parentheses if we have a different
					// conjunction string.
					wrap = (prev.Kind() == kind) && (conj != q.conj(cl))
				}

				if wrap {
					buf.WriteByte(')')
				}

				buf.WriteString(conj)

				if wrap {
					buf.WriteByte('(')
				}
			} else {
				if kind == _WhereClause {
					buf.WriteByte(')')
				}
				buf.WriteByte(' ')
			}
		}

		if i == end && kind == _WhereClause {
			buf.WriteByte(')')
		}
	}
	return buf.String()
}

// Args returns a slice of all the arguments that have been added to the given
// query.
func (q Query) Args() []interface{} { return q.args }

// Build builds up the query. It will initially create a query using ? as the
// placeholder for arguments. Once built up it will replace the ? with $n where
// n is the number of the argument.
func (q Query) Build() string {
	s := q.buildInitial()

	query := make([]byte, 0, len(s))
	param := int64(0)

	for i := strings.Index(s, "?"); i != -1; i = strings.Index(s, "?") {
		param++

		query = append(query, s[:i]...)
		query = append(query, '$')
		query = strconv.AppendInt(query, param, 10)

		s = s[i+1:]
	}
	return string(append(query, []byte(s)...))
}
