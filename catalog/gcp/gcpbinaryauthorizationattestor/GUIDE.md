# GcpBinaryAuthorizationAttestor Guide

The judgment this guide protects: an attestor is only as trustworthy as its signing key, and only useful if it can read its own note. Keep the private key in Cloud KMS, let the block create and grant the note, and never change the note or name of an attestor policies depend on without re-applying them.

## How attestation works

1. A pipeline that approves an image signs the image digest with the attestor's private key and records the signature as an occurrence of the attestor's note (gcloud `beta container binauthz attestations sign-and-create`, or Google's signing tools).
2. A `GcpBinaryAuthorizationPolicy` rule with `REQUIRE_ATTESTATION` names the attestor.
3. At pod creation, Binary Authorization reads the occurrences with the attestor's service account and admits the image if one of `publicKeys` verifies a signature.

This block creates the attestor and the trust anchor; the signatures are written by pipelines.

## The note

Set `note` and the block creates an ATTESTATION_AUTHORITY note in the attestor's project and grants the attestor's service account (`delegation_service_account_email`) `roles/containeranalysis.notes.occurrences.viewer` on it -- Google requires that grant before the attestor can verify anything. Use `attestationAuthorityNote.noteReference` only for a note someone else owns; grant the role on it yourself. The note is immutable on the attestor: changing it replaces the attestor.

## Keys

- **Cloud KMS (recommended).** `pkixPublicKey.kmsKeyVersion` names an `ASYMMETRIC_SIGN` key version -- a `GcpKmsKey` reference reads `initial_version_name`. Both modules read the version's public key and algorithm from Cloud KMS and set the key `id` to `//cloudkms.googleapis.com/v1/{version}`, the ID gcloud puts in signatures. Signers need `roles/cloudkms.signer` on the key.
- **PEM.** `publicKeyPem` with its `signatureAlgorithm`, for a key pair held elsewhere. Leave `id` empty for Google's digest-based default or give an RFC 3986 URI.
- **PGP.** `asciiArmoredPgpPublicKey`, the whole `gpg --export --armor` output; leave `id` empty (Google computes the fingerprint).

Rotate by adding the new key, re-signing, then removing the old one. `ML_DSA_65` (post-quantum) is accepted.

## Across projects

A policy in another project needs that project's Binary Authorization service agent granted `roles/binaryauthorization.attestorsVerifier` on this attestor.
