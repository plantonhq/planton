variable "metadata" {
  description = "Cloud resource metadata"
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
  description = "AwsInternetGateway specification"
  type = object({
    # AWS region the internet gateway is created in. Must match the region of the
    # VPC it attaches to. Example: "us-west-2", "eu-west-1". This drives provider
    # construction, so it is required even though the gateway logically inherits
    # the VPC's region.
    region = string

    # The VPC this internet gateway attaches to. Supply a literal vpc-id or
    # reference an AwsVpc and the platform resolves its vpc_id output. Unlike a
    # subnet's vpc_id, this is updatable: changing it detaches the gateway from
    # the old VPC and attaches it to the new one (the gateway itself is not
    # replaced). A VPC may have only one internet gateway attached at a time.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    vpc_id = string
  })
}
