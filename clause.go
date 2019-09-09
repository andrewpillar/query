package query

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

type alias struct {
	name string
}

type clause interface {
	build(buf *bytes.Buffer)

	cat() string

	kind() clauseKind
}

type clauseKind uint8

type column struct {
	col   string
	op    string
	val   interface{}
	cat_  string
	kind_ clauseKind
}

type list struct {
	kind_ clauseKind
	items []string
}

type order struct {
	col string
	dir string
}

type portion struct {
	kind_ clauseKind
	n     int64
}

type table struct {
	kind_ clauseKind
	item  string
}

type values struct {
	vals []interface{}
}

const (
	anon_ clauseKind = iota
	from_
	into_
	where_
	order_
	set_
	as_
	limit_
	offset_
	values_
	columns_
	returning_
)

func (k clauseKind) build(buf *bytes.Buffer) {
	switch k {
	case from_:
		buf.WriteString("FROM ")
		return
	case into_:
		buf.WriteString("INTO ")
		return
	case where_:
		buf.WriteString("WHERE ")
		return
	case order_:
		buf.WriteString("ORDER BY ")
		return
	case set_:
		buf.WriteString("SET ")
	case as_:
		buf.WriteString("AS ")
		return
	case limit_:
		buf.WriteString("LIMIT ")
		return
	case offset_:
		buf.WriteString("OFFSET ")
		return
	case values_:
		buf.WriteString("VALUES ")
		return
	case returning_:
		buf.WriteString("RETURNING ")
		return
	}
}

func (a alias) build(buf *bytes.Buffer) {
	buf.WriteString(a.name)
}

func (a alias) cat() string {
	return ","
}

func (a alias) kind() clauseKind {
	return as_
}

func (c column) build(buf *bytes.Buffer) {
	fmt.Fprintf(buf, "%s %s %v", c.col, c.op, c.val)
}

func (c column) cat() string {
	if c.kind_ == where_ {
		return " " + c.cat_ + " "
	}

	return c.cat_ + " "
}

func (c column) kind() clauseKind {
	return c.kind_
}

func (l list) build(buf *bytes.Buffer) {
	if l.kind_ == values_ {
		buf.WriteString("(")
	}

	buf.WriteString(strings.Join(l.items, ", "))

	if l.kind_ == values_ {
		buf.WriteString(")")
	}
}

func (l list) cat() string {
	return ","
}

func (l list) kind() clauseKind {
	return l.kind_
}

func (o order) build(buf *bytes.Buffer) {
	buf.WriteString(o.col)
	buf.WriteString(" ")
	buf.WriteString(o.dir)
}

func (o order) cat() string {
	return ","
}

func (o order) kind() clauseKind {
	return order_
}

func (p portion) build(buf *bytes.Buffer) {
	buf.WriteString(strconv.FormatInt(p.n, 10))
}

func (p portion) cat() string {
	return ""
}

func (p portion) kind() clauseKind {
	return p.kind_
}

func (t table) build(buf *bytes.Buffer) {
	buf.WriteString(t.item)
}

func (t table) cat() string {
	return ","
}

func (t table) kind() clauseKind {
	return t.kind_
}
