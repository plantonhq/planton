# Overview

The **AzureKeyVaultSecret** component stores a secret -- a password, API key, connection string, or any small sensitive string -- inside an Azure Key Vault. Secrets are versioned: every value change creates a new version, and consumers choose whether to follow updates (the versionless reference) or pin a frozen version (the versioned reference). The secret's value is a sensitive, reference-resolved input -- it is never embedded in manifests, and it is never exposed as an output.

## Purpose

- **Secrets as declarative infrastructure**: which secrets exist, where they live, and when they expire is reviewed and versioned -- only the values stay out of the manifests.
- **The handoff point between platforms**: databases, CI systems, and applications read credentials from the vault at runtime instead of carrying them in configuration.
- **Rotation without redeployment**: consumers referencing the versionless ID pick up new values automatically when the secret rotates.

## Key Features

- Full azurerm v5 surface: value, content type, activation/expiry attributes, and tags (Key Vault's own 15-tag cap mirrored in validation).
- The value is a sensitive reference-resolved field -- reference a managed secret or another resource's output; plaintext never lands in a manifest.
- Chart-ready: publishes the versioned and versionless data-plane IDs plus both ARM resource IDs.

## Use Cases

- **Database credentials**: store the admin password once; applications and pipelines read it from the vault.
- **Third-party API keys** with expiry attributes that make rotation auditable.
- **Connection strings** consumed by App Service, Functions, and AKS workloads via Key Vault references.

## Future Enhancements

- The value's source references widen as more credential-emitting kinds land in the catalog.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
