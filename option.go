package query

import (
	"fmt"
	"strings"
)

func expand(vals ...interface{}) string {
	s := make([]string, 0, len(vals))

	for _ = range vals {
		s = append(s, "?")
	}

	return "(" + strings.Join(s, ", ") + ")"
}

func realOrder(col, dir string) Option {
	return func(q Query) Query {
		o := order{
			col: col,
			dir: dir,
		}

		q.clauses = append(q.clauses, o)

		return q
	}
}

func realPortion(kind clauseKind, n int64) Option {
	return func(q Query) Query {
		p := portion{
			kind_: kind,
			n:     n,
		}

		q.clauses = append(q.clauses, p)

		return q
	}
}

func realWhere(cat, col, op string, val interface{}) Option {
	return func(q Query) Query {
		c := column{
			col:   col,
			op:    op,
			val:   val,
			cat_:  cat,
			kind_: where_,
		}

		q.clauses = append(q.clauses, c)

		return q
	}
}

func As(name string) Option {
	return func(q Query) Query {
		a := alias{
			name: name,
		}

		q.clauses = append(q.clauses, a)

		return q
	}
}

func Columns(cols ...string) Option {
	return func(q Query) Query {
		if q.stmt == insert_ {
			end := len(cols) - 1

			cols[0] = "(" + cols[0]
			cols[end] = cols[end] + ")"
		}

		l := list{
			kind_: columns_,
			items: cols,
		}

		q.clauses = append(q.clauses, l)

		return q
	}
}

func From(item string) Option {
	return func(q Query) Query {
		if q.stmt == select_ || q.stmt == delete_ {
			t := table{
				kind_: from_,
				item:  item,
			}

			q.clauses = append(q.clauses, t)
		}

		return q
	}
}

func Into(item string) Option {
	return func(q Query) Query {
		if q.stmt == insert_ {
			t := table{
				kind_: into_,
				item:  item,
			}

			q.clauses = append(q.clauses, t)
		}

		return q
	}
}

func Limit(n int64) Option {
	return func(q Query) Query {
		return realPortion(limit_, n)(q)
	}
}

func Offset(n int64) Option {
	return func(q Query) Query {
		return realPortion(offset_, n)(q)
	}
}

func Options(opts ...Option) Option {
	return func(q Query) Query {
		for _, opt := range opts {
			q = opt(q)
		}

		return q
	}
}

func OrderAsc(col string) Option {
	return func(q Query) Query {
		return realOrder(col, "ASC")(q)
	}
}

func OrderDesc(col string) Option {
	return func(q Query) Query {
		return realOrder(col, "DESC")(q)
	}
}

func OrWhere(col, op string, vals ...interface{}) Option {
	return func(q Query) Query {
		var val interface{} = "?"

		q.args = append(q.args, vals...)

		if len(vals) > 1 {
			val = expand(vals...)
		}

		return realWhere("OR", col, op, val)(q)
	}
}

func OrWhereQuery(col, op string, q2 Query) Option {
	return func(q1 Query) Query {
		val := "(" + q2.buildInitial() + ")"

		q1.args = append(q1.args, q2.Args()...)

		return realWhere("OR", col, op, val)(q1)
	}
}

func Returning(cols ...string) Option {
	return func(q Query) Query {
		if q.stmt == insert_ || q.stmt == update_ || q.stmt == delete_ {
			l := list{
				kind_: returning_,
				items: cols,
			}

			q.clauses = append(q.clauses, l)
		}

		return q
	}
}

func Set(col string, val interface{}) Option {
	return func(q Query) Query {
		q.args = append(q.args, val)

		return SetRaw(col, "?")(q)
	}
}

func SetRaw(col string, val interface{}) Option {
	return func(q Query) Query {
		if q.stmt == update_ {
			c := column{
				col:   col,
				op:    "=",
				val:   val,
				kind_: set_,
				cat_:  ",",
			}

			q.clauses = append(q.clauses, c)
		}

		return q
	}
}

func Table(item string) Option {
	return func(q Query) Query {
		if q.stmt == update_ {
			t := table{
				kind_: anon_,
				item:  item,
			}

			q.clauses = append(q.clauses, t)
		}

		return q
	}
}

func Values(vals ...interface{}) Option {
	return func(q Query) Query {
		if q.stmt == insert_ {
			items := make([]string, 0, len(vals))

			for _ = range vals {
				items = append(items, "?")
			}

			l := list{
				kind_: values_,
				items: items,
			}

			q.args = append(q.args, vals...)
			q.clauses = append(q.clauses, l)
		}

		return q
	}
}

func Where(col, op string, vals ...interface{}) Option {
	return func(q Query) Query {
		var val interface{} = "?"

		q.args = append(q.args, vals...)

		if len(vals) > 1 || op == "IN" {
			val = expand(vals...)
		}

		return realWhere("AND", col, op, val)(q)
	}
}

func WhereRaw(col, op string, vals ...interface{}) Option {
	return func(q Query) Query {
		var val interface{}

		if len(vals) > 1 || op == "IN" {
			s := make([]string, 0, len(vals))

			for _, v := range vals {
				s = append(s, fmt.Sprintf("%v", v))
			}

			val = "(" + strings.Join(s, ", ") + ")"
		} else {
			val = vals[0]
		}

		return realWhere("AND", col, op, val)(q)
	}
}

func OrWhereRaw(col, op string, vals ...interface{}) Option {
	return func(q Query) Query {
		var val interface{}

		if len(vals) > 1 {
			s := make([]string, 0, len(vals))

			for _, v := range vals {
				s = append(s, fmt.Sprintf("%v", v))
			}

			val = "(" + strings.Join(s, ", ") + ")"
		} else {
			val = vals[0]
		}

		return realWhere("OR", col, op, val)(q)
	}
}

func WhereQuery(col, op string, q2 Query) Option {
	return func(q1 Query) Query {
		val := "(" + q2.buildInitial() + ")"

		q1.args = append(q1.args, q2.Args()...)

		return realWhere("AND", col, op, val)(q1)
	}
}
