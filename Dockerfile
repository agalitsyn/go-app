FROM golang:1.13-stretch AS build
WORKDIR /src/github.com/agalitsyn/goapi
ADD . .
RUN make


FROM debian:stretch

ENV DEBIAN_FRONTEND=noninteractive \
    TERM=xterm

MAINTAINER agalitsyn

LABEL name=goapi
LABEL version=1.0.0
LABEL architecrture=amd64
LABEL source="https://github.com/agalitsyn/goapi.git"

COPY --from=build /src/github.com/agalitsyn/goapi/bin /usr/local/bin
COPY ./docs /usr/local/share/doc/goapi
