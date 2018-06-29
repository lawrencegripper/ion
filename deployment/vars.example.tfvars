// Provide the Client ID of a service principal for use by AKS
client_id = "00000000-0000-0000-0000-000000000000"

// Provide the Client Secret of a service principal for use by AKS
client_secret = "00000000-0000-0000-0000-000000000000"

// The resource group you would like to deploy too
resource_group_name = "temp-iontf"

// The location of all resources
resource_group_location = "westeurope"

// The number of dedicated AzureBatch Compute nodes to use
batch_dedicated_node_count = "1"

// The number of low priority AzureBatch Compute nodes to use
batch_low_priority_node_count = "2"

// The number of nodes to create in the AKS cluster
aks_node_count = "2"
