# Pulumi Module: DigitalOcean Database Connection Pool

Provisions a PgBouncer connection pool on a DigitalOcean managed PostgreSQL cluster -- the complete `digitalocean_database_connection_pool` resource surface, at 100% behavioral parity with the Terraform module (same arguments, same outputs).

## Layout

- `main.go` -- entrypoint (`package main`), loads the IaC input and calls the module
- `module/main.go` -- orchestration: locals, provider, resource
- `module/connection_pool.go` -- the `DatabaseConnectionPool` resource and output exports
- `module/locals.go` -- target handle (the resource has no tag surface, so no label set applies)
- `module/outputs.go` -- output key constants (the `DigitalOceanDatabaseConnectionPoolOutputs` contract)

## Behavior notes

- EVERY argument is create-only upstream; any change replaces the pool and drops its live connections.
- Omitted `user` creates DigitalOcean's inbound-user pool (clients bring their own credentials; the password output is legitimately empty).
