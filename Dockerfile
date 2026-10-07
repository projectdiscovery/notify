FROM alpine:3.18.2

LABEL org.opencontainers.image.authors="ProjectDiscovery"
LABEL org.opencontainers.image.description="Notify is a Go-based assistance package that enables you to stream the output of several tools (or read from a file) and publish it to a variety of supported platforms."
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.title="notify"
LABEL org.opencontainers.image.url="https://github.com/projectdiscovery/notify"

RUN apk -U upgrade --no-cache \
    && apk add --no-cache bind-tools ca-certificates

ARG TARGETPLATFORM
COPY $TARGETPLATFORM/notify /usr/local/bin/

ENTRYPOINT ["notify"]
