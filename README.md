[![Build Status](https://travis-ci.org/agalitsyn/goapi.svg?branch=master)](https://travis-ci.org/agalitsyn/goapi)

# goapi

This is example of golang app, which has minimal skeleton for web app and some dev tools.

Inspired by [article](https://medium.com/@kelseyhightower/12-fractured-apps-1080c73d481c#.ihna7diaw).
Partially based on repo https://github.com/kelseyhightower/app

Don't use for production.

## Usage

Download:

```
go get https://github.com/agalitsyn/goapi
```

### Review settings

```
$ cp .env.default .env
$ vi .env
```

### Build and Run

```
$ make

```

For database you can use docker image https://hub.docker.com/_/postgres/

### Test with cURL

```
$ curl --cacert ./ca.pem --key ./client-key.pem --cert ./client.pem https://127.0.0.1:5000/
```
