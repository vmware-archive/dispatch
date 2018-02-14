FROM golang:1.10 as builder

COPY driver.go .
RUN go build -o timer driver.go

FROM scratch
COPY --from=builder timer .
CMD ["timer"]
