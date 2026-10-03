# Cloudflare Secrets Store Secret

Deploys one secret inside the account-level Cloudflare Secrets Store, readable by exactly the surfaces its scopes allow — Workers, AI Gateway, DEX, and Access. The value is write-only at Cloudflare: never returned, never drift-detected, so your managed secret store is the real system of record. Consumers reference the secret by store and name, which makes rotation the whole point of this kind — change the value in one place and every Worker binding and gateway picks it up on the next read, no consumer redeploys.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Store Secret** -- one `cloudflare_secrets_store_secret` inside the referenced store, with the value marked sensitive in state and the declared scopes controlling which Cloudflare surfaces may read it

## Before You Deploy

### Planton Setup

- **Cloudflare Provider Connection** -- an active connection in the Connect module with a Cloudflare API token that has Secrets Store Edit on the account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### Cloudflare Account

- **The account Secrets Store** -- a Cloudflare Secrets Store (one per account) must exist; wire `storeId` to it by reference or paste the store ID directly.
- **The secret value in a managed secret** -- `value` takes a reference the platform resolves just-in-time at deploy. Since Cloudflare never returns the value, whatever holds the source copy is the only recoverable record of it.

## Deploy

### Console

Open the deployment store, find **Cloudflare Secrets Store Secret**, and click **Deploy**. The creation wizard walks you through the account and store, the secret's name, the value (as a managed-secret reference), and the scopes. Start from the **Shared provider API key** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareSecretsStoreSecret
metadata:
  name: openai-api-key
  org: acme-corp
  env: prod
spec:
  accountId: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  storeId:
    value: "7b0a3d5c1e9f42c68d1a2b3c4d5e6f70"
  name: openai-api-key
  value:
    value: "${secrets-group/prod-ai/openai-key}"
  scopes:
    - ai_gateway
    - workers
```

```shell
planton apply -f secret.yaml
```

This creates a secret named `openai-api-key` in the store, readable by AI Gateway and Workers only, with the value resolved from a managed secret at deploy time. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the store to a Secrets Store managed in the same InfraPipeline:

```yaml
spec:
  storeId:
    valueFrom:
      kind: CloudflareSecretsStore
      name: account-secrets
      fieldPath: status.outputs.store_id
```

The InfraPipeline resolves the dependency graph, creates the store first, then provisions the secret inside it.

## Key Configuration

These are the most important decisions when configuring a store secret. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The value is write-only — your secret store is the system of record.** Cloudflare never returns it: not on read, not on import, not on refresh. Drift on the value is undetectable (a dashboard edit behind IaC's back wins silently until the next apply overwrites it), an import lands with an empty value that must be re-asserted from configuration, and losing the source value means the secret can only be replaced, never recovered.

**The name is the contract with consumers.** Rotate by changing `value` — an in-place update every consumer picks up without redeploying. Renaming, by contrast, replaces the secret under a new ID and breaks every consumer referencing the old name. Treat `name` like an API contract; `accountId` and `storeId` are create-only too.

**Scopes are alphabetical, or they drift forever.** The API returns scopes sorted no matter how you send them, and the provider models the list as ordered — an out-of-order config re-plans a change on every run, forever. The spec walls the order at validation time: `[access, ai_gateway, dex, workers]` is the full canonical sequence, and you list your subset in that order. If validation rejects your list, sort it — the wall exists to keep your plans clean.

**Scope minimally.** `scopes` is an access-control list, not a formality: a secret scoped `[workers]` cannot be read by AI Gateway, and vice versa. Grant the surfaces that actually read it — widening later is a plain in-place update.

**Destroy is a real delete, and consumers fail at runtime.** A Worker binding or gateway referencing the deleted name fails its reads when they happen, not at apply time. Re-point or retire consumers first.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **CloudflareSecretsStore** | `storeId` | `status.outputs.store_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `secret_id` | The secret's ID within its store | Worker secrets-store bindings that reference the store/secret pair |
| `store_id` | The store holding the secret | Paired with `secret_id` by consumers that need both halves of the reference |

The secret's value is deliberately absent from the outputs — Cloudflare never returns it.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Shared provider API key** -- An AI provider's key stored once and readable by both Workers and the AI Gateway — the classic Bring-Your-Own-Keys shape. The value arrives as a managed-secret reference and rotates in one place for every consumer. Start from the **Shared provider API key** preset.

**Worker-only credential** -- A signing key or upstream token scoped `[workers]`, so nothing outside the Workers runtime can read it. The narrowest scope that works is the right one.

**Rotation without redeploys** -- Update the managed secret, re-apply, and every consumer reads the new value on its next access. The manifest never changes shape — only the resolved value does.

## Works With

- [**Cloudflare Secrets Store**](/infra-catalog/cloudflare-secrets-store) -- the one-per-account vault this secret lives in; wire `storeId` via ValueFromRef.
- [**Cloudflare Worker**](/infra-catalog/cloudflare-worker) -- secrets-store bindings that read the secret at runtime.
- [**Cloudflare AI Gateway**](/infra-catalog/cloudflare-ai-gateway) -- BYO-keys authentication backed by secrets like this one.
