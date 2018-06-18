variable "resource_group_name" {
  type        = "string"
  description = "Name of the azure resource group."
  default     = "akc-rg"
}

variable "resource_group_location" {
  type        = "string"
  description = "Location of the azure resource group."
  default     = "eastus"
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

resource "azurerm_servicebus_namespace" "ion" {
  name                = "ionsb-${random_string.name.result}"
  location            = "${var.resource_group_location}"
  resource_group_name = "${var.resource_group_name}"
  sku                 = "basic"

  tags {
    source = "terraform"
  }
}

output "key" {
  value = "${azurerm_servicebus_namespace.ion.default_primary_key}"
}

output "name" {
  value = "${azurerm_servicebus_namespace.ion.name}"
}
