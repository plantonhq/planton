# DigitalOcean Droplet Autoscale Pool

Runs a fleet of identical droplets that DigitalOcean keeps at your target size -- or grows and shrinks automatically on CPU/memory utilization between the bounds you set. Unhealthy members are replaced to hold the target, and template changes roll through the fleet by replacing members with the new shape. The scaling mode is a strict either/or -- a static pool holds an exact count, a dynamic pool scales between min and max on utilization targets -- and the spec makes a mixed shape unrepresentable. Destroying the pool destroys its member droplets: that is DigitalOcean's only delete for pools.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Droplet Autoscale Pool** -- the pool with your scaling mode (static count, or dynamic bounds plus utilization targets); creation waits for the pool AND every member to reach `active`
- **Member Droplets** -- provisioned and owned by the pool from your template (size, region, image, SSH keys, networking); member names are generated from the pool name, and every member carries your template `tags` plus the standard Planton labels both engines always apply

## Before You Deploy

### Planton Setup

- **DigitalOcean Provider Connection** -- an active connection in the Connect module with a DigitalOcean API token. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.
- **An SSH key** -- a DigitalOceanSshKey resource (or a literal numeric key id) for `dropletTemplate.sshKeys`; the API requires at least one key on every pool template, because an autoscaled droplet has no other first-boot access path.
- **A VPC and Project (optional)** -- referenced in the template when you want members placed explicitly; otherwise members land in the region's default VPC and the account's default project.

### DigitalOcean Account

- **Droplet quota and budget** -- the pool needs quota for its maximum size, and every member is a real droplet billing its size's hourly rate from the moment the pool provisions it.
- **Destroy needs a second pass at the current provider** -- the first destroy reports `unexpected state 'deleting'` after DigitalOcean has already accepted the deletion; the pool and its members are removed within seconds. Terraform: destroy again. Pulumi: refresh, then destroy. The GUIDE explains the upstream defect, tracked as [digitalocean/terraform-provider-digitalocean#1605](https://github.com/digitalocean/terraform-provider-digitalocean/issues/1605).

## Deploy

### Console

Open the deployment store, find **DigitalOcean Droplet Autoscale Pool**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Static Web Fleet** preset in the [Presets](#presets) tab to hold a fixed two-droplet fleet with health-based replacement.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDropletAutoscalePool
metadata:
  name: web-fleet
  org: acme-corp
  env: prod
spec:
  poolName: web-fleet
  static:
    targetInstances: 2
  dropletTemplate:
    size: s-1vcpu-1gb
    region: nyc3
    image: ubuntu-24-04-x64
    sshKeys:
      - value: "263654"
    tags:
      - web
    withDropletAgent: true
```

```shell
planton apply -f do-autoscale-pool.yaml
```

This holds exactly two identical Ubuntu 24.04 droplets in `nyc3`, tagged `web`, with the monitoring agent installed and your SSH key injected at first boot -- DigitalOcean replaces any member that fails health checks. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the template to resources deployed in the same InfraPipeline:

```yaml
spec:
  dropletTemplate:
    sshKeys:
      - valueFrom:
          kind: DigitalOceanSshKey
          name: ops-key
          fieldPath: status.outputs.ssh_key_id
    vpc:
      valueFrom:
        kind: DigitalOceanVpc
        name: app-vpc
        fieldPath: status.outputs.vpc_id
    projectId:
      valueFrom:
        kind: DigitalOceanProject
        name: platform
        fieldPath: status.outputs.project_id
```

The InfraPipeline resolves the dependency graph, deploys the SSH key, VPC, and project first, then provisions the pool with the resolved ids.

## Key Configuration

These are the most important decisions when configuring a droplet autoscale pool. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Static or dynamic -- pick exactly one** -- The `static` and `dynamic` blocks are a strict either/or. A static pool holds `targetInstances` droplets around the clock; a dynamic pool scales between `minInstances` and `maxInstances` and requires at least one utilization target (`targetCpuUtilization` and/or `targetMemoryUtilization`, fractions in (0, 1] -- `0.7` keeps the pool near 70% load). The provider itself never validates this pairing; the spec rejects a mixed shape before anything reaches DigitalOcean.

**`maxInstances` is your cost ceiling** -- A dynamic pool bills between its bounds, and the max bound is what you pay under sustained load or a runaway feedback loop. Set it from budget, not optimism. `minInstances` is the floor of the bill -- that many droplets run even at zero load. `cooldownMinutes` throttles scaling events, so short spikes may ride on existing members.

**Keep the monitoring agent on for dynamic pools** -- Utilization targets are evaluated from the droplet monitoring agent's telemetry (`withDropletAgent: true`). Without it, memory-based scaling has no data source at all. Static pools can skip the agent, though replacement health still benefits from it.

**Target the fleet with tags, never droplet ids** -- Member droplet ids churn with every scale event; any firewall rule or load-balancer target list naming them goes stale immediately. The template's `tags` follow the membership automatically -- tag-targeted firewall rules and load-balancer tag targets are the only reliable way to address the fleet.

**Private members** -- `dropletTemplate.publicNetworking: false` creates members with no public interface, reachable only inside the VPC -- the right shape for a fleet behind a load balancer. Unset means DigitalOcean's default (public on).

**Template changes roll the fleet** -- Editing the template (size, image, `userData`, `publicNetworking`) applies in place on the pool, and DigitalOcean replaces members to converge on the new shape. Plan template edits like deployments: capacity dips while members roll. One read-back quirk: DigitalOcean reports the image as a numeric id even when you configured a slug; the modules keep your configured value, but a freshly imported pool shows an image diff on its first plan.

**Members are cattle -- keep state off them** -- The pool creates and destroys members on its own schedule. Anything on a member's local disk is one scale-in from gone. Point members at managed databases, Spaces, or volumes owned elsewhere, and use `userData` (cloud-init) to bootstrap every member identically.

**Destroy destroys the members** -- DigitalOcean's only delete for a pool terminates every droplet it owns; there is no adopt-the-members teardown. Drain traffic first, as if terminating that many droplets by hand. The inverse hazard also holds: a forgotten pool keeps real droplets billing indefinitely.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **DigitalOceanSshKey** (at least one) | `dropletTemplate.sshKeys[]` | `status.outputs.ssh_key_id` |
| **DigitalOceanVpc** (optional) | `dropletTemplate.vpc` | `status.outputs.vpc_id` |
| **DigitalOceanProject** (optional) | `dropletTemplate.projectId` | `status.outputs.project_id` |

### What This Component Provides

After provisioning, `status.outputs` carries `pool_id` (the pool's UUID -- its API identity and import id). The pool's health is deliberately not an output: a status captured at apply time goes stale the moment DigitalOcean changes it (a member fails, the pool scales), so live health is read from the API (`GET /v2/droplets/autoscale/{pool_id}`), never from stored outputs. `pool_id` is not a wiring surface for downstream Cloud Resources either: member droplet ids churn by design, so firewalls and load balancers address the fleet through the template's `tags`, which follow the membership as it scales.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Fixed-size fleet with self-healing** -- a static pool as a "keep N of these alive" primitive: DigitalOcean rebuilds any member that dies without an operator paging anyone. The right shape when capacity is known and steady. Start from the **Static Web Fleet** preset.

**Elastic worker fleet** -- a dynamic pool for batch processors, queue consumers, and bursty backends: grows on CPU load, shrinks when it falls, with a cooldown so it doesn't thrash. Trades a variable bill (bounded by `maxInstances`) for not paying peak capacity around the clock. Start from the **CPU-Scaled Workers** preset.

## Works With

- [**DigitalOcean SSH Key**](/cloud-catalog/digital-ocean-ssh-key) -- the required first-boot access key injected into every member
- [**DigitalOcean VPC**](/cloud-catalog/digital-ocean-vpc) -- explicit private-network placement for members
- [**DigitalOcean Project**](/cloud-catalog/digital-ocean-project) -- the project members are created in
- [**DigitalOcean Load Balancer**](/cloud-catalog/digital-ocean-load-balancer) -- routes traffic to the fleet by tag, surviving member churn
- [**DigitalOcean Cloud Firewall**](/cloud-catalog/digital-ocean-firewall) -- secures the fleet with tag-targeted rules that follow membership
