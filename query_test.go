package query

import "testing"

type testQuery struct {
	expected string
	query    Query
}

func checkQueries(qq []testQuery, t *testing.T) {
	for _, q := range qq {
		built := q.query.Build()

		if built != q.expected {
			t.Fatalf(
				"query not as expected:\n\texpected = '%s'\n\t  actual = '%s'",
				q.expected,
				built,
			)
		}
	}
}

func TestSelect(t *testing.T) {
	queries := []testQuery{
		{
			"SELECT * FROM users WHERE (username = $1)",
			Select(Columns("*"), From("users"), Where("username", "=", "me")),
		},
		{
			"SELECT * FROM users WHERE (username = $1 OR email = $2)",
			Select(
				Columns("*"),
				From("users"),
				Where("username", "=", "me"),
				OrWhere("email", "=", "email@domain.com"),
			),
		},
		{
			"SELECT * FROM users WHERE (username = $1 OR email = $2) AND (registered = $3)",
			Select(
				Columns("*"),
				From("users"),
				Where("username", "=", "me"),
				OrWhere("email", "=", "email@domain.com"),
				Where("registered", "=", true),
			),
		},
		{
			"SELECT * FROM posts WHERE (title LIKE $1) LIMIT 25 OFFSET 2",
			Select(
				Columns("*"),
				From("posts"),
				Where("title", "LIKE", "%foo%"),
				Limit(int64(25)),
				Offset(int64(2)),
			),
		},
		{
			"SELECT * FROM posts WHERE (user_id = $1 AND id IN (SELECT post_id FROM tags WHERE (title LIKE $2)))",
			Select(
				Columns("*"),
				From("posts"),
				Where("user_id", "=", 1),
				WhereQuery("id", "IN",
					Select(
						Columns("post_id"),
						From("tags"),
						Where("title", "LIKE", "%foo%"),
					),
				),
			),
		},
		{
			"SELECT * FROM posts WHERE (id IN (SELECT post_id FROM tags WHERE (title LIKE $1)) AND category_id IN (SELECT id FROM categories WHERE (name LIKE $2)))",
			Select(
				Columns("*"),
				From("posts"),
				WhereQuery("id", "IN",
					Select(
						Columns("post_id"),
						From("tags"),
						Where("title", "LIKE", "%foo%"),
					),
				),
				WhereQuery("category_id", "IN",
					Select(
						Columns("id"),
						From("categories"),
						Where("name", "LIKE", "%bar%"),
					),
				),
			),
		},
		{
			"SELECT * FROM users WHERE (id IN ($1, $2, $3, $4, $5))",
			Select(
				Columns("*"),
				From("users"),
				Where("id", "IN", 1, 2, 3, 4, 5),
			),
		},
		{
			"SELECT * FROM builds WHERE (status = $1 AND namespace_id IN (SELECT id FROM namespaces WHERE (root_id IN (SELECT namespace_id FROM collaborators WHERE (user_id = $2))))) OR (user_id = $3) AND (status = $4) ORDER BY created_at DESC",
			Select(
				Columns("*"),
				From("builds"),
				Where("status", "=", "running"),
				Options(
					WhereQuery("namespace_id", "IN",
						Select(
							Columns("id"),
							From("namespaces"),
							WhereQuery("root_id", "IN",
								Select(
									Columns("namespace_id"),
									From("collaborators"),
									Where("user_id", "=", 1),
								),
							),
						),
					),
					OrWhere("user_id", "=", 1),
				),
				Where("status", "=", "running"),
				OrderDesc("created_at"),
			),
		},
		{
			"SELECT * FROM files WHERE (deleted_at IS NOT NULL)",
			Select(
				Columns("*"),
				From("files"),
				WhereRaw("deleted_at", "IS NOT", "NULL"),
			),
		},
		{
			"SELECT * FROM users WHERE (id IN ($1))",
			Select(
				Columns("*"),
				From("users"),
				Where("id", "IN", 1),
			),
		},
		{
			"SELECT COUNT(*) FROM users",
			Select(
				Count("*"),
				From("users"),
			),
		},
		{
			"SELECT * FROM variables WHERE (namespace_id IN (SELECT id FROM namespaces WHERE (root_id IN (SELECT namespace_id FROM collaborators WHERE (user_id = $1) UNION SELECT id FROM namespaces WHERE (user_id = $2)))) OR user_id = $3)",
			Select(
				Columns("*"),
				From("variables"),
				WhereQuery("namespace_id", "IN",
					Select(
						Columns("id"),
						From("namespaces"),
						WhereQuery("root_id", "IN",
							Union(
								Select(
									Columns("namespace_id"),
									From("collaborators"),
									Where("user_id", "=", 2),
								),
								Select(
									Columns("id"),
									From("namespaces"),
									Where("user_id", "=", 2),
								),
							),
						),
					),
				),
				OrWhere("user_id", "=", 2),
			),
		},
	}

	checkQueries(queries, t)
}

func TestInsert(t *testing.T) {
	queries := []testQuery{
		{
			"INSERT INTO users (email, username, password) VALUES ($1, $2, $3)",
			Insert(
				Into("users"),
				Columns("email", "username", "password"),
				Values("me@exmaple.com", "me", "secret"),
			),
		},
		{
			"INSERT INTO users (email, username, password) VALUES ($1, $2, $3) RETURNING id, created_at",
			Insert(
				Into("users"),
				Columns("email", "username", "password"),
				Values("me@exmaple.com", "me", "secret"),
				Returning("id", "created_at"),
			),
		},
	}

	checkQueries(queries, t)
}

func TestUpdate(t *testing.T) {
	queries := []testQuery{
		{
			"UPDATE users SET email = $1, updated_at = NOW() WHERE (id = $2)",
			Update(
				Table("users"),
				Set("email", "me@example.com"),
				SetRaw("updated_at", "NOW()"),
				Where("id", "=", 1),
			),
		},
	}

	checkQueries(queries, t)
}

func TestDelete(t *testing.T) {
	queries := []testQuery{
		{
			"DELETE FROM users WHERE (id = $1)",
			Delete(
				From("users"),
				Where("id", "=", 1),
			),
		},
	}

	checkQueries(queries, t)
}
