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
			"SELECT * FROM users WHERE id IN ($1, $2, $3, $4, $5)",
			Select(
				Columns("*"),
				Table("users"),
				WhereIn("id", 1, 2, 3, 4, 5),
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
			"UPDATE posts SET deleted_at = NOW() WHERE user_id IN (SELECT id FROM users WHERE username = $1)",
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
