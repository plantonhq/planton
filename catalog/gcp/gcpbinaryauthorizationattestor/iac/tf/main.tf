# The provider's default project, read only when the manifest names none.
data "google_client_config" "current" {
  count = local.needs_client_project ? 1 : 0
}

# Binary Authorization serves the attestor; Artifact Analysis stores the
# note and the attestations signed against it. disable_on_destroy is false:
# policies and other attestors in the project depend on both APIs.
resource "google_project_service" "apis" {
  for_each = toset(["binaryauthorization.googleapis.com", "containeranalysis.googleapis.com"])
  project  = local.project
  service  = each.value

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The attestor's own ATTESTATION_AUTHORITY note, when the spec asks for one.
resource "google_container_analysis_note" "this" {
  count = local.creates_note ? 1 : 0

  project           = local.project
  name              = local.note_name
  short_description = var.spec.note.short_description != "" ? var.spec.note.short_description : null
  long_description  = var.spec.note.long_description != "" ? var.spec.note.long_description : null
  expiration_time   = var.spec.note.expiration_time != "" ? var.spec.note.expiration_time : null
  related_note_names = (
    length(var.spec.note.related_note_names) > 0 ? var.spec.note.related_note_names : null
  )
  deletion_policy = local.deletion_policy

  attestation_authority {
    hint {
      human_readable_name = var.spec.note.human_readable_name
    }
  }

  dynamic "related_url" {
    for_each = var.spec.note.related_url
    content {
      url   = related_url.value.url
      label = related_url.value.label != "" ? related_url.value.label : null
    }
  }

  depends_on = [google_project_service.apis]
}

# The public key and algorithm of each Cloud KMS key version the spec names.
data "google_kms_crypto_key_version" "kms" {
  for_each = local.kms_key_versions

  crypto_key = regex("^(.+)/cryptoKeyVersions/[0-9]+$", each.value)[0]
  version    = tonumber(regex("/cryptoKeyVersions/([0-9]+)$", each.value)[0])
}

resource "google_binary_authorization_attestor" "this" {
  project         = local.project
  name            = local.attestor_name
  description     = local.description
  deletion_policy = local.deletion_policy

  attestation_authority_note {
    note_reference = local.creates_note ? google_container_analysis_note.this[0].id : local.authority_note.note_reference

    dynamic "public_keys" {
      for_each = { for index, key in local.public_keys : tostring(index) => key }
      content {
        # A Cloud KMS key's ID is the key version's full name, the ID gcloud
        # and Google's signing tools put in signatures.
        id = (
          public_keys.value.id != "" ? public_keys.value.id :
          contains(keys(local.kms_key_versions), public_keys.key) ? data.google_kms_crypto_key_version.kms[public_keys.key].id :
          null
        )
        comment                      = public_keys.value.comment != "" ? public_keys.value.comment : null
        ascii_armored_pgp_public_key = public_keys.value.ascii_armored_pgp_public_key != "" ? public_keys.value.ascii_armored_pgp_public_key : null

        dynamic "pkix_public_key" {
          for_each = public_keys.value.pkix_public_key != null ? [public_keys.value.pkix_public_key] : []
          content {
            public_key_pem = (
              contains(keys(local.kms_key_versions), public_keys.key) ?
              data.google_kms_crypto_key_version.kms[public_keys.key].public_key[0].pem :
              pkix_public_key.value.public_key_pem
            )
            signature_algorithm = (
              contains(keys(local.kms_key_versions), public_keys.key) ?
              data.google_kms_crypto_key_version.kms[public_keys.key].public_key[0].algorithm :
              pkix_public_key.value.signature_algorithm
            )
          }
        }
      }
    }
  }

  depends_on = [google_project_service.apis]
}

# Google requires the attestor's service account to read occurrences of its
# note before it can verify attestations; for a note this block created,
# the module grants it. A referenced note's owner grants it there.
resource "google_container_analysis_note_iam_member" "attestor_reads_note" {
  count = local.creates_note ? 1 : 0

  project = local.project
  note    = google_container_analysis_note.this[0].name
  role    = "roles/containeranalysis.notes.occurrences.viewer"
  member  = "serviceAccount:${google_binary_authorization_attestor.this.attestation_authority_note[0].delegation_service_account_email}"
}
