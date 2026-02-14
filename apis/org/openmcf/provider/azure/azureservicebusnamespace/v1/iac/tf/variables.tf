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
  description = "Azure Service Bus Namespace specification"
  type = object({
    # The Azure region
    region = string

    # The Azure Resource Group name
    resource_group = string

    # The namespace name (globally unique)
    name = string

    # SKU tier: "Basic", "Standard", "Premium"
    sku = optional(string, "Standard")

    # Messaging units for Premium SKU (1, 2, 4, 8, 16)
    capacity = optional(number)

    # Premium messaging partitions (1, 2, 4)
    premium_messaging_partitions = optional(number)

    # Enable zone redundancy (Premium only)
    zone_redundant = optional(bool, false)

    # Minimum TLS version
    minimum_tls_version = optional(string, "1.2")

    # Whether the namespace is publicly accessible
    public_network_access_enabled = optional(bool, true)

    # Queues within the namespace
    queues = optional(list(object({
      name                                 = string
      max_size_in_megabytes                = optional(number)
      partitioning_enabled                 = optional(bool, false)
      default_message_ttl                  = optional(string)
      lock_duration                        = optional(string, "PT1M")
      max_delivery_count                   = optional(number, 10)
      requires_duplicate_detection         = optional(bool, false)
      requires_session                     = optional(bool, false)
      dead_lettering_on_message_expiration = optional(bool, false)
      forward_to                           = optional(string)
      forward_dead_lettered_messages_to    = optional(string)
    })), [])

    # Topics within the namespace
    topics = optional(list(object({
      name                         = string
      max_size_in_megabytes        = optional(number)
      partitioning_enabled         = optional(bool, false)
      default_message_ttl          = optional(string)
      requires_duplicate_detection = optional(bool, false)
      support_ordering             = optional(bool, false)
    })), [])
  })
}
