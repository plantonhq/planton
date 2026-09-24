# DigitalOcean Droplet

Deploys a DigitalOcean Droplet with configurable sizing, region and VPC placement, SSH key injection, automated backups with a policy window, IPv6 and public-networking toggles, the monitoring and web-console agents, block storage volume attachments, cloud-init user data, tags, and GPU partitioning. The create-time choices carry the weight: image, region, VPC placement, SSH keys, and cloud-init user data are all fixed at creation, while size, backups, tags, and volume attachments can change over the Droplet's life.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **DigitalOcean Droplet** -- a virtual machine using the configured size slug, base image, and optional cloud-init user data; placed in the specified region and VPC, or DigitalOcean's choice of region and its default VPC when omitted
- **SSH Key Injection** -- created only when `sshKeys` are provided; injects account-registered keys at first boot
- **Automated Backups** -- enabled by `enableBackups`, with an optional `backupPolicy` window (daily, or weekly with weekday and hour)
- **Block Storage Volume Attachments** -- created only when `volumeIds` are provided; attaches existing volumes to the Droplet (volumes must reside in the same region)
- **Monitoring and Web-Console Agents** -- opt-in via `monitoring` and `dropletAgent`
- **DigitalOcean Tags** -- user-supplied tags plus the standard Planton labels, applied for firewall targeting and resource organization

## Before You Deploy

### Planton Setup

- **DigitalOcean Provider Connection** -- an active connection in the Connect module with a DigitalOcean API token. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline API token authentication.

### DigitalOcean Account

- **A VPC network** (optional) -- provide the VPC UUID directly or reference a DigitalOceanVpc Cloud Resource via ValueFromRef; omit to use the region's default VPC.
- **SSH keys** (recommended) -- registered on your account (`doctl compute ssh-key list`); keys are injected at create only.
- **A valid Droplet size slug** (e.g., `"s-2vcpu-4gb"`) -- check available sizes via `doctl compute size list`.
- **A valid image slug** (e.g., `"ubuntu-24-04-x64"`) -- check available images via `doctl compute image list --public`.
- **Block storage volumes** (optional) -- must be pre-created in the same region if you plan to attach them via `volumeIds`.

## Deploy

### Console

Open the deployment store, find **DigitalOcean Droplet**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Production Droplet** preset in the [Presets](#presets) tab to pre-populate a working configuration with SSH keys, backups, and VPC isolation.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDroplet
metadata:
  name: app-server
  org: acme-corp
  env: prod
spec:
  dropletName: app-server
  region: nyc3
  size: s-2vcpu-4gb
  image: ubuntu-24-04-x64
  sshKeys:
    - value: "3b:16:bf:e4:8b:00:8b:b8:59:8c:a9:d3:f0:19:45:fa"
  vpc:
    value: "abc12345-6789-def0-1234-567890abcdef"
```

```shell
planton apply -f do-droplet.yaml
```

This creates a Droplet with the specified size and image in the NYC3 region, reachable with your SSH key. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the Droplet to a VPC deployed in the same InfraPipeline:

```yaml
spec:
  vpc:
    valueFrom:
      kind: DigitalOceanVpc
      name: app-network
      fieldPath: status.outputs.vpc_id
```

The InfraPipeline resolves the dependency graph, deploys the VPC first, then provisions the Droplet within it.

## Key Configuration

These are the most important decisions when configuring a Droplet. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**SSH keys** -- each `sshKeys` entry references a `DigitalOceanSshKey` (its `ssh_key_id` output) or carries a literal ID or fingerprint of a key already registered on the account. The list is applied at creation only: the API never reports the injected keys back, and later edits to the list are ignored rather than recreating the Droplet. A Droplet without keys falls back to a root password email.

**Sizing** -- The `size` field sets the Droplet's CPU and memory (e.g., `"s-1vcpu-1gb"` for development, `"s-2vcpu-4gb"` for production web servers, `"c-4vcpu-8gb"` for CPU-intensive workloads). Resizing powers the Droplet off briefly; whether the disk also grows (permanently) is governed by `resizeDisk`, which defaults ON.

**Backups** -- Set `enableBackups: true` for automated snapshots, and pin the window with `backupPolicy` (`daily`, or `weekly` with `weekday` and an `hour` on the 0/4/8/12/16/20 grid). A policy without the toggle is rejected. Known provider defect at the pinned version: backups are enabled and the policy applied, but every plan re-proposes `backups: false -> true` because the provider reads the state from a `features` list DigitalOcean no longer populates -- harmless noise, tracked upstream ([#1525](https://github.com/digitalocean/terraform-provider-digitalocean/issues/1525)).

**Cloud-init user data** -- The `userData` field accepts a cloud-init script (up to 32 KiB) for bootstrapping the Droplet on first boot. Applied at creation only: DigitalOcean stores only a hash, and later edits are ignored rather than recreating the Droplet (to run new user data, replace the Droplet deliberately). Adopting an existing Droplet whose manifest carries `userData`, `sshKeys`, or `dropletAgent` is therefore safe -- the first apply plans no replacement.

**Volume attachments** -- `volumeIds` attaches existing block storage volumes (UUIDs or ValueFromRef references) and updates in place: moving a volume between Droplets is an edit to the Droplets' manifests, never a volume recreation. Volumes must live in the same region as the Droplet.

**Networking one-way doors** -- `enableIpv6` turns on in place, but turning it off recreates the Droplet. `publicNetworking: false` (create-only) produces a Droplet with no public interface at all -- reachable only inside its VPC -- so it belongs behind a load balancer or bastion that already exists.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **DigitalOceanVpc** (optional) | `vpc` | `status.outputs.vpc_id` |
| **DigitalOceanVolume** (optional) | `volumeIds` | `status.outputs.volume_id` |
| **DigitalOceanSshKey** (optional) | `sshKeys` | `status.outputs.ssh_key_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `droplet_id` | Unique Droplet identifier on DigitalOcean | Load balancer backend attachment, firewall rules |
| `ipv4_address` | Public IPv4 address | DNS records, application configuration |
| `ipv6_address` | Public IPv6 address (if IPv6 was enabled) | IPv6 DNS records |
| `ipv4_address_private` | Private IPv4 address inside the VPC | Internal service wiring, private DNS |
| `urn` | Uniform resource name (`do:droplet:<id>`) | DigitalOcean project resources |
| `vpc_uuid` | VPC the Droplet landed in (the region's default when `vpc` was omitted) | Verifying network placement, downstream VPC wiring |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Production server** -- 2 vCPU / 4 GB instance with SSH keys, weekly backups in a fixed window, VPC isolation, monitoring, and tags for firewall targeting. Start from the **Production Droplet** preset.

**Development instance** -- 1 vCPU / 1 GB instance with no backups and no pinned region or VPC — the smallest real Droplet. Start from the **Development Droplet** preset.

## Works With

- [**DigitalOcean VPC**](/cloud-catalog/digital-ocean-vpc) -- provides the private network for Droplet placement
- [**DigitalOcean Volume**](/cloud-catalog/digital-ocean-volume) -- provides block storage volumes attached to the Droplet
- [**DigitalOcean Load Balancer**](/cloud-catalog/digital-ocean-load-balancer) -- routes traffic to Droplets by ID or tag
- [**DigitalOcean Cloud Firewall**](/cloud-catalog/digital-ocean-firewall) -- secures Droplets by ID or tag
