---
title: "Where Secrets Live"
sidebar_title: "Where Secrets Live"
description: "Provider-native storage under readable, collision-proof names — your values in your store, labeled, deep-linked, and never held hostage"
icon: shield
order: 40
tags:
  - Secrets
  - Storage
  - Security
---

# Where Secrets Live

Most secrets managers store your values in their own format, in their own database. Planton does the opposite: a secret stored through Planton into a real backend lives **in your store, as itself**. This page explains exactly what lands in your vault, under what name, and why that model means Planton can never hold your secrets hostage.

## Your Store Holds the Real Value

When a secret's backend is AWS Secrets Manager, GCP Secret Manager, Azure Key Vault, HashiCorp Vault, or OpenBAO:

- A **text secret** is stored byte-for-byte. What you wrote is exactly what `gcloud secrets versions access`, the AWS console, or `vault kv get` returns.
- A **key-value secret** is stored as one canonical JSON object (Vault-family stores hold native KV fields). An External Secrets Operator extractor or a Cloud Run `secretKeyRef` reads each key directly.

There is no Planton encryption wrapper around your values in your store — the provider's own encryption, IAM, and audit apply, because the value is a native citizen of the provider. Anything that can read your store can read the value with no Planton dependency, which is precisely the point: **the day you stop using Planton, every secret is already where you need it, in the format your tools expect.**

The one deliberate exception: a single-machine local instance's built-in backend stores values envelope-encrypted in the instance's own database, with the encryption key held in the operating system's keychain — because there, Planton's database IS the store.

## Readable, Collision-Proof Names

Every managed secret has a **logical path** that encodes its full identity:

```
{prefix}/{org}/org/{slug}                  # organization-scoped
{prefix}/{org}/env/{environment}/{slug}    # environment-scoped
```

The prefix defaults to `planton` and is configurable per backend. The path is rendered into each provider's naming rules:

| Provider | Rendering | Example |
|----------|-----------|---------|
| AWS Secrets Manager | verbatim (slashes make a hierarchy in the AWS console) | `planton/acme/env/prod/db-password` |
| GCP Secret Manager | segments joined with `_` | `planton_acme_env_prod_db-password` |
| Azure Key Vault | segments joined with `--` | `planton--acme--env--prod--db-password` |
| Vault / OpenBAO | native KV path | `planton/acme/env/prod/db-password` |

Because the environment is part of the address, same-named secrets in two environments are two remote secrets — structurally, not by convention. Anyone reading your store can tell at a glance what a secret is, which organization and environment it belongs to, and that Planton manages it.

## Provenance Labels

Every managed remote secret carries labels (GCP), tags (AWS, Azure), or custom metadata (Vault-family):

| Label | Value |
|-------|-------|
| `managed-by` | `planton` |
| `org` | the organization slug |
| `scope` | `organization` or `environment` |
| `env` | the environment slug (environment scope only) |
| `secret-id` | the Planton record id |

Filter your own store by `managed-by=planton` to see exactly what Planton manages — with your own tools, no Planton API involved.

## Remote Identity on Every Secret

The moment a secret is created, its record captures the **remote identity**: the real rendered name, the logical path, and a deep link to the secret in your provider's own console. The console detail page, `planton secret describe`, and agent tools all surface it. Share the deep link with a teammate who has provider IAM access and they can read or update the value right in the provider console — out-of-band edits are first-class (see [Version History](/docs/secrets/versions)).

## Secrets a Resource Generates

Some resources create a credential when they deploy: an Auth0 application's client secret, an AWS IAM user's secret access key, a container registry's admin password, a database's connection string. Planton stores each one in your organization's secret store the moment the deploy finishes, before the value leaves your runner, and the resource's outputs carry a reference instead of the value:

```
$secret/@prod/auth0-client-outputs-checkout/client_secret
```

- **One key-value secret per resource**, named `<kind>-outputs-<resource>`, in the resource's environment, with one key per secret output. It is labeled as managed by Planton for that resource, and its page links back to it.
- **A re-deploy adds a version only when a value changed**, so the secret's history shows real rotations, not every deploy.
- **Another resource reads it with `valueFrom`**, exactly as it reads any output. The reference travels, and the runner resolves it at deploy, so the value reaches the workload and nowhere else. Because the value is a secret, a `valueFrom` of a secret output belongs in a sensitive field (a workload's `env.secrets`, for example); written into a plain field it is refused, naming the field.
- **Deleting the resource deletes its secret.** While another resource still reads it, the delete is refused with the reader named; destroy or re-point the reader first, or delete with the force flag.

Which outputs are secrets is declared in each kind's schema: the reference pages mark them `(sensitive)` in the Outputs table.

## Signed In Through Your Connection

To read or write a secret in your AWS Secrets Manager, Google Cloud Secret Manager, or Azure Key Vault, Planton has to sign in to that store. The recommended way is through one of your organization's cloud connections — an AWS connection for Secrets Manager, a Google Cloud connection for Secret Manager, an Azure connection for Key Vault:

```bash
planton secret backend create team-secrets --type aws \
  --aws-region us-east-1 \
  --aws-connection prod-aws
```

With a [keyless connection](/docs/connections/keyless-cloud-connections), **nothing about the backend is stored anywhere**. Every time the backend needs fresh credentials, Planton mints a short-lived token for that connection and your cloud exchanges it against the trust you wrote once in your own account:

- **AWS** — the backend assumes the connection's role, in a session named `planton-secret-backend-<backend>`, so your CloudTrail shows exactly which backend read a secret.
- **Google Cloud** — the token is exchanged at Google's token service and the connection's service account is impersonated.
- **Azure** — the token is the client assertion of the connection's app registration.

Each token lives minutes and is never written down. A stored-key connection works too: the backend reads the key from the connection's own secret, and the key exists only there. Connections that sign in on a runner, through a Vault broker, or as a person through the browser cannot serve a secret backend, because Planton's control plane reads and writes your secrets itself, on your organization's behalf.

The connection's identity needs permission to manage secrets. Choose the **Secrets Manager** capability on an AWS connection or the **Secret Manager** capability on a Google Cloud connection; on Azure, give the app registration a role that manages secrets on the vault, such as Key Vault Secrets Officer.

Three guarantees come with it:

- **One hop, never two.** A connection's own secrets — a stored key, or an Azure app registration's client ID — must live in a backend that does not itself sign in through a connection. Planton refuses the link where you make it, and checks again at every read, so reaching a secret never needs a second sign-in.
- **The connection cannot disappear under your secrets.** Its page lists the **Backends Signing In** through it (`planton connection get aws prod-aws`), and deleting it is refused while any does, naming each backend.
- **Switching keeps every secret where it is.** A backend can move from a stored key to a connection in place; secret names depend only on the backend's type and prefix, and the stored key is deleted once the change is saved.

Prove the sign-in end to end, before or after you create the backend, with `planton secret backend verify team-secrets` (or `-f` with a manifest). When your cloud refuses the keyless exchange, the verdict names the exact subject and issuer your trust must admit. On Planton's hosted service, a backend cannot sign in as Planton's own identity: choose a connection or a stored key. See [What Your Cloud Trusts](/docs/security/what-your-cloud-trusts) for the trust itself.

## What Planton's Own Database Holds

Metadata only: the secret's name, scope, backend binding, remote identity, version authorship records, and the read-audit trail. A backend that signs in through a keyless connection adds nothing to that list: no key, no token. **Never the value**, including for secrets a resource generates: its outputs, its deploy records and the saved inputs of every resource that reads it hold only the reference. A full copy of Planton's database yields no secret values for provider-backed secrets — they are in your store, under your IAM, and nowhere else.

## Related Documentation

- [Secret Backends](/docs/secrets/backends) — Configuring and verifying where secrets are stored
- [Version History](/docs/secrets/versions) — The live provider-backed timeline
- [Consuming Secrets Anywhere](/docs/secrets/integrations) — Ready-to-paste snippets for your runtime
- [Security Overview](/docs/security) — Platform-wide security architecture
