FROM golang:1.10.0 as builder

# Download and install the latest release of dep
ADD https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep

# Copy common dependencies
WORKDIR /go/src/github.com/lawrencegripper/ion/common
COPY ./common .

WORKDIR /go/src/github.com/lawrencegripper/ion/dispatcher/
COPY ./dispatcher/Gopkg.toml ./dispatcher/Gopkg.lock ./
RUN dep ensure -v --vendor-only

COPY ./dispatcher/ .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o dispatcher .

FROM alpine:3.7
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /go/src/github.com/lawrencegripper/ion/dispatcher .
ENTRYPOINT ["./dispatcher"]
