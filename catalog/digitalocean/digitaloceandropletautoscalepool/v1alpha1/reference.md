# DigitalOceanDropletAutoscalePool

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

DigitalOceanDropletAutoscalePoolSpec models the full
digitalocean_droplet_autoscale resource surface: a pool of identical
droplets that DigitalOcean keeps at a fixed size or scales between bounds
on CPU/memory utilization -- the closest thing DigitalOcean has to a
managed instance group.

The scaling mode is a strict either/or: a static pool holds an exact
member count, a dynamic pool scales between min and max on utilization
targets. The provider validates NONE of this pairing (it sends every
field it is given and lets the API sort it out); the oneof below makes a
mixed shape unrepresentable instead.

DESTROY DESTROYS THE MEMBER DROPLETS: the API's only delete for a pool is
the "dangerous" variant that terminates every droplet the pool owns.
There is no way to keep the members and delete only the pool.

## Example

```yaml
# Reference manifest for DigitalOceanDropletAutoscalePool --
# protovalidate-valid, embedded as the reference page's Example block, and
# the documents the offline tofu plan renders. Two documents so the plan
# proves BOTH scaling branches' rendering: a static pool and a dynamic
# pool with the full template surface.
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDropletAutoscalePool
metadata:
  name: web-static-pool
spec:
  poolName: web-static
  static:
    targetInstances: 2
  dropletTemplate:
    size: s-1vcpu-1gb
    region: nyc3
    image: ubuntu-24-04-x64
    # Literal numeric SSH key ids; use valueFrom to reference
    # DigitalOceanSshKey resources instead.
    sshKeys:
      - value: "12345678"
---
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDropletAutoscalePool
metadata:
  name: web-dynamic-pool
spec:
  poolName: web-dynamic
  dynamic:
    minInstances: 1
    maxInstances: 5
    targetCpuUtilization: 0.7
    targetMemoryUtilization: 0.8
    cooldownMinutes: 10
  dropletTemplate:
    size: s-1vcpu-1gb
    region: nyc3
    image: ubuntu-24-04-x64
    sshKeys:
      - value: "12345678"
    vpc:
      value: aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee
    projectId:
      value: ffffffff-1111-2222-3333-444444444444
    tags:
      - web
      - autoscaled
    # Dynamic scaling decides on agent-supplied CPU/memory metrics.
    withDropletAgent: true
    ipv6: true
    userData: |
      #cloud-config
      packages:
        - nginx
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.poolName` | `string` | yes |  |  |
| `spec.static` | `DigitalOceanDropletAutoscalePoolStaticScale` |  |  |  |
| `spec.static.targetInstances` | `uint32` | yes |  |  |
| `spec.dynamic` | `DigitalOceanDropletAutoscalePoolDynamicScale` |  |  |  |
| `spec.dynamic.minInstances` | `uint32` | yes |  |  |
| `spec.dynamic.maxInstances` | `uint32` | yes |  |  |
| `spec.dynamic.targetCpuUtilization` | `double` |  |  |  |
| `spec.dynamic.targetMemoryUtilization` | `double` |  |  |  |
| `spec.dynamic.cooldownMinutes` | `uint32` |  |  |  |
| `spec.dropletTemplate` | `DigitalOceanDropletAutoscalePoolTemplate` | yes |  |  |
| `spec.dropletTemplate.size` | `string` | yes |  |  |
| `spec.dropletTemplate.region` | `enum` | yes |  |  |
| `spec.dropletTemplate.image` | `string` | yes |  |  |
| `spec.dropletTemplate.sshKeys` | `[]string \| valueFrom` | yes |  | DigitalOceanSshKey (`status.outputs.ssh_key_id`) |
| `spec.dropletTemplate.vpc` | `string \| valueFrom` |  |  | DigitalOceanVpc (`status.outputs.vpc_id`) |
| `spec.dropletTemplate.projectId` | `string \| valueFrom` |  |  | DigitalOceanProject (`status.outputs.project_id`) |
| `spec.dropletTemplate.tags` | `[]string` |  |  |  |
| `spec.dropletTemplate.withDropletAgent` | `bool` |  |  |  |
| `spec.dropletTemplate.ipv6` | `bool` |  |  |  |
| `spec.dropletTemplate.userData` | `string` |  |  |  |

## Field Details

### spec.poolName

`string` · required

Name of the autoscale pool. Updates in place.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.static

`DigitalOceanDropletAutoscalePoolStaticScale`

Hold the pool at an exact member count.

### spec.static.targetInstances

`uint32` · required

The exact number of droplets to maintain in the pool.

- rule: {"required":true,"uint32":{"gte":1}}

### spec.dynamic

`DigitalOceanDropletAutoscalePoolDynamicScale`

Scale between bounds on average CPU/memory utilization.

- rule: min_instances must not exceed max_instances
- rule: dynamic scaling requires target_cpu_utilization and/or target_memory_utilization

### spec.dynamic.minInstances

`uint32` · required

The fewest droplets the pool may shrink to.

- rule: {"required":true,"uint32":{"gte":1}}

### spec.dynamic.maxInstances

`uint32` · required

The most droplets the pool may grow to.

- rule: {"required":true,"uint32":{"gte":1}}

### spec.dynamic.targetCpuUtilization

`double` · optional (explicit presence)

(Optional) Average CPU load to maintain, as a fraction in (0, 1] --
e.g. 0.7 scales to keep the pool near 70% CPU. The floor excludes zero
because the wire format drops zero values: an explicit 0 can never
reach the API and would silently mean "no CPU target".

- rule: {"double":{"lte":1,"gt":0}}

### spec.dynamic.targetMemoryUtilization

`double` · optional (explicit presence)

(Optional) Average memory load to maintain, as a fraction in (0, 1].
Same zero-exclusion rationale as target_cpu_utilization.

- rule: {"double":{"lte":1,"gt":0}}

### spec.dynamic.cooldownMinutes

`uint32` · optional (explicit presence)

(Optional) Minutes to wait between scaling events. When unset,
DigitalOcean applies its own default cooldown.

- rule: {"uint32":{"gte":1}}

### spec.dropletTemplate

`DigitalOceanDropletAutoscalePoolTemplate` · required

The template every pool member is created from. Template changes roll
through the pool by replacing members with the new shape.

- rule: {"required":true}

### spec.dropletTemplate.size

`string` · required

Droplet size slug for every member (e.g. "s-1vcpu-1gb").

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.dropletTemplate.region

`enum` · required

The DigitalOcean region every member is created in.

- rule: {"required":true}

Allowed values (use exactly as shown):

- `digital_ocean_region_unspecified` -- 0: default / unspecified region
- `nyc3` -- new york 3
- `sfo3` -- san francisco 3
- `fra1` -- frankfurt 1
- `sgp1` -- singapore 1
- `lon1` -- london 1
- `tor1` -- toronto 1
- `blr1` -- bangalore 1
- `ams3` -- amsterdam 3
- `nyc1` -- new york 1
- `nyc2` -- new york 2
- `sfo2` -- san francisco 2
- `syd1` -- sydney 1
- `atl1` -- atlanta 1

### spec.dropletTemplate.image

`string` · required

Image slug (e.g. "ubuntu-24-04-x64") or numeric image ID for every
member. DigitalOcean reports the image back as a numeric ID even when
a slug was supplied; both provisioners keep the configured value.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.dropletTemplate.sshKeys

`[]string | valueFrom` · required

SSH keys injected into every member -- required by the API (an
autoscaled droplet has no other first-boot access path). Each entry is
a literal numeric SSH key ID or a reference to a DigitalOceanSshKey
resource.

- references: DigitalOceanSshKey (`status.outputs.ssh_key_id`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanSshKey, name: <that resource's name>, fieldPath: status.outputs.ssh_key_id}} -- a bare string does not parse

### spec.dropletTemplate.vpc

`string | valueFrom`

(Optional) The VPC every member joins. Use a literal VPC UUID or a
reference to a DigitalOceanVpc resource. When unset, members join the
region's default VPC -- both provisioners look that VPC up and send it
explicitly, because DigitalOcean reports it back on every read and an
implicit value would otherwise re-plan on every apply.

- references: DigitalOceanVpc (`status.outputs.vpc_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

### spec.dropletTemplate.projectId

`string | valueFrom`

(Optional) The project members are created in. Use a literal project
UUID or a reference to a DigitalOceanProject resource. When unset,
members land in the account's default project.

- references: DigitalOceanProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.dropletTemplate.tags

`[]string`

(Optional) Tags applied to every member, in addition to the standard
Planton labels both provisioners always apply. Tags are how firewalls
and load balancers target droplet groups.

- rule: {"repeated":{"unique":true,"items":{"string":{"pattern":"^[a-zA-Z0-9:\\-_]{1,255}$"}}}}

### spec.dropletTemplate.withDropletAgent

`bool`

(Optional) Install the DigitalOcean monitoring agent on every member.
The agent supplies the CPU/memory utilization metrics dynamic scaling
decides on -- leave it enabled for any dynamic pool.

### spec.dropletTemplate.ipv6

`bool`

(Optional) Enable IPv6 networking on every member.

### spec.dropletTemplate.userData

`string`

(Optional) Cloud-init user data executed on each member's first boot.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanDropletAutoscalePool, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.pool_id` | `string` | UUID of the autoscale pool (the resource's API identity and its import id). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.dropletTemplate.sshKeys` | DigitalOceanSshKey | `status.outputs.ssh_key_id` |
| `spec.dropletTemplate.vpc` | DigitalOceanVpc | `status.outputs.vpc_id` |
| `spec.dropletTemplate.projectId` | DigitalOceanProject | `status.outputs.project_id` |

## See Also

- [Overview](../README.md)
