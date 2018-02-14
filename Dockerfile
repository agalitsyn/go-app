FROM golang:1.9-alpine AS build-env

ENV GOPATH=/ \
    GOOS=linux \
    GOARCH=amd64 \
    CGO_ENABLED=0

RUN apk update && \
    apk add git && \
    apk add make

WORKDIR /src/github.com/agalitsyn/goapi
ADD . .

RUN make    


FROM alpine:3.7
MAINTAINER agalitsyn

LABEL name=goapi
LABEL version=1.0.0
LABEL architecrture=amd64
LABEL source="ssh://git@github.com:agalitsyn/goapi.git"

COPY --from=build-env /src/github.com/agalitsyn/goapi/bin/goapi /usr/local/bin/goapi
COPY ./docs /usr/local/share/doc/goapi

EXPOSE 5000
ENTRYPOINT ["/usr/local/bin/goapi"]
