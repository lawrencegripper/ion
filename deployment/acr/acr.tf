variable "resource_group_name" {
  type        = "string"
  description = "Name of the azure resource group."
}

variable "resource_group_location" {
  type        = "string"
  description = "Location of the azure resource group."
}

resource "random_string" "name" {
  keepers = {
    # Generate a new id each time we switch to a new resource group
    group_name = "${var.resource_group_name}"
  }

  length  = 8
  upper   = false
  special = false
  number  = false
}

resource "azurerm_container_registry" "deploy" {
  name = "ionacr${random_string.name.result}"

  location            = "${var.resource_group_location}"
  resource_group_name = "${var.resource_group_name}"

  admin_enabled = true
  sku           = "Standard"
}

output "username" {
  value = "${azurerm_container_registry.deploy.admin_username}"
}

output "password" {
  value = "${azurerm_container_registry.deploy.admin_password}"
}

output "login_server" {
  value = "${azurerm_container_registry.deploy.login_server}"
}
