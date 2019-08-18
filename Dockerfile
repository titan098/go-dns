FROM golang:1.12-alpine AS builder
RUN apk add --no-cache git

WORKDIR /src/github.com/titan098/go-dns/
COPY . .

RUN go install -v ./cmd/...

FROM alpine:latest
RUN apk add --no-cache ca-certificates bash

WORKDIR /root
COPY --from=builder /go/bin/go-dns .
COPY config.toml .

EXPOSE 53/udp
CMD ["./go-dns"]
