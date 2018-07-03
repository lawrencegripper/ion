.PHONY: all
all: dependencies checks test compile-grpc dispatcher handler frontapi management ioncli example-modules

dependencies:
	dep ensure -v --vendor-only

test:
	go test -short ./...

integration:
	go test ./...

ioncli:
	make -f build/ion/Makefile.Docker

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

checks:
	gometalinter --vendor --disable-all --enable=errcheck --enable=vet --enable=gofmt --enable=golint --enable=deadcode --enable=varcheck --enable=structcheck --deadline=15m ./...

plan-tf:
	terraform plan -var-file=./deployment/vars.example.tfvars ./deployment

compile-grpc:
	cd ./internal/pkg/management/module && rm *.pb.go && protoc -I . module.proto --go_out=plugins=grpc:. && cd -