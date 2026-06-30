# OCI Vault Secret

Deploys an Oracle Cloud Infrastructure Vault secret — a named piece of sensitive data (credential, certificate, API key) stored in a KMS vault and encrypted by a master encryption key. Supports explicit base64 content, OCI-managed auto-generation (bytes, passphrase, SSH key), lifecycle rules for expiry and reuse, and scheduled rotation against Autonomous Database or OCI Functions targets.

## What Gets Created

When you deploy an OciVaultSecret resource, Planton provisions:

- **Vault Secret** — a `vault.Secret` resource in the specified compartment, vault, and encryption key. The secret is created with either explicit content or auto-generated content, and includes optional lifecycle rules and rotation configuration. Content updates create new secret versions automatically.

## Prerequisites

- **OCI credentials** configured via environment variables or Planton provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the secret will be created — either a literal value or a reference to an OciCompartment resource
- **A vault OCID** — the OciKmsVault resource that will contain this secret, either as a literal value or via `valueFrom`
- **A KMS key OCID** — a symmetric encryption key within the vault for encrypting the secret, either as a literal value or via `valueFrom` referencing an OciKmsKey resource
- **An Autonomous Database OCID** (for ADB rotation only) — if configuring scheduled rotation against an Autonomous Database
- **A Functions function OCID** (for function rotation only) — if configuring scheduled rotation via a custom OCI Functions function

## Quick Start

Create a file `secret.yaml`:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciVaultSecret
metadata:
  name: my-secret
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciVaultSecret.my-secret
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  secretName: "my-app-secret"
  vaultId:
    value: "ocid1.vault.oc1..example"
  keyId:
    value: "ocid1.key.oc1..example"
  secretContent:
    content: "c2VjcmV0LXZhbHVl"
```

Deploy:

```shell
planton apply -f secret.yaml
```

This creates a secret with explicit base64-encoded content, encrypted by the specified KMS key. The secret OCID and current version number are exported as stack outputs.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the secret will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `secretName` | `string` | Name of the secret. Must be unique within the vault. Immutable after creation. | Min length 1 |
| `vaultId` | `StringValueOrRef` | OCID of the vault that will contain this secret. Immutable after creation. Can reference an OciKmsVault resource via `valueFrom` using `status.outputs.vaultId`. | Required |
| `keyId` | `StringValueOrRef` | OCID of the master encryption key. Must be a symmetric key in the specified vault. Immutable after creation. Can reference an OciKmsKey resource via `valueFrom` using `status.outputs.keyId`. | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | — | Brief description of the secret. |
| `secretContent` | `SecretContent` | — | Explicit secret content (base64-encoded). Mutually exclusive with auto-generation. Updating creates a new secret version. |
| `enableAutoGeneration` | `bool` | `false` | Enable OCI-managed secret content generation. Mutually exclusive with `secretContent`. |
| `secretGenerationContext` | `SecretGenerationContext` | — | Configuration for auto-generation. Required when `enableAutoGeneration` is `true`; must not be set otherwise. |
| `secretRules` | `SecretRule[]` | — | Lifecycle rules for expiry and content reuse. See below. |
| `rotationConfig` | `RotationConfig` | — | Scheduled rotation configuration against a target system. See below. |
| `secretMetadata` | `map<string, string>` | — | Additional metadata key-value pairs for administrative context (e.g., rotation notes). |

### SecretContent

| Field | Type | Description |
|-------|------|-------------|
| `content` | `string` | Base64-encoded secret data. |
| `name` | `string` | Optional version name. Must be unique across versions. |
| `stage` | `string` | Rotation state. Values: `""` or `"CURRENT"` (default), `"PENDING"`. |

### SecretGenerationContext

| Field | Type | Description |
|-------|------|-------------|
| `generationType` | `enum` | Type of content to generate. Values: `bytes`, `passphrase`, `ssh_key`. |
| `generationTemplate` | `string` | Name of the generation template (provider-defined, varies by `generationType`). |
| `passphraseLength` | `int32` | Length of the passphrase to generate. Required when `generationType` is `passphrase`. Must be > 0. |
| `secretTemplate` | `string` | Optional template structure with placeholders for generated values. |

### SecretRule

| Field | Type | Description |
|-------|------|-------------|
| `ruleType` | `enum` | Rule type. Values: `secret_expiry_rule`, `secret_reuse_rule`. |
| `isSecretContentRetrievalBlockedOnExpiry` | `bool` | Block retrieval after version expires. Applies to `secret_expiry_rule`. |
| `secretVersionExpiryInterval` | `string` | Duration after which each version expires (ISO 8601, e.g., `"P30D"`). Range: 1-90 days. Applies to `secret_expiry_rule`. |
| `timeOfAbsoluteExpiry` | `string` | Absolute expiry timestamp (RFC 3339). Range: 1-365 days from creation. Applies to `secret_expiry_rule`. |
| `isEnforcedOnDeletedSecretVersions` | `bool` | Enforce the reuse rule even on deleted versions. Applies to `secret_reuse_rule`. |

### RotationConfig

| Field | Type | Description |
|-------|------|-------------|
| `isScheduledRotationEnabled` | `bool` | Enable scheduled automatic rotation. |
| `rotationInterval` | `string` | Rotation interval (ISO 8601, e.g., `"P30D"`). Range: 1-360 days. Required when scheduled rotation is enabled. |
| `targetSystemDetails` | `TargetSystemDetails` | Target system that will be updated during rotation. Required. |

### TargetSystemDetails

| Field | Type | Description |
|-------|------|-------------|
| `targetSystemType` | `enum` | Type of target system. Values: `adb` (Autonomous Database), `function` (OCI Functions). |
| `adbId` | `StringValueOrRef` | OCID of the Autonomous Database. Required when `targetSystemType` is `adb`. Can reference an OciAutonomousDatabase resource via `valueFrom`. |
| `functionId` | `StringValueOrRef` | OCID of the OCI Functions function invoked during rotation. Required when `targetSystemType` is `function`. |

## Examples

### Explicit Content

A secret with base64-encoded content — suitable for storing a database password or API key:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciVaultSecret
metadata:
  name: db-password
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciVaultSecret.db-password
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  secretName: "prod-db-password"
  vaultId:
    value: "ocid1.vault.oc1..example"
  keyId:
    value: "ocid1.key.oc1..example"
  description: "Production database admin password"
  secretContent:
    content: "c2VjcmV0LXZhbHVl"
```

### Auto-Generated Passphrase with Expiry

A passphrase auto-generated by OCI with a 30-day expiry rule:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciVaultSecret
metadata:
  name: app-passphrase
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciVaultSecret.app-passphrase
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  secretName: "app-passphrase"
  vaultId:
    valueFrom:
      kind: OciKmsVault
      name: prod-vault
      fieldPath: status.outputs.vaultId
  keyId:
    valueFrom:
      kind: OciKmsKey
      name: prod-key
      fieldPath: status.outputs.keyId
  enableAutoGeneration: true
  secretGenerationContext:
    generationType: passphrase
    generationTemplate: "SECRET_TEMPLATE_DBAAS"
    passphraseLength: 32
  secretRules:
    - ruleType: secret_expiry_rule
      secretVersionExpiryInterval: "P30D"
      isSecretContentRetrievalBlockedOnExpiry: true
```

### Scheduled Rotation Against Autonomous Database

A database credential that automatically rotates every 30 days against an Autonomous Database:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciVaultSecret
metadata:
  name: adb-credential
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciVaultSecret.adb-credential
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  secretName: "adb-admin-credential"
  vaultId:
    value: "ocid1.vault.oc1..example"
  keyId:
    value: "ocid1.key.oc1..example"
  enableAutoGeneration: true
  secretGenerationContext:
    generationType: passphrase
    generationTemplate: "SECRET_TEMPLATE_DBAAS"
    passphraseLength: 24
  rotationConfig:
    isScheduledRotationEnabled: true
    rotationInterval: "P30D"
    targetSystemDetails:
      targetSystemType: adb
      adbId:
        valueFrom:
          kind: OciAutonomousDatabase
          name: prod-adb
          fieldPath: status.outputs.autonomousDatabaseId
```

### Auto-Generated SSH Key with Reuse Rule

An auto-generated SSH key pair with a reuse rule preventing content reuse even on deleted versions:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciVaultSecret
metadata:
  name: ssh-key
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciVaultSecret.ssh-key
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  secretName: "bastion-ssh-key"
  vaultId:
    value: "ocid1.vault.oc1..example"
  keyId:
    value: "ocid1.key.oc1..example"
  enableAutoGeneration: true
  secretGenerationContext:
    generationType: ssh_key
    generationTemplate: "2048"
  secretRules:
    - ruleType: secret_reuse_rule
      isEnforcedOnDeletedSecretVersions: true
  secretMetadata:
    purpose: "bastion-access"
    team: "platform"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `secret_id` | `string` | OCID of the Vault Secret |
| `current_version_number` | `string` | Version number of the currently active secret version |

## Related Components

- [OciKmsVault](/docs/catalog/oci/ocikmsvault) — provides the vault referenced by `vaultId` via `valueFrom`
- [OciKmsKey](/docs/catalog/oci/ocikmskey) — provides the encryption key referenced by `keyId` via `valueFrom`
- [OciCompartment](/docs/catalog/oci/ocicompartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
- [OciAutonomousDatabase](/docs/catalog/oci/ociautonomousdatabase) — rotation target referenced by `adbId` via `valueFrom`
