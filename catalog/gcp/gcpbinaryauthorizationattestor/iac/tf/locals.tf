locals {
  # Empty project means the provider's default project, read from the
  # provider's own configuration (google_client_config, no API call) --
  # identical to the Pulumi module's GetClientConfig fallback.
  needs_client_project = var.spec.project_id == ""
  project = (
    var.spec.project_id != "" ? trimprefix(var.spec.project_id, "projects/") :
    data.google_client_config.current[0].project
  )

  attestor_name   = var.spec.attestor_name != "" ? var.spec.attestor_name : var.metadata.name
  description     = var.spec.description != "" ? var.spec.description : null
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # The attestor's note: created here (Google's one-note-per-attestor shape)
  # or an existing one by reference.
  creates_note = var.spec.note != null
  note_name    = local.creates_note ? (var.spec.note.note_name != "" ? var.spec.note.note_name : "${local.attestor_name}-note") : null

  authority_note = var.spec.attestation_authority_note
  public_keys    = local.authority_note != null ? local.authority_note.public_keys : []

  # PKIX keys held in Cloud KMS, by position: their public key and algorithm
  # are read from the key version, the way Google's own example wires an
  # attestor to Cloud KMS.
  kms_key_versions = {
    for index, key in local.public_keys :
    tostring(index) => key.pkix_public_key.kms_key_version
    if key.pkix_public_key != null && try(key.pkix_public_key.kms_key_version, "") != ""
  }
}
