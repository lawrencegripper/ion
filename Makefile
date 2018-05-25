all: dependencies checks test dispatcher handler

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

checks:
	gometalinter --vendor --exclude=modules/helpers/Go/* --disable-all --enable=errcheck --enable=vet --enable=gofmt --enable=golint --enable=deadcode --enable=varcheck --enable=structcheck --deadline=15m ./...
