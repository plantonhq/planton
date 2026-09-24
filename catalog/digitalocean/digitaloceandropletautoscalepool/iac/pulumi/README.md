# Pulumi Module: DigitalOcean Droplet Autoscale Pool

Provisions a pool of identical droplets with static or utilization-driven scaling -- the complete `digitalocean_droplet_autoscale` resource surface. Behavioral parity with the Terraform module is the contract.

## Resources

| Resource | Purpose |
|---|---|
| `digitalocean.DropletAutoscale` | The pool: name + scaling config + the member droplet template |

## Inputs

`DigitalOceanDropletAutoscalePoolStackInput`: the target `DigitalOceanDropletAutoscalePool` resource and the DigitalOcean provider config (API token).

## Outputs

Exactly the `DigitalOceanDropletAutoscalePoolStackOutputs` contract: `pool_id` (Pulumi's resource id). The pool's health is deliberately not exported (an apply-time status goes stale; live health is read from the API).

## Behavior notes

- The SDK's `Config` and `DropletTemplate` are singular objects (no one-element-array bridge quirk); only the chosen scaling branch's leaves are set -- nil pointers never reach the API.
- Member tags are spec.droplet_template.tags ∪ the standard Planton labels -- the exact set the Terraform module applies.
- `public_networking` is sent only when the manifest states it (omitted when unset, so DigitalOcean's default applies); explicit `false` creates members with no public interface -- the Terraform module's null coalescing.
- SSH key, VPC, and project references resolve to literal ids before the module runs. When the VPC is unset, `LookupVpc{Region}` resolves the region's default VPC and the module sends its UUID explicitly (DigitalOcean reads it back on every GET; an omitted value would re-plan forever on a refreshed plan -- measured live).
- **Known upstream defect (bridge v4.79.1 through v4.80.1 / provider v2.100.1 through v2.101.1; [digitalocean/terraform-provider-digitalocean#1605](https://github.com/digitalocean/terraform-provider-digitalocean/issues/1605)):** the delete waiter accepts only `OK` -> `Not Found`, but the API reports `deleting` while members are terminated, so `pulumi destroy` fails `unexpected state 'deleting'` 5-6 seconds in although DigitalOcean completes the deletion. Recover with `pulumi refresh` (the pool reads 404 and leaves state) and then destroy; a second destroy WITHOUT the refresh calls delete on a gone pool, DigitalOcean answers 404, and the provider errors on that too. Fix upstream: accept `deleting` as a pending state.
