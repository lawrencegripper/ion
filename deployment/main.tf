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

  node_count = "${var.aks_node_count}"
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

  dedicated_node_count    = "${var.batch_dedicated_node_count}"
  low_priority_node_count = "${var.batch_low_priority_node_count}"
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

  subscription_id = "${data.azurerm_client_config.current.subscription_id}"
  tenant_id       = "${data.azurerm_client_config.current.tenant_id}"

  resource_group_location = "${azurerm_resource_group.batchrg.location}"
  resource_group_name     = "${azurerm_resource_group.batchrg.name}"

  cluster_client_key         = "${module.aks.cluster_client_key}"
  cluster_client_certificate = "${module.aks.cluster_client_certificate}"
  cluster_ca                 = "${module.aks.cluster_ca}"
  cluster_host               = "${module.aks.host}"

  client_id     = "${var.client_id}"
  client_secret = "${var.client_secret}"

  batch_account_name = "${module.azurebatch.name}"

  servicebus_key  = "${module.servicebus.key}"
  servicebus_name = "${module.servicebus.name}"

  storage_name = "${module.storage.name}"
  storage_key  = "${module.storage.key}"

  cosmos_key     = "${module.cosmos.key}"
  cosmos_name    = "${module.cosmos.name}"
  cosmos_db_name = "iondb"

  acr_url      = "${module.acr.login_server}"
  acr_username = "${module.acr.username}"
  acr_password = "${module.acr.password}"

  managementapi_docker_image = "${var.docker_root}/ion-management:${var.docker_tag}"
  dispatcher_docker_image    = "${var.docker_root}/ion-dispatcher:${var.docker_tag}"
  frontapi_docker_image      = "${var.docker_root}/ion-frontapi:${var.docker_tag}"

  app_insights_key = "${module.appinsights.instrumentation_key}"
}

output "kubeconfig" {
  value     = "${module.aks.kubeconfig}"
  sensitive = true
}

output "cluster_name" {
  value = "${module.aks.cluster_name}"
}

output "resource_group_name" {
  value = "${var.resource_group_name}"
}

output "acr_url" {
  value = "${module.acr.login_server}"
}

output "acr_username" {
  value = "${module.acr.username}"
}

output "acr_password" {
  sensitive = true
  value     = "${module.acr.password}"
}

output "client_cert" {
  value     = "${module.ion.client_cert}"
  sensitive = true
}

output "client_key" {
  value     = "${module.ion.client_key}"
  sensitive = true
}

output "cluster_ca" {
  value     = "${module.ion.cluster_ca}"
  sensitive = true
}

output "server_cert" {
  value     = "${module.ion.server_cert}"
  sensitive = true
}

output "server_key" {
  value     = "${module.ion.server_key}"
  sensitive = true
}

output "ion_management_endpoint" {
  value = "${module.ion.ion_management_endpoint}"
}
