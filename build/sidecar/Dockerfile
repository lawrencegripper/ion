# BUILDER
FROM golang:1.10 AS builder
ARG SERVICE=sidecar
COPY . /go/src/github.com/lawrencegripper/ion
WORKDIR /go/src/github.com/lawrencegripper/ion/cmd/$SERVICE
RUN go get -insecure -d -v
RUN rm -rf /go/src/github.com/lawrencegripper/pod2docker/vendor
RUN CGO_ENABLED=0 GOOS=linux go install -a -installsuffix cgo  -v

# RUNNER
FROM alpine:3.7
ARG SERVICE=sidecar
RUN apk --no-cache --update add \
    ca-certificates
WORKDIR /usr/local
COPY --from=builder /go/bin/$SERVICE ./bin

# it does accept the variable ${etc_dir} in the parameters
CMD ["./sidecar"]
