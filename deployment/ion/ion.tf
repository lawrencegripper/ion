provider "kubernetes" {
  host = "${var.cluster_host}"

  client_certificate     = "${var.cluster_client_certificate}"
  client_key             = "${var.cluster_client_key}"
  cluster_ca_certificate = "${var.cluster_ca}"
}

resource "kubernetes_deployment" "ion-management-api" {
  metadata {
    name = "ion-management-api"
  }

  spec {
    selector {
      app = "ion-management-api"
    }

    template {
      metadata {
        labels {
          app = "ion-management-api"
        }
      }

      spec {
        container {
          name  = "managementapi"
          image = "${var.managementapi_docker_image}"

          args = [
            "tbd",
          ]

          port {
            container_port = 10250
            protocol       = "TCP"
            name           = "kubeletport"
          }

          volume_mount {
            name       = "azure-credentials"
            mount_path = "/etc/aks/azure.json"
          }

          env = [
            {
              name  = "AZURE_BATCH_ACCOUNT_LOCATION"
              value = "${var.resource_group_location}"
            },
            {
              name  = "AZURE_BATCH_ACCOUNT_NAME"
              value = "${var.batch_account_name}"
            },
          ]
        }

        volume {
          name = "azure-credentials"

          host_path {
            path = "/etc/kubernetes/azure.json"
          }
        }
      }
    }
  }
}
