curl -L -o - https://github.com/sl1pm4t/terraform-provider-kubernetes/releases/download/v1.0.7-custom/terraform-provider-kubernetes_linux-amd64.gz | gunzip > terraform-provider-kubernetes
chmod +x terraform-provider-kubernetes

mkdir -p $GOPATH/src/github.com/terraform-providers; cd $GOPATH/src/github.com/terraform-providers
git clone git@github.com:terraform-providers/terraform-provider-tls
cd $GOPATH/src/github.com/terraform-providers/terraform-provider-tls
make build