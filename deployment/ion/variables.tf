variable "certificate_mount_path" {
  type        = "string"
  description = "Path to mount certificates"
  default     = "/etc/config/"
}

variable "client_id" {
  type        = "string"
  description = "Client ID"
}

variable "client_secret" {
  type        = "string"
  description = "Client secret."
}

variable "subscription_id" {
  type        = "string"
  description = "Azure subscription ID"
}

variable "tenant_id" {
  type        = "string"
  description = "Azure tenant ID"
}

variable "cluster_client_certificate" {
  type        = "string"
  description = "Cluster client Certificate"
  default     = "eastus"
}

variable "cluster_client_key" {
  type        = "string"
  description = "Cluster client Certificate Key"
}

variable "cluster_ca" {
  type        = "string"
  description = "Cluster Certificate Authority"
}

variable "cluster_host" {
  type        = "string"
  description = "Cluster Admin API host"
}

variable "batch_account_name" {
  type        = "string"
  description = "The name of the Azure Batch account to use"
}

variable "azure_batch_pool_id" {
  type        = "string"
  description = "The PoolID to use in Azure batch"
  default     = "pool1"
}

variable "resource_group_location" {
  description = "Resource group location"
  type        = "string"
}

variable "resource_group_name" {
  description = "Resource group location"
  type        = "string"
}

variable "servicebus_key" {
  type = "string"
}

variable "servicebus_name" {
  type = "string"
}

variable "cosmos_name" {
  type = "string"
}

variable "cosmos_db_name" {
  type = "string"
}

variable "cosmos_key" {
  type = "string"
}

variable "storage_name" {
  type = "string"
}

variable "storage_key" {
  type = "string"
}

variable "acr_url" {}

variable "acr_username" {}

variable "acr_password" {}

variable "managementapi_docker_image" {
  description = "The docker image for the ion management api"
}

variable "frontapi_docker_image" {
  description = "The docker image for the ion front api"
}

variable "dispatcher_docker_image" {
  description = "The docker image for the ion dispatcher service"
}

variable "app_insights_key" {}
