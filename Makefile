.PHONY: dependencies management checks test dispatcher handler frontapi ioncli example-modules
all: dependencies checks test dispatcher handler frontapi management ioncli example-modules

dependencies:
	# dep ensure -v --vendor-only
	echo "skipping vendor until vcabbage PR is merged"

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
	cd ./internal/pkg/management/module && rm -f *.pb.go && protoc -I . module.proto --go_out=plugins=grpc:. && cd -
	cd ./internal/pkg/management/trace && rm -f *.pb.go && protoc -I . trace.proto --go_out=plugins=grpc:. && cd -
	make -f build/management/Makefile.Docker
	
frontapi:
	make -f build/frontapi/Makefile.Docker

example-modules:
	 make -f modules/transcode/Makefile.Docker
	 make -f modules/example/Makefile.Docker
	 make -f modules/downloadfile/Makefile.Docker

check-tf:
	terraform init ./deployment && terraform validate -var-file=./deployment/vars.example.tfvars ./deployment/

checks:
	gometalinter --vendor --disable-all --enable=errcheck --enable=vet --enable=gofmt --enable=golint --enable=deadcode --enable=varcheck --enable=structcheck --deadline=15m ./...

plan-tf:
	terraform plan -var-file=./deployment/vars.example.tfvars ./deployment
