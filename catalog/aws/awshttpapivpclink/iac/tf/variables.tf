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
  description = "AwsHttpApiVpcLink specification"
  type = object({
    # The AWS region where the VPC link will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Subnets in which AWS provisions the VPC link's network interfaces.
    # Immutable after creation (changing the set replaces the link). Spread
    # the link across at least two availability zones for high availability --
    # the link can only reach targets in AZs it has an ENI in. Accepts direct
    # subnet IDs or references to AwsSubnet resources.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnet_ids = list(string)

    # Security groups applied to the VPC link's network interfaces. Immutable
    # after creation (changing the set replaces the link). These groups control
    # what the link can reach inside the VPC: allow egress to the private
    # ALB/NLB listener ports (and make the target's security group admit
    # ingress from these groups). When omitted, the link is created without
    # security groups and AWS applies no filtering on the link side.
    # Accepts direct security group IDs or references to AwsSecurityGroup
    # resources.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    security_group_ids = optional(list(string), [])
  })
}
