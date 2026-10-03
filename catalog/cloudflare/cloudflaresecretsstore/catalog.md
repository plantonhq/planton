# Cloudflare Secrets Store

Deploys the account-level Cloudflare Secrets Store: the vault that Worker secrets-store bindings, AI Gateway authentication, and other Cloudflare surfaces consume secrets from. The store itself is a named container — the secrets inside it are separate Cloudflare Secrets Store Secret resources referencing this store. Two hard provider facts shape everything about this kind: Cloudflare allows exactly one store per account, and every field is create-only, so renaming the store replaces it and destroys every secret it holds. It is permanent infrastructure wearing a two-field spec.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Secrets Store** -- one `cloudflare_secrets_store` in the account: the singleton vault every store secret, Worker binding, and AI Gateway key reference

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has Secrets Store Edit on the account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **No existing store** -- Cloudflare allows one Secrets Store per account, and a create against an account that already has one (dashboard-created, or another team's) fails at the API. Adopt an existing store by import instead of creating a second one, and decide early which manifest owns it — everything else references it.

## Deploy

### Console

Open the deployment store, find **Cloudflare Secrets Store**, and click **Deploy**. The creation wizard walks you through the account and the store's name — the whole spec, by design. Start from the **Account vault** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareSecretsStore
metadata:
  name: account-secrets
  org: acme-corp
  env: prod
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: account-secrets
```

```shell
planton apply -f secrets-store.yaml
```

This creates the account's single Secrets Store, empty and ready for Cloudflare Secrets Store Secret resources to fill. An Infra Job tracks the provisioning in real time.

## Key Configuration

These are the most important decisions when configuring a Secrets Store. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**One per account: create once, adopt otherwise.** The store is a singleton. If the account already has one, the create fails honestly at the API — the correct move is adopt-by-import, never a second create. In an organization, the store belongs to exactly one manifest; every other team references its `store_id`.

**Renaming is destruction.** Both fields force replacement (the provider's update path is an empty stub), and replacing the store destroys every secret inside it — every Worker binding and AI Gateway key breaks at once. Pick a boring, permanent name like `account-secrets` and never touch it. If you truly must rename: re-create every secret under the new store and re-point every consumer before the old store goes away.

**Destroy last, if ever.** The store's blast radius is every secret it holds. In teardown flows, destroy secrets and their consumers first and leave the store for the very end — or leave it standing; it costs nothing.

## Outputs and Dependencies

### What This Kind Consumes

This kind has no foreign key dependencies. The account ID travels as a literal value, and the store is the root of the secrets dependency chain — everything else references it.

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `store_id` | The store's ID | Referenced by Cloudflare Secrets Store Secret resources, Worker secrets-store bindings, and AI Gateway BYO-keys authentication |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Account vault** -- The account's single store under an unremarkable, stable name, created the first time any Worker binding or AI Gateway in the account needs a shared secret. Start from the **Account vault** preset.

**Adopt the dashboard-created store** -- When someone already clicked the store into existence, import it rather than fighting the singleton limit; from then on the manifest is the system of record and secrets are layered in as separate resources.

**Store now, secrets on their own cadence** -- Deploy the store as permanent shared infrastructure and let each team manage its own Cloudflare Secrets Store Secret resources against it — each secret rotates independently without ever touching the container.

## Works With

- [**Cloudflare Secrets Store Secret**](/infra-catalog/cloudflare-secrets-store-secret) -- the secrets inside this store, each a separate resource wired to `store_id`.
- [**Cloudflare Worker**](/infra-catalog/cloudflare-worker) -- secrets-store bindings that read secrets from this vault at runtime.
- [**Cloudflare AI Gateway**](/infra-catalog/cloudflare-ai-gateway) -- BYO-keys authentication backed by secrets in this store.
