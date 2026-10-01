# GCP Binary Authorization Attestor

A Binary Authorization attestor with the Artifact Analysis note its attestations are stored under: the trusted signer a policy rule requires. A build or QA pipeline signs an image digest with a private key and records the signature against the note; at deploy time the policy admits the image only if one of the attestor's public keys verifies it. Keys can be PGP, a PEM, or a Cloud KMS signing key whose private half never leaves KMS.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `binaryauthorization.googleapis.com` and `containeranalysis.googleapis.com` on the project (never disabled on destroy)
- **Note** (when `note` is set) -- the attestor's own `container_analysis_note` (ATTESTATION_AUTHORITY)
- **Attestor** -- one `binary_authorization_attestor` with its public keys
- **Note grant** (when `note` is set) -- `roles/containeranalysis.notes.occurrences.viewer` on the note for the attestor's service account, which Google requires

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Binary Authorization attestor admin (`roles/binaryauthorization.attestorsAdmin`) and Artifact Analysis notes editor (`roles/containeranalysis.notes.editor`) permissions on the project, and Cloud KMS public-key viewer on any KMS key it names.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpKmsKey`** -- an `ASYMMETRIC_SIGN` key whose first version verifies signatures (`pkixPublicKey.kmsKeyVersion`).
- **`GcpProject`** -- the project, by reference (`projectId`).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBinaryAuthorizationAttestor
metadata:
  name: built-by-ci
spec:
  note:
    humanReadableName: CI build pipeline
  attestationAuthorityNote:
    publicKeys:
      - pkixPublicKey:
          kmsKeyVersion:
            valueFrom:
              kind: GcpKmsKey
              name: ci-signing-key
              fieldPath: status.outputs.initial_version_name
```

```shell
planton apply -f binary-authorization-attestor.yaml
```

## Configuration Reference

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `string` / ref | provider project | The attestor's project (`GcpProject` ref). |
| `attestorName` | `string` | `metadata.name` | The attestor's ID. Immutable. |
| `description` | `string` | none | What the attestor vouches for. |
| `note` | `object` | none | Create the attestor's note: `humanReadableName` (required), `noteName` (default `{attestorName}-note`), descriptions, `expirationTime`, `relatedNoteNames`, `relatedUrl[]`. |
| `attestationAuthorityNote.noteReference` | `string` | none | An existing note instead: `projects/{project}/notes/{note}`. |
| `attestationAuthorityNote.publicKeys` | `object[]` | none | Verifying keys: `asciiArmoredPgpPublicKey`, or `pkixPublicKey` with `publicKeyPem` + `signatureAlgorithm` or `kmsKeyVersion` (`GcpKmsKey` ref); optional `id`, `comment`. |
| `deletionPolicy` | `string` | `DELETE` | For the attestor and a created note: `DELETE`, `PREVENT`, `ABANDON`. |

### Validation Rules

- Exactly one of `note` or `attestationAuthorityNote.noteReference`.
- Each key is exactly one of PGP or PKIX; a PGP key has no `id`.
- A PKIX key is exactly one of `publicKeyPem` (with `signatureAlgorithm`) or `kmsKeyVersion`.
- `signatureAlgorithm` takes Google's values; `kmsKeyVersion` names a key version.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `attestor_id` | `string` | `projects/{project}/attestors/{name}` -- what policies require |
| `attestor_name` | `string` | The attestor's ID |
| `note_reference` | `string` | The note signatures are recorded against |
| `delegation_service_account_email` | `string` | The identity the attestor reads attestations with |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **One note per attestor.** Google's recommended shape; `note` creates it and grants the attestor's service account the read access it needs. A referenced note's owner grants it there.
- **Cross-project policies.** A policy in another project needs `roles/binaryauthorization.attestorsVerifier` on this attestor for that project's Binary Authorization service agent.
- **Changing the note or the name replaces the attestor.** Re-apply every policy that names it afterwards.
- **No keys, no verification.** An attestor without public keys verifies nothing, so every rule requiring it denies.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpBinaryAuthorizationPolicy** -- requires the attestor
- **GcpKmsKey** -- an asymmetric signing key held in Cloud KMS

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
