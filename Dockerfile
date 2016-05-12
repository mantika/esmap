FROM gliderlabs/alpine:latest

RUN apk add --update curl jq && \
    rm -rf /var/cache/apk/*

ENTRYPOINT ["/usr/bin/esmap"]
