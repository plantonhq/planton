# Existing Note With PEM Key

## Use Case

An attestor over a note a central security team already owns, verifying signatures made with a key pair held outside Cloud KMS.

## When to Use

- A shared note in a dedicated notes project
- A signer whose key lives in an external HSM or signing service

## What This Creates

- The Binary Authorization and Artifact Analysis APIs
- The attestor over the existing note, verifying an ECDSA P-256 PEM key

Grant the attestor's `delegation_service_account_email` output `roles/containeranalysis.notes.occurrences.viewer` on the note.

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `attestationAuthorityNote.noteReference` | `security-notes/qa-approved` | Your note. |
| `pkixPublicKey.publicKeyPem` | example | Your public key. |
| `pkixPublicKey.signatureAlgorithm` | `ECDSA_P256_SHA256` | Match your key. |
