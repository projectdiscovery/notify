FROM golang:1.17.5-alpine as build-env
RUN GO111MODULE=on go get -v github.com/projectdiscovery/notify/cmd/notify

FROM alpine:latest
COPY --from=build-env /go/bin/notify /usr/local/bin/notify

ENTRYPOINT ["notify"]
