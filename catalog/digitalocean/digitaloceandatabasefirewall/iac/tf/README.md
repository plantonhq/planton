# Terraform Module: DigitalOcean Database Firewall

Provisions the inbound trusted-sources rule set of a DigitalOcean managed database cluster -- the complete `digitalocean_database_firewall` resource surface.

## Resources

| Resource | Purpose |
|---|---|
| `digitalocean_database_firewall.firewall` | The cluster's full rule set, fanned out from the spec's five typed lists |

## Inputs

Generated `variables.tf` mirrors the `DigitalOceanDatabaseFirewallSpec` proto: `cluster` (resolved reference -- arrives as the literal cluster UUID) and the five typed lists `ip_rules`, `droplet_ids`, `kubernetes_cluster_ids`, `app_ids`, `tags` (reference lists arrive as literal id strings). Authentication uses `digitalocean_token` (sensitive).

## Outputs

Exactly the `DigitalOceanDatabaseFirewallOutputs` contract: `cluster_id` -- the rule set's only durable identity.

## Behavior notes

- locals.tf fans the five typed lists out to the provider's polymorphic `{type, value}` rows (`ip_addr`, `droplet`, `k8s`, `app`, `tag`).
- Every update PUTs the FULL rule set; destroy PUTs an EMPTY set (the cluster then accepts connections from anywhere).
- The provider mints a random state id at create -- deliberately non-deterministic; the cluster UUID is the identity.
- Import: `terraform import ... <cluster_id>` (bare; see `iac/import-map.yaml`).
