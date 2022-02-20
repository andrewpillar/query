package query

import (
	"strconv"
	"strings"
)

// clause is a Query expression that will typically represent one of the
// following SQL clauses, FROM, LIMIT, OFFSET, ORDER BY, UNION, VALUES,
// WHERE, RETURNING, and SET.
type clause interface {
	Expr

	// kind returns the kind of the current clause.
	kind() clauseKind
}

type clauseKind uint

//go:generate stringer -type clauseKind -linecomment
const (
	_FromClause clauseKind = iota // FROM
	_LimitClause                  // LIMIT
	_OffsetClause                 // OFFSET
	_OrderClause                  // ORDER BY
	_UnionClause                  // UNION
	_ValuesClause                 // VALUES
	_WhereClause                  // WHERE
	_ReturningClause              // RETURNING
	_SetClause                    // SET
)

func realWhere(conjunction string, left Expr, op string, right Expr) Option {
	return func(q Query) Query {
		leftArgs := left.Args()
		rightArgs := right.Args()

		args := make([]interface{}, 0, len(leftArgs) + len(rightArgs))
		args = append(args, leftArgs...)
		args = append(args, rightArgs...)

		if q1, ok := right.(Query); ok {
			right = Lit("(" + q1.buildInitial() + ")")
		}

		q.clauses = append(q.clauses, whereClause{
			conjunction: conjunction,
			op:          op,
			left:        left,
			right:       right,
		})
		q.args = append(q.args, args...)
		return q
	}
}

// Where appends a WHERE clause to the Query. This will append the arguments
// of the given expression to the Query too. By default this will use AND for
// conjoining multiple WHERE clauses.
func Where(col, op string, expr Expr) Option {
	return func(q Query) Query {
		return realWhere("AND", Ident(col), op, expr)(q)
	}
}

// OrWhere appends a WHERE clause to the Query. This will append the arguments
// of the given expression to the Query too. This will use OR for conjoining
// with a preceding WHERE clause.
func OrWhere(col, op string, expr Expr) Option {
	return func(q Query) Query {
		return realWhere("OR", Ident(col), op, expr)(q)
	}
}

// From appends a FROM clause for the given table to the Query.
func From(table string) Option {
	return func(q Query) Query {
		q.clauses = append(q.clauses, fromClause{
			table: table,
		})
		return q
	}
}

// Limit appends a LIMIT clause with the given amount to the Query.
func Limit(n int64) Option {
	return func(q Query) Query {
		q.clauses = append(q.clauses, limitClause(n))
		return q
	}
}

// Offset appends an OFFSET clause with the given value to the Query.
func Offset(n int64) Option {
	return func(q Query) Query {
		q.clauses = append(q.clauses, offsetClause(n))
		return q
	}
}

// OrderAsc appends an ORDER BY [column,...] ASC clause for the given columns
// to the Query.
func OrderAsc(cols ...string) Option {
	return func(q Query) Query {
		q.clauses = append(q.clauses, orderClause{
			cols: cols,
			dir:  "ASC",
		})
		return q
	}
}

// OrderDesc appends an ORDER BY [column,...] DESC clause for the given columns
// to the Query.
func OrderDesc(cols ...string) Option {
	return func(q Query) Query {
		q.clauses = append(q.clauses, orderClause{
			cols: cols,
			dir:  "DESC",
		})
		return q
	}
}

// Returning appends a RETURNING [column,...] clause for the given columns to
// the Query.
func Returning(cols ...string) Option {
	return func(q Query) Query {
		q.clauses = append(q.clauses, returningClause{
			cols: cols,
		})
		return q
	}
}

// Set appends a SET clause for the given column and expression to the Query.
func Set(col string, expr Expr) Option {
	return func(q Query) Query {
		if q.stmt == _Update {
			q.clauses = append(q.clauses, setClause{
				col:  col,
				expr: Lit(expr.Build()),
			})
			q.args = append(q.args, expr.Args()...)
		}
		return q
	}
}

// Values appends a VALUES clause for the given values to the Query. Each
// given value will use the ? placeholder when built.
func Values(vals ...interface{}) Option {
	items := make([]string, 0, len(vals))

	for range vals {
		items = append(items, "?")
	}

	return func(q Query) Query {
		q.clauses = append(q.clauses, valuesClause{
			items: items,
			args:  vals,
		})
		q.args = append(q.args, vals...)
		return q
	}
}

type fromClause struct {
	table string
}

var _ clause = (*fromClause)(nil)

func (c fromClause) Args() []interface{} { return nil }
func (c fromClause) Build() string       { return c.table }
func (c fromClause) kind() clauseKind    { return _FromClause }

type limitClause int64

var _ clause = (*limitClause)(nil)

func (c limitClause) Args() []interface{} { return nil }
func (c limitClause) Build() string       { return strconv.FormatInt(int64(c), 10) }
func (c limitClause) kind() clauseKind    { return _LimitClause }

type offsetClause int64

var _ clause = (*offsetClause)(nil)

func (c offsetClause) Args() []interface{} { return nil }
func (c offsetClause) Build() string       { return strconv.FormatInt(int64(c), 10) }
func (c offsetClause) kind() clauseKind    { return _OffsetClause }

type orderClause struct {
	cols []string
	dir  string
}

var _ clause = (*orderClause)(nil)

func (c orderClause) Args() []interface{} { return nil }
func (c orderClause) Build() string       { return strings.Join(c.cols, ", ") + " " + c.dir }
func (c orderClause) kind() clauseKind    { return _OrderClause }

type returningClause struct {
	cols []string
}

var _ clause = (*returningClause)(nil)

func (c returningClause) Args() []interface{} { return nil }
func (c returningClause) Build() string       { return strings.Join(c.cols, ", ") }
func (c returningClause) kind() clauseKind    { return _ReturningClause }

type setClause struct {
	col  string
	expr Expr
}

var _ clause = (*setClause)(nil)

func (c setClause) Args() []interface{} { return nil }
func (c setClause) Build() string       { return c.col + " = " + c.expr.Build() }
func (c setClause) kind() clauseKind    { return _SetClause }

type unionClause struct {
	q Query
}

var _ clause = (*unionClause)(nil)

func (c unionClause) Args() []interface{}  { return nil }
func (c unionClause) Build() string        { return c.q.buildInitial() }
func (c unionClause) kind() clauseKind     { return _UnionClause }

type valuesClause struct {
	items []string
	args  []interface{}
}

var _ clause = (*valuesClause)(nil)

func (c valuesClause) Args() []interface{} { return c.args  }
func (c valuesClause) Build() string       { return "(" + strings.Join(c.items, ", ") + ")" }
func (c valuesClause) kind() clauseKind    { return _ValuesClause }

type whereClause struct {
	conjunction string
	op          string
	left        Expr
	right       Expr
}

var _ clause = (*whereClause)(nil)

func (c whereClause) Args() []interface{} { return nil }

func (c whereClause) Build() string {
	return c.left.Build() + " " + c.op + " " + c.right.Build()
}

func (c whereClause) kind() clauseKind { return _WhereClause }
