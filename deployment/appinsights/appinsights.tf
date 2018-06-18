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

resource "azurerm_application_insights" "ai" {
  name = "ionappinsight${random_string.name.result}"

  location            = "${var.resource_group_location}"
  resource_group_name = "${var.resource_group_name}"

  application_type = "Other"
}

output "instrumentation_key" {
  value = "${azurerm_application_insights.ai.instrumentation_key}"
}

output "app_id" {
  value = "${azurerm_application_insights.ai.app_id}"
}
