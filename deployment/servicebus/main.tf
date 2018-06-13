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

variable "servicebus_name" {
  description = "Input your unique Azure service bus name"
}

resource "azurerm_servicebus_namespace" "ion" {
  name                = "${var.servicebus_name}"
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
  value = "${var.servicebus_name}"
}
