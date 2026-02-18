variable "metadata" {
  description = "metadata for the resource\nid format \"<id-prefix>-<env-id>-<normalized-resource-name>\""
  type = object({

    # name of the resource
    name = string

    # id of the resource
    id = string

    # id of the organization to which the api-resource belongs to
    org = string

    # environment to which the resource belongs to
    env = string

    # labels for the resource
    labels = object({

      # Description for key
      key = string

      # Description for value
      value = string
    })

    # annotations for the resource
    annotations = object({

      # Description for key
      key = string

      # Description for value
      value = string
    })

    # tags for the resource
    tags = list(string)
  })
}

variable "spec" {
  description = "aws-route53-zone spec"
  type = object({

    # The AWS region where the resource will be created.
    region = string

    # The DNS records that are added to the zone.
    # Each record represents a DNS resource record, such as A, AAAA, CNAME, MX, TXT, etc.
    # These records define how your domain or subdomains are routed to your resources.
    records = list(object({

      # The DNS record type.
      # This specifies the type of DNS record, such as A, AAAA, CNAME, MX, TXT, etc.
      # The record type determines how the DNS query is processed and what kind of data is returned.
      record_type = string

      # The name of the DNS record, e.g., "example.com." or "dev.example.com.".
      # This is the domain name or subdomain for which the DNS record applies.
      # The value should always end with a dot, following DNS standards to denote a fully qualified domain name.
      name = string

      # The values for the DNS record.
      # This field contains the data associated with the DNS record type.
      # For example, for an A record, it would be the IP address(es) the domain resolves to.
      # If the record type is CNAME, each value in the list should end with a dot to denote a fully qualified domain name.
      values = list(string)

      # The Time To Live (TTL) for the DNS record, in seconds.
      # TTL specifies how long DNS resolvers should cache the DNS record before querying again.
      ttl_seconds = number
    }))
  })
}