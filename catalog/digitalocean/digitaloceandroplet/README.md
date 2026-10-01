# DigitalOcean Droplet

A DigitalOcean virtual machine described once in a Planton manifest: base image and sizing, region and VPC placement, SSH key injection, automated backups with a policy window, IPv6 and public-network toggles, the monitoring and web-console agents, block volume attachments, cloud-init user data, tags, GPU partitioning, and the resize and shutdown behavior flags.

## What this component models

The spec maps one-to-one onto DigitalOcean's droplet:

| Spec field | What it controls |
|---|---|
| `dropletName` | The droplet's name and hostname (hostname-style, dots allowed; renames update in place) |
| `region` | Data-center region; omit to let DigitalOcean choose one with available capacity |
| `size` | Size slug (`s-1vcpu-1gb`, `g-8vcpu-32gb`, ...); changing it resizes the droplet |
| `image` | OS image slug, custom image ID, or snapshot ID; create-only |
| `vpc` | Optional VPC placement — a literal UUID or a reference to a `DigitalOceanVpc`; omit to use the region's default |
| `sshKeys` | SSH key IDs or fingerprints injected at create — the standard access path; applied at creation only, later edits ignored |
| `enableIpv6` | Public IPv6 networking (disabling later forces recreation) |
| `enableBackups` / `backupPolicy` | Automated backups and their daily/weekly window (plan, weekday, hour); see the known provider diff below |
| `monitoring` | The DigitalOcean monitoring agent (enhanced graphs, alert policies); create-only |
| `dropletAgent` | The web-console agent: unset lets DigitalOcean decide per image, explicit values are enforced; applied at creation only, later edits ignored |
| `volumeIds` | Block storage volumes to attach, by literal UUID or `DigitalOceanVolume` reference |
| `tags` | Tags — how firewalls and load balancers target droplet groups |
| `userData` | Cloud-init script executed on first boot (max 32 KiB); applied at creation only, later edits ignored, stored hashed |
| `gracefulShutdown` | ACPI power-off (letting the OS flush) before destroy, instead of immediate power-off |
| `resizeDisk` | Whether a size change also grows the disk permanently (DigitalOcean defaults ON; `false` keeps resizes reversible) |
| `publicNetworking` | Explicit `false` creates a droplet with no public network interface at all; create-only |
| `gpuPartitionMode` | GPU partitioning for GPU sizes that support it; create-only |

## Quick start

The smallest real droplet — DigitalOcean picks the region and default VPC:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanDroplet
metadata:
  name: dev-box
spec:
  dropletName: dev-box
  size: s-1vcpu-1gb
  image: ubuntu-24-04-x64
```

```shell
planton apply -f dev-box.yaml
```

## Outputs

Both provisioners export the identical output set:

| Output | Description |
|---|---|
| `droplet_id` | The droplet's integer id as a string (import id for `digitalocean_droplet`) |
| `ipv4_address` | Public IPv4 address |
| `ipv6_address` | Public IPv6 address (empty unless `enableIpv6`) |
| `ipv4_address_private` | Private IPv4 address inside the droplet's VPC |
| `urn` | `do:droplet:<id>` — the member reference projects and firewalls accept |
| `vpc_uuid` | The VPC the droplet landed in (the region's default when `vpc` was omitted) |

## Behavior worth knowing

- **SSH keys are create-only.** They must already be registered on the DigitalOcean account (`doctl compute ssh-key list`). Set them at create — a droplet without keys falls back to a root password email.
- **`sshKeys`, `userData`, and `dropletAgent` are applied at creation only and later edits are ignored.** The API never reports them back and the provider's only response to a change is to recreate the droplet; both provisioners skip changes to them instead, so a manifest edit never silently replaces a running machine and adopting an existing droplet is safe. To run new cloud-init or inject different keys, replace the droplet deliberately.
- **The remaining identity fields force recreation**: `image`, `region`, `vpc`, `monitoring`, `publicNetworking`, `gpuPartitionMode`. `dropletName`, `size`, `tags`, `volumeIds`, and backups update in place.
- **`enableBackups: true` re-plans on every apply (provider defect, harmless).** Backups are really enabled and the policy really applied, but the provider reads the on/off state from the droplet's `features` list, which DigitalOcean no longer populates for backups; every plan shows `backups: false → true` and every apply sends a no-op enable. Check `GET /v2/droplets/{id}/backups/policy` for the truth. Upstream: [#1525](https://github.com/digitalocean/terraform-provider-digitalocean/issues/1525).
- **`resizeDisk` defaults ON provider-side.** A disk-growing resize is permanent — you can never pick a smaller disk afterward. Set `false` to scale CPU/RAM reversibly.
- **Disabling IPv6 on a running droplet recreates it**; enabling it updates in place.
- **`publicNetworking: false` and `gpuPartitionMode` deploy on both provisioners.** Both are create-only; `gpuPartitionMode` only means anything on a GPU size. See the [GUIDE](GUIDE.md).

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
