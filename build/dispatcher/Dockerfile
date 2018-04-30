# BUILDER
FROM golang:1.10 AS builder
ARG SERVICE=dispatcher
COPY . /go/src/github.com/lawrencegripper/ion
WORKDIR /go/src/github.com/lawrencegripper/ion/cmd/$SERVICE
RUN go get -insecure -d -v
RUN rm -rf /go/src/github.com/lawrencegripper/pod2docker/vendor
RUN CGO_ENABLED=0 GOOS=linux go install -a -installsuffix cgo  -v

# RUNNER
FROM alpine:3.7
ARG SERVICE=dispatcher
RUN apk --no-cache --update add \
    ca-certificates
WORKDIR /usr/local
COPY --from=builder /go/bin/$SERVICE ./bin
COPY --from=builder /go/src/github.com/lawrencegripper/ion/configs ./etc
#RUN chmod +x $SERVICE

# it does accept the variable ${etc_dir} in the parameters
CMD ["./dispatcher", "--config", "/usr/local/etc"]