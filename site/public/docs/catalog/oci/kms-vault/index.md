---
title: "KMS Vault"
description: "KMS Vault deployment documentation"
icon: "package"
order: 100
componentName: "ocikmsvault"
---

# OCI KMS Vault

Deploys an Oracle Cloud Infrastructure Key Management Service vault — an HSM-backed container for encryption keys used by Compute, Block Volume, Object Storage, Database, and other OCI services. Supports shared (Default), dedicated (Virtual Private), and external (BYOK/EKMS) vault types.

## What Gets Created

When you deploy an OciKmsVault resource, OpenMCF provisions:

- **KMS Vault** — a `kms.Vault` resource in the specified compartment with configurable vault type (shared HSM, dedicated HSM, or external key manager). The vault exposes crypto and management endpoints consumed by downstream OciKmsKey resources.

## Prerequisites

- **OCI credentials** configured via environment variables or OpenMCF provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the vault will be created — either a literal value or a reference to an OciCompartment resource
- **IDCS OAuth credentials** (for external vaults only) — a registered IDCS client app ID, secret, and account URL for connecting to the third-party key manager
- **A KMS private endpoint OCID** (for external vaults only) — a pre-existing private endpoint for network connectivity to the external HSM

## Quick Start

Create a file `vault.yaml`:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciKmsVault
metadata:
  name: my-vault
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciKmsVault.my-vault
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  vaultType: default_vault
```

Deploy:

```shell
openmcf apply -f vault.yaml
```

This creates a shared-HSM vault suitable for most workloads. The vault OCID, crypto endpoint, and management endpoint are exported as stack outputs.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the vault will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `vaultType` | `enum` | Type of vault to create. Immutable after creation. Values: `default_vault` (shared HSM, lower cost), `virtual_private` (dedicated HSM, higher throughput), `external` (third-party HSM via IDCS OAuth). | Required, not `unspecified` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `displayName` | `string` | metadata name | Display name for the vault in the OCI Console. |
| `externalKeyManagerMetadata` | `ExternalKeyManagerMetadata` | — | Connection configuration for an external key manager. Required when `vaultType` is `external`; must not be set otherwise. All sub-fields are immutable after creation. |

### ExternalKeyManagerMetadata

| Field | Type | Description |
|-------|------|-------------|
| `externalVaultEndpointUrl` | `string` | URI of the vault on the external key manager system. |
| `oauthMetadata` | `OAuthMetadata` | IDCS OAuth credentials for authenticating with the external key manager. Required. |
| `privateEndpointId` | `string` | OCID of a KMS private endpoint for network connectivity to the external HSM. |

### OAuthMetadata

| Field | Type | Description |
|-------|------|-------------|
| `clientAppId` | `string` | Application ID of the client app registered in IDCS. |
| `clientAppSecret` | `string` | Secret of the client app registered in IDCS. Sensitive — not returned by the API after creation. |
| `idcsAccountNameUrl` | `string` | Base URL of the IDCS account (e.g., `"https://idcs-xxx.identity.oraclecloud.com"`). |

## Examples

### Shared HSM Vault

A default vault with shared HSM partition — suitable for most encryption use cases:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciKmsVault
metadata:
  name: shared-vault
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.OciKmsVault.shared-vault
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  vaultType: default_vault
```

### Dedicated HSM Vault

A virtual private vault with a dedicated HSM partition for high-throughput cryptographic operations:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciKmsVault
metadata:
  name: dedicated-vault
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciKmsVault.dedicated-vault
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  displayName: "prod-dedicated-vault"
  vaultType: virtual_private
```

### External Key Manager Vault

A BYOK/EKMS vault connecting to a third-party HSM via IDCS OAuth and a KMS private endpoint:

```yaml
apiVersion: oci.openmcf.org/v1
kind: OciKmsVault
metadata:
  name: external-vault
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.OciKmsVault.external-vault
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  vaultType: external
  externalKeyManagerMetadata:
    externalVaultEndpointUrl: "https://ekm.corp.example.com/vault/prod"
    oauthMetadata:
      clientAppId: "abcdef1234567890"
      clientAppSecret: "secret-value-here"
      idcsAccountNameUrl: "https://idcs-abc123.identity.oraclecloud.com"
    privateEndpointId: "ocid1.kmsendpoint.oc1..example"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `vault_id` | `string` | OCID of the KMS vault |
| `crypto_endpoint` | `string` | Service endpoint for cryptographic operations (encrypt, decrypt, sign, verify) |
| `management_endpoint` | `string` | Service endpoint for key management operations (create, import, rotate keys) |

## Related Components

- [OciCompartment](/docs/catalog/oci/compartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
- [OciKmsKey](/docs/catalog/oci/kms-key) — creates encryption keys within this vault using the `managementEndpoint` output
- [OciVaultSecret](/docs/catalog/oci/vault-secret) — stores secrets encrypted by keys in this vault using the `vaultId` output
