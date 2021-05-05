package storage

import (
	"context"
	"errors"
)

var ErrArticleNotFound = errors.New("not found")

type Article struct {
	ID    int
	Title string
	Slug  string
}

type ArticleFilter struct {
	Limit uint64
}

type ArticleRepository interface {
	FilterArticles(ctx context.Context, params ArticleFilter) ([]Article, error)
	StoreArticles(ctx context.Context, articles []Article) error
	DeleteArticles(ctx context.Context, articles []Article) error
}
