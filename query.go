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
	col string
	op  string
	val interface{}
	cat string
}

type Query struct {
	stmt   statement
	table  string
	cols   []string
	wheres []where
	sets   []set
	order  order
	limit  int
	ret    []string
	args   []interface{}
}

const (
	asc  direction = iota
	desc

	_select statement = iota
	_insert
	_update
	_delete
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

func (q Query) buildWheres(buf *bytes.Buffer) {
	if len(q.wheres) == 0 {
		return
	}

	buf.WriteString(" WHERE ")

	wheres := make([]string, 0)
	end := len(q.wheres) - 1

	for i, w := range q.wheres {
		wheres = append(wheres, fmt.Sprintf("%s %s %v", w.col, w.op, w.val))

		if i != end {
			next := q.wheres[i + 1]

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

func (q Query) Build() string {
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
				sets[i] = fmt.Sprintf("%s = %v", s.col, s.val)
			}

			buf.WriteString(" SET ")
			buf.WriteString(strings.Join(sets, ", "))

			q.buildWheres(buf)

			if q.limit > 0 {
				buf.WriteString(" LIMIT ")
				buf.WriteString(strconv.FormatInt(int64(q.limit), 10))
			}

			return buf.String()
		case _delete:
			buf.WriteString("DELETE FROM ")
			buf.WriteString(q.table)

			q.buildWheres(buf)

			if q.limit > 0 {
				buf.WriteString(" LIMIT ")
				buf.WriteString(strconv.FormatInt(int64(q.limit), 10))
			}

			return buf.String()
		default:
			return buf.String()
	}
}
