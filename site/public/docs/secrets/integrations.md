---
title: "Consuming Secrets Anywhere"
sidebar_title: "Integrations"
description: "Ready-to-paste snippets for External Secrets Operator, Kubernetes CSI, Cloud Run, ECS, provider CLIs, and Planton manifests — computed from the secret's real remote identity"
icon: plug
order: 50
tags:
  - Secrets
  - Integrations
  - Kubernetes
---

# Consuming Secrets Anywhere

Because a Planton-managed secret lives provider-native in your own store under a [stable, readable name](/docs/secrets/where-secrets-live), anything that can read your store can consume it — with or without Planton in the path. The platform makes this concrete: every secret offers a catalog of **ready-to-paste integration snippets**, computed on the server from the secret's real remote identity and its backend's actual coordinates, so every surface hands you the same correct text.

Find them on the secret's detail page (**Use This Secret**), on the creation success screen, from the CLI, or through agent tools:

```bash
# The full catalog, grouped
planton secret snippet db-password

# One target, pipe-safe
planton secret snippet db-password --target eso -o plain > external-secret.yaml
```

## Through Planton

- **Planton manifest reference** — the `$secret/...` reference for service and infrastructure manifests, resolved just-in-time on the Runner at deployment.
- **Connection field reference** — the typed reference for connection credential fields (organization-scoped secrets).
- **CLI one-liner** — `planton secret get <slug> --reveal -o plain` for scripts and local shells. It exits 3 with an empty stdout when the secret does not exist, and 1 when the instance could not be asked.

## Straight From Your Store

These read your store directly — no Planton dependency at runtime:

- **External Secrets Operator** — a `SecretStore` + `ExternalSecret` pair for your provider. Key-value secrets extract every key automatically (the payoff of canonical JSON storage); text secrets map the single value.
- **Kubernetes CSI Secrets Store** — the provider class and object definition for mounting the secret into pods.
- **Cloud Run** — the `--update-secrets` binding for the secret's real GCP name.
- **ECS** — the task-definition `valueFrom` entry (with an honest note about cross-account ARN forms).
- **Provider CLI** — the exact `gcloud` / `aws` / `az` / `vault` read command for the secret's real name.

Snippets are honest by construction: provider-native targets appear only for stores your own tools can reach (never for a local instance's built-in store), and `<key>` placeholders appear exactly where a key name cannot be known without reading the value — which snippet rendering never does. Snippets render even when the backend's credentials are broken, because rendering never contacts the provider.

## Related Documentation

- [Where Secrets Live](/docs/secrets/where-secrets-live) — The naming and storage model these snippets rely on
- [Managing Secrets](/docs/secrets/managing-secrets) — The `$secret/` reference grammar
- [CI/CD: What is a Service](/docs/ci-cd/what-is-a-service) — Referencing secrets from service manifests
