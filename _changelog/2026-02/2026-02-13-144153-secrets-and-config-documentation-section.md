# Secrets and Config Documentation Section

**Date**: February 13, 2026
**Type**: Feature
**Components**: Documentation

## Summary

Created a new top-level "Secrets and Config" documentation section (4 pages) covering Planton's built-in Secrets Manager and configuration variable management. Deleted the superseded `service-hub/secrets-and-variables.md` page and updated all cross-references. This is the first public documentation of the Config Manager feature, including envelope encryption, customer-managed encryption keys (CMEK), secret backends, and the just-in-time decryption model.

## Problem Statement / Motivation

Planton's Config Manager — a significant shipped feature with envelope encryption, 6 pluggable secret backends, CMEK support, and immutable secret versioning — had zero public documentation. The existing `service-hub/secrets-and-variables.md` page covered the old SecretsGroup/VariablesGroup API, which is being superseded by the new Config Manager API with fundamentally different capabilities (envelope encryption, secret backends, versioned secrets, CMEK).

### Pain Points

- No documentation for envelope encryption, secret backends, or CMEK — enterprise-critical features
- The old SecretsGroup/VariablesGroup page described a superseded API
- No documentation for the just-in-time decryption model (secrets decrypted only within the Planton Runner)
- No documentation for the scoping model (organization vs environment secrets/variables)
- No documentation for the 6 pluggable backend options

## Solution / What's New

### New Section: `public/docs/secrets-and-config/` (4 pages)

**index.md** — Overview page establishing the Secrets Manager and configuration variable system. Covers the problem (scattered secrets), the solution (single secure home), the scoping model (GitHub Organization Secrets analogy), three security guarantees (always encrypted, decrypted only in customer infrastructure, customer can own the encryption key), backend options, and the platform integration roadmap.

**secrets.md** — Full secrets lifecycle documentation. Covers creating secrets, immutable versioning (why each version gets its own encryption key), scoping, the just-in-time decryption flow (with Mermaid sequence diagram), referencing secrets from services, and future platform integrations (connection fields, infra component inputs, sensitive outputs).

**variables.md** — Variables and Variable Groups documentation. Covers creating variables, scoping, dynamic references to infrastructure outputs (ValueFromRef), Variable Groups for configuration reuse, referencing from services, .env file generation, and future platform integrations.

**secret-backends.md** — Backend options and encryption deep dive. Covers the built-in backend (OpenBAO), 5 bring-your-own options (AWS Secrets Manager, GCP Secret Manager, Azure Key Vault, HashiCorp Vault, self-hosted OpenBAO), envelope encryption (two-tier key hierarchy with Mermaid diagram), zero-trust encryption model, CMEK (AWS KMS, GCP Cloud KMS, Azure Key Vault), and a decision matrix for choosing the right configuration.

### Deleted Page

**service-hub/secrets-and-variables.md** — Removed. Covered the old SecretsGroup/VariablesGroup API which is superseded by the Config Manager.

### Updated Cross-References (3 pages)

- `service-hub/index.md` — Updated link and description to point to new section
- `service-hub/deployment-targets.md` — Updated related docs link
- `service-hub/what-is-a-service.md` — Updated resource table (SecretsGroup/VariablesGroup replaced with Secret/Variable descriptions) and link

## Implementation Details

### Key Terminology Decisions

- **"Connections"** used instead of "credentials" when referring to Connect APIs (per project owner direction)
- **"Secrets Manager"** as the user-facing product name (not "Config Manager" which is the internal domain name)
- **OpenBAO** named explicitly (open-source, credibility value) but described as "built-in backend" in most contexts
- No protobuf field names, message types, or enum values in user-facing prose

### Source Verification

Content verified against:
- 5 Config Manager protobuf API resources (Variable, VariableGroup, Secret, SecretVersion, SecretBackend)
- 7 ADRs (Config Manager domain, pluggable backends, envelope encryption, CMEK providers, provider abstraction, secret-zero bootstrap)
- 6 changelog entries documenting Config Manager implementation steps
- CLI commands from Go source (`planton service secrets`, `planton service variables`, `planton service env-vars`, `planton service dot-env`)
- Web console routes (`/orgs/[org]/secrets`, `/orgs/[org]/variables`, secret/variable detail pages)

### Design-Intent Documentation

Per project owner direction, the following are documented as committed design direction with appropriate disclaimers:
- Secrets decrypted just-in-time only within the Planton Runner (customer infrastructure)
- Future integration of secret/variable references into connection fields
- Future integration into Infra Component input fields
- Future routing of sensitive outputs to Secrets Manager

## Benefits

- **Enterprise readiness** — Envelope encryption and CMEK are now documented for the first time, enabling sales and security conversations
- **Developer clarity** — Clear documentation of secrets vs variables, scoping, and the runtime decryption flow
- **Platform vision** — Future integration roadmap gives users confidence in the direction
- **Clean architecture** — Old superseded API page removed, preventing confusion between old and new systems

## Impact

- 4 new documentation pages (~780 lines total)
- 1 page deleted (~190 lines)
- 3 pages updated with corrected cross-references
- New top-level section in docs sidebar (order 45, after Service Hub)

## Related Work

- [CP01-CP07] Previous documentation overhaul sessions established the docs quality bar and philosophy
- Envelope Encryption ADR (2026-02-09) — Primary source for encryption documentation
- Config Manager implementation (February 2026) — The shipped feature being documented

---

**Status**: Live
**Timeline**: Single session
