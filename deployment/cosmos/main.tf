resource "random_integer" "ri" {
  min = 10000
  max = 99999
}

resource "azurerm_cosmosdb_account" "db" {
  name                = "iondb-${random_integer.ri.result}"
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
