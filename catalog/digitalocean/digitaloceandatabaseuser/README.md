# DigitalOcean Database User

Built for 100% parity with the Terraform DigitalOcean provider's `digitalocean_database_user` resource at the pinned provider version.

## What this component models

An additional user on a DigitalOcean managed database cluster: the user's name, the MySQL authentication plugin choice, and the Kafka / OpenSearch access-control lists. DigitalOcean generates the password server-side (and, on Kafka clusters, a mutual-TLS certificate pair); they surface only as secret stack outputs.

The component covers the provider's full argument surface:

- `cluster` -- the owning cluster, by literal UUID or by reference to a `DigitalOceanDatabaseCluster` (create-only; a cluster change replaces the user, which mints a NEW password)
- `user_name` -- the user's API identity within the cluster (create-only)
- `mysql_auth_plugin` -- MySQL clusters only; `caching_sha2_password` (DigitalOcean's default) or the legacy `mysql_native_password`; updates apply in place through a password-preserving auth reset
- `settings.kafka_acls` -- per-topic Kafka permissions (`admin`, `consume`, `produce`, `produceconsume`), with wildcard topic patterns
- `settings.opensearch_acls` -- per-index OpenSearch permissions (`deny`, `admin`, `read`, `write`, `readwrite`), with wildcard index patterns

## Quick start

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDatabaseUser
metadata:
  name: orders-service-user
spec:
  cluster:
    valueFrom:
      kind: DigitalOceanDatabaseCluster
      name: orders-postgres
      fieldPath: status.outputs.cluster_id
  userName: orders-service
```

Deploy with either provisioner; both produce identical resources and outputs.

## Outputs

| Output | Description |
|---|---|
| `cluster_id` | UUID of the owning cluster (half of the user's API identity) |
| `user_name` | The user's name (the other half of its identity) |
| `role` | Role DigitalOcean assigned (normally `normal`) |
| `password` | Server-generated password (secret; MongoDB returns it only at create) |
| `access_cert` | Kafka only: PEM mTLS certificate (secret) |
| `access_key` | Kafka only: PEM mTLS key (secret) |

## Behavior worth knowing

- **ACLs are write-only upstream.** The provisioners record `settings` only from the create response and never refresh them, so the manifest is the source of truth for what the user may do, and imports cannot recover ACLs.
- **Set `settings` by engine.** PostgreSQL users declare `settings: {}` (DigitalOcean stores a settings object for every PostgreSQL user; without the block the first re-plan proposes removing it). MySQL users leave it out (the API refuses a settings update on MySQL, so even an empty block fails every apply after the first). Kafka and OpenSearch users declare their ACLs.
- **User operations serialize per cluster.** The provider locks user creation/deletion per cluster, so many users on one cluster deploy sequentially by design.
- **Renaming replaces.** `user_name` and `cluster` are create-only; a change creates a new user with a new password and deletes the old one.
- **ACL read-back is normalized.** Kafka permission spellings canonicalize (e.g. `read_write` becomes `produceconsume`) and each ACL row gets a server-side id -- both are provisioning noise the modules absorb.

## Module layout

- `iac/tf/` -- OpenTofu/Terraform module (provider pinned `~> 2.99`)
- `iac/pulumi/` -- Pulumi module (Go, pulumi-digitalocean SDK)
- Both engines wire the same spec fields and export the same outputs; behavioral parity is the contract.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
