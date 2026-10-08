---
title: "Secret Backends"
sidebar_title: "Backends"
description: "Choose where secrets live — from a zero-config default to your own AWS, GCP, Azure, Vault, or OpenBAO — signed in through your cloud connection, and proven by a verify-before-save health probe"
icon: database
order: 30
tags:
  - Secrets
  - Backends
  - Storage
---

# Secret Backends

Every secret in Planton is stored in a **secret backend** — the store that holds its value. The backend determines where your secrets physically live, and with a real provider backend they live there [as native citizens](/docs/secrets/where-secrets-live): real values, readable names, your IAM, your audit.

## The Built-In Default

Every organization starts with a working default and zero setup:

- **Hosted organizations** get a Planton-operated [OpenBAO](https://openbao.org/) vault. Values are stored as native KV entries with per-organization isolation.
- **A local desktop instance** gets a built-in local store: values envelope-encrypted in the instance's own database, with the encryption key minted into the operating system's keychain — the one backend where Planton's database is the store, because there is nothing else on a laptop.

For teams getting started, the default is the right choice. Connect your own store when you want secrets in your own account.

## Bring Your Own Backend

| Backend | When to Use |
|---------|-------------|
| **AWS Secrets Manager** | Your organization standardizes on AWS, or compliance requires secrets in your AWS account |
| **GCP Secret Manager** | Your organization standardizes on GCP, or you need secrets co-located with GCP workloads |
| **Azure Key Vault** | Your organization standardizes on Azure, or policy requires secrets in your tenant |
| **HashiCorp Vault** | You already operate a Vault cluster |
| **Self-hosted OpenBAO** | The same technology as the hosted default, operated by you |

### How a Cloud Backend Signs In

**Cloud Connection** (recommended) — the backend signs in through one of your organization's cloud connections (`--aws-connection`, `--gcp-connection`, `--azure-connection`). With a keyless connection nothing about the backend is stored; with a stored-key connection the key stays in the connection's own secret. A connection's own secrets must live in a backend that does not itself sign in through a connection, and a connection a backend signs in through cannot be deleted. See [Signed In Through Your Connection](/docs/secrets/where-secrets-live#signed-in-through-your-connection).

**Inline Credentials** — provide static credentials for the store (an AWS access key pair, a GCP service account key, an Azure service principal). They are moved into the platform's internal credential store on save and masked in every response — they never sit in the backend record.

**Ambient Identity** — provide no credentials at all. The deployment authenticates as itself: the cluster's workload identity on a self-hosted install, or your own signed-in developer CLI on a local desktop instance. On a machine signed into several identities, per-backend handles pin one: an AWS profile, a gcloud configuration, or an az subscription — authenticated through the host's own CLI. Planton's hosted service refuses it, because its identity is Planton's own.

The vault-family backends sign in with a token stored with the backend.

### Coordinates and Naming

- Region, project, vault URL, and tenant coordinates accept a literal value **or a reference to an org-level variable**, so a coordinate maintained in one place serves many backends.
- An optional **name prefix** (default `planton`) leads every remote name the backend renders; it is immutable after creation because it is part of every stored secret's address.
- The backend type is likewise immutable — a backend is a stable address, never a moving target. How it signs in can change in place: every secret stays where it is, and a stored key is deleted once the backend signs in another way.

### Verify Before You Save

Backend creation ends with a **verification probe** you can run on the not-yet-saved backend: credential resolution, store reachability, and a real write/read/delete round-trip under the backend's own prefix. The verdict is per-check data, not a boolean — and when an ambient login is broken, the failing check names the exact remedy (`gcloud auth application-default login`, `aws sso login`, `az login`, or the specific config-file problem). The same probe is available any time:

```bash
planton secret backend verify my-backend    # exit 1 on unhealthy — scriptable
```

### The Default Backend

Each organization has a default backend, used when a secret does not name one explicitly. Change it to any configured backend. A secret's backend binding is immutable after creation — to move a secret, create it on the target backend and copy the value.

## Choosing the Right Backend

| Scenario | Backend |
|----------|---------|
| **Getting started** | Built-in default |
| **Secrets must stay in your account** | Your cloud's secret manager |
| **Existing Vault estate** | HashiCorp Vault or OpenBAO |
| **Zero-dependency laptop development** | The local instance's built-in store |

## Related Documentation

- [Where Secrets Live](/docs/secrets/where-secrets-live) — What lands in your store, and under what name
- [Managing Secrets](/docs/secrets/managing-secrets) — Creating and referencing secrets
- [Variables](/docs/variables) — Non-sensitive configuration management
- [Connections: Cloud Providers](/docs/connections/cloud-providers) — Connecting the cloud accounts your backends live in
