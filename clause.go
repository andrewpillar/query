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
	col         string
	op          string
	val         interface{}
	conjunction string
	clauseKind  clauseKind
}

type count struct {
	expr string
}

type list struct {
	clauseKind clauseKind
	items      []string
}

type order struct {
	col string
	dir string
}

type portion struct {
	clauseKind clauseKind
	n          int64
}

type table struct {
	clauseKind clauseKind
	item       string
}

type union struct {
	query Query
}

type values struct {
	vals []interface{}
}

const (
	noneKind clauseKind = iota
	fromKind
	intoKind
	whereKind
	orderKind
	setKind
	asKind
	limitKind
	offsetKind
	valuesKind
	columnsKind
	countKind
	returningKind
	unionKind
)

func (k clauseKind) build(buf *bytes.Buffer) {
	switch k {
	case fromKind:
		buf.WriteString("FROM ")
	case intoKind:
		buf.WriteString("INTO ")
	case whereKind:
		buf.WriteString("WHERE ")
	case orderKind:
		buf.WriteString("ORDER BY ")
	case setKind:
		buf.WriteString("SET ")
	case asKind:
		buf.WriteString("AS ")
	case limitKind:
		buf.WriteString("LIMIT ")
	case offsetKind:
		buf.WriteString("OFFSET ")
	case valuesKind:
		buf.WriteString("VALUES ")
	case countKind:
		buf.WriteString("COUNT")
	case returningKind:
		buf.WriteString("RETURNING ")
	}
}

// -- alias implement --

func (a alias) build(buf *bytes.Buffer) {
	buf.WriteString(a.name)
}

func (a alias) cat() string {
	return ","
}

func (a alias) kind() clauseKind {
	return asKind
}

// -- column implement --

func (c column) build(buf *bytes.Buffer) {
	fmt.Fprintf(buf, "%s %s %v", c.col, c.op, c.val)
}

func (c column) cat() string {
	if c.clauseKind == whereKind {
		return " " + c.conjunction + " "
	}

	return c.conjunction + " "
}

func (c column) kind() clauseKind {
	return c.clauseKind
}

// -- count implement --

func (c count) build(buf *bytes.Buffer) {
	buf.WriteString(c.expr)
}

func (c count) cat() string {
	return ""
}

func (c count) kind() clauseKind {
	return countKind
}

// -- list implement --

func (l list) build(buf *bytes.Buffer) {
	if l.clauseKind == valuesKind {
		buf.WriteString("(")
	}

	buf.WriteString(strings.Join(l.items, ", "))

	if l.clauseKind == valuesKind {
		buf.WriteString(")")
	}
}

func (l list) cat() string {
	return ","
}

func (l list) kind() clauseKind {
	return l.clauseKind
}

// -- order implement --

func (o order) build(buf *bytes.Buffer) {
	buf.WriteString(o.col)
	buf.WriteString(" ")
	buf.WriteString(o.dir)
}

func (o order) cat() string {
	return ","
}

func (o order) kind() clauseKind {
	return orderKind
}

// -- portion implement --

func (p portion) build(buf *bytes.Buffer) {
	buf.WriteString(strconv.FormatInt(p.n, 10))
}

func (p portion) cat() string {
	return ""
}

func (p portion) kind() clauseKind {
	return p.clauseKind
}

// -- table implement --

func (t table) build(buf *bytes.Buffer) {
	buf.WriteString(t.item)
}

func (t table) cat() string {
	return ","
}

func (t table) kind() clauseKind {
	return t.clauseKind
}

// -- union implement --

func (u union) build(buf *bytes.Buffer) {
	buf.WriteString(u.query.buildInitial())
}

func (u union) cat() string {
	return " UNION "
}

func (u union) kind() clauseKind {
	return unionKind
}
