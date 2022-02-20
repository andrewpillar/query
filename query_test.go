package query

import "testing"

func Test_Query(t *testing.T) {
	tests := []struct {
		expected string
		q        Query
	}{
		{
			"SELECT SUM(size) FROM objects WHERE (user_id = $1)",
			Select(Sum("size"), From("objects"), Where("user_id", "=", Arg(1))),
		},
		{
			"SELECT COUNT(*) FROM users",
			Select(Count("*"), From("users")),
		},
		{
			"SELECT * FROM users WHERE (username = $1)",
			Select(Columns("*"), From("users"), Where("username", "=", Arg("andrew"))),
		},
		{
			"SELECT * FROM keys WHERE (name = $1 AND namespace_id IS NULL)",
			Select(
				Columns("*"),
				From("keys"),
				Where("name", "=", Arg("id_rsa")),
				Where("namespace_id", "IS", Lit("NULL")),
			),
		},
		{
			"SELECT * FROM users WHERE (email = $1 OR username = $2)",
			Select(
				Columns("*"),
				From("users"),
				Where("email", "=", Arg("me@example.com")),
				OrWhere("username", "=", Arg("andrew")),
			),
		},
		{
			"SELECT * FROM users WHERE (email = $1 OR username = $2) AND (registered = $3)",
			Select(
				Columns("*"),
				From("users"),
				Where("email", "=", Arg("me@example.com")),
				OrWhere("username", "=", Arg("andrew")),
				Where("registered", "=", Arg(true)),
			),
		},
		{
			"SELECT * FROM posts WHERE (title LIKE $1) LIMIT 25 OFFSET 2",
			Select(
				Columns("*"),
				From("posts"),
				Where("title", "LIKE", Arg("%foo%")),
				Limit(int64(25)),
				Offset(int64(2)),
			),
		},
		{
			"SELECT * FROM posts WHERE (user_id = $1 AND id IN (SELECT post_id FROM tags WHERE (name LIKE $2)))",
			Select(
				Columns("*"),
				From("posts"),
				Where("user_id", "=", Arg(1)),
				Where("id", "IN", Select(
					Columns("post_id"),
					From("tags"),
					Where("name", "LIKE", Arg("%foo%")),
				)),
			),
		},
		{
			"SELECT * FROM posts WHERE (id IN (SELECT post_id FROM tags WHERE (name LIKE $1)) AND category_id IN (SELECT id FROM categories WHERE (name LIKE $2)))",
			Select(
				Columns("*"),
				From("posts"),
				Where("id", "IN",
					Select(Columns("post_id"), From("tags"), Where("name", "LIKE", Arg("%foo%"))),
				),
				Where("category_id", "IN",
					Select(Columns("id"), From("categories"), Where("name", "LIKE", Arg("%bar%"))),
				),
			),
		},
		{
			"SELECT * FROM users WHERE (id IN ($1))",
			Select(Columns("*"), From("users"), Where("id", "IN", List(1))),
		},
		{
			"SELECT * FROM users WHERE (id IN ($1, $2, $3, $4, $5))",
			Select(
				Columns("*"),
				From("users"),
				Where("id", "IN", List(1, 2, 3, 4, 5)),
			),
		},
		{
			"SELECT * FROM builds WHERE (status = $1 AND namespace_id IN (SELECT id FROM namespaces WHERE (root_id IN (SELECT namespace_id FROM namespace_collaborators WHERE (user_id = $2))))) OR (user_id = $3) AND (status = $4) ORDER BY created_at DESC",
			Select(
				Columns("*"),
				From("builds"),
				Where("status", "=", Arg("running")),
				Options(
					Where("namespace_id", "IN",
						Select(
							Columns("id"),
							From("namespaces"),
							Where("root_id", "IN",
								Select(
									Columns("namespace_id"),
									From("namespace_collaborators"),
									Where("user_id", "=", Arg(1)),
								),
							),
						),
					),
					OrWhere("user_id", "=", Arg(1)),
				),
				Where("status", "=", Arg("running")),
				OrderDesc("created_at"),
			),
		},
		{
			"SELECT * FROM files WHERE (deleted_at IS NOT NULL)",
			Select(Columns("*"), From("files"), Where("deleted_at", "IS NOT", Lit("NULL"))),
		},
		{
			"SELECT * FROM variables WHERE (namespace_id IN (SELECT id FROM namespaces WHERE (root_id IN (SELECT namespace_id FROM namespace_collaborators WHERE (user_id = $1) UNION SELECT id FROM namespaces WHERE (user_id = $2)))) OR user_id = $3)",
			Select(
				Columns("*"),
				From("variables"),
				Where("namespace_id", "IN",
					Select(
						Columns("id"),
						From("namespaces"),
						Where("root_id", "IN",
							Union(
								Select(
									Columns("namespace_id"),
									From("namespace_collaborators"),
									Where("user_id", "=", Arg(2)),
								),
								Select(
									Columns("id"),
									From("namespaces"),
									Where("user_id", "=", Arg(2)),
								),
							),
						),
					),
				),
				OrWhere("user_id", "=", Arg(2)),
			),
		},
		{
			"INSERT INTO users (email, username, password) VALUES ($1, $2, $3)",
			Insert(
				"users",
				Columns("email", "username", "password"),
				Values("me@example.com", "me", "secret"),
			),
		},
		{
			"INSERT INTO users (email, username, password) VALUES ($1, $2, $3) RETURNING id, created_at",
			Insert(
				"users",
				Columns("email", "username", "password"),
				Values("me@example.com", "me", "secret"),
				Returning("id", "created_at"),
			),
		},
		{
			"UPDATE users SET email = $1, updated_at = NOW() WHERE (id = $2)",
			Update(
				"users",
				Set("email", Arg("me@example.com")),
				Set("updated_at", Lit("NOW()")),
				Where("id", "=", Arg(1)),
			),
		},
		{
			"DELETE FROM users WHERE (id = $1)",
			Delete("users", Where("id", "=", Arg(10))),
		},
		{
			"DELETE FROM namespaces WHERE (id IN ($1) OR parent_id IN ($2) OR root_id IN ($3))",
			Delete(
				"namespaces",
				Where("id", "IN", List(1)),
				OrWhere("parent_id", "IN", List(1)),
				OrWhere("root_id", "IN", List(1)),
			),
		},
		{
			"INSERT INTO notes (title, comment) VALUES ($1, $2), ($3, $4), ($5, $6)",
			Insert(
				"notes",
				Columns("title", "comment"),
				Values("note 1", "some comment"),
				Values("note 2", "some other comment"),
				Values("note 3", "another comment"),
			),
		},
		{
			"SELECT * FROM posts ORDER BY created_at DESC, author ASC",
			Select(
				Columns("*"),
				From("posts"),
				OrderDesc("created_at"),
				OrderAsc("author"),
			),
		},
	}

	for i, test := range tests {
		built := test.q.Build()

		if test.expected != built {
			t.Errorf("tests[%d]:\n\texpected = %q\n\tgot      = %q\n", i, test.expected, built)
		}
	}
}
