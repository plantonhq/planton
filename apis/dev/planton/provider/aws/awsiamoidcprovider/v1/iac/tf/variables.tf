variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name = string
    id = optional(string, "")
    org = optional(string, "")
    env = optional(string, "")
    labels = optional(map(string), {})
    annotations = optional(map(string), {})
    tags = optional(list(string), [])
  })
}

variable "spec" {
  description = "AwsIamOidcProvider specification"
  type = object({
    region = string
    # url arrives pre-resolved: the orchestrator replaces a valueFrom reference
    # (e.g. an AwsEksCluster's oidc_issuer_url) with the literal issuer URL
    # before the module runs.
    url = string
    client_id_list = list(string)
    thumbprint_list = optional(list(string), [])
  })
}
