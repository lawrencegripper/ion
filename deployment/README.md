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
7. In kubectl run `kubectl port-forward ion-management-api-*** 9000:9000` (replace *** with your api pods name)
8. Connect to the API using the client and create modules!

## Deploying and running module

The cli can be used to deploy modules into the cluster. 

1. Forward the port 9000 to your management API instance `kubectl port-forward ion-management-api-**** 9000:9000` and `kubectl port-forward ion-front-api**** 9001:9001` (replace `****`'s with the name show in your cluster)
2. Deploy a module as follows: `docker run --network host ion-cli module create -i frontapi.new_link -o file_downloaded -n downloader -m ion-module-download-file -p kubernetes`
3. Run `curl --header "Content-Type: application/json"   --request POST   --data '{"url": "http://google.co.uk"}'   http://localhost:9001/`
