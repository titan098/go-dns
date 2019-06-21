FROM golang AS builder

WORKDIR /src/bitbucket.org/titan098/go-dns/
COPY . .

RUN go get -d -v ./...
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o go-dns .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root
COPY --from=builder /src/bitbucket.org/titan098/go-dns/go-dns .
COPY config.toml .

EXPOSE 53/udp
CMD ["./go-dns"]
