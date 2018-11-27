FROM golang:1.11 as builder

WORKDIR /go/src/cron

COPY driver.go .
COPY vendor vendor/

RUN CGO_ENABLED=0 GOOS=linux go build -a -o /cron

FROM scratch
COPY --from=builder /cron /
ENTRYPOINT ["/cron"]