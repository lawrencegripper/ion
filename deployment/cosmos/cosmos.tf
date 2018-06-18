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

variable "db_name" {
  type = "string"
}

variable "failover_location" {
  default = "northeurope"
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

resource "azurerm_cosmosdb_account" "db" {
  name                = "iondb-${random_string.name.result}"
  location            = "${var.resource_group_location}"
  resource_group_name = "${var.resource_group_name}"
  offer_type          = "Standard"
  kind                = "MongoDB"

  consistency_policy {
    consistency_level = "BoundedStaleness"
  }

  geo_location {
    location          = "${var.resource_group_location}"
    failover_priority = 0
  }
}

output "key" {
  value = "${azurerm_cosmosdb_account.db.primary_master_key}"
}

output "name" {
  value = "${azurerm_cosmosdb_account.db.name}"
}
