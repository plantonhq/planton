variable "metadata" {
  description = "Catalog object metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "AwsRestApiDomain specification"
  type = object({
    region          = string
    domain_name     = string
    certificate_arn = optional(string, "")
    uploaded_certificate = optional(object({
      name        = string
      body        = string
      chain       = optional(string, "")
      private_key = string
    }))
    endpoint_configuration = optional(object({
      type            = optional(string, "")
      ip_address_type = optional(string, "")
    }))
    endpoint_access_mode = optional(string, "")
    security_policy      = optional(string, "")
    mutual_tls = optional(object({
      truststore_uri     = string
      truststore_version = optional(string, "")
    }))
    ownership_verification_certificate_arn = optional(string, "")
    policy                                 = optional(any)
    routing_mode                           = optional(string, "")
    base_path_mappings = optional(list(object({
      base_path   = optional(string, "")
      rest_api_id = string
      stage_name  = optional(string, "")
    })), [])
    access_associations = optional(list(object({
      vpc_endpoint_id = string
    })), [])
  })
}