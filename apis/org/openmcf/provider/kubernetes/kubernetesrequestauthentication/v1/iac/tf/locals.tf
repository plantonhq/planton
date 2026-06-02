locals {
  request_authentication_name = var.metadata.name
  namespace                   = var.spec.namespace

  labels = {
    "app.kubernetes.io/name"       = "request-authentication"
    "app.kubernetes.io/instance"   = var.metadata.name
    "app.kubernetes.io/managed-by" = "openmcf"
    "app.kubernetes.io/component"  = "request-authentication"
  }

  # Workload selector, mapped to its camelCase CRD form and omitted entirely when
  # no labels are provided. The nested match_labels is read through a ?: conditional
  # (which only evaluates the taken branch) so a null selector never triggers an
  # attribute access on null.
  selector_match_labels = var.spec.selector != null ? var.spec.selector.match_labels : null
  selector = local.selector_match_labels != null ? {
    matchLabels = local.selector_match_labels
  } : null

  # Target references, proto snake_case mapped to camelCase, with unset optional
  # fields pruned per element so they do not reach the manifest.
  target_refs = var.spec.target_refs != null ? [
    for ref in var.spec.target_refs : merge(
      { kind = ref.kind, name = ref.name },
      ref.group != null ? { group = ref.group } : {},
      ref.namespace != null ? { namespace = ref.namespace } : {},
    )
  ] : null

  # JWT rules. Each rule's optional fields are null-pruned so only what the user
  # set reaches the manifest; nested from_headers and output_claim_to_headers
  # lists are built the same way.
  jwt_rules = var.spec.jwt_rules != null ? [
    for rule in var.spec.jwt_rules : merge(
      { issuer = rule.issuer },
      rule.audiences != null ? { audiences = rule.audiences } : {},
      rule.jwks_uri != null ? { jwksUri = rule.jwks_uri } : {},
      rule.jwks != null ? { jwks = rule.jwks } : {},
      rule.from_headers != null ? { fromHeaders = [
        for header in rule.from_headers : merge(
          { name = header.name },
          header.prefix != null ? { prefix = header.prefix } : {},
        )
      ] } : {},
      rule.from_params != null ? { fromParams = rule.from_params } : {},
      rule.from_cookies != null ? { fromCookies = rule.from_cookies } : {},
      rule.output_payload_to_header != null ? { outputPayloadToHeader = rule.output_payload_to_header } : {},
      rule.forward_original_token != null ? { forwardOriginalToken = rule.forward_original_token } : {},
      rule.output_claim_to_headers != null ? { outputClaimToHeaders = [
        for claim in rule.output_claim_to_headers : {
          header = claim.header
          claim  = claim.claim
        }
      ] } : {},
      rule.timeout != null ? { timeout = rule.timeout } : {},
    )
  ] : null

  # Assemble the full RequestAuthentication spec, omitting unset optional blocks
  # (and empty lists) so upstream/controller behavior flows through.
  request_authentication_spec = merge(
    local.selector != null ? { selector = local.selector } : {},
    local.target_refs != null && length(coalesce(local.target_refs, [])) > 0 ? { targetRefs = local.target_refs } : {},
    local.jwt_rules != null && length(coalesce(local.jwt_rules, [])) > 0 ? { jwtRules = local.jwt_rules } : {},
  )
}
