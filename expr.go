package query

import (
	"fmt"
	"strings"
)

// callExpr is the expression used for calling functions within PostgreSQL
type callExpr struct {
	name string
	args []Expr
}

type listExpr struct {
	items []string
	wrap  bool
	args  []interface{}
}

type identExpr string

type argExpr struct {
	val interface{}
}

type litExpr struct {
	val interface{}
}

// Expr is an expression that exists within the Query being built. This would
// typically be an identifier, literal, argument, function call, or list
// values in queries.
type Expr interface {
	// Args returns the arguments that were given to the Query expression.
	Args() []interface{}

	// Build returns the built up Query expression.
	Build() string
}

var (
	_ Expr = (*listExpr)(nil)
	_ Expr = (*identExpr)(nil)
	_ Expr = (*argExpr)(nil)
	_ Expr = (*litExpr)(nil)
	_ Expr = (*callExpr)(nil)
)

// Columns returns a list expression of the given column names. This will not
// be wrapped in parentheses when built.
func Columns(cols ...string) listExpr {
	return listExpr{
		items: cols,
		wrap:  false,
	}
}

// Sum returns a call expression for the SUM function on the given column.
func Sum(col string) callExpr {
	return callExpr{
		name: "SUM",
		args: []Expr{Lit(col)},
	}
}

// Count returns a call expression for the COUNT function on the given columns.
func Count(cols ...string) callExpr {
	exprs := make([]Expr, 0, len(cols))

	for _, col := range cols {
		exprs = append(exprs, Lit(col))
	}

	return callExpr{
		name: "COUNT",
		args: exprs,
	}
}

// List returns a list expression of the given values. Each item in the
// given list will use the ? placeholder. This will be wrapped in parentheses
// when built.
func List(vals ...interface{}) listExpr {
	items := make([]string, 0, len(vals))

	for range vals {
		items = append(items, "?")
	}
	return listExpr{
		items: items,
		wrap:  true,
		args:  vals,
	}
}

// Ident returns an identifier expression for the given string. When built
// this will simply use the initial string that was given.
func Ident(s string) identExpr { return identExpr(s) }

// Arg returns an argument expression for the given value. When built this will
// use ? as the placeholder for the argument value.
func Arg(val interface{}) argExpr {
	return argExpr{
		val: val,
	}
}

// Lit returns a literal expression for the given value. This will place the
// literal value into the built up expression string itself, and not use the ?
// placeholder. For example using Lit like so,
//
//     Where("deleted_at", "IS NOT", Lit("NULL"))
//
// would result in a WHERE clause being built up like this,
//
//     WHERE (deleted_at IS NOT NULL)
func Lit(val interface{}) litExpr {
	return litExpr{
		val: val,
	}
}

func (e listExpr) Args() []interface{} { return e.args }

func (e listExpr) Build() string {
	items := strings.Join(e.items, ", ")

	if e.wrap {
		return "(" + items + ")"
	}
	return items
}

func (e identExpr) Args() []interface{} { return nil }
func (e identExpr) Build() string       { return string(e) }

func (e argExpr) Args() []interface{} { return []interface{}{e.val} }
func (e argExpr) Build() string       { return "?" }

func (e litExpr) Args() []interface{} { return nil }
func (e litExpr) Build() string       { return fmt.Sprintf("%v", e.val) }

func (e callExpr) Args() []interface{} {
	vals := make([]interface{}, 0)

	for _, expr := range e.args {
		vals = append(vals, expr.Args()...)
	}
	return vals
}

func (e callExpr) Build() string {
	args := make([]string, 0, len(e.args))

	for _, arg := range e.args {
		args = append(args, arg.Build())
	}
	return e.name + "(" + strings.Join(args, ", ") + ")"
}
