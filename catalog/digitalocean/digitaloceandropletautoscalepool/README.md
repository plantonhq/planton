# DigitalOcean Droplet Autoscale Pool

Built for 100% parity with the Terraform DigitalOcean provider's `digitalocean_droplet_autoscale` resource at the pinned provider version.

## What this component models

A pool of identical droplets DigitalOcean keeps at a fixed size or scales between bounds on CPU/memory utilization -- the closest thing DigitalOcean has to a managed instance group.

- `pool_name` -- the pool's name (updates in place)
- `static` XOR `dynamic` -- the scaling mode: an exact member count, or min/max bounds with utilization targets and a cooldown (the provider validates none of this pairing; the spec's either/or makes mixed shapes unrepresentable)
- `droplet_template` -- the shape of every member: size, region, image, SSH keys (by reference to DigitalOceanSshKey -- required by the API), optional VPC and project references, tags, monitoring agent, IPv6, user data

## Quick start

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDropletAutoscalePool
metadata:
  name: web-pool
spec:
  poolName: web
  dynamic:
    minInstances: 2
    maxInstances: 6
    targetCpuUtilization: 0.7
  dropletTemplate:
    size: s-1vcpu-1gb
    region: nyc3
    image: ubuntu-24-04-x64
    sshKeys:
      - valueFrom:
          kind: DigitalOceanSshKey
          name: ops-key
          fieldPath: status.outputs.ssh_key_id
    withDropletAgent: true
```

Deploy with either provisioner; both produce identical resources and outputs.

## Outputs

| Output | Description |
|---|---|
| `pool_id` | UUID of the autoscale pool (its API identity and import id) |

The pool's health is deliberately not an output: a status captured at apply time goes stale the moment DigitalOcean changes it. Read live health from the control panel or `GET /v2/droplets/autoscale/{pool_id}`.

## Behavior worth knowing

- **DESTROY DESTROYS THE MEMBERS.** The API's only delete for a pool is the dangerous variant that terminates every member droplet. There is no keep-the-droplets teardown.
- **The first destroy reports an error DigitalOcean has already accepted.** At the current provider the delete waiter does not expect the pool's `deleting` status and fails 5-6 seconds in, while DigitalOcean finishes deleting the pool and its members regardless. Terraform: run destroy again. Pulumi: `pulumi refresh`, then destroy. Tracked upstream as [digitalocean/terraform-provider-digitalocean#1605](https://github.com/digitalocean/terraform-provider-digitalocean/issues/1605); details in the GUIDE.
- **Unset `vpc` means the region's default VPC, sent explicitly.** Both modules look it up and send its UUID so the plan matches what DigitalOcean reads back.
- **Every member bills.** Members are real droplets at the template size's hourly rate -- a static pool bills `target_instances` around the clock.
- **Create waits for the whole pool.** The provider polls until the pool AND every member reach active (up to 15 minutes).
- **Dynamic scaling needs the agent.** Utilization metrics come from the droplet monitoring agent (`withDropletAgent`); a dynamic pool without it scales blind on memory.

## Module layout

- `iac/tf/` -- OpenTofu/Terraform module (provider pinned `~> 2.99`)
- `iac/pulumi/` -- Pulumi module (Go, pulumi-digitalocean SDK)
- Both engines wire the same spec fields, apply the same Planton labels as member tags, and export the same outputs; behavioral parity is the contract.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
