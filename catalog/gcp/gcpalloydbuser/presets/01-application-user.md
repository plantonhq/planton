# Application User (ALLOYDB_BUILT_IN)

This preset creates a classic username/password AlloyDB user for an application service.

## When to Use

- Applications that connect with a stored credential (outside IAM proxy flows)
- One user per service with its own rotatable password

## Key Configuration Choices

- **ALLOYDB_BUILT_IN (default)** — password-authenticated database role
- **databaseRoles: [alloydbiamuser]** — standard application role; adjust for your privilege model

## Related Presets

- **02-iam-user** — passwordless IAM-authenticated user

## Related Kinds

- [GcpAlloydbCluster](/docs/catalog/gcp/gcpalloydbcluster) — the cluster this user lives on
