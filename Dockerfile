FROM golang:1.16 as builder

ENV GOBIN=/go/src/app/bin

WORKDIR /go/src/app

ADD . .
RUN make


FROM debian:buster

ENV DEBIAN_FRONTEND=noninteractive \
    TERM=xterm

RUN apt-get update && \
    apt-get install --yes --no-install-recommends \
        ca-certificates && \
    apt-get clean

COPY --from=builder /go/src/app/bin/api /usr/local/bin/api
COPY ./docs /usr/local/share/doc/go-app

CMD ["api"]
