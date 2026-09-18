# One external issuer inside a Workload Identity Pool — the piece that makes
# keyless federation work. The pool is referenced by its bare ID; IAM
# principals and token audiences are built from this provider's resource name.
#
# workload_identity_pool_id, workload_identity_pool_provider_id, and project
# are immutable (ForceNew): changing any of them destroys and recreates the
# provider, invalidating tokens minted for the old audience. The issuer arm's
# contents, attribute_mapping, attribute_condition, display_name, description,
# and disabled all update in place. The issuer TYPE (aws vs oidc vs saml vs
# x509) cannot change on a live provider — the API rejects cross-type updates,
# so switching issuers means a new provider.
#
# GCP soft-deletes providers: after destroy, the provider remains DELETED for
# ~30 days and its ID cannot be reused until permanent deletion (no
# undelete-on-create). Prefer `disabled` for temporary shutoffs.
resource "google_iam_workload_identity_pool_provider" "this" {
  workload_identity_pool_id          = var.spec.workload_identity_pool_id
  workload_identity_pool_provider_id = var.spec.workload_identity_pool_provider_id
  project                            = local.project_id
  display_name                       = var.spec.display_name
  description                        = var.spec.description
  # Always sent: `disabled` is the documented kill switch, and an explicit
  # false keeps re-enable flows (true -> false) working.
  disabled            = var.spec.disabled
  attribute_mapping   = local.attribute_mapping
  attribute_condition = var.spec.attribute_condition

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # Exactly one issuer arm renders (enforced by the variable validation); the
  # API enforces the same exclusivity server-side.
  dynamic "aws" {
    for_each = var.spec.aws != null ? [var.spec.aws] : []
    content {
      account_id = aws.value.account_id
    }
  }

  dynamic "oidc" {
    for_each = var.spec.oidc != null ? [var.spec.oidc] : []
    content {
      issuer_uri = oidc.value.issuer_uri
      # An empty audience list means "audience must equal the provider's own
      # canonical resource name" — the safest default; only send overrides.
      allowed_audiences = length(oidc.value.allowed_audiences) > 0 ? oidc.value.allowed_audiences : null
      # Unset JWKS means keys are fetched from the issuer's .well-known
      # discovery document — the normal path for public issuers. An empty
      # string is unset: sending it would hand the API an empty key set.
      jwks_json = oidc.value.jwks_json != "" ? oidc.value.jwks_json : null
    }
  }

  dynamic "saml" {
    for_each = var.spec.saml != null ? [var.spec.saml] : []
    content {
      idp_metadata_xml = saml.value.idp_metadata_xml
    }
  }

  dynamic "x509" {
    for_each = var.spec.x509 != null ? [var.spec.x509] : []
    content {
      trust_store {
        dynamic "trust_anchors" {
          for_each = x509.value.trust_store.trust_anchors
          content {
            pem_certificate = trust_anchors.value.pem_certificate
          }
        }

        dynamic "intermediate_cas" {
          for_each = x509.value.trust_store.intermediate_cas
          content {
            pem_certificate = intermediate_cas.value.pem_certificate
          }
        }
      }
    }
  }
}
