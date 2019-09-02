package query

import "testing"

type testQuery struct {
	expected string
	query    Query
}

func checkQueries(tqq []testQuery, t *testing.T) {
	for _, tq := range tqq {
		built := tq.query.Build()

		if built != tq.expected {
			t.Fatalf(
				"query not as expected:\n\texpected = '%s'\n\t  actual = '%s'",
				tq.expected,
				built,
			)
		}
	}
}

func TestSelect(t *testing.T) {
	testQueries := []testQuery{
		{
			"SELECT * FROM users WHERE username = $1",
			Select(Columns("*"), Table("users"), WhereEq("username", "me")),
		},
		{
			"SELECT * FROM users WHERE username = $1 OR email = $2",
			Select(
				Columns("*"),
				Table("users"),
				Or(
					WhereEq("username", "me"),
					WhereEq("email", "me@example.com"),
				),
			),
		},
		{
			"SELECT * FROM users WHERE username = $1 AND email != $2 AND age > $3 AND age <= $4 AND created_at >= $5 AND created_at < $6 AND name LIKE $7",
			Select(
				Columns("*"),
				Table("users"),
				Where("username", OpEq, "me"),
				Where("email", OpNotEq, "me@example.com"),
				Where("age", OpGt, 20),
				Where("age", OpLtOrEq, 100),
				Where("created_at", OpGtOrEq, 2018),
				Where("created_at", OpLt, 2019),
				Where("name", OpLike, "some category"),
			),
		},
		{
			"SELECT * FROM posts WHERE user_id IN (SELECT id FROM users WHERE username = $1)",
			Select(
				Columns("*"),
				Table("posts"),
				WhereInQuery("user_id",
					Select(
						Columns("id"),
						Table("users"),
						WhereEq("username", "me"),
					),
				),
			),
		},
		{
			"SELECT * FROM posts WHERE title LIKE $1 LIMIT 5 OFFSET 2",
			Select(
				Columns("*"),
				Table("posts"),
				WhereLike("title", "%foo%"),
				Limit(5),
				Offset(2),
			),
		},
		{
			"SELECT * FROM posts ORDER BY created_at DESC",
			Select(
				Columns("*"),
				Table("posts"),
				OrderDesc("created_at"),
			),
		},
		{
			"SELECT * FROM posts WHERE user_id = $1 AND id IN (SELECT post_id FROM tags WHERE title LIKE $2)",
			Select(
				Columns("*"),
				Table("posts"),
				WhereEq("user_id", 1234),
				WhereInQuery("id",
					Select(
						Columns("post_id"),
						Table("tags"),
						WhereLike("title", "some title"),
					),
				),
			),
		},
		{
			"SELECT * FROM posts WHERE id IN (SELECT post_id FROM tags WHERE title LIKE $1) AND category_id IN (SELECT id FROM categories WHERE name LIKE $2) AND user_id = $3",
			Select(
				Columns("*"),
				Table("posts"),
				WhereInQuery("id",
					Select(
						Columns("post_id"),
						Table("tags"),
						WhereLike("title", "some title"),
					),
				),
				WhereInQuery("category_id",
					Select(
						Columns("id"),
						Table("categories"),
						WhereLike("name", "some category"),
					),
				),
				WhereEq("user_id", 1234),
			),
		},
		{
			"SELECT * FROM users WHERE id IN ($1, $2, $3, $4, $5)",
			Select(
				Columns("*"),
				Table("users"),
				WhereIn("id", 1, 2, 3, 4, 5),
			),
		},
		{
			"SELECT * FROM objects WHERE deleted_at IS $1 AND namespace_id IN (SELECT id FROM namespaces WHERE root_id IN (SELECT namespace_id FROM collaborators WHERE user_id = $2)) OR user_id = $3 AND foo = $4",
			Select(
				Columns("*"),
				Table("objects"),
				WhereIs("deleted_at", "NULL"),
				Or(
					WhereInQuery("namespace_id",
						Select(
							Columns("id"),
							Table("namespaces"),
							WhereInQuery("root_id",
								Select(
									Columns("namespace_id"),
									Table("collaborators"),
									WhereEq("user_id", 2),
								),
							),
						),
					),
					WhereEq("user_id", 2),
				),
				WhereEq("foo", "bar"),
			),
		},
	}

	checkQueries(testQueries, t)
}

func TestInsert(t *testing.T) {
	testQueries := []testQuery{
		{
			"INSERT INTO users (email, username, password) VALUES ($1, $2, $3)",
			Insert(
				Columns("email", "username", "password"),
				Table("users"),
				Values("me@example.com", "me", "secret"),
			),
		},
		{
			"INSERT INTO users (email, username, password) VALUES ($1, $2, $3) RETURNING id, created_at",
			Insert(
				Columns("email", "username", "password"),
				Table("users"),
				Values("me@example.com", "me", "secret"),
				Returning("id", "created_at"),
			),
		},
	}

	checkQueries(testQueries, t)
}

func TestUpdate(t *testing.T) {
	testQueries := []testQuery{
		{
			"UPDATE users SET email = $1 WHERE id = $2",
			Update(
				Table("users"),
				Set("email", "me@example.com"),
				WhereEq("id", 1234),
			),
		},
		{
			"UPDATE posts SET deleted_at = NOW() WHERE user_id IN (SELECT id FROM users WHERE username = $1) RETURNING updated_at",
			Update(
				Table("posts"),
				SetRaw("deleted_at", "NOW()"),
				WhereInQuery("user_id",
					Select(
						Columns("id"),
						Table("users"),
						WhereEq("username", "me"),
					),
				),
				Returning("updated_at"),
			),
		},
	}

	checkQueries(testQueries, t)
}

func TestDelete(t *testing.T) {
	testQueries := []testQuery{
		{
			"DELETE FROM posts WHERE id = $1",
			Delete(
				Table("posts"),
				WhereEq("id", 1234),
			),
		},
	}

	checkQueries(testQueries, t)
}
