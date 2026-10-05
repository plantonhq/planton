# Cloudflare Secrets Store

## Overview

`CloudflareSecretsStore` creates the account-level Secrets Store: the vault that holds secrets consumed by Workers bindings, AI Gateway authentication, and other Cloudflare surfaces. The store itself is just a named container -- the secrets inside it are managed by `CloudflareSecretsStoreSecret` resources referencing this store.

Two hard provider facts shape this resource: EVERY field is create-only (a name change replaces the store, destroying every secret inside it), and Cloudflare currently allows ONE store per account -- creating a second fails at the API.

## Key Features

- **Account-wide vault** -- one store, referenced by every secret, Worker binding, and AI Gateway in the account
- **Centralized rotation** -- rotate a secret's value once; every consumer picks it up without redeploying
- **Free** -- no plan gate on the store itself

## Use Cases

**Ideal for:**

- The single account vault behind Worker `secrets_store_secrets` bindings
- Holding provider API keys the AI Gateway presents (Bring Your Own Keys)

**Not ideal for:**

- Per-Worker secrets that never need sharing -- Worker secret bindings cover those
- The secrets themselves -- those are `CloudflareSecretsStoreSecret` resources

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `account_id` | string | Yes | The Cloudflare account (32-hex). |
| `name` | string | Yes | The store's name. Create-only: renaming replaces the store AND every secret it holds. |

### Outputs

| Field | Description |
|-------|-------------|
| `store_id` | The store's ID -- what store secrets, Worker bindings, and AI Gateway authentication reference |

## Example Manifest

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareSecretsStore
metadata:
  name: account-secrets
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  name: account-secrets
```

## Destroy Semantics

Destroy is a real delete -- and it takes every secret in the store with it. Treat the store as permanent infrastructure.

## The one-store limit

Cloudflare allows one store per account. If your account already has one (dashboard-created), adopt it by import instead of creating: the import identity is `{account_id}/{store_id}`.

## Related Resources

- **CloudflareSecretsStoreSecret** -- the secrets inside this store
- **CloudflareWorker** -- secrets-store bindings read from this store
- **CloudflareAiGateway** -- BYO-keys authentication reads from this store

## Further Reading

For operational judgment -- why the store is permanent infrastructure, the one-store limit, adopt-by-import -- see GUIDE.md.

## References

- [Cloudflare Secrets Store](https://developers.cloudflare.com/secrets-store/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
