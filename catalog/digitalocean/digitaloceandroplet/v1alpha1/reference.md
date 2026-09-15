# DigitalOceanDroplet

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `digital-ocean.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

DigitalOceanDropletSpec models the full digitalocean_droplet resource
surface: base image and sizing, region and VPC placement, SSH key
injection, automated backups with a weekly/daily policy window, IPv6 and
public-network toggles, the monitoring and web-console agents, block
volume attachments, cloud-init user data, tags, GPU partitioning, and the
resize/shutdown behavior flags.

## Example

```yaml
# Example DigitalOceanDroplet manifests.
#
# Deploy with: planton apply -f manifest.yaml
#
# The first document is the smallest real droplet (name + size + image;
# DigitalOcean chooses the region and uses its default VPC). The second
# exercises the full surface: explicit region and VPC, SSH keys, weekly
# backups with a policy window, the monitoring and web-console agents,
# IPv6, tags, cloud-init user data, graceful shutdown, and the reversible
# resize mode.
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDroplet
metadata:
  name: example-dodrop-minimal
spec:
  dropletName: example-dodrop-minimal
  size: s-1vcpu-1gb
  image: ubuntu-24-04-x64
---
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDroplet
metadata:
  name: example-dodrop-full
spec:
  dropletName: web-1.example.com
  region: nyc3
  size: s-2vcpu-4gb
  image: ubuntu-24-04-x64
  vpc:
    value: b5648f9e-a28a-4760-bb87-b2fad07ae295
  sshKeys:
    - "12345678"
    - "3b:16:bf:e4:8b:00:8b:b8:59:8c:a9:d3:f0:19:45:fa"
  enableIpv6: true
  enableBackups: true
  backupPolicy:
    plan: weekly
    weekday: SUN
    hour: 4
  monitoring: true
  dropletAgent: true
  gracefulShutdown: true
  resizeDisk: false
  tags:
    - web
    - env:prod
  userData: |
    #cloud-config
    package_update: true
    runcmd:
      - apt-get install -y nginx
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.dropletName` | `string` | yes |  |  |
| `spec.region` | `enum` |  |  |  |
| `spec.size` | `string` | yes |  |  |
| `spec.image` | `string` | yes |  |  |
| `spec.vpc` | `string \| valueFrom` |  |  | DigitalOceanVpc (`status.outputs.vpc_id`) |
| `spec.enableIpv6` | `bool` |  |  |  |
| `spec.enableBackups` | `bool` |  |  |  |
| `spec.volumeIds` | `[]string \| valueFrom` |  |  | DigitalOceanVolume (`status.outputs.volume_id`) |
| `spec.tags` | `[]string` |  |  |  |
| `spec.userData` | `string` |  |  |  |
| `spec.monitoring` | `bool` |  |  |  |
| `spec.sshKeys` | `[]string` |  |  |  |
| `spec.backupPolicy` | `DigitalOceanDropletBackupPolicy` |  |  |  |
| `spec.backupPolicy.plan` | `string` |  |  |  |
| `spec.backupPolicy.weekday` | `string` |  |  |  |
| `spec.backupPolicy.hour` | `int32` |  |  |  |
| `spec.dropletAgent` | `bool` |  |  |  |
| `spec.gracefulShutdown` | `bool` |  |  |  |
| `spec.resizeDisk` | `bool` |  | `true` |  |
| `spec.publicNetworking` | `bool` |  |  |  |
| `spec.gpuPartitionMode` | `string` |  |  |  |

## Field Details

### spec.dropletName

`string` · required

Droplet name, shown in the control panel and used as the instance
hostname. DigitalOcean accepts hostname-style names: letters, digits,
hyphens, and dots, up to 255 characters. Renaming updates in place.

- rule: {"required":true,"string":{"maxLen":"255","pattern":"^[a-zA-Z0-9]([a-zA-Z0-9.-]*[a-zA-Z0-9])?$"}}

### spec.region

`enum`

(Optional) The region to create the droplet in. When unset, DigitalOcean
chooses a region with available capacity. Cannot be changed after
creation.

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

### spec.size

`string` · required

Droplet size slug, e.g. "s-1vcpu-1gb" or "g-8vcpu-32gb". Changing the
size resizes the droplet (it is powered off during the resize); whether
the resize permanently grows the disk is governed by resize_disk.

- rule: {"required":true,"string":{"pattern":"^[a-z0-9]+(-[a-z0-9]+)+$"}}

### spec.image

`string` · required

Base image: an OS image slug (e.g. "ubuntu-24-04-x64"), a custom image
ID, or a droplet snapshot ID (both numeric). Cannot be changed after
creation.

- rule: {"required":true,"string":{"pattern":"^[a-z0-9]([-a-z0-9]*[a-z0-9])?$"}}

### spec.vpc

`string | valueFrom`

(Optional) Reference to the DigitalOcean VPC to attach the droplet's
private network interface to. When unset, the droplet lands in the
region's default VPC (the resulting UUID is exported as the vpc_uuid
output either way). Cannot be changed after creation.

- references: DigitalOceanVpc (`status.outputs.vpc_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

### spec.enableIpv6

`bool`

Enable public IPv6 networking. Enabling on an existing droplet updates
in place; disabling it forces the droplet to be recreated.

### spec.enableBackups

`bool`

Enable automated backups. Toggling updates in place. The backup window
is configured via backup_policy; without one, DigitalOcean defaults to
a daily plan.
KNOWN PROVIDER DEFECT at the pinned version (~> 2.99): the provider
decides whether backups are on by looking for "backups" in the droplet's
`features` list, and DigitalOcean's current backup system no longer
reports it there (the truth lives at GET /v2/droplets/{id}/backups/policy,
which shows backup_enabled: true). Backups ARE enabled and the policy IS
applied, but every apply re-plans enable_backups false -> true and sends
an idempotent enable action. Harmless, noisy, and upstream:
digitalocean/terraform-provider-digitalocean#1525.

### spec.volumeIds

`[]string | valueFrom`

Block storage volumes to attach, referencing DigitalOceanVolume
resources in the same region. Attachment changes update in place.

- references: DigitalOceanVolume (`status.outputs.volume_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: DigitalOceanVolume, name: <that resource's name>, fieldPath: status.outputs.volume_id}} -- a bare string does not parse

### spec.tags

`[]string`

(Optional) Tags applied to the droplet in DigitalOcean, in addition to
the standard Planton labels both provisioners always apply. Tags are
how firewalls and load balancers target droplet groups.

- rule: {"repeated":{"unique":true,"items":{"string":{"pattern":"^[a-zA-Z0-9:\\-_]{1,255}$"}}}}

### spec.userData

`string`

(Optional) Cloud-init user data executed on first boot (<= 32 KiB).
Applied at creation ONLY; later edits are ignored (both provisioners
skip changes to it, because the only alternative the provider offers is
recreating the droplet). DigitalOcean stores only a hash of it and never
reports it back. To run new user data, replace the droplet deliberately.

- rule: {"string":{"maxBytes":"32768"}}

### spec.monitoring

`bool`

Install the DigitalOcean monitoring agent for enhanced graphs and
monitor alert policies. Defaults OFF, matching the provider. Cannot be
changed after creation.

### spec.sshKeys

`[]string`

(Optional) SSH keys to inject at creation — the standard access path to
a droplet. Each entry is the ID or fingerprint of an SSH key already
registered on the DigitalOcean account. Applied at creation ONLY; later
edits are ignored (both provisioners skip changes to the list, because
the only alternative the provider offers is recreating the droplet), and
the API never reports the injected keys back. Rotate access inside the
OS (authorized_keys) or replace the droplet deliberately.

- rule: {"repeated":{"unique":true,"items":{"string":{"minLen":"1"}}}}

### spec.backupPolicy

`DigitalOceanDropletBackupPolicy`

(Optional) When and how often automated backups run. Requires
enable_backups; omitted with backups enabled, DigitalOcean defaults to
a daily plan in a window it picks.

### spec.backupPolicy.plan

`string`

Backup plan: "daily" or "weekly".

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["daily","weekly"]}}

### spec.backupPolicy.weekday

`string`

Day of the week a weekly backup runs, as the API's three-letter
uppercase token.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["SUN","MON","TUE","WED","THU","FRI","SAT"]}}

### spec.backupPolicy.hour

`int32`

Hour of the day the backup window starts (0-20, mirroring the
provider's own bounds). The API schedules windows on a four-hour grid:
0, 4, 8, 12, 16, or 20.

- rule: {"int32":{"lte":20,"gte":0}}

### spec.dropletAgent

`bool` · optional (explicit presence)

(Optional) Install the DigitalOcean agent that powers the web console
in the control panel. Unset, DigitalOcean installs it where the image
supports it and silently skips otherwise; explicit true makes an
installation failure fatal; explicit false prevents installation.
Applied at creation ONLY; later edits are ignored (both provisioners
skip changes to it, because the only alternative the provider offers is
recreating the droplet), and the API never reports it back.

### spec.gracefulShutdown

`bool`

Gracefully shut the droplet down (ACPI power-off, letting the OS flush
and stop services) before it is destroyed, instead of the default
immediate power-off. Updates in place; it only affects destroy-time
behavior.

### spec.resizeDisk

`bool` · optional (explicit presence)

Whether a size change also permanently grows the disk. DigitalOcean
defaults this ON (unset defers to that): the resize applies fully and
cannot be reverted to a smaller size later. Set false for a CPU/RAM-only
resize that stays reversible.

- default: `true`

### spec.publicNetworking

`bool` · optional (explicit presence)

(Optional) Public networking is enabled on every new droplet by
default; set explicit false to create a droplet with NO public network
interface at all (reachable only inside its VPC). Cannot be changed
after creation.

### spec.gpuPartitionMode

`string`

(Optional) Partition mode for a GPU droplet, only supported on GPU
sizes that advertise it. Omit for a full GPU (equivalent to
PARTITION_MODE_SPX_NPS1). Cannot be changed after creation.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["PARTITION_MODE_SPX_NPS1","PARTITION_MODE_DPX_NPS2"]}}

## Validation Rules

- `backup_policy_requires_backups`: backup_policy can only be set when enable_backups is true

## Outputs

Reference an output from another manifest as `valueFrom: {kind: DigitalOceanDroplet, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.droplet_id` | `string` | droplet unique identifier (DigitalOcean's integer id, as a string) |
| `status.outputs.ipv4_address` | `string` | public IPv4 address |
| `status.outputs.ipv6_address` | `string` | public IPv6 address (empty unless enable_ipv6 is set) |
| `status.outputs.ipv4_address_private` | `string` | private IPv4 address inside the droplet's VPC |
| `status.outputs.urn` | `string` | uniform resource name, e.g. "do:droplet:12345" — the form other DigitalOcean APIs (projects, firewalls) accept as a member reference |
| `status.outputs.vpc_uuid` | `string` | UUID of the VPC the droplet landed in — the region's default VPC when spec.vpc was omitted |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.vpc` | DigitalOceanVpc | `status.outputs.vpc_id` |
| `spec.volumeIds` | DigitalOceanVolume | `status.outputs.volume_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| DigitalOceanDatabaseFirewall | `spec.dropletIds` | `status.outputs.droplet_id` |
| DigitalOceanFirewall | `spec.inboundRules[].sourceDropletIds` | `status.outputs.droplet_id` |
| DigitalOceanFirewall | `spec.outboundRules[].destinationDropletIds` | `status.outputs.droplet_id` |
| DigitalOceanFirewall | `spec.dropletIds` | `status.outputs.droplet_id` |
| DigitalOceanLoadBalancer | `spec.dropletIds` | `status.outputs.droplet_id` |
| DigitalOceanMonitorAlert | `spec.dropletIds` | `status.outputs.droplet_id` |
| DigitalOceanReservedIp | `spec.droplet` | `status.outputs.droplet_id` |

## See Also

- [Overview](../README.md)
