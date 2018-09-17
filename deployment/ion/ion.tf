provider "kubernetes" {
  host             = "${var.cluster_host}"
  load_config_file = false

  client_certificate     = "${var.cluster_client_certificate}"
  client_key             = "${var.cluster_client_key}"
  cluster_ca_certificate = "${var.cluster_ca}"
}

resource "kubernetes_deployment" "ion-front-api" {
  metadata {
    name = "ion-front-api"
  }

  spec {
    selector {
      app = "ion-front-api"
    }

    template {
      metadata {
        labels {
          app = "ion-front-api"
        }
      }

      spec {
        container {
          name  = "frontapi"
          image = "${var.frontapi_docker_image}"

          args = [
            "serve",
            "--printconfig",
            "--loglevel=debug",
          ]

          port {
            container_port = 9001
            protocol       = "TCP"
            name           = "apiport"
          }

          env = [
            {
              name  = "PRINTCONFIG"
              value = "true"
            },
            {
              name  = "PORT"
              value = "9001"
            },
            {
              name  = "MODULENAME"
              value = "frontapi"
            },
            {
              name  = "SUBSCRIBESTOEVENT"
              value = "none"
            },
            {
              name  = "EVENTSPUBLISHED"
              value = "frontapi.new_link"
            },
            {
              name  = "CLIENTID"
              value = "${var.client_id}"
            },
            {
              name  = "CLIENTSECRET"
              value = "${var.client_secret}"
            },
            {
              name  = "SUBSCRIPTIONID"
              value = "${var.subscription_id}"
            },
            {
              name  = "TENANTID"
              value = "${var.tenant_id}"
            },
            {
              name  = "RESOURCEGROUP"
              value = "${var.resource_group_name}"
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
              name  = "SERVICEBUSNAMESPACE"
              value = "${var.servicebus_name}"
            },
            {
              name  = "IMAGE_REGISTRY_URL"
              value = "${var.acr_url}"
            },
            {
              name  = "IMAGE_REGISTRY_USERNAME"
              value = "${var.acr_username}"
            },
            {
              name  = "IMAGE_REGISTRY_PASSWORD"
              value = "${var.acr_password}"
            },
            {
              name  = "LOGLEVEL"
              value = "DEBUG"
            },
          ]
        }
      }
    }
  }
}

resource "kubernetes_secret" "ion-management-api" {
  metadata {
    name = "generic"
  }

  data {
    certificate     = "${var.server_cert}"
    certificate_key = "${var.server_key}"
    certificate_ca  = "${var.root_ca}"
  }
}

resource "kubernetes_service" "ion-management-api" {
  metadata {
    name = "ion-management-api"

    annotations {
      "service.beta.kubernetes.io/azure-dns-label-name" = "${var.prefix}"
    }
  }

  spec {
    selector {
      app = "${kubernetes_deployment.ion-management-api.metadata.0.labels.app}"
    }

    port {
      port        = 9000
      target_port = 9000
    }

    type = "LoadBalancer"
  }
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
              name  = "CERTFILE"
              value = "${var.certificate_mount_path}/certificate"
            },
            {
              name  = "KEYFILE"
              value = "${var.certificate_mount_path}/certificate_key"
            },
            {
              name  = "CACERTFILE"
              value = "${var.certificate_mount_path}/certificate_ca"
            },
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
              name  = "AZURE_BATCH_REQUIRES_GPU"
              value = "true"
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
              name  = "IMAGE_REGISTRY_URL"
              value = "${var.acr_url}"
            },
            {
              name  = "IMAGE_REGISTRY_USERNAME"
              value = "${var.acr_username}"
            },
            {
              name  = "IMAGE_REGISTRY_PASSWORD"
              value = "${var.acr_password}"
            },
            {
              name  = "LOGGING_APPINSIGHTS"
              value = "${var.app_insights_key}"
            },
          ]

          volume_mount {
            mount_path = "${var.certificate_mount_path}"
            name       = "ion-management-api"
            read_only  = true
          }
        }

        volume {
          name = "ion-management-api"

          secret {
            secret_name = "${kubernetes_secret.ion-management-api.metadata.0.name}"
          }
        }
      }
    }
  }
}

output "service_ip" {
  value = "${kubernetes_service.ion-management-api.load_balancer_ingress.0.ip}"
}
