# Pulumi Module: DigitalOcean Database Db

Provisions an additional logical database inside a DigitalOcean managed database cluster -- the complete `digitalocean_database_db` resource surface, at 100% behavioral parity with the Terraform module (same arguments, same outputs).

## Layout

- `main.go` -- entrypoint (`package main`), loads the IaC input and calls the module
- `module/main.go` -- orchestration: locals, provider, resource
- `module/database_db.go` -- the `DatabaseDb` resource and output exports
- `module/locals.go` -- target handle (the resource has no tag surface, so no label set applies)
- `module/outputs.go` -- output key constants (the `DigitalOceanDatabaseDbOutputs` contract)

## Behavior notes

- Both arguments are create-only upstream: any change replaces the logical database and drops its data.
