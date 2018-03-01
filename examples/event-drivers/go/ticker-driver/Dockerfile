FROM golang:1.10 as builder

COPY driver.go .
RUN go build -o ticker driver.go

FROM scratch
COPY --from=builder ticker .
ENTRYPOINT ["/ticker"]
