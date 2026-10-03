# Terraform Module: DigitalOcean Database User

Provisions an additional user on a DigitalOcean managed database cluster -- the complete `digitalocean_database_user` resource surface.

## Resources

| Resource | Purpose |
|---|---|
| `digitalocean_database_user.user` | The user: name, MySQL auth plugin, Kafka/OpenSearch ACLs |

## Inputs

Generated `variables.tf` mirrors the `DigitalOceanDatabaseUserSpec` proto: `cluster` (resolved reference -- arrives as the literal cluster UUID), `user_name`, optional `mysql_auth_plugin`, optional `settings` with `kafka_acls` / `opensearch_acls` lists. Authentication uses `digitalocean_token` (sensitive).

## Outputs

Exactly the `DigitalOceanDatabaseUserOutputs` contract: `cluster_id`, `user_name`, `role`, and the secrets `password`, `access_cert`, `access_key` (Kafka only).

## Behavior notes

- ACL settings are write-only upstream (returned only at create); the configuration is the source of truth and imports never recover them.
- `mysql_auth_plugin` left null defers to DigitalOcean's `caching_sha2_password` default without drift (the provider suppresses the default-vs-unset diff).
- Import: `terraform import ... <cluster_id>,<user_name>` (see `iac/import-map.yaml`).
