# Pulumi Module: DigitalOcean SSH Key

Registers an SSH public key on the DigitalOcean account -- the complete `digitalocean_ssh_key` resource surface, at 100% behavioral parity with the Terraform module (same arguments, same outputs).

## Layout

- `main.go` -- entrypoint (`package main`), loads the IaC input and calls the module
- `module/main.go` -- orchestration: locals, provider, resource
- `module/ssh_key.go` -- the `SshKey` resource and output exports
- `module/locals.go` -- target handle (a key has no tag surface, so no label set applies)
- `module/outputs.go` -- output key constants (the `DigitalOceanSshKeyOutputs` contract)

## Behavior notes

- `public_key` is create-only upstream: any in-line change replaces the key, rotating the id and fingerprint.
- `ssh_key_id` is exported from the resource id (the numeric key id as a string) -- the same identity imports require.
