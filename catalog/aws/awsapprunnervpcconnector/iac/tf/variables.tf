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
  description = "AwsAppRunnerVpcConnector specification"
  type = object({
    # The AWS region where the VPC connector will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Subnets in which AWS provisions the connector's network interfaces.
    # Immutable after creation (changing the set replaces the connector).
    # Provide subnets in at least two availability zones for high
    # availability -- App Runner routes egress only through AZs the connector
    # has an ENI in. All subnets must belong to the same VPC. Accepts direct
    # subnet IDs or references to AwsSubnet resources.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnet_ids = list(string)

    # Security groups applied to the connector's network interfaces.
    # Immutable after creation (changing the set replaces the connector).
    # These groups control what VPC resources the connected services can
    # reach: allow egress to your databases' and caches' ports, and make the
    # targets' security groups admit ingress from these groups. AWS requires
    # at least one group on a connector. Accepts direct security group IDs or
    # references to AwsSecurityGroup resources.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    security_group_ids = list(string)
  })
}
