# Per-Service Application User

This preset creates one dedicated database user for one application service, referencing the owning cluster by name. DigitalOcean generates the password server-side; read it from the `password` stack output (a secret) -- it never appears in the manifest.

## When to Use

- One user per service, so credentials rotate and revoke independently
- Replacing shared use of the cluster's default `doadmin` user in application configs
- Composing with a DigitalOceanDatabaseDb of the same service's name

## Key Configuration Choices

- **Cluster by reference** (`valueFrom`) -- wires the user to a DigitalOceanDatabaseCluster in the same chart or environment; swap in a literal UUID for an existing cluster.
- **No auth plugin set** -- on MySQL clusters this defers to DigitalOcean's modern `caching_sha2_password` default; on other engines the field does not apply.
- **`settings: {}` because the cluster is PostgreSQL** -- every PostgreSQL user comes back from DigitalOcean with a settings object, so the empty block keeps the user idempotent from the first apply. On a MySQL cluster delete the line: MySQL users never carry settings and the API refuses a settings update there.

## What You Get

A user visible in the cluster's Users & Databases tab, with `password` (and on Kafka, `access_cert`/`access_key`) exported as secret stack outputs for application wiring.
