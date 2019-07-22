package query

import (
	"fmt"
	"strconv"
	"strings"
)

type Option func(q Query) Query

func param(i int) string {
	return "$" + strconv.FormatInt(int64(i), 10)
}

func Columns(cols ...string) Option {
	return func(q Query) Query {
		q.cols = cols

		return q
	}
}

func Limit(l int) Option {
	return func(q Query) Query {
		q.limit = l

		return q
	}
}

func Offset(o int) Option {
	return func(q Query) Query {
		q.offset = o

		return q
	}
}

func Or(opts ...Option) Option {
	return func(q Query) Query {
		l := len(q.wheres)

		for _, opt := range opts {
			q = opt(q)
		}

		diff := len(q.wheres) - l

		if l == diff {
			return q
		}

		wheres := make([]where, 0)

		for i, w := range q.wheres {
			if i >= l {
				w.cat = " OR "
			}

			wheres = append(wheres, w)
		}

		q.wheres = wheres

		return q
	}
}

func OrderAsc(cols ...string) Option {
	return func(q Query) Query {
		q.order.cols = cols
		q.order.dir = asc

		return q
	}
}

func OrderDesc(cols ...string) Option {
	return func(q Query) Query {
		q.order.cols = cols
		q.order.dir = desc

		return q
	}
}

func Returning(cols ...string) Option {
	return func(q Query) Query {
		q.ret = cols

		return q
	}
}

func Set(col string, val interface{}) Option {
	return func(q Query) Query {
		q = SetRaw(col, nil)(q)
		q.args = append(q.args, val)

		return q
	}
}

func SetRaw(col string, val interface{}) Option {
	return func(q Query) Query {
		if q.stmt != _update {
			return q
		}

		s := set{
			col: col,
			val: val,
		}

		q.sets = append(q.sets, s)

		return q
	}
}

func Table(table string) Option {
	return func(q Query) Query {
		q.table = table

		return q
	}
}

func Values(vals ...interface{}) Option {
	return func(q Query) Query {
		if q.stmt == _insert {
			q.args = vals
		}

		return q
	}
}

func Where(col string, op string, val interface{}) Option {
	return func(q Query) Query {
		w := where{
			col: col,
			op:  op,
			cat: " AND ",
		}
		q.wheres = append(q.wheres, w)
		switch op {
		case OpEq, OpNotEq, OpGt, OpGtOrEq, OpLt, OpLtOrEq, OpLike:
			q.args = append(q.args, val)
		case OpIs:
		}
		return q
	}
}

func WhereEq(col string, val interface{}) Option {
	return Where(col, OpEq, val)
}

func WhereIs(col string, val interface{}) Option {
	return Where(col, OpIs, val)
}

func WhereLike(col string, val interface{}) Option {
	return Where(col, OpLike, val)
}

func WhereIn(col string, vals ...interface{}) Option {
	return func(q Query) Query {
		if len(vals) == 0 {
			return q
		}

		w := where{
			col: col,
			op:  "IN",
			cat: " AND ",
		}

		q.wheres = append(q.wheres, w)
		q.args = append(q.args, vals...)

		return q
	}
}

func WhereInRaw(col string, vals ...interface{}) Option {
	return func(q Query) Query {
		if len(vals) == 0 {
			return q
		}

		in := make([]string, len(vals), len(vals))

		for i, v := range vals {
			in[i] = fmt.Sprintf("%v", v)
		}

		val := "(" + strings.Join(in, ", ") + ")"

		w := where{
			col: col,
			op:  "IN",
			val: val,
			cat: " AND ",
		}

		q.wheres = append(q.wheres, w)

		return q
	}
}

func WhereInQuery(col string, q2 Query) Option {
	return func(q1 Query) Query {
		w := where{
			col:   col,
			op:    "IN",
			cat:   " AND ",
			query: q2,
		}

		q1.wheres = append(q1.wheres, w)
		q1.args = append(q1.args, q2.args...)

		return q1
	}
}
