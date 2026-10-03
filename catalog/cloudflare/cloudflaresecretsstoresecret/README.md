# Cloudflare Secrets Store Secret

## Overview

`CloudflareSecretsStoreSecret` manages one secret inside the account Secrets Store. The secret's VALUE is write-only at Cloudflare -- the API never returns it -- so a value change is proven by the deployment, never by reading back. Consumers (Worker secrets-store bindings, AI Gateway authentication) reference the secret by store and name; rotating the value here rotates it for every consumer without touching them.

The `scopes` list is walled to Cloudflare's canonical alphabetical order by this spec: the API always returns scopes sorted, the provider models the list as ordered, and any other configured order would show as permanent plan drift (a defect documented in the provider's own tests). The wall makes that drift unreachable.

## Key Features

- **Write-only value** -- the secret never leaves Cloudflare once written; provide it as a managed-secret reference
- **Rotate in one place** -- consumers reference the name, not the value
- **Scoped access** -- `workers`, `ai_gateway`, `dex`, `access` decide which surfaces may read it
- **Up to 64 KiB** per value

## Use Cases

**Ideal for:**

- Provider API keys Workers call out with (read via secrets-store bindings)
- The keys AI Gateway presents upstream (BYO keys via the gateway's store link)
- Any credential shared by several Workers that must rotate centrally

**Not ideal for:**

- Per-Worker secrets that never need sharing -- Worker secret bindings cover those
- Certificates and key material consumed by TLS surfaces -- those have their own upload kinds

## API Specification

### Required Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `account_id` | string | Yes | The Cloudflare account (32-hex). |
| `store_id` | StringValueOrRef | Yes | The Secrets Store holding this secret (references a `CloudflareSecretsStore`'s `store_id` output). |
| `name` | string | Yes | The secret's name -- what consumers reference. Create-only: renaming replaces the secret under a new ID. |
| `value` | StringValueOrRef (sensitive) | Yes | The secret value (<= 64 KiB). WRITE-ONLY: never returned, never drift-detected. Provide a managed-secret reference. |
| `scopes[]` | list | Yes | The surfaces allowed to read it: `access`, `ai_gateway`, `dex`, `workers` -- listed ALPHABETICALLY (the spec enforces the canonical order the API returns). |

### Optional Fields

| Field | Type | Description |
|-------|------|-------------|
| `comment` | string | A free-form note (shown in the dashboard). |

### Outputs

| Field | Description |
|-------|-------------|
| `secret_id` | The secret's ID within its store |
| `store_id` | The store holding the secret (echoed for consumers needing the pair) |

## Example Manifest

```yaml
apiVersion: cloudflare.planton.dev/v1alpha1
kind: CloudflareSecretsStoreSecret
metadata:
  name: openai-api-key
spec:
  account_id: "0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d"
  store_id:
    valueFrom:
      kind: CloudflareSecretsStore
      name: account-secrets
      fieldPath: status.outputs.store_id
  name: openai-api-key
  value:
    value: ${secrets-group/prod-ai/openai-key}
  scopes:
    - ai_gateway
    - workers
```

## Destroy Semantics

Destroy is a real delete; consumers referencing the name start failing reads at runtime, not at apply time. Re-point or retire consumers first.

## Related Resources

- **CloudflareSecretsStore** -- the vault this secret lives in
- **CloudflareWorker** -- secrets-store bindings read this secret at runtime
- **CloudflareAiGateway** -- BYO-keys authentication reads from the same store

## Further Reading

For operational judgment -- the write-only contract, scope ordering, rotation flow -- see GUIDE.md.

## References

- [Cloudflare Secrets Store](https://developers.cloudflare.com/secrets-store/)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
