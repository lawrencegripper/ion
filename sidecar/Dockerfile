FROM golang:1.10.0 as builder

WORKDIR /go/src/github.com/lawrencegripper/mlops/sidecar/
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o sidecar .

FROM alpine:3.7
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /go/src/github.com/lawrencegripper/mlops/sidecar .
ENTRYPOINT ["./sidecar"]
