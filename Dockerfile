FROM golang:1.14.2-alpine3.11 AS builder

COPY . /tmp/build
WORKDIR /tmp/build

RUN apk add --no-cache \
        make

RUN make build

FROM alpine:3.10
COPY --from=builder /tmp/build/drlm-core /
CMD [ "/drlm-core" ]