FROM golang

WORKDIR /go/src/bitbucket.org/titan098/go-dns/
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

EXPOSE 53/udp
CMD ["go-dns"]