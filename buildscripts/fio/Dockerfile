FROM alpine:3.22.1

RUN apk update && \
    apk add --no-cache fio && \
    rm -rf /var/cache/apk/*

ENTRYPOINT ["fio"]
