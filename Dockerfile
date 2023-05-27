# Base
FROM golang:1.20.3-alpine AS builder
RUN apk add --no-cache build-base
WORKDIR /app
COPY . /app
RUN go mod download
RUN go build ./cmd/notify

# Release
FROM alpine:3.18.0
RUN apk -U upgrade --no-cache \
    && apk add --no-cache bind-tools ca-certificates
COPY --from=builder /app/notify /usr/local/bin/

ENTRYPOINT ["notify"]