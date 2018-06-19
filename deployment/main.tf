resource "azurerm_resource_group" "batchrg" {
  name     = "${var.resource_group_name}"
  location = "${var.resource_group_location}"
}

module "aks" {
  source = "aks"

  //Defaults to using current ssh key: recomend changing as needed
  linux_admin_username      = "aks"
  linux_admin_ssh_publickey = "${file("~/.ssh/id_rsa.pub")}"

  client_id     = "${var.client_id}"
  client_secret = "${var.client_secret}"

  resource_group_name     = "${azurerm_resource_group.batchrg.name}"
  resource_group_location = "${azurerm_resource_group.batchrg.location}"
}

module "storage" {
  source                     = "storage"
  pool_bootstrap_script_path = "./scripts/poolstartup.sh"

  resource_group_name     = "${azurerm_resource_group.batchrg.name}"
  resource_group_location = "${azurerm_resource_group.batchrg.location}"
}

module "azurebatch" {
  source = "azurebatch"

  storage_account_id        = "${module.storage.id}"
  pool_bootstrap_script_url = "${module.storage.pool_boostrap_script_url}"

  resource_group_name     = "${azurerm_resource_group.batchrg.name}"
  resource_group_location = "${azurerm_resource_group.batchrg.location}"

  dedicated_node_count    = 0
  low_priority_node_count = 0
}

module "servicebus" {
  source = "servicebus"

  resource_group_name     = "${azurerm_resource_group.batchrg.name}"
  resource_group_location = "${azurerm_resource_group.batchrg.location}"
}

module "cosmos" {
  source = "cosmos"

  resource_group_name     = "${azurerm_resource_group.batchrg.name}"
  resource_group_location = "${azurerm_resource_group.batchrg.location}"
  db_name                 = "ion-cosmos"
}

module "acr" {
  source = "acr"

  resource_group_name     = "${azurerm_resource_group.batchrg.name}"
  resource_group_location = "${azurerm_resource_group.batchrg.location}"
}

module "appinsights" {
  source = "appinsights"

  resource_group_name     = "${azurerm_resource_group.batchrg.name}"
  resource_group_location = "${azurerm_resource_group.batchrg.location}"
}

data "azurerm_client_config" "current" {}

module "ion" {
  source = "ion"

  subscription_id = "${data.azurerm_subscription.current.subscription_id}"
  tenant_id       = "${data.azurerm_subscription.current.tenant_id}"

  resource_group_location = "${azurerm_resource_group.batchrg.location}"

  cluster_client_key         = "${module.aks.cluster_client_key}"
  cluster_client_certificate = "${module.aks.cluster_client_certificate}"
  cluster_ca                 = "${module.aks.cluster_ca}"
  cluster_host               = "${module.aks.host}"

  client_id     = "${var.client_id}"
  client_secret = "${var.client_secret}"

  managementapi_docker_image = "ion-management"

  batch_account_name = "${module.azurebatch.name}"

  servicebus_key  = "${module.servicebus.key}"
  servicebus_name = "${module.servicebus.name}"

  cosmos_key     = "${module.cosmos.key}"
  cosmos_name    = "${module.cosmos.name}"
  cosmos_db_name = "${module.cosmos.db_name}"
}
