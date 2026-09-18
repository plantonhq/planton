# Terraform Module: DigitalOcean Droplet Autoscale Pool

Provisions a pool of identical droplets with static or utilization-driven scaling -- the complete `digitalocean_droplet_autoscale` resource surface.

## Resources

| Resource | Purpose |
|---|---|
| `digitalocean_droplet_autoscale.pool` | The pool: name + scaling config + the member droplet template |

## Inputs

Generated `variables.tf` mirrors the `DigitalOceanDropletAutoscalePoolSpec` proto: `pool_name`, the scaling oneof as two optional objects (`static` / `dynamic`), and `droplet_template` with flattened reference strings for `ssh_keys` / `vpc` / `project_id`. Authentication uses `digitalocean_token` (sensitive).

## Outputs

Exactly the `DigitalOceanDropletAutoscalePoolStackOutputs` contract: `pool_id`. The pool's health is deliberately not exported (an apply-time status goes stale; live health is read from the API).

## Behavior notes

- The config block renders ONLY the chosen scaling branch's leaves (null attributes are omitted, matching the API's zero-means-unset wire behavior); the spec oneof makes mixed shapes unrepresentable.
- Member tags are spec.droplet_template.tags ∪ the standard Planton labels -- the exact set the Pulumi module applies.
- `public_networking` is sent only when the manifest states it (null when unset, so DigitalOcean's default applies); explicit `false` creates members with no public interface.
- When `spec.droplet_template.vpc` is unset, `data.digitalocean_vpc.region_default` resolves the region's default VPC and the module sends its UUID explicitly: DigitalOcean reads the UUID back on every GET and the provider's `vpc_uuid` is Optional but not Computed, so an omitted value would re-plan forever (measured live).
- Create waits for the pool AND every member to reach active (up to 15 minutes upstream); delete is `DeleteDangerous` -- it destroys the member droplets and polls up to 1 minute for the 404.
- **Known upstream defect (provider v2.100.1):** the delete waiter accepts only `OK` -> `Not Found`, but the API reports `deleting` while members are terminated, so `tofu destroy` fails `unexpected state 'deleting'` 5-6 seconds in although DigitalOcean completes the deletion. Run destroy again: the refresh reads the 404 and drops the pool from state. Fix upstream: accept `deleting` as a pending state.
- Import: `terraform import ... <pool_id>` (see `iac/import-map.yaml`; the template image reads back as a numeric id -- the first post-import plan updates it back to the configured slug, a no-op on DigitalOcean's side, tolerated as write-normalized in the import catalog).
