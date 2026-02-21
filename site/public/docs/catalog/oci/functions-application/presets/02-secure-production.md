---
title: "Secure Production"
description: "This preset creates an OCI Functions Application with container image signature verification enforced, NSG-protected networking, and APM tracing. Only container images signed by the specified KMS key..."
type: "preset"
rank: "02"
presetSlug: "02-secure-production"
componentSlug: "functions-application"
componentTitle: "Functions Application"
provider: "oci"
icon: "package"
order: 2
---

# Secure Production

This preset creates an OCI Functions Application with container image signature verification enforced, NSG-protected networking, and APM tracing. Only container images signed by the specified KMS key can be deployed as functions in this application. Unsigned or incorrectly signed images are rejected at deploy time, preventing unauthorized or tampered code from running in the production environment.

## When to Use

- Production environments in regulated industries (finance, healthcare, government) where code provenance must be cryptographically verified
- Organizations with supply chain security requirements mandating signed container images
- Environments where only CI/CD pipeline-produced artifacts (signed during build) should be deployable
- Multi-team organizations where platform teams control which images are approved for production via signing keys

## Key Configuration Choices

- **Image signature verification** (`imagePolicyConfig.isPolicyEnabled: true`) -- the OCI Functions service validates container image signatures against the specified KMS keys before allowing deployment. Images must be signed using `oci artifacts container image sign` or an equivalent CI/CD signing step with the same KMS key. This is the primary security control that distinguishes this preset from the standard one.
- **KMS signing key** (`keyDetails` with `kmsKeyId`) -- the RSA or ECDSA key used to verify image signatures. Use the same key that your CI/CD pipeline uses to sign images after building them. Multiple keys can be listed to support key rotation (images signed by any listed key are accepted). Create a dedicated signing key using the `OciKmsKey` 02-rsa-4096-hsm-signing preset.
- **x86 architecture, NSG protection, APM tracing** -- same configuration as the standard x86 preset. These settings are independent of image policy enforcement.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the application will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<private-subnet-ocid>` | OCID of the private subnet where functions will execute | OCI Console > Networking > Subnets, or `OciSubnet` status outputs |
| `<functions-nsg-ocid>` | OCID of the NSG controlling function network access | OCI Console > Networking > NSGs, or `OciSecurityGroup` status outputs |
| `<image-signing-key-ocid>` | OCID of the KMS key used to sign container images in CI/CD | `OciKmsKey` status outputs (`keyId`), or OCI Console > Identity & Security > Vault > Keys |
| `<apm-domain-ocid>` | OCID of the APM domain for distributed tracing | OCI Console > Observability & Management > Application Performance Monitoring > Domains |

## Related Presets

- **01-standard-x86** -- Use instead when image signature verification is not required (development, staging)
