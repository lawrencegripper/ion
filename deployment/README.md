# Deployment

# Intro

This folder contains the terraform template for deploying the required ion infrastructure. This includes:

- MongoDB via CosmosDB for data plane
- AMQP via ServiceBus for Eventing
- Azure Blob Storage for data storage
- Azure Batch for GPU Compute
- Application Insights for Logging
- Azure Container repository for a private docker image registry

# Gettings Started

1. Setup Terraform for Azure following [this guide here](https://docs.microsoft.com/en-us/azure/virtual-machines/linux/terraform-install-configure)
2. From the commandline move to the deployment folder `cd ./deployment` then edit `vars.example.tfvars` adding in your Service Principal details and adjusting node codes as required.
3. Download the latest version of the Community Kubernetes Provider for Terraform. Get the correct link [from here](https://github.com/sl1pm4t/terraform-provider-kubernetes/releases) and use it as follows: (Current official Terraform K8s provider doesn't support `Deployments`) or use the scripts `bootstrap_linux.sh`, `bootstrap_mac.sh` and `bootstrap_windows.ps1`


```shell
curl -L -o - PUT_RELASE_BINARY_LINK_YOU_FOUND_HERE | gunzip > terraform-provider-kubernetes
chmod +x ./terraform-provider-kubernetes
```

4. Use `terraform init` to initialize the template
5. Use `terraform plan -var-file=./vars.example.tfvars` and `terraform apply -var-file=./vars.example.tfvars` to deploy the template
6. Use `az aks get-credentials` to connect to your new AKS cluster
7. Grab the client and ca certificates
```shell
terraform output ion_ca > ~/ca.pem
terraform output ion_client_cert > ~/client.crt
terraform output ion_client_key > ~/client.key
```
8. Grab the management API fqdn
```
export ION_FQDN=$(terraform output ion_fqdn)
```
9. Connect to the API using the CLI
```
ion module list --endpoint "$ION_FQDN:9000" --certfile ~/client.crt --keyfile ~/client.key --cacertfile ~/ca.pem
```
10. Manage your Ion resources via the CLI

## Deploying and running module

The cli can be used to deploy modules into the cluster.

1. Deploy a module as follows: `docker run --network host ion-cli module create -i frontapi.new_link -o file_downloaded -n downloader -m ion-module-download-file -p kubernetes`
2. Run `curl --header "Content-Type: application/json"   --request POST   --data '{"url": "http://google.co.uk"}'   http://localhost:9001/`
