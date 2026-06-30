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
  description = "spec holds the core configuration: the issuer URL, allowed client IDs, and optional thumbprints."
  type = object({

    # The AWS region used to configure the provider.
    region = string

    # url is the URL of the OIDC identity provider (the issuer / `iss` claim).
    # Modeled as a StringValueOrRef: either an inline value or a reference to
    # another resource's field (e.g. an AwsEksCluster's oidc_issuer_url).
    url = object({

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

    # client_id_list is the set of client IDs (audiences) allowed to authenticate.
    client_id_list = list(string)

    # thumbprint_list is the optional set of SHA-1 thumbprints of the issuer's root CA.
    thumbprint_list = list(string)
  })
}
