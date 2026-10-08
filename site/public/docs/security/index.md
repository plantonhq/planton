---
title: "Security"
description: "How Planton protects credentials, encrypts data, authorizes access, and tracks every change across your organization"
icon: shield
order: 70
tags:
  - Security
  - Encryption
  - Authorization
  - Credentials
  - Audit
---

# Security

Planton is a self-service cloud platform that touches some of the most sensitive parts of your infrastructure: cloud credentials, deployment secrets, production access controls, and runtime operations. The security model is designed around one premise — you should not have to trust Planton with more than you want to give it.

This page is an overview of how security works across the platform. Each section links to the detailed documentation where the full implementation is described.

## Credential Isolation

The defining security property of Planton is that your cloud credentials do not need to leave your infrastructure.

When Planton needs to interact with your cloud provider — creating a VPC, deploying a container, querying a database — it sends the instruction through a secure tunnel to a [Runner](/docs/runner) deployed in your environment. The Runner executes the operation using credentials it holds locally. Planton's control plane never receives, stores, or transmits your cloud credentials in plaintext.

This is not encryption at rest (though that exists too). It is architectural isolation. The credentials physically stay in your infrastructure. A breach of Planton's control plane does not expose your cloud credentials, because Planton does not have them.

The authentication method of each connection sets its security posture:

- **Inline credentials** — You provide credentials to Planton (encrypted at rest). Simplest to set up, suitable for development environments.
- **Runner-delegated authentication** — The Runner uses its own cloud identity (IRSA, Workload Identity, Managed Identity). No credential material is stored anywhere. The strongest isolation.
- **Keyless (OIDC)** — Your AWS account, Google Cloud project, or Azure subscription trusts Planton's identity issuer. For each operation, Planton signs a short-lived token for one connection, and your cloud exchanges it for short-lived credentials. No credential is stored in Planton, and no Planton cloud account is part of the trust. See [What Your Cloud Trusts](/docs/security/what-your-cloud-trusts).

For full details, see [Runner Security Model](/docs/runner/security-model), [Connections: Cloud Providers](/docs/connections/cloud-providers), and [Keyless Cloud Connections](/docs/connections/keyless-cloud-connections).

## Encryption

### Data at Rest

Secrets stored through Planton's [Secrets Manager](/docs/secrets) live in the secret backend you choose — your AWS Secrets Manager, GCP Secret Manager, Azure Key Vault, HashiCorp Vault, or OpenBAO — as native values protected by that store's own encryption, IAM, and audit. Planton's own database holds metadata, never values, so a full copy of it yields no secrets.

Because the values are native citizens of your store, your encryption controls apply directly: customer-managed keys are your provider's own setting (for example CMEK on GCP Secret Manager or a KMS key on AWS Secrets Manager), not a platform layer between you and your data.

The one deliberate exception is a single-machine local instance's built-in store, where the instance's own database is the storage medium: those values are envelope-encrypted with an AES-256 key held in the machine's OS keychain, outside the database.

For full details, see [Where Secrets Live](/docs/secrets/where-secrets-live) and [Secret Backends](/docs/secrets/backends).

### Data in Transit

All communication between the Runner and Planton's control plane is protected by mutual TLS (mTLS). Both sides present certificates and verify each other's identity. The Runner's certificate is bound to its organization, meaning the control plane cryptographically validates that a Runner belongs to the organization it claims before accepting any connection.

All web console and API traffic uses standard TLS encryption.

## Authentication

Planton supports three authentication methods for users and automation:

**Interactive login** — Users authenticate through an OAuth identity provider using the PKCE flow (no client secrets). The web console handles this automatically. The CLI initiates a browser-based login:

```bash
planton auth login
```

**API keys** — For CI/CD pipelines, scripts, and automation that cannot use a browser. API keys are scoped to the user who created them and carry that user's permissions. Keys are stored as hashes — Planton never persists the plaintext key after creation. Keys support expiration dates and track last-used timestamps.

```bash
planton iam apikey new --name "ci-pipeline"
```

**Service accounts** — For machine-to-machine authentication. Service accounts are dedicated machine identities with their own API keys, purpose-built for [Runners](/docs/runner) and other automated systems. When a runner enrolls with a runner token, Planton auto-provisions a service account and mints an API key — the runner uses it to authenticate with the control plane and resolve secrets at runtime.

```bash
planton sa create --org acme --name deploy-runner
```

For full details on roles and permissions, see [Authentication and Authorization](/docs/security/authentication-and-authorization).

## Authorization

Planton uses [OpenFGA](https://openfga.dev/) — an open-source, fine-grained authorization engine — for all access control decisions. This is not a simple role list. It is a relationship-based authorization model where permissions flow through the resource hierarchy.

The key properties:

- **Hierarchical inheritance** — Grant a user admin access to an organization, and they automatically have admin access to every environment and resource within it. No need to assign permissions at each level individually.
- **Resource-scoped roles** — Roles are specific to resource types. An "admin" on an organization is different from an "admin" on a service. Five standard roles apply across most resources: owner, admin, IAM admin, viewer, and member.
- **Team-based access** — Permissions granted to a team flow to all team members, including nested teams. Add someone to the Platform Engineering team, and they inherit every permission that team holds.
- **Environment authorization for credentials** — Connections must be explicitly authorized for each environment before they can be used there. A production AWS credential cannot be accidentally used in development.

For the full authorization model, see [Authentication and Authorization](/docs/security/authentication-and-authorization).

## Audit Trails

Every resource change in Planton is captured as an immutable version record. The audit system records:

- **Full state snapshots** — The complete YAML representation of the resource before and after the change
- **Unified diffs** — Git-style diffs showing exactly what changed
- **Who and when** — The identity account that made the change and the timestamp
- **Event type** — Whether the resource was created, updated, deleted, or restored
- **Deployment linkage** — For infrastructure changes, the version links to the Infra Job that executed the deployment, including whether it succeeded or failed

Version records are append-only and immutable. They form a linked chain — you can walk backward through any resource's complete history.

The web console provides a versions list, full version detail view, and side-by-side diff visualization. Access to audit data is authorization-gated — you can only view versions for resources you have permission to see.

For full details, see [Audit Trails](/docs/security/audit-trails).

## Network Security

The Runner's network model eliminates the need for inbound firewall rules, VPN tunnels, or network peering:

- **Outbound-only connectivity** — The Runner initiates all connections. It connects outbound to Planton's control plane over HTTPS (port 443). No inbound ports need to be opened.
- **No VPN required** — The secure tunnel is established over standard HTTPS, which passes through corporate firewalls and proxies without special configuration.
- **mTLS authentication** — Both the Runner and the control plane verify each other's identity with every connection. Man-in-the-middle attacks require compromising the CA private key.
- **Organization-scoped identity** — Each Runner's cryptographic identity is bound to its organization. A Runner registered to Organization A cannot receive requests intended for Organization B.

## Platform Security Summary

| Layer | Mechanism | Details |
|-------|-----------|---------|
| **Credentials** | Architectural isolation via Runner | [Runner Security Model](/docs/runner/security-model) |
| **Keyless cloud trust** | Your cloud pins Planton's issuer, the audience, and one connection's subject; no stored credential | [What Your Cloud Trusts](/docs/security/what-your-cloud-trusts) |
| **Secrets** | Provider-native storage in your own vault; read-audited; never stored in the platform's database | [Where Secrets Live](/docs/secrets/where-secrets-live) |
| **Authentication** | OAuth + PKCE (interactive), API keys (automation), service accounts (machines) | [Authentication and Authorization](/docs/security/authentication-and-authorization) |
| **Authorization** | OpenFGA relationship-based access control | [Authentication and Authorization](/docs/security/authentication-and-authorization) |
| **Audit** | Immutable version records with diffs | [Audit Trails](/docs/security/audit-trails) |
| **Network** | Outbound-only mTLS tunnel | [Runner](/docs/runner) |
| **Credential scoping** | Environment authorization + defaults | [Connections](/docs/connections) |
| **Team management** | Hierarchical teams with inherited permissions | [Teams and Access](/docs/teams-and-access) |

## Related Documentation

- [Authentication and Authorization](/docs/security/authentication-and-authorization) — How users authenticate and how permissions are evaluated
- [What Your Cloud Trusts](/docs/security/what-your-cloud-trusts) — The exact values a keyless cloud connection pins, and how to revoke it
- [Audit Trails](/docs/security/audit-trails) — Immutable change tracking and version history
- [Runner Security Model](/docs/runner/security-model) — Credential isolation, mTLS, and authentication modes
- [Secrets](/docs/secrets) — Provider-native storage, live version history, and secret backends
- [Connections](/docs/connections) — Credential management and environment authorization
- [Teams and Access](/docs/teams-and-access) — Membership, teams, and role-based access
