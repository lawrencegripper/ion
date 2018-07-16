resource "random_string" "batchname" {
  keepers = {
    # Generate a new id each time we switch to a new resource group
    group_name = "${var.resource_group_name}"
  }

  length  = 8
  upper   = false
  special = false
  number  = false
}

resource "azurerm_kubernetes_cluster" "aks" {
  name       = "ionaks-${random_string.batchname.result}"
  location   = "${var.resource_group_location}"
  dns_prefix = "ionaks-${random_string.batchname.result}"

  resource_group_name = "${var.resource_group_name}"
  kubernetes_version  = "1.9.2"

  linux_profile {
    admin_username = "${var.linux_admin_username}"

    ssh_key {
      key_data = "${var.linux_admin_ssh_publickey}"
    }
  }

  agent_pool_profile {
    name    = "agentpool"
    count   = "${var.node_count}"
    vm_size = "${var.node_sku}"
    os_type = "Linux"
  }

  service_principal {
    client_id     = "${var.client_id}"
    client_secret = "${var.client_secret}"
  }

  tags {
    source = "terraform"
  }
}

output "cluster_client_certificate" {
  value = "${base64decode(azurerm_kubernetes_cluster.aks.kube_config.0.client_certificate)}"
}

output "cluster_client_key" {
  value = "${base64decode(azurerm_kubernetes_cluster.aks.kube_config.0.client_key)}"
}

output "cluster_ca" {
  value = "${base64decode(azurerm_kubernetes_cluster.aks.kube_config.0.cluster_ca_certificate)}"
}

output "host" {
  value = "${azurerm_kubernetes_cluster.aks.kube_config.0.host}"
}

output "kubeconfig" {
  value = "${azurerm_kubernetes_cluster.aks.kube_config_raw}"
}

output "cluster_name" {
  value = "ionaks-${random_string.batchname.result}"
}
