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

### Install dev tools

```
$ make install-tools
```

### Build and Run

```
$ make generate-certificates
$ make start

```
It will start web service with TLS.

Or you can simply run `goapi` and get simple HTTP service.

### Test with cURL

```
# https
$ curl --cacert ./ca.pem --key ./client-key.pem --cert ./client.pem https://127.0.0.1:5000/

# http
$ curl http://127.0.0.1:5000
```
