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
  description = "AwsIamRole specification"
  type = object({
    # The AWS region used by the provider while managing this role.
    # IAM is a global service -- the role is assumable in every region -- but
    # every AWS API call is still made against a regional endpoint, so a region
    # is required.
    # Example: "us-west-2", "eu-west-1".
    region = string

    # An optional human-readable description of the role's purpose, shown in
    # the IAM console. Updatable in place. Maximum 1000 characters; AWS rejects
    # typographic ("curly") quotes here, so stick to plain ASCII quoting.
    description = optional(string, "")

    # The IAM path for the role, used to organize and match roles in IAM
    # policies (e.g. grant iam:PassRole only for
    # "arn:aws:iam::<acct>:role/service-roles/*"). Must begin and end with "/"
    # (e.g. "/service-roles/"). Defaults to "/" when omitted. Immutable:
    # changing the path replaces the role.
    path = optional(string, "")

    # The trust policy as free-form JSON: the statement of WHO may assume this
    # role (service principals, AWS accounts, federated identities) and under
    # what conditions. This is the security-critical half of the role -- prefer
    # exact principals and add conditions (aws:SourceAccount, aws:SourceArn,
    # sts:ExternalId) to prevent confused-deputy access. Updatable in place.
    # Example:
    #   Version: "2012-10-17"
    #   Statement:
    #     - Effect: Allow
    #       Principal: { Service: lambda.amazonaws.com }
    #       Action: sts:AssumeRole
    trust_policy = optional(any)

    # Typed federated trust against an IAM OIDC provider: the IaC modules
    # compose the sts:AssumeRoleWithWebIdentity trust document from the
    # provider's outputs and the subject/audience conditions declared here.
    # The form that makes keyless workload identity composable -- provider,
    # role, and consumer can deploy in one run with the trust wired by
    # reference.
    oidc_trust = optional(object({
      # The IAM OIDC provider this role trusts -- the `Federated` principal of
      # the composed trust policy. Reference an AwsIamOidcProvider's
      # provider_arn output or pass a literal provider ARN
      # (arn:aws:iam::<account>:oidc-provider/<issuer-host-and-path>).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      provider_arn = string

      # The provider's issuer URL with the scheme stripped (e.g.
      # "oidc.eks.us-west-2.amazonaws.com/id/EXAMPLED") -- the prefix of the
      # `sub`/`aud` condition keys in the composed document. Reference the SAME
      # AwsIamOidcProvider's provider_url output as provider_arn so the pair can
      # never drift apart.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      provider_url = string

      # Exact-match `sub` claim values this role accepts (StringEquals). For EKS
      # IRSA the subject is "system:serviceaccount:<namespace>:<serviceaccount>".
      # At least one subject -- exact or wildcard -- is required.
      subjects = optional(list(string), [])

      # Wildcard `sub` claim patterns this role accepts (StringLike; `*` and `?`
      # wildcards) -- the CI-federation shape, e.g. "repo:my-org/my-repo:*" for
      # GitHub Actions. Rendered as its own statement so exact and wildcard
      # subjects are ORed, never ANDed (see the message comment).
      wildcard_subjects = optional(list(string), [])

      # `aud` claim values this role accepts (StringEquals on the audience
      # condition key). Empty defaults to ["sts.amazonaws.com"] -- the audience
      # EKS IRSA and GitHub Actions both present. Set explicitly only for
      # providers registered with a different client id.
      audiences = optional(list(string), [])
    }))

    # Managed policies to attach, each a reference to an AwsIamPolicy's
    # policy_arn output or a literal ARN (literals are how AWS-managed policies
    # like arn:aws:iam::aws:policy/ReadOnlyAccess attach). Attachments are
    # reconciled in place: adding or removing an entry attaches or detaches
    # without touching the role. Permissions unique to this role belong in
    # inline_policies instead.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    managed_policy_arns = optional(list(string), [])

    # Inline policies embedded in this role: a map of policy name to a
    # free-form JSON permission document. An inline policy lives and dies with
    # the role, so use it for permissions that make no sense anywhere else
    # (e.g. access to this service's own queue); anything reused across
    # principals belongs in a first-class AwsIamPolicy attached via
    # managed_policy_arns.
    inline_policies = optional(any, {})

    # The maximum duration, in seconds, of sessions assumed on this role
    # (the ceiling for the AssumeRole DurationSeconds parameter). Between 3600
    # (1 hour, the AWS default when unset) and 43200 (12 hours). Raise it for
    # long-running human or CI sessions; keep the default for service roles.
    # Updatable in place.
    max_session_duration = optional(number, 0)

    # An optional permissions boundary: a managed policy whose grants cap the
    # maximum permissions this role can ever have -- effective permissions are
    # the INTERSECTION of the boundary and the role's permission policies.
    # Reference an AwsIamPolicy's policy_arn output or pass a literal policy
    # ARN. Setting or changing the boundary is in-place; clearing it removes
    # the ceiling.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    permissions_boundary = optional(string, "")

    # Whether deleting the role force-detaches any policies still attached to
    # it (including attachments made outside this resource). Off by default:
    # deletion fails if out-of-band attachments exist, surfacing them instead
    # of silently severing another owner's wiring. Turn on for ephemeral or
    # CI-owned roles where teardown must always succeed.
    force_detach_policies = optional(bool, false)
  })
}
