FROM golang:1.18.3-alpine as build-env
RUN go install -v github.com/projectdiscovery/notify/cmd/notify@latest

FROM alpine:latest
COPY --from=build-env /go/bin/notify /usr/local/bin/notify

ENTRYPOINT ["notify"]
