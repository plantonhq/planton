# GCP Binary Authorization Attestor

Declares a trusted signer for container images. Your build or QA pipeline signs each image it approves; a Binary Authorization policy that requires this attestor admits only images it has signed. Use a Cloud KMS key so the private signing key never leaves Google's HSM-backed key service.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `binaryauthorization.googleapis.com` and `containeranalysis.googleapis.com`
- **Note** -- the attestor's own `containeranalysis.Note`, when `note` is set
- **Attestor** -- one `binaryauthorization.Attestor`
- **Note grant** -- read access to the note for the attestor's service account, when `note` is set

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Binary Authorization attestor and Artifact Analysis note admin permissions on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

## Deploy

### Console

Open the deployment store, find **GCP Binary Authorization Attestor**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Cloud KMS Signer** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpBinaryAuthorizationAttestor
metadata:
  name: built-by-ci
  org: acme-corp
  env: prod
spec:
  description: Images built from main by the CI pipeline
  note:
    humanReadableName: CI build pipeline
  attestationAuthorityNote:
    publicKeys:
      - comment: Signing key held in Cloud KMS
        pkixPublicKey:
          kmsKeyVersion:
            value: projects/shop-security/locations/global/keyRings/binauthz/cryptoKeys/ci-signer/cryptoKeyVersions/1
```

```shell
planton apply -f binary-authorization-attestor.yaml
```

This creates the CI attestor and its note, verifying signatures made with the KMS key. A Stack Job tracks the provisioning in real time.

### InfraChart

Reference a `GcpKmsKey` (purpose `ASYMMETRIC_SIGN`) from `pkixPublicKey.kmsKeyVersion` with its `status.outputs.initial_version_name`, and the attestor's `status.outputs.attestor_id` from the `GcpBinaryAuthorizationPolicy` rules that require it.

## Key Configuration

These are the most important decisions when configuring an attestor. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The note** -- create the attestor's own note (recommended), or reference one that exists.

**The keys** -- a Cloud KMS key version (the private key never leaves KMS), a PEM with its algorithm, or a PGP key.

**Who uses it** -- policies in this project, or in others with the verifier grant.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpKmsKey** | `attestationAuthorityNote.publicKeys[].pkixPublicKey.kmsKeyVersion` | `status.outputs.initial_version_name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `attestor_id` | The attestor's full name | `GcpBinaryAuthorizationPolicy` rules |
| `note_reference` | The note attestations are recorded against | Signing pipelines |
| `delegation_service_account_email` | The identity the attestor reads with | Grants on a referenced note |
| `attestor_name` | The attestor's ID | Reporting |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Cloud KMS signer** -- an attestor verifying a KMS signing key. Start from the **Cloud KMS Signer** preset.

**Existing note with a PEM key** -- an attestor over a note a security team already owns. Start from the **Existing Note With PEM Key** preset.

## Works With

- [**GCP Binary Authorization Policy**](/cloud-catalog/gcp-binary-authorization-policy) -- requires the attestor
- [**GCP KMS Key**](/cloud-catalog/gcp-kms-key) -- the signing key
