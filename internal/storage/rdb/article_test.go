package rdb

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agalitsyn/go-app/internal/pkg/postgres"
	"github.com/agalitsyn/go-app/internal/storage"
)

func TestArticleStorage_DeleteArticles(t *testing.T) {
	t.Parallel()

	db, teardown := setupTestDB(t)
	defer teardown()

	err := migrateArticle(db)
	require.NoError(t, err)

	ctx := context.Background()
	store := NewArticleStorage(db)

	err = loadArticles(store, []storage.Article{{
		ID:    1,
		Title: "Foo",
		Slug:  "foo",
	}})
	require.NoError(t, err)

	err = store.DeleteArticles(ctx, []storage.Article{{
		Slug: "Foo",
	}})
	require.NoError(t, err)

	c, err := countArticles(db)
	require.NoError(t, err)

	assert.Equal(t, 0, c)
}

func TestArticleStorage_FilterArticles(t *testing.T) {
	t.Parallel()

	db, teardown := setupTestDB(t)
	defer teardown()

	err := migrateArticle(db)
	require.NoError(t, err)

	ctx := context.Background()
	store := NewArticleStorage(db)

	err = loadArticles(store, []storage.Article{
		{
			ID:    1,
			Title: "Foo",
			Slug:  "foo",
		},
		{
			ID:    2,
			Title: "Bar",
			Slug:  "bar",
		},
	})
	require.NoError(t, err)

	t.Run("limit", func(t *testing.T) {
		articles, err := store.FilterArticles(ctx, storage.ArticleFilter{Limit: 1})
		require.NoError(t, err)

		assert.Equal(t, 1, len(articles))
	})

	t.Run("ordering", func(t *testing.T) {
		articles, err := store.FilterArticles(ctx, storage.ArticleFilter{Limit: 1})
		require.NoError(t, err)
		require.Greater(t, len(articles), 0)

		assert.Equal(t, "foo", articles[0].Slug)
	})

	t.Run("unlimited", func(t *testing.T) {
		articles, err := store.FilterArticles(ctx, storage.ArticleFilter{})
		require.NoError(t, err)

		assert.Equal(t, 2, len(articles))
	})
}

func TestArticleStorage_StoreArticles(t *testing.T) {
	t.Parallel()

	db, teardown := setupTestDB(t)
	defer teardown()

	err := migrateArticle(db)
	require.NoError(t, err)

	ctx := context.Background()
	store := NewArticleStorage(db)

	err = loadArticles(store, []storage.Article{
		{
			ID:    1,
			Title: "Foo",
			Slug:  "foo",
		},
	})
	require.NoError(t, err)

	t.Run("update by slug", func(t *testing.T) {
		err := store.StoreArticles(ctx, []storage.Article{
			{
				Title: "Bar",
				Slug:  "foo",
			},
		})
		require.NoError(t, err)

		c, err := countArticles(db)
		require.NoError(t, err)

		assert.Equal(t, 1, c)
	})

	t.Run("new", func(t *testing.T) {
		err := store.StoreArticles(ctx, []storage.Article{
			{
				Title: "Foobar",
				Slug:  "foobar",
			},
		})
		require.NoError(t, err)

		c, err := countArticles(db)
		require.NoError(t, err)

		assert.Equal(t, 2, c)
	})
}

func countArticles(db *postgres.DB) (int, error) {
	return countRows(context.Background(), "article", db)
}

func migrateArticle(db *postgres.DB) error {
	// language=PostgreSQL
	const schema = `
		CREATE TABLE article (
			id          SERIAL      PRIMARY KEY,
			title       text        NOT NULL,
			slug        text        UNIQUE NOT NULL
		);
	`
	_, err := db.Session.Exec(context.Background(), schema)
	return err
}

func loadArticles(s *ArticleStorage, articles []storage.Article) error {
	for _, v := range articles {
		// language=PostgreSQL
		const query = `INSERT INTO article ( id, title, slug ) VALUES ( $1, $2, $3 )`
		_, err := s.db.Session.Exec(context.Background(), query, v.ID, v.Title, v.Slug)
		if err != nil {
			return err
		}
	}

	return nil
}
