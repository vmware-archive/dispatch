FROM golang:1.10 as builder

WORKDIR /go/src/ticker

COPY driver.go .
COPY Gopkg.toml .

RUN go get -u github.com/golang/dep/cmd/dep
RUN dep ensure
RUN go build -o /ticker driver.go
RUN chmod +x /ticker

FROM scratch
COPY --from=builder /ticker /ticker
ENTRYPOINT ["/ticker"]
