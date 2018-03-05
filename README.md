# Sample golang RESTful Web API

[![Build Status](https://travis-ci.org/agalitsyn/goapi.svg?branch=master)](https://travis-ci.org/agalitsyn/goapi)

## Quickstart

```bash
$ docker-compose start -d # will start database
$ make install && goapi

# test it
$ curl -X PUT localhost:5000/1.0/articles/1 --data '{"title": "new book", "slug": "new-book"}'
$ curl  localhost:5000/1.0/articles
```
