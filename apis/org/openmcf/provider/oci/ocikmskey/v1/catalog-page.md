# OCI KMS Key

Deploys an Oracle Cloud Infrastructure Key Management Service encryption key inside a KMS vault. Supports AES, RSA, and ECDSA algorithms with configurable key length, HSM/software/external protection modes, and optional automatic key rotation.

## What Gets Created

When you deploy an OciKmsKey resource, OpenMCF provisions:

- **KMS Key** — a `kms.Key` resource in the specified compartment and vault with configurable algorithm, key length, protection mode, and optional auto-rotation schedule. The key is created in ENABLED state and an initial key version is generated automatically.

## Prerequisites

- **OCI credentials** configured via environment variables or OpenMCF provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the key will be created — either a literal value or a reference to an OciCompartment resource
- **A vault management endpoint** — the `managementEndpoint` output from an OciKmsVault resource, either as a literal URL or via `valueFrom`
- **An external key ID** (for external protection mode only) — the identifier of the key on the third-party key manager

## Quick Start

Create a file `key.yaml`:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciKmsKey
metadata:
  name: my-key
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciKmsKey.my-key
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  managementEndpoint:
    value: "https://xxx-management.kms.us-ashburn-1.oraclecloud.com"
  keyShape:
    algorithm: aes
    length: 32
```

Deploy:

```shell
openmcf apply -f key.yaml
```

This creates a 256-bit AES key with HSM protection (the default). The key OCID and current key version OCID are exported as stack outputs.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the key will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `managementEndpoint` | `StringValueOrRef` | Vault management endpoint URL. Can reference an OciKmsVault resource via `valueFrom` using `status.outputs.managementEndpoint`. | Required |
| `keyShape` | `KeyShape` | Cryptographic properties of the key. Immutable after creation. | Required |
| `keyShape.algorithm` | `enum` | Encryption algorithm. Values: `aes`, `rsa`, `ecdsa`. | Required, not `unspecified` |
| `keyShape.length` | `int32` | Key length in bytes. AES: 16/24/32. RSA: 256/384/512. ECDSA: 32/48/66. | > 0 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `displayName` | `string` | metadata name | Display name for the key in the OCI Console. |
| `protectionMode` | `enum` | `hsm` | Where key material is stored. Values: `hsm` (FIPS 140-2 Level 3 HSM), `software` (software-based, lower cost), `external` (third-party key manager). Immutable. |
| `keyShape.curveId` | `enum` | — | Elliptic curve for ECDSA keys. Values: `nist_p256`, `nist_p384`, `nist_p521`. Required for ECDSA; must not be set for AES or RSA. |
| `isAutoRotationEnabled` | `bool` | `false` | Enables automatic key rotation on a schedule. |
| `autoKeyRotationDetails` | `AutoKeyRotationDetails` | — | Schedule configuration for auto-rotation. Only valid when `isAutoRotationEnabled` is `true`. |
| `externalKeyReference` | `ExternalKeyReference` | — | Reference to a key on an external key manager. Required when `protectionMode` is `external`; must not be set otherwise. |

### AutoKeyRotationDetails

| Field | Type | Description |
|-------|------|-------------|
| `rotationIntervalInDays` | `int32` | Rotation interval in days. When omitted, OCI uses its default rotation interval. |
| `timeOfScheduleStart` | `string` | RFC 3339 timestamp for when the first rotation should occur. When omitted, OCI schedules based on key creation time. |

### ExternalKeyReference

| Field | Type | Description |
|-------|------|-------------|
| `externalKeyId` | `string` | Identifier of the key on the external key manager. |

## Examples

### AES-256 Key with HSM Protection

A 256-bit AES symmetric key in an HSM — the most common choice for data-at-rest encryption:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciKmsKey
metadata:
  name: aes-key
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciKmsKey.aes-key
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  managementEndpoint:
    value: "https://xxx-management.kms.us-ashburn-1.oraclecloud.com"
  keyShape:
    algorithm: aes
    length: 32
```

### RSA-4096 Key with Auto-Rotation

A 4096-bit RSA asymmetric key with automatic rotation every 90 days, using `valueFrom` to reference a vault:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciKmsKey
metadata:
  name: rsa-signing-key
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciKmsKey.rsa-signing-key
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-security
      fieldPath: status.outputs.compartmentId
  managementEndpoint:
    valueFrom:
      kind: OciKmsVault
      name: prod-vault
      fieldPath: status.outputs.managementEndpoint
  keyShape:
    algorithm: rsa
    length: 512
  isAutoRotationEnabled: true
  autoKeyRotationDetails:
    rotationIntervalInDays: 90
```

### ECDSA P-384 Key with Software Protection

An ECDSA P-384 key with software-based protection — lower cost for non-regulatory workloads:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciKmsKey
metadata:
  name: ecdsa-key
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: staging.OciKmsKey.ecdsa-key
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  managementEndpoint:
    value: "https://xxx-management.kms.us-ashburn-1.oraclecloud.com"
  keyShape:
    algorithm: ecdsa
    length: 48
    curveId: nist_p384
  protectionMode: software
```

### External Key (BYOK)

A key backed by an external key manager for regulatory BYOK requirements:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciKmsKey
metadata:
  name: byok-key
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciKmsKey.byok-key
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  managementEndpoint:
    value: "https://xxx-management.kms.us-ashburn-1.oraclecloud.com"
  keyShape:
    algorithm: aes
    length: 32
  protectionMode: external
  externalKeyReference:
    externalKeyId: "ekm-key-uuid-12345"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `key_id` | `string` | OCID of the KMS key |
| `current_key_version` | `string` | OCID of the currently active key version |

## Related Components

- [OciKmsVault](/docs/catalog/oci/ocikmsvault) — provides the `managementEndpoint` consumed by this component via `valueFrom`
- [OciCompartment](/docs/catalog/oci/ocicompartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
- [OciBlockVolume](/docs/catalog/oci/ociblockvolume) — uses this key for volume encryption via `kmsKeyId`
- [OciObjectStorageBucket](/docs/catalog/oci/ociobjectstoragebucket) — uses this key for bucket encryption via `kmsKeyId`
- [OciVaultSecret](/docs/catalog/oci/ocivaultsecret) — uses this key to encrypt secrets via `keyId`
