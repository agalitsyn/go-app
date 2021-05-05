package rdb

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/Masterminds/squirrel"

	"github.com/agalitsyn/go-app/internal/pkg/postgres"
	"github.com/agalitsyn/go-app/internal/storage"
)

type ArticleStorage struct {
	db *postgres.DB
}

func NewArticleStorage(db *postgres.DB) *ArticleStorage {
	return &ArticleStorage{db: db}
}

func (s *ArticleStorage) FilterArticles(ctx context.Context, params storage.ArticleFilter) ([]storage.Article, error) {
	qb := squirrel.Select("id", "title", "slug").
		From("article").
		OrderBy("id ASC")

	if params.Limit > 0 {
		qb = qb.Limit(params.Limit)
	}

	query, args := qb.PlaceholderFormat(squirrel.Dollar).MustSql()
	rows, err := s.db.Session.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("could not perform query: %w", err)
	}
	defer rows.Close()

	res := make([]storage.Article, 0, params.Limit)
	for rows.Next() {
		var article storage.Article
		err = rows.Scan(
			&article.ID,
			&article.Title,
			&article.Slug,
		)
		if err != nil {
			return nil, fmt.Errorf("could not scan row: %w", err)
		}
		res = append(res, article)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("could not iterate rows: %w", err)
	}

	return res, nil
}

func (s *ArticleStorage) StoreArticles(ctx context.Context, articles []storage.Article) error {
	if len(articles) == 0 {
		return nil
	}

	// dedup for preventing ON CONFLICT loop
	// sort for preventing deadlocks
	unique := make(map[string]storage.Article, len(articles))
	ordered := make([]string, 0, len(articles))
	for _, obj := range articles {
		if obj.Slug == "" {
			continue
		}

		slug := strings.ToLower(obj.Slug)
		_, visited := unique[slug]
		if visited {
			continue
		}
		unique[slug] = obj
		ordered = append(ordered, slug)
	}
	sort.Strings(ordered)

	qb := squirrel.Insert("article").Columns("title", "slug")
	for _, slug := range ordered {
		obj := unique[slug]
		qb = qb.Values(obj.Title, slug)
	}
	const onConflict = `ON CONFLICT (slug) DO UPDATE SET
		title = excluded.title,
		slug = excluded.slug
	`
	qb = qb.Suffix(onConflict)

	query, args, err := qb.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("could not build query: %w", err)
	}

	if _, err = s.db.Session.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("could not perform query: %w", err)
	}

	return nil
}

func (s *ArticleStorage) DeleteArticles(ctx context.Context, articles []storage.Article) error {
	if len(articles) == 0 {
		return nil
	}

	slugs := make([]string, 0, len(articles))
	for i := range articles {
		if articles[i].Slug != "" {
			slugs = append(slugs, strings.ToLower(articles[i].Slug))
		}
	}

	// language=PostgreSQL
	const query = `DELETE FROM article WHERE slug = ANY($1);`
	_, err := s.db.Session.Exec(ctx, query, slugs)
	if err != nil {
		return fmt.Errorf("could not perform query: %w", err)
	}
	return nil
}
