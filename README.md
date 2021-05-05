# go-app

* Ready for creating Go applications in one repository.
* Includes example of RESTful API service, but you can add more service types, for example workers, consumers, etc.
* Go 1.16 + docker + docker-compose - easy setup on own server. No CI server required, just `make docker-export` and `scp` to your server.

## Repository structure

* Structure supports multiple services:
    * Each service has own directory in `cmd` folder for `main` function, which typically consists of dependencies and service start.
    * Service logic with tests are in `internal/app`.
* Shared packages are in `internal/pkg`.
* Storage entities and interfaces are in `internal/storage`. Interfaces can have multiple implementations of storages types like `internal/storage/{rdb,memory,kv,object,document}`.
* All applications build in one docker image for easy distribution, so you should define explicit `command` in `docker-compose` or in k8s files for each service.

## Migrations

This template includes `PostgreSQL` connector, with [tern](https://github.com/jackc/tern) migrator. Check `migrations` folder.

## Quickstart

```bash
make docker-build
docker-compose up
make migrate

# test it
curl -i -X POST localhost:8080/1.0/articles --data '{"title": "New Book", "slug": "new-book"}'
curl -i -X GET localhost:8080/1.0/articles
```

## Local development

```bash
cp .env.example .env.local
docker-compose up postgres
make migrate
make run
```

### Testing

* Service tests with mocked storages are in `internal/app`.
* Storage tests are in `internal/storage`.

For testing postgres storages run `docker-compose up postgres`. Can be skipped with `make test-short`.
