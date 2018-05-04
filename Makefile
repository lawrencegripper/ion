all: dependencies test dispatcher sidecar

dependencies:
	dep ensure -v --vendor-only

test:
	go test -v -short ./...

dispatcher:
	make -f build/dispatcher/Makefile.Docker
	
sidecar:
	make -f build/sidecar/Makefile.Docker
	
