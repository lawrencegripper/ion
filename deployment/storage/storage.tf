variable "resource_group_name" {
  description = "Resource group name"
  type        = "string"
}

variable "resource_group_location" {
  description = "Resource group location"
  type        = "string"
}

variable "pool_bootstrap_script_path" {
  description = "The filepath of the pool boostrapping script"
  type        = "string"
}

resource "random_string" "storage" {
  keepers = {
    # Generate a new id each time we switch to a new resource group
    group_name = "${var.resource_group_name}"
  }

  length  = 5
  upper   = false
  special = false
  number  = false
}

resource "azurerm_storage_account" "batchstorage" {
  name                     = "ionstorage${lower(random_string.storage.result)}"
  resource_group_name      = "${var.resource_group_name}"
  location                 = "${var.resource_group_location}"
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags {
    source = "terraform"
  }
}

resource "azurerm_storage_container" "boostrapscript" {
  name                  = "scripts"
  resource_group_name   = "${var.resource_group_name}"
  storage_account_name  = "${azurerm_storage_account.batchstorage.name}"
  container_access_type = "private"
}

resource "azurerm_storage_blob" "initscript" {
  name = "init.sh"

  resource_group_name    = "${var.resource_group_name}"
  storage_account_name   = "${azurerm_storage_account.batchstorage.name}"
  storage_container_name = "${azurerm_storage_container.boostrapscript.name}"

  type   = "block"
  source = "${var.pool_bootstrap_script_path}"
}

data "azurerm_storage_account_sas" "scriptaccess" {
  connection_string = "${azurerm_storage_account.batchstorage.primary_connection_string}"
  https_only        = true

  resource_types {
    service   = false
    container = false
    object    = true
  }

  services {
    blob  = true
    queue = false
    table = false
    file  = false
  }

  start  = "${timestamp()}"
  expiry = "${timeadd(timestamp(), "8776h")}"

  permissions {
    read    = true
    write   = false
    delete  = false
    list    = false
    add     = false
    create  = false
    update  = false
    process = false
  }
}

output "pool_boostrap_script_url" {
  value = "${azurerm_storage_blob.initscript.url}${data.azurerm_storage_account_sas.scriptaccess.sas}"
}

output "name" {
  value = "${azurerm_storage_account.batchstorage.name}"
}

output "key" {
  value = "${azurerm_storage_account.batchstorage.primary_access_key}"
}

output "id" {
  value = "${azurerm_storage_account.batchstorage.id}"
}
