all: dependencies checks test dispatcher handler frontapi management example-modules

dependencies:
	dep ensure -v --vendor-only

test:
	go test -short ./...

integration:
	go test ./...

dispatcher:
	make -f build/dispatcher/Makefile.Docker

handler:
	make -f build/handler/Makefile.Docker

management:
	make -f build/management/Makefile.Docker
	
frontapi:
	make -f build/frontapi/Makefile.Docker

example-modules:
	 make -f modules/example/Makefile.Docker && make -f modules/downloadfile/Makefile.Docker

check-tf:
	terraform init ./deployment && terraform validate -var-file=./deployment/vars.example.tfvars ./deployment/

plan-tf:
	terraform plan -var-file=./deployment/vars.example.tfvars ./deployment

checks:
	gometalinter --vendor --disable-all --enable=errcheck --enable=vet --enable=gofmt --enable=golint --enable=deadcode --enable=varcheck --enable=structcheck --deadline=15m ./...

.PHONY: dependencies checks test dispatcher handler frontapi management