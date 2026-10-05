# Pulumi Module: DigitalOcean Spaces Key

Provisions a Spaces access-key pair with optional per-bucket grants -- the complete `digitalocean_spaces_key` resource surface. Behavioral parity with the Terraform module is the contract.

## Resources

| Resource | Purpose |
|---|---|
| `digitalocean.SpacesKey` | The access-key pair, carrying the spec's grant rows |

## Inputs

`DigitalOceanSpacesKeyIacInput`: the target `DigitalOceanSpacesKey` resource and the DigitalOcean provider config (API token).

## Outputs

Exactly the `DigitalOceanSpacesKeyOutputs` contract: `access_key`, `secret_key` (exported through `pulumi.ToSecret` -- the SDK does not secret-flag it).

## Behavior notes

- A fullaccess grant renders an empty `Bucket` string -- the provider's own account-wide grammar; spec validation guarantees the pairing.
- `secret_key` is write-once (create response only); the ToSecret wrap keeps it encrypted in state and masked in every output surface.
- Name and grants update in place (the grant list is replaced wholesale); the key material never changes.
