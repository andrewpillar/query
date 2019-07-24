package query

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

type direction uint8

type order struct {
	cols []string
	dir  direction
}

type set struct {
	col string
	val interface{}
}

type statement uint8

type where struct {
	col   string
	op    string
	val   interface{}
	cat   string
	query Query
}

type Query struct {
	stmt   statement
	table  string
	cols   []string
	wheres []where
	sets   []set
	order  order
	limit  int
	offset int
	ret    []string
	args   []interface{}
	bind   int
}

const (
	asc direction = iota
	desc

	_select statement = iota
	_insert
	_update
	_delete

	OpEq     = "="
	OpNotEq  = "!="
	OpGt     = ">"
	OpGtOrEq = ">="
	OpLt     = "<"
	OpLtOrEq = "<="
	OpLike   = "LIKE"
	OpIs     = "IS"
)

func Delete(opts ...Option) Query {
	q := Query{
		stmt: _delete,
	}

	for _, opt := range opts {
		q = opt(q)
	}

	return q
}

func Insert(opts ...Option) Query {
	q := Query{
		stmt: _insert,
	}

	for _, opt := range opts {
		q = opt(q)
	}

	return q
}

func Select(opts ...Option) Query {
	q := Query{
		stmt: _select,
	}

	for _, opt := range opts {
		q = opt(q)
	}

	return q
}

func Update(opts ...Option) Query {
	q := Query{
		stmt: _update,
	}

	for _, opt := range opts {
		q = opt(q)
	}

	return q
}

func (d direction) String() string {
	switch d {
	case asc:
		return "ASC"
	case desc:
		return "DESC"
	default:
		return ""
	}
}

func (o order) isZero() bool {
	return len(o.cols) == 0 && o.dir == direction(0)
}

func (q Query) Args() []interface{} {
	return q.args
}

func (q Query) buildReturning(buf *bytes.Buffer) {
	if len(q.ret) == 0 {
		return
	}

	buf.WriteString(" RETURNING ")
	buf.WriteString(strings.Join(q.ret, ", "))
}

func (q *Query) buildWheres(buf *bytes.Buffer) {
	if len(q.wheres) == 0 {
		return
	}

	buf.WriteString(" WHERE ")

	wheres := make([]string, 0)
	end := len(q.wheres) - 1

	for i, w := range q.wheres {
		where := bytes.NewBufferString(w.col)
		where.WriteString(" ")
		where.WriteString(w.op)
		where.WriteString(" ")

		if w.val == nil {
			if !w.query.isZero() {
				w.query.bind += q.bind
				w.val = "(" + w.query.Build() + ")"
				q.bind = w.query.bind
			} else {
				if w.op == "IN" {
					in := make([]string, 0)

					for q.bind != len(q.args) {
						q.bind++
						in = append(in, param(q.bind))
					}

					w.val = "(" + strings.Join(in, ", ") + ")"
				} else {
					q.bind++
					w.val = param(q.bind)
				}
			}
		}

		fmt.Fprintf(where, "%v", w.val)

		wheres = append(wheres, where.String())

		if i != end {
			next := q.wheres[i+1]

			if next.cat != w.cat {
				buf.WriteString("(")
				buf.WriteString(strings.Join(wheres, w.cat))
				buf.WriteString(")")
				buf.WriteString(next.cat)

				wheres = make([]string, 0)
			}

			continue
		}

		buf.WriteString(strings.Join(wheres, w.cat))
	}
}

func (q Query) isZero() bool {
	return q.stmt == statement(0) &&
		q.table == "" &&
		len(q.cols) == 0 &&
		len(q.wheres) == 0 &&
		len(q.sets) == 0 &&
		len(q.wheres) == 0 &&
		len(q.sets) == 0 &&
		q.order.isZero() &&
		q.limit == 0 &&
		q.offset == 0 &&
		len(q.ret) == 0 &&
		len(q.args) == 0
}

func (q *Query) Build() string {
	buf := bytes.NewBufferString("")

	switch q.stmt {
	case _select:
		buf.WriteString("SELECT ")
		buf.WriteString(strings.Join(q.cols, ", "))
		buf.WriteString(" FROM ")
		buf.WriteString(q.table)

		q.buildWheres(buf)

		if !q.order.isZero() {
			buf.WriteString(" ORDER BY ")
			buf.WriteString(strings.Join(q.order.cols, ", "))
			buf.WriteString(" ")
			buf.WriteString(q.order.dir.String())
		}

		if q.limit > 0 {
			buf.WriteString(" LIMIT ")
			buf.WriteString(strconv.FormatInt(int64(q.limit), 10))
		}

		if q.offset > 0 {
			buf.WriteString(" OFFSET ")
			buf.WriteString(strconv.FormatInt(int64(q.offset), 10))
		}

		return buf.String()
	case _insert:
		buf.WriteString("INSERT INTO ")
		buf.WriteString(q.table)
		buf.WriteString(" (")
		buf.WriteString(strings.Join(q.cols, ", "))

		vals := make([]string, len(q.cols), len(q.cols))

		for i := range q.cols {
			vals[i] = param(i + 1)
		}

		buf.WriteString(") VALUES (")
		buf.WriteString(strings.Join(vals, ", "))
		buf.WriteString(")")

		q.buildReturning(buf)

		return buf.String()
	case _update:
		buf.WriteString("UPDATE ")
		buf.WriteString(q.table)

		sets := make([]string, len(q.sets), len(q.sets))

		for i, s := range q.sets {
			if s.val == nil {
				q.bind++
				s.val = param(q.bind)
			}

			sets[i] = fmt.Sprintf("%s = %v", s.col, s.val)
		}

		buf.WriteString(" SET ")
		buf.WriteString(strings.Join(sets, ", "))

		q.buildWheres(buf)

		if q.limit > 0 {
			buf.WriteString(" LIMIT ")
			buf.WriteString(strconv.FormatInt(int64(q.limit), 10))
		}

		if q.offset > 0 {
			buf.WriteString(" OFFSET ")
			buf.WriteString(strconv.FormatInt(int64(q.offset), 10))
		}

		q.buildReturning(buf)

		return buf.String()
	case _delete:
		buf.WriteString("DELETE FROM ")
		buf.WriteString(q.table)

		q.buildWheres(buf)

		if q.limit > 0 {
			buf.WriteString(" LIMIT ")
			buf.WriteString(strconv.FormatInt(int64(q.limit), 10))
		}

		if q.offset > 0 {
			buf.WriteString(" OFFSET ")
			buf.WriteString(strconv.FormatInt(int64(q.offset), 10))
		}

		return buf.String()
	default:
		return buf.String()
	}
}
