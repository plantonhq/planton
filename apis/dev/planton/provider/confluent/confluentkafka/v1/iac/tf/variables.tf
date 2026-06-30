variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Confluent Kafka cluster specification"
  type = object({
    # Cloud provider (AWS, AZURE, GCP)
    # https://www.pulumi.com/registry/packages/confluentcloud/api-docs/kafkacluster/#cloud_yaml
    cloud = string

    # Cloud-specific region (e.g., us-east-2, us-central1, eastus)
    # https://www.pulumi.com/registry/packages/confluentcloud/api-docs/kafkacluster/#region_yaml
    region = string

    # Availability configuration (SINGLE_ZONE, MULTI_ZONE, LOW, HIGH)
    # https://www.pulumi.com/registry/packages/confluentcloud/api-docs/kafkacluster/#availability_yaml
    availability = string

    # Confluent Cloud environment ID (parent container for clusters)
    # https://www.pulumi.com/registry/packages/confluentcloud/api-docs/kafkacluster/#environment_yaml
    environment_id = string

    # Cluster type (BASIC, STANDARD, ENTERPRISE, DEDICATED)
    # Default: STANDARD if not specified
    cluster_type = optional(string, "STANDARD")

    # Dedicated cluster configuration (required when cluster_type is DEDICATED)
    dedicated_config = optional(object({
      # Confluent Kafka Units (CKU) for provisioned capacity
      # Minimum: 1 CKU
      # https://www.pulumi.com/registry/packages/confluentcloud/api-docs/kafkacluster/#cku_yaml
      cku = number
    }))

    # Network configuration for private networking (PrivateLink, VNet Peering, Private Service Connect)
    # Only available for ENTERPRISE and DEDICATED cluster types
    network_config = optional(object({
      # ID of the Confluent Cloud network resource
      # Must be pre-created in the same environment
      network_id = string
    }))

    # Display name shown in Confluent Cloud UI
    # Optional: defaults to metadata.name if not specified
    display_name = optional(string)
  })
}

variable "confluent_api_key" {
  description = "Confluent Cloud API Key for authentication"
  type        = string
  sensitive   = true
}

variable "confluent_api_secret" {
  description = "Confluent Cloud API Secret for authentication"
  type        = string
  sensitive   = true
}