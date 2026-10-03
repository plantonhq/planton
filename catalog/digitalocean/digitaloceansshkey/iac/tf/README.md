# Terraform Module: DigitalOcean SSH Key

Registers an SSH public key on the DigitalOcean account -- the complete `digitalocean_ssh_key` resource surface.

## Resources

| Resource | Purpose |
|---|---|
| `digitalocean_ssh_key.ssh_key` | The registered public key: name + material |

## Inputs

Generated `variables.tf` mirrors the `DigitalOceanSshKeySpec` proto: `key_name` and `public_key`. Authentication uses `digitalocean_token` (sensitive).

## Outputs

Exactly the `DigitalOceanSshKeyOutputs` contract: `ssh_key_id` (the numeric id as a string) and `fingerprint`.

## Behavior notes

- `public_key` is create-only upstream: any in-line change replaces the key (the provider trims only leading/trailing whitespace before comparing).
- Only the name updates in place.
- Import: `terraform import ... <ssh_key_id>` -- the NUMERIC id; fingerprints do not import (see `iac/import-map.yaml`).
