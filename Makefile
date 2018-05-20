all: dependencies test dispatcher handler

dependencies:
	dep ensure -v --vendor-only

test:
	go test -v -short ./...

integration:
	go test ./...

dispatcher:
	make -f build/dispatcher/Makefile.Docker
	
handler:
	make -f build/handler/Makefile.Docker
	
