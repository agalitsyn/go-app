package article

import (
	"database/sql"

	"github.com/lib/pq"
	"github.com/pkg/errors"
)

var ErrNotFound = errors.New("not found")

type Article struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Slug  string `json:"slug"`
}

type Manager struct {
	db *sql.DB
}

func NewManager(db *sql.DB) *Manager {
	return &Manager{db: db}
}

func (m *Manager) Save(a *Article) error {
	_, err := m.db.Exec("INSERT INTO article(title, slug) VALUES ($1, $2);", a.Title, a.Slug)
	if err != nil {
		return errors.Wrap(err, "could not save article")
	}
	return nil
}

func (m *Manager) Update(a *Article) error {
	_, err := m.db.Exec("UPDATE article SET title = $2, slug = $3 WHERE id = $1;", a.ID, a.Title, a.Slug)
	if err != nil {
		return errors.Wrap(err, "could not update article")
	}
	return nil
}

func (m *Manager) Delete(a *Article) error {
	_, err := m.db.Exec("DELETE FROM article WHERE id = $1;", a.ID)
	if err != nil {
		return errors.Wrap(err, "could not delete article")
	}
	return nil
}

func (m *Manager) ByID(id string) (*Article, error) {
	articles, err := m.ByIDs([]string{id})
	if err != nil {
		return nil, err
	}
	if len(articles) == 0 {
		return nil, ErrNotFound
	}
	return articles[0], nil
}

func (m *Manager) ByIDs(ids []string) ([]*Article, error) {
	rows, err := m.db.Query("SELECT id, title, slug FROM article WHERE id = ANY($1);", pq.Array(ids))
	if err != nil {
		return nil, errors.Wrap(err, "could not get articles by ids")
	}
	defer rows.Close()

	var articles []*Article
	for rows.Next() {
		device, err := scan(rows)
		if err != nil {
			return nil, err
		}
		articles = append(articles, device)
	}
	err = rows.Err()
	if err != nil {
		return nil, errors.Wrap(err, "could not get articles by ids")
	}
	return articles, nil
}

func (m *Manager) All() ([]*Article, error) {
	rows, err := m.db.Query("SELECT id, title, slug FROM article;")
	if err != nil {
		return nil, errors.Wrap(err, "could not get articles")
	}
	defer rows.Close()

	var articles []*Article
	for rows.Next() {
		device, err := scan(rows)
		if err != nil {
			return nil, err
		}
		articles = append(articles, device)
	}
	err = rows.Err()
	if err != nil {
		return nil, errors.Wrap(err, "could not get articles")
	}
	return articles, nil
}

func scan(rows *sql.Rows) (*Article, error) {
	var a Article
	err := rows.Scan(&a.ID, &a.Title, &a.Slug)
	if err != nil {
		return nil, errors.Wrapf(err, "could not scan row to article model")
	}
	return &a, nil
}
