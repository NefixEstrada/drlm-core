FROM golang:1.13.6-alpine3.10 AS builder

COPY . /tmp/build
WORKDIR /tmp/build

RUN apk add --no-cache \
        make

RUN make build

FROM alpine:3.10
COPY --from=builder /tmp/build/drlm-core /
CMD [ "/drlm-core" ]