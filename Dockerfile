FROM alpine:3.5

RUN apk add --no-cache bash ca-certificates

COPY ./bin/pusher /usr/local/bin/pusher
COPY ./docs /usr/local/share/doc/pusher

EXPOSE 5000
