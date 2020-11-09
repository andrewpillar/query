# query

>**Note:** This library is still very much a work in progress, expect breaking
changes.

query is a simple PostgreSQL query builder for Go. It works by using first
class functions to allow for queries to be built up. This does not support the
entire dialect for PostgreSQL, only a subset that would be necessary for CRUD
operations.

For full examples as to how queries of different types can be built up, see the
`query_test.go` file.

Let's look at how a simple SELECT query can be built up using this library,

    q := query.Select(
        query.Columns("*"),
        query.From("posts"),
        query.Where("user_id", "=", query.Arg(10)),
        query.OrderDesc("created_at"),
    )

we can then use the above to pass a query to our database driver, along with
that arguments passed to the query,

    rows, err := db.Query(q.Build(), q.Args()...)

Calling Build on the query will build up the query string and correctly set
the parameter arguments in the query. Args will return an interface slice
containing the arguments given to the query via query.Arg.

We can do even more complex SELECT queries too,

    q := query.Select(
        query.Columns("*"),
        query.From("posts"),
        query.Where("id", "IN",
            query.Select(
                query.Columns("id"),
                query.From("post_tags"),
                query.Where("name", "LIKE", query.Arg("%sql%")),
            ),
        ),
        query.OrderDesc("created_at"),
    )

This library makes use of the type Option which is a first class function,
which takes the current Query an Option is being applied to and returns it.

    type Option func(Query) Query

With this we can define our own Option functions to clean up some of the
queries we want to build.

    func Search(col, pattern string) query.Option {
        return func(q query.Query) query.Query {
            if pattern == "" {
                return q
            }
            return query.Where(col, "LIKE", query.Arg("%" + pattern + "%"))(q)
        }
    }

    q := query.Select(
        query.Columns("*"),
        query.From("posts"),
        Search("title", "query builder"),
        query.OrderDesc("created_at"),
    )
