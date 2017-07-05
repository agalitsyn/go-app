package article

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/agalitsyn/goapi/log"
	"github.com/agalitsyn/goapi/router"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestListHandler(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	m := &Manager{db: db}

	mock.ExpectQuery("SELECT id, title, slug FROM article;").
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug"}).
			AddRow(1, "Новая", "new"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)

	r := router.New(router.WithLogging(log.New("", "", ioutil.Discard)))
	r.Get("/", makeHandler(m, listHandler))
	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("unexpected status: %v", resp.StatusCode)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	if string(body) == "" {
		t.Errorf("unexpected body: %v", string(body))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestDeleteHandler(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	m := &Manager{db: db}

	// delete first time
	mock.ExpectQuery("SELECT id, title, slug FROM article WHERE id = ANY(.+);").
		WithArgs(`{"1"}`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug"}).
			AddRow(1, "Новая", "new"))
	mock.ExpectExec("DELETE FROM article WHERE id = \\$1;").WithArgs("1").WillReturnResult(sqlmock.NewResult(0, 1))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "http://example.com/1", nil)

	r := router.New(router.WithLogging(log.New("", "", ioutil.Discard)))
	r.Delete("/:articleID", makeHandler(m, deleteHandler))
	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("unexpected status: %v", resp.StatusCode)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	if string(body) != "" {
		t.Errorf("unexpected body: %v", string(body))
	}

	// check that article was deleted and not found now
	mock.ExpectQuery("SELECT id, title, slug  FROM article WHERE id = ANY(.+);").
		WithArgs(`{"1"}`).
		WillReturnRows(sqlmock.NewRows([]string{}))

	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	nextResp := w.Result()
	if nextResp.StatusCode != http.StatusNotFound {
		t.Errorf("unexpected status: %v", nextResp.StatusCode)
	}
	nextBody, _ := ioutil.ReadAll(nextResp.Body)
	if string(nextBody) == "" {
		t.Errorf("unexpected body: %v", string(nextBody))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestPutHandler_Update(t *testing.T) {
	toUpdate := `{
		"title": "Не новая",
		"slug": "not-new"
	}`
	buf := bytes.NewBufferString(toUpdate)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "http://example.com/1", buf)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT id, title, slug FROM article WHERE id = ANY(.+);").
		WithArgs(`{"1"}`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug"}).
			AddRow(1, "Новая", "new"))

	mock.ExpectExec("UPDATE article SET title = \\$2, slug = \\$3 WHERE id = \\$1;").
		WithArgs("1", "Не новая", "not-new").
		WillReturnResult(sqlmock.NewResult(0, 1))

	m := &Manager{db: db}

	r := router.New(router.WithLogging(log.New("", "", ioutil.Discard)))
	r.Put("/:articleID", makeHandler(m, putHandler))
	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("unexpected status: %v", resp.StatusCode)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	if string(body) == "" {
		t.Errorf("unexpected body: %v", string(body))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}

func TestPutHandler_Create(t *testing.T) {
	toCreate := `{
		"title": "Новая",
		"slug": "new"
	}`
	buf := bytes.NewBufferString(toCreate)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "http://example.com/1", buf)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT id, title, slug FROM article WHERE id = ANY(.+);").
		WithArgs(`{"1"}`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "slug"}))

	mock.ExpectExec("INSERT INTO article").
		WithArgs("Новая", "new").
		WillReturnResult(sqlmock.NewResult(0, 1))

	m := &Manager{db: db}

	r := router.New(router.WithLogging(log.New("", "", ioutil.Discard)))
	r.Put("/:articleID", makeHandler(m, putHandler))
	r.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("unexpected status: %v", resp.StatusCode)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	if string(body) == "" {
		t.Errorf("unexpected body: %v", string(body))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expections: %s", err)
	}
}
