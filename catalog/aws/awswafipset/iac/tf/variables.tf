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
  description = "AwsWafIpSet specification"
  type = object({
    # The AWS region where the IP set is created. Must match the scope of the
    # web ACLs that will reference it: a REGIONAL set lives in the same region
    # as the resources it protects (ALBs, API Gateway stages, Cognito user
    # pools), while a CLOUDFRONT set must be created in us-east-1 (the WAF
    # global region).
    # Example: "us-west-2", "us-east-1"
    region = string

    # Whether the set is usable by REGIONAL web ACLs (protecting ALBs, API
    # Gateway, AppSync, Cognito, App Runner, Verified Access) or CLOUDFRONT
    # web ACLs (protecting CloudFront distributions). Create-time immutable
    # (ForceNew). A web ACL can only reference sets of its own scope.
    scope = string

    # The IP address family this set holds: "IPV4" or "IPV6". Create-time
    # immutable (ForceNew). A set holds exactly one family — deployments that
    # match both families use two sets (one per family) referenced by two
    # rules or one rule with an OR statement.
    ip_address_version = string

    # The addresses in the set, in CIDR notation — AWS accepts ONLY CIDR
    # ranges, never bare addresses: a single IPv4 host is "192.0.2.44/32"
    # (not "192.0.2.44"), a single IPv6 host is "2001:db8::1/128". IPv4
    # supports /1 through /32; IPv6 supports /1 through /128. Up to 10,000
    # entries. An empty list is valid (a placeholder set that rules can
    # reference before the ranges are known) — an empty set matches nothing.
    addresses = optional(list(string), [])

    # Description of what the set represents and who maintains it. AWS
    # restricts the character set: letters, digits, whitespace, and
    # _ + = : # @ / - , . only (notably NO parentheses), 3-256 characters —
    # WAF rejects anything else at create time, so the constraint is enforced
    # here where the failure is immediate and readable.
    description = optional(string, "")
  })
}
