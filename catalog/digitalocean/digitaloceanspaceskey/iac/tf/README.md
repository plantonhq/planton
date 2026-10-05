# Terraform Module: DigitalOcean Spaces Key

Provisions a Spaces access-key pair with optional per-bucket grants -- the complete `digitalocean_spaces_key` resource surface.

## Resources

| Resource | Purpose |
|---|---|
| `digitalocean_spaces_key.key` | The access-key pair, carrying the spec's grant rows |

## Inputs

Generated `variables.tf` mirrors the `DigitalOceanSpacesKeySpec` proto: `key_name` and the `grants` list (each row a flattened bucket reference string plus `permission`). Authentication uses `digitalocean_token` (sensitive) -- no Spaces credentials are needed to MANAGE keys.

## Outputs

Exactly the `DigitalOceanSpacesKeyOutputs` contract: `access_key`, `secret_key` (sensitive).

## Behavior notes

- A fullaccess grant renders `bucket = ""` -- the provider's own account-wide grammar; spec validation guarantees the permission/bucket pairing before the module runs.
- `secret_key` is write-once: it exists only in the create response, and the provider's Read never touches it -- state carries the only copy.
- Name and grants update in place (the grant list is replaced wholesale); the key material never changes.
- Import: none -- the resource declares no importer at the pinned provider (see `iac/import-map.yaml` for the recorded exclusion).
