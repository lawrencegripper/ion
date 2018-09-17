# Deployment

# Intro
This folder contains the terraform template for deploying the required ion infrastructure. This includes:

- MongoDB via CosmosDB for data plane
- AMQP via ServiceBus for Eventing
- Azure Blob Storage for data storage
- Azure Batch for GPU Compute
- Application Insights for Logging
- Azure Container repository for a private docker image registry