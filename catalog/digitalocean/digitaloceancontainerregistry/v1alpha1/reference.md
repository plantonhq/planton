# DigitalOceanContainerRegistry

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

DigitalOceanContainerRegistrySpec defines the configuration for a DigitalOcean Container
Registry (DOCR), modeling the provider's full surface: the registry itself plus the optional
Docker credentials DigitalOcean can mint for it. A DigitalOcean account can hold exactly ONE
container registry, and registry names share a global namespace across all DigitalOcean
accounts.

## Example

```yaml
# Example DigitalOceanContainerRegistry manifests. Deploy with:
#   planton apply -f manifest.yaml
#
# Document 1 -- the smallest real registry: starter (free) tier, DigitalOcean
# chooses the region. Remember: one registry per account, and names are
# globally unique across ALL DigitalOcean accounts.
#
# Document 2 -- a production-shaped registry: basic tier, pinned region, and
# minted docker credentials (push access, 30-day expiry) exported through the
# docker_credentials output.
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanContainerRegistry
metadata:
  name: example-docr-minimal
spec:
  name: acme-registry
  subscriptionTier: starter
---
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanContainerRegistry
metadata:
  name: example-docr-full
spec:
  name: acme-registry-prod
  subscriptionTier: basic
  region: nyc3
  dockerCredentials:
    write: true
    expirySeconds: 2592000
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.name` | `string` | yes |  |  |
| `spec.subscriptionTier` | `enum` | yes |  |  |
| `spec.region` | `enum` |  |  |  |
| `spec.dockerCredentials` | `DigitalOceanContainerRegistryDockerCredentials` |  |  |  |
| `spec.dockerCredentials.write` | `bool` |  |  |  |
| `spec.dockerCredentials.expirySeconds` | `int32` |  |  |  |

## Field Details

### spec.name

`string` · required

Registry name (globally unique across ALL DigitalOcean accounts, not just yours).
1-63 characters, lowercase letters, numbers, and hyphens; must start and end with an
alphanumeric. The name is the resource identity and cannot be changed after creation.

- rule: {"required":true,"string":{"minLen":"1","maxLen":"63","pattern":"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"}}

### spec.subscriptionTier

`enum` · required

Subscription tier slug (defines storage limits and pricing). The tier can be changed
after creation; everything else on the registry is create-only.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digitalocean_container_registry_tier_unspecified`
- `starter`
- `basic`
- `professional`

### spec.region

`enum`

(Optional) Region slug where registry data is stored (e.g. "nyc3", "sfo3").
When omitted, DigitalOcean chooses a region and the chosen slug is reported back.
The region cannot be changed after creation.

Allowed values (use exactly as shown):

- `digital_ocean_region_unspecified` -- 0: default / unspecified region
- `nyc3` -- new york 3
- `sfo3` -- san francisco 3
- `fra1` -- frankfurt 1
- `sgp1` -- singapore 1
- `lon1` -- london 1
- `tor1` -- toronto 1
- `blr1` -- bangalore 1
- `ams3` -- amsterdam 3
- `nyc1` -- new york 1
- `nyc2` -- new york 2
- `sfo2` -- san francisco 2
- `syd1` -- sydney 1
- `atl1` -- atlanta 1

### spec.dockerCredentials

`DigitalOceanContainerRegistryDockerCredentials`

(Optional) Docker credentials to mint for this registry. When set, both provisioners
create a credential (a base64-encoded Docker `config.json`) exported through the
`docker_credentials` output. When omitted, no credential is created -- the secure
default, since an unconfigured credential would otherwise live for ~50 years.

### spec.dockerCredentials.write

`bool`

Allow push (write) access. Defaults to false: a read-only pull credential.

### spec.dockerCredentials.expirySeconds

`int32` · optional (explicit presence)

(Optional) Credential lifetime in seconds. When unset, DigitalOcean uses the API maximum
(1576800000 seconds, roughly 50 years -- effectively non-expiring). Changing the lifetime
re-mints the credential in place.

- rule: {"int32":{"lte":1576800000,"gte":0}}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanContainerRegistry, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.registry_name` | `string` | The registry name (also the registry's resource identifier in DigitalOcean). |
| `status.outputs.server_url` | `string` | The registry host, always "registry.digitalocean.com". |
| `status.outputs.endpoint` | `string` | The full endpoint for docker push/pull, i.e. "registry.digitalocean.com/<registry_name>". |
| `status.outputs.region` | `string` | Region slug where the registry is hosted (reported by DigitalOcean, which also covers the case where the region was left unset and DigitalOcean chose one). |
| `status.outputs.docker_credentials` | `string` (sensitive) | Base64-encoded Docker `config.json` for this registry -- a SECRET. Populated only when the spec's docker_credentials block is set; empty otherwise. |
| `status.outputs.credential_expiration_time` | `string` | RFC 3339 timestamp at which the minted docker credentials expire. Populated only when the spec's docker_credentials block is set; empty otherwise. |

## See Also

- [Overview](../README.md)
