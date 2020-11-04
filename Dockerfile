FROM golang:1.14-alpine AS builder
RUN apk add --no-cache git
RUN GO111MODULE=auto go get -u -v github.com/projectdiscovery/notify/cmd/notify

FROM alpine:latest
COPY --from=builder /go/bin/notify /usr/local/bin/

ENTRYPOINT ["notify"]
