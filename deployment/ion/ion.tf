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
            "start",
          ]

          port {
            container_port = 9000
            protocol       = "TCP"
            name           = "apiport"
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
            {
              name  = "AZURE_BATCH_POOL_ID"
              value = "${var.azure_batch_pool_id}"
            },
            {
              name  = "AZURE_CLIENT_ID"
              value = "${var.client_id}"
            },
            {
              name  = "AZURE_CLIENT_SECRET"
              value = "${var.client_secret}"
            },
            {
              name  = "AZURE_SUBSCRIPTION_ID"
              value = "${var.subscription_id}"
            },
            {
              name  = "AZURE_TENANT_ID"
              value = "${var.tenant_id}"
            },
            {
              name  = "AZURE_RESOURCE_GROUP"
              value = "${var.resource_group_name}"
            },
            {
              name  = "AZURE_STORAGE_ACCOUNT_KEY"
              value = "${var.storage_key}"
            },
            {
              name  = "AZURE_STORAGE_ACCOUNT_NAME"
              value = "${var.storage_name}"
            },
            {
              name  = "DISPATCHER_IMAGE_NAME"
              value = "${var.dispatcher_docker_image}"
            },
            {
              name  = "DISPATCHER_IMAGE_TAG"
              value = "latest"
            },
            {
              name  = "MANAGEMENT_PORT"
              value = "9000"
            },
            {
              name  = "MONGODB_COLLECTION"
              value = "${var.cosmos_db_name}"
            },
            {
              name  = "MONGODB_NAME"
              value = "${var.cosmos_name}"
            },
            {
              name  = "MONGODB_PASSWORD"
              value = "${var.cosmos_key}"
            },
            {
              name  = "MONGODB_PORT"
              value = "10255"
            },
            {
              name  = "MONGODB_USERNAME"
              value = "${var.cosmos_name}"
            },
            {
              name  = "AZURE_SERVICEBUS_NAMESPACE"
              value = "${var.servicebus_name}"
            },
            {
              name  = "NAMESPACE"
              value = "default"
            },
            {
              name  = "LOGLEVEL"
              value = "DEBUG"
            },
            {
              name  = "PROVIDER"
              value = "kubernetes"
            },
            {
              name  = "CONTAINER_IMAGE_REGISTRY_URL"
              value = "${var.acr_url}"
            },
            {
              name  = "CONTAINER_IMAGE_REGISTRY_USERNAME"
              value = "${var.acr_username}"
            },
            {
              name  = "CONTAINER_IMAGE_REGISTRY_PASSWORD"
              value = "${var.acr_password}"
            },
          ]
        }
      }
    }
  }
}
