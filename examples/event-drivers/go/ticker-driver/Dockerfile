FROM golang:1.11 as builder

WORKDIR /go/src/ticker

COPY driver.go .
COPY vendor vendor/

RUN CGO_ENABLED=0 GOOS=linux go build -a -o /ticker

FROM scratch
COPY --from=builder /ticker /
ENTRYPOINT ["/ticker"]
