# Cloud KMS Signer

## Use Case

An attestor for the CI pipeline whose signing key lives in Cloud KMS, so the private key can sign but never be copied.

## When to Use

- The standard production signer
- Pipelines that sign with `gcloud beta container binauthz attestations sign-and-create --keyversion=...`

## What This Creates

- The Binary Authorization and Artifact Analysis APIs
- The attestor's note and the grant that lets the attestor read it
- The attestor verifying the first version of the `ci-signing-key` `GcpKmsKey` (purpose `ASYMMETRIC_SIGN`)

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `attestationAuthorityNote.publicKeys[].pkixPublicKey.kmsKeyVersion` | `ci-signing-key` | Your signing key's `GcpKmsKey`. |
| `note.humanReadableName` | `CI build pipeline` | Name the signer the way your team does. |
