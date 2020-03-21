# Go API

Example of RESTful API written in Go.

Purpose was to make kind of copy&paste template for other projects. But this is outdated, actually I wrote differently in Go nowadays. What should be fixed:
* Migrate from `github.com/lib/pq` to `github.com/jackc/pgx`
* Add `github.com/jmoiron/sqlx`
* Move all shared packages off root folder
* Keep specific packages in `cmd/internal`
* Maybe add swagger

## Quickstart

```bash
$ make docker-build
$ docker-compose up

# test it
$ curl -i -X PUT localhost:8080/1.0/articles/1 --data '{"title": "new book", "slug": "new-book"}'
$ curl -i localhost:8080/1.0/articles
```
