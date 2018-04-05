FROM golang:1.10.0 as builder

# Download and install the latest release of dep
ADD https://github.com/golang/dep/releases/download/v0.4.1/dep-linux-amd64 /usr/bin/dep
RUN chmod +x /usr/bin/dep
RUN go get -u github.com/alecthomas/gometalinter
RUN gometalinter --install

WORKDIR /go/src/github.com/lawrencegripper/ion/

#Restore dep for sidecar
WORKDIR /go/src/github.com/lawrencegripper/ion/sidecar
COPY ./sidecar/Gopkg.lock .
COPY ./sidecar/Gopkg.toml .
RUN dep ensure -v -vendor-only
COPY ./sidecar .
RUN go test -v -race -short ./...
RUN gometalinter --vendor --disable-all --enable=errcheck --enable=vet --enable=gofmt --enable=golint --enable=deadcode --enable=varcheck --enable=structcheck --deadline=15m ./...

