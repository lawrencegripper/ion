# BUILDER
FROM golang:1.10 AS builder
ARG SERVICE=sidecar
ADD https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep
COPY . /go/src/github.com/lawrencegripper/ion
WORKDIR /go/src/github.com/lawrencegripper/ion/cmd/$SERVICE
RUN dep ensure -v --vendor-only
RUN rm -rf /go/src/github.com/lawrencegripper/pod2docker/vendor
RUN CGO_ENABLED=0 GOOS=linux go install -a -installsuffix cgo  -v

# RUNNER
FROM alpine:3.7
ARG SERVICE=sidecar
RUN apk --no-cache --update add \
    ca-certificates
WORKDIR /usr/local/bin
COPY --from=builder /go/bin/$SERVICE .

# it does accept the variable $SERVICE
ENTRYPOINT ["sidecar"]
