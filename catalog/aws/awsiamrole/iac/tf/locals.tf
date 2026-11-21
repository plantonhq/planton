locals {
  # Resource-identity tags, matching the Pulumi module key-for-key.
  aws_tags = {
    "Name"                     = var.metadata.name
    "planton.ai/resource"      = "true"
    "planton.ai/organization"  = var.metadata.org
    "planton.ai/environment"   = var.metadata.env
    "planton.ai/resource-kind" = "AwsIamRole"
    "planton.ai/resource-id"   = var.metadata.id
  }

  # The trust oneof: exactly one of trust_policy (free-form JSON) or
  # oidc_trust (typed federated trust) is set — spec-enforced.
  oidc_trust = try(var.spec.oidc_trust, null)

  # The audience condition defaults to sts.amazonaws.com — the audience EKS
  # IRSA and GitHub Actions both present — so the common case needs no
  # explicit audiences.
  oidc_audiences = local.oidc_trust != null ? (
    length(local.oidc_trust.audiences) > 0 ? local.oidc_trust.audiences : ["sts.amazonaws.com"]
  ) : []

  # IAM's condition keys are the bare issuer (<host>/<path>:sub). A provider
  # URL handed over with its scheme (typed by hand, or read from a provider
  # built before its output dropped the scheme) is stripped, or the condition
  # would never match and every sign-in would be refused.
  oidc_issuer = local.oidc_trust == null ? "" : replace(local.oidc_trust.provider_url, "/^https?:\\/\\//", "")

  # One statement PER CONDITION OPERATOR: IAM ANDs condition operators inside
  # a statement and ORs across statements, so exact subjects (StringEquals)
  # and wildcard subjects (StringLike) must ride separate statements — mixing
  # both on the `sub` key in one statement would require a token to satisfy
  # both at once. The jsondecode(jsonencode(...)) seam keeps the two
  # differently-shaped statement objects out of plan-time type unification
  # (a bare concat of differently-shaped objects fails the plan).
  oidc_statements = local.oidc_trust == null ? [] : concat(
    length(local.oidc_trust.subjects) > 0 ? [jsondecode(jsonencode({
      Effect = "Allow"
      Principal = {
        Federated = local.oidc_trust.provider_arn
      }
      Action = "sts:AssumeRoleWithWebIdentity"
      Condition = {
        StringEquals = {
          "${local.oidc_issuer}:sub" = local.oidc_trust.subjects
          "${local.oidc_issuer}:aud" = local.oidc_audiences
        }
      }
    }))] : [],
    length(local.oidc_trust.wildcard_subjects) > 0 ? [jsondecode(jsonencode({
      Effect = "Allow"
      Principal = {
        Federated = local.oidc_trust.provider_arn
      }
      Action = "sts:AssumeRoleWithWebIdentity"
      Condition = {
        # The audience stays an exact match even on the wildcard-subject
        # statement — only the subject pattern is fuzzy.
        StringEquals = {
          "${local.oidc_issuer}:aud" = local.oidc_audiences
        }
        StringLike = {
          "${local.oidc_issuer}:sub" = local.oidc_trust.wildcard_subjects
        }
      }
    }))] : []
  )

  # aws_iam_role wants assume_role_policy as a JSON string. The oidc_trust
  # arm composes the web-identity document from the provider's outputs
  # (resolved references by the time the module runs); the trust_policy arm
  # encodes the user's free-form JSON object (google.protobuf.Struct) as-is.
  trust_policy_json = local.oidc_trust != null ? jsonencode({
    Version   = "2012-10-17"
    Statement = local.oidc_statements
  }) : jsonencode(var.spec.trust_policy)

  # inline_policies is free-form JSON (map<string, google.protobuf.Struct>), typed `any` in
  # variables.tf because its entries have heterogeneous shapes. Encode each policy document to a
  # JSON string here so the result is a homogeneous map(string): aws_iam_role_policy.for_each
  # accepts a map/set, and converting a heterogeneous object to a map would otherwise fail with
  # "all map elements must have the same type".
  inline_policies_json = {
    for k, v in var.spec.inline_policies : k => jsonencode(v)
  }
}
