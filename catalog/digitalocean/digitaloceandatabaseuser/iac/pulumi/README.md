# Pulumi Module: DigitalOcean Database User

Provisions an additional user on a DigitalOcean managed database cluster -- the complete `digitalocean_database_user` resource surface, at 100% behavioral parity with the Terraform module (same arguments, same outputs).

## Layout

- `main.go` -- entrypoint (`package main`), loads the stack input and calls the module
- `module/main.go` -- orchestration: locals, provider, resource
- `module/database_user.go` -- the `DatabaseUser` resource and output exports
- `module/locals.go` -- target handle (the resource has no tag surface, so no label set applies)
- `module/outputs.go` -- output key constants (the `DigitalOceanDatabaseUserStackOutputs` contract)

## Behavior notes

- The spec's singular `settings` message wraps into the SDK's one-element settings array (mirroring the Terraform module's single dynamic block).
- ACLs are write-only upstream; the configuration is the source of truth.
- `password` and `access_key` are exported as secrets; `access_cert` (the public half) is unwrapped from the bridged SDK's secret flag.
