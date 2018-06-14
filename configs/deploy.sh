az group create -g temp-ion1 -l westeurope

az group deployment create -g temp-ion1 --template-file ./azurebatch-gpu.json --parameters @./azurebatch-gpu-params.json
