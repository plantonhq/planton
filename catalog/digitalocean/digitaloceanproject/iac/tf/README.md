# Terraform Module: DigitalOcean Project

Provisions a DigitalOcean project -- the complete `digitalocean_project` resource surface, with membership carried on the project itself.

## Resources

| Resource | Purpose |
|---|---|
| `digitalocean_project.project` | The project: name, description, purpose, environment, default flag, member URNs |

## Inputs

Generated `variables.tf` mirrors the `DigitalOceanProjectSpec` proto: `project_name`, optional `description`/`purpose`/`environment`, `is_default`, and `resources` (resolved references -- each arrives as a literal URN string). Authentication uses `digitalocean_token` (sensitive).

## Outputs

Exactly the `DigitalOceanProjectStackOutputs` contract: `project_id`, `owner_uuid`, `owner_id` (stringified from the provider's number), and `resource_urns` (the provider's read-back of membership, sorted from its unordered set so both engines export the same list).

## Behavior notes

- Empty optional strings become null so the provider's defaults apply (purpose defaults to "Web Application" upstream).
- An empty `resources` list stays null: membership is then unmanaged and out-of-band assignments are left alone (the attribute is Optional+Computed upstream).
- Destroy relocates member resources to the account's default project and retries through the API's 412 responses while the moves settle -- nothing inside is destroyed. The resource carries `timeouts { delete = "10m" }` because the moves were measured to outlast the provider's 3-minute default once; a destroy that still fails on the 412 has moved the members already and succeeds when run again.
- Import: `terraform import ... <project_id>` (see `iac/import-map.yaml`).
