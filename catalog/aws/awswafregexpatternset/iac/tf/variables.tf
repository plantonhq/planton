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
  description = "AwsWafRegexPatternSet specification"
  type = object({
    # The AWS region where the pattern set is created. Must match the scope of
    # the web ACLs that will reference it: a REGIONAL set lives in the same
    # region as the resources it protects, while a CLOUDFRONT set must be
    # created in us-east-1 (the WAF global region).
    # Example: "us-west-2", "us-east-1"
    region = string

    # Whether the set is usable by REGIONAL web ACLs (protecting ALBs, API
    # Gateway, AppSync, Cognito, App Runner, Verified Access) or CLOUDFRONT
    # web ACLs (protecting CloudFront distributions). Create-time immutable
    # (ForceNew). A web ACL can only reference sets of its own scope.
    scope = string

    # The regular expressions in the set, each up to 200 characters. AWS's
    # default quota is 10 expressions per set (adjustable via Service Quotas).
    # AWS supports a subset of PCRE — no backreferences (\1), no lookaround
    # ((?=...)/(?<=...)); such patterns fail at AWS create time. At least one
    # expression is required — unlike IP sets, AWS rejects an empty pattern
    # set.
    regular_expressions = list(string)

    # Description of what the patterns match and who maintains them. AWS
    # restricts the character set: letters, digits, whitespace, and
    # _ + = : # @ / - , . only (notably NO parentheses), 3-256 characters —
    # WAF rejects anything else at create time, so the constraint is enforced
    # here where the failure is immediate and readable.
    description = optional(string, "")
  })
}
