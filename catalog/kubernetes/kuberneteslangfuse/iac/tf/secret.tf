# The module-owned `<metadata.name>-secrets` Secret: every declared secret
# value plus the credentials generated below for the ones the spec leaves
# empty (Pulumi twin: secrets.go). The chart's values and the additionalEnv
# channels reference its keys by secretKeyRef, so no secret ever lands in
# the Helm release.
#
# Every generator ignores its generation-shape arguments after creation, so
# an IMPORTED credential never silently regenerates: rotation stays an
# explicit verb, never plan fallout. Letters and digits only - symbols are
# the one class that can need quoting in a URL or a shell.

# Salt for API-key hashes: 48 letters and digits (~285 bits).
resource "random_password" "salt" {
  count   = local.generate.salt ? 1 : 0
  length  = 48
  special = false

  lifecycle {
    ignore_changes = [length, special, upper, lower, numeric, min_lower, min_numeric, min_special, min_upper, override_special]
  }
}

# Encryption key: Langfuse requires exactly 64 hex characters - 32 random
# bytes, hex-encoded (`openssl rand -hex 32`). The Secret also carries the
# same bytes in standard base64 (`encryption-key-std-b64`), the form the
# random_bytes importer takes.
resource "random_bytes" "encryption_key" {
  count  = local.generate.encryption_key ? 1 : 0
  length = 32

  lifecycle {
    ignore_changes = [length]
  }
}

resource "random_password" "nextauth_secret" {
  count   = local.generate.nextauth_secret ? 1 : 0
  length  = 48
  special = false

  lifecycle {
    ignore_changes = [length, special, upper, lower, numeric, min_lower, min_numeric, min_special, min_upper, override_special]
  }
}

resource "random_password" "slack_state_secret" {
  count   = local.generate.slack_state_secret ? 1 : 0
  length  = 48
  special = false

  lifecycle {
    ignore_changes = [length, special, upper, lower, numeric, min_lower, min_numeric, min_special, min_upper, override_special]
  }
}

# The seeded admin's password: 32 letters and digits.
resource "random_password" "init_user_password" {
  count   = local.generate.admin_password ? 1 : 0
  length  = 32
  special = false

  lifecycle {
    ignore_changes = [length, special, upper, lower, numeric, min_lower, min_numeric, min_special, min_upper, override_special]
  }
}

# The seeded project key pair, in Langfuse's key format: the fixed
# pk-lf-/sk-lf- prefix and 32 random letters and digits. The Secret also
# carries each generated part alone (`*-suffix`), the value the
# random_password importer takes.
resource "random_password" "init_project_public_key" {
  count   = local.generate.project_keys ? 1 : 0
  length  = 32
  special = false

  lifecycle {
    ignore_changes = [length, special, upper, lower, numeric, min_lower, min_numeric, min_special, min_upper, override_special]
  }
}

resource "random_password" "init_project_secret_key" {
  count   = local.generate.project_keys ? 1 : 0
  length  = 32
  special = false

  lifecycle {
    ignore_changes = [length, special, upper, lower, numeric, min_lower, min_numeric, min_special, min_upper, override_special]
  }
}

resource "kubernetes_secret_v1" "instance" {
  metadata {
    name      = local.instance_secret_name
    namespace = local.namespace
    labels    = local.labels
  }

  type = "Opaque"

  # ONE merge, declared values over generated ones: a conditional that
  # indexed a count-0 generator would fail the plan.
  data = merge(
    { for k, v in {
      "salt"                           = try(random_password.salt[0].result, null)
      "encryption-key"                 = try(random_bytes.encryption_key[0].hex, null)
      "encryption-key-std-b64"         = try(random_bytes.encryption_key[0].base64, null)
      "nextauth-secret"                = try(random_password.nextauth_secret[0].result, null)
      "slack-state-secret"             = try(random_password.slack_state_secret[0].result, null)
      "init-user-password"             = try(random_password.init_user_password[0].result, null)
      "init-project-public-key"        = try("pk-lf-${random_password.init_project_public_key[0].result}", null)
      "init-project-public-key-suffix" = try(random_password.init_project_public_key[0].result, null)
      "init-project-secret-key"        = try("sk-lf-${random_password.init_project_secret_key[0].result}", null)
      "init-project-secret-key-suffix" = try(random_password.init_project_secret_key[0].result, null)
    } : k => v if v != null },
    local.declared_secrets,
  )

  depends_on = [kubernetes_namespace_v1.langfuse]
}
