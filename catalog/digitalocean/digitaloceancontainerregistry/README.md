# DigitalOcean Container Registry

A DigitalOcean Container Registry (DOCR) described once in a Planton manifest: the account's private Docker registry with its subscription tier and region, plus optionally minted Docker credentials (read-only or write, with a controlled lifetime) exported as an output. A DigitalOcean account holds exactly ONE registry, and registry names are globally unique across all DigitalOcean accounts.

## What this kind models

The spec maps onto DigitalOcean's `digitalocean_container_registry` plus the optional `digitalocean_container_registry_docker_credentials`:

| Spec field | What it controls |
|---|---|
| `name` | The registry's name — globally unique across ALL DigitalOcean accounts, create-only, and the resource identity |
| `subscriptionTier` | `starter` (free) / `basic` / `professional` — the one attribute that can change after creation |
| `region` | Where registry data is stored; omit to let DigitalOcean choose (create-only either way) |
| `dockerCredentials` | When set, mints a Docker credential: `write` (push access, default read-only) and `expirySeconds` (default: the API maximum, ~50 years) |

## Quick start

The smallest real registry:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanContainerRegistry
metadata:
  name: my-registry
spec:
  name: acme-registry
  subscriptionTier: starter
```

A production registry with push credentials that expire:

```yaml
spec:
  name: acme-registry-prod
  subscriptionTier: basic
  region: nyc3
  dockerCredentials:
    write: true
    expirySeconds: 2592000 # 30 days
```

## Behavior worth knowing

- **One registry per account** — a second manifest fails at create; the registry is an account-level singleton.
- **Names are globally unique** — across every DigitalOcean account, like domain names; an "already exists" error may mean another account holds the name.
- **Credentials are minted only when asked** — no `dockerCredentials` block, no long-lived token. When set, the credential lands in the `docker_credentials` output (a secret) as a base64 Docker `config.json`.
- **Credential knobs are write-only** — the API never reports `write`/`expirySeconds` back, and the credentials resource cannot be imported at the current provider pin (its importer is defective; recorded in the import map).
- **The registry imports by name** — the `registry_name` output is the resource identity.

## Outputs

| Output | Meaning |
|---|---|
| `registry_name` | The registry's name — also its resource identifier (`status.outputs.registry_name`) |
| `server_url` | The registry host, always `registry.digitalocean.com` |
| `endpoint` | The full docker push/pull endpoint: `registry.digitalocean.com/<name>` |
| `region` | The region slug DigitalOcean reports (covers the DigitalOcean-chooses case) |
| `docker_credentials` | Base64 Docker `config.json` — a SECRET; empty when credentials are not configured |
| `credential_expiration_time` | RFC 3339 expiry of the minted credential; empty when not configured |

## See also

- `GUIDE.md` — operational judgment calls (the singleton constraint, credential lifetimes, tier changes)
- `presets/` — professional-tier starting point
- `v1alpha1/reference.md` — the generated field-by-field contract

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
