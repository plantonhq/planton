# GcpBinaryAuthorizationAttestor — Pulumi Implementation

This directory contains the Pulumi implementation for a Binary Authorization attestor from the Planton spec: `gcp.projects.Service` for the Binary Authorization and Artifact Analysis APIs, an optional `gcp.containeranalysis.Note` with its `NoteIamMember` grant, `kms.GetKMSCryptoKeyVersion` reads for Cloud KMS keys, and one `gcp.binaryauthorization.Attestor`.

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, and `attestor` |
| `module/locals.go` | Stack input holder |
| `module/attestor.go` | Project resolution, APIs, the note, the public keys (with the Cloud KMS reads), the attestor, the note grant, the outputs |
| `module/outputs.go` | Output key constants |

## Send Posture (parity with Terraform)

- **`Project`** -- the spec's project, or the provider's project from `organizations.GetClientConfig`.
- **`Name`** -- `attestor_name`, defaulting to `metadata.name` (always sent, never auto-named); the created note defaults to `{attestor_name}-note`.
- **Cloud KMS keys** -- PEM and algorithm from `kms.GetKMSCryptoKeyVersion`; `Id` defaults to its `//cloudkms.googleapis.com/v1/...` ID.
- **`DeletionPolicy`** -- only when set, to the attestor and the created note.
- **`DelegationServiceAccountEmail`** -- read, never set.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
