variable "metadata" {
  description = "metadata captures identifying information (name, org, version, etc.)\nand must pass standard validations for resource naming."
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
  description = "spec holds the core configuration data defining how the ECS service is deployed."
  type = object({

    # The AWS region where the resource will be created.
    region = string

    # primaryDomainName is a required field representing the main (apex or wildcard) domain name.
    # Examples include "example.com" or "*.example.com" (wildcard).
    # This domain will be set as the 'DomainName' in the AWS ACM certificate.
    # 
    # The pattern enforces a domain-like structure, allowing an optional wildcard prefix.
    # The string is mandatory, so users must always supply a primary domain.
    primary_domain_name = string

    # alternateDomainNames is an optional list of Subject Alternative Names (SANs) for the certificate.
    # Each entry must follow the same pattern as primary_domain_name and cannot contain duplicates.
    # Primary domain should not be added to this list.
    alternate_domain_names = list(string)

    # route53_hosted_zone_id is the identifier of the Route53 hosted zone
    # where DNS validation records will be created automatically.
    # Must be a public hosted zone matching the domain names.
    # Example: "Z123456ABCXYZ".
    route53_hosted_zone_id = object({

      # Description for value
      value = string

      # Description for value_from
      value_from = object({

        # Description for kind
        kind = string

        # Description for env
        env = string

        # Description for name
        name = string

        # Description for field_path
        field_path = string
      })
    })

    # validation_method indicates how ACM verifies domain ownership.
    # By default, DNS is recommended.
    validation_method = string
  })
}