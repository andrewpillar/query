package query

import (
	"bytes"
	"strconv"
	"strings"
)

type Option func(q Query) Query

type Query struct {
	stmt    statement
	clauses []clause
	args    []interface{}
}

type statement uint8

const (
	none_ statement = iota
	select_
	insert_
	update_
	delete_
)

var paren map[clauseKind]struct{} = map[clauseKind]struct{}{
	where_: struct{}{},
}

func Delete(opts ...Option) Query {
	q := Query{
		stmt: delete_,
	}

	for _, opt := range opts {
		q = opt(q)
	}

	return q
}

func Insert(opts ...Option) Query {
	q := Query{
		stmt: insert_,
	}

	for _, opt := range opts {
		q = opt(q)
	}

	return q
}

func Select(opts ...Option) Query {
	q := Query{
		stmt: select_,
	}

	for _, opt := range opts {
		q = opt(q)
	}

	return q
}

func Update(opts ...Option) Query {
	q := Query{
		stmt: update_,
	}

	for _, opt := range opts {
		q = opt(q)
	}

	return q
}

func (q Query) Args() []interface{} {
	return q.args
}

func (q Query) buildInitial() string {
	buf := &bytes.Buffer{}

	switch q.stmt {
	case select_:
		buf.WriteString("SELECT ")
		break
	case insert_:
		buf.WriteString("INSERT ")
		break
	case update_:
		buf.WriteString("UPDATE ")
		break
	case delete_:
		buf.WriteString("DELETE ")
		break
	}

	clauses := make(map[clauseKind]struct{})
	end := len(q.clauses) - 1

	for i, c := range q.clauses {
		var (
			prev clause
			next clause
		)

		if i > 0 {
			prev = q.clauses[i - 1]
		}

		if i < end {
			next = q.clauses[i + 1]
		}

		kind := c.kind()

		if _, ok := clauses[kind]; !ok {
			clauses[kind] = struct{}{}

			kind.build(buf)

			if _, ok := paren[kind]; ok {
				buf.WriteString("(")
			}
		}

		c.build(buf)

		if next != nil {
			cat := next.cat()

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
				buf.WriteString(" ")
			}

			if next.kind() != kind {
				if _, ok := paren[kind]; ok {
					buf.WriteString(") ")
				}
			}
		}

		if i == end {
			if _, ok := paren[kind]; ok {
				buf.WriteString(")")
			}
		}
	}

	return buf.String()
}

func (q Query) Build() string {
	built := q.buildInitial()

	query := make([]byte, 0, len(built))
	param := int64(0)

	for i := strings.Index(built, "?"); i != -1; i = strings.Index(built, "?") {
		param++

		query = append(query, built[:i]...)
		query = append(query, '$')
		query = strconv.AppendInt(query, param, 10)

		built = built[i + 1:]
	}

	query = append(query, []byte(strings.TrimPrefix(built, " "))...)

	return string(query)
}
