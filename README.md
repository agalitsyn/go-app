# Sample golang RESTful Web API

## Quickstart

```bash
$ make docker-build
$ docker-compose up # will start database

# test it
$ curl -i -X PUT localhost:8080/1.0/articles/1 --data '{"title": "new book", "slug": "new-book"}'
$ curl -i localhost:8080/1.0/articles
```
