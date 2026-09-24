# DigitalOcean Droplet -- Terraform Module

Deploys a `digitalocean_droplet` from a `DigitalOceanDroplet` spec: image and sizing, region and VPC placement, SSH keys, backups with a policy window, IPv6 and public-networking toggles, the monitoring and web-console agents, volume attachments, tags, cloud-init user data, GPU partitioning, graceful shutdown, and resize behavior. Provider pin is `~> 2.99`.

`variables.tf` is generated (`planton tofu generate-variables DigitalOceanDroplet`). Do not hand-edit it. The API token lives in `credentials.tf`.

## Prerequisites

- OpenTofu or Terraform 1.5+
- DigitalOcean API token (`digitalocean_token`)

## Usage

```hcl
module "droplet" {
  source = "./path/to/module"

  metadata = {
    name = "web-1"
  }

  spec = {
    droplet_name = "web-1"
    region       = "nyc3"
    size         = "s-2vcpu-4gb"
    image        = "ubuntu-24-04-x64"
    vpc          = "b5648f9e-a28a-4760-bb87-b2fad07ae295"
    ssh_keys     = ["3b:16:bf:e4:8b:00:8b:b8:59:8c:a9:d3:f0:19:45:fa"]
    enable_backups = true
    backup_policy = {
      plan    = "weekly"
      weekday = "SUN"
      hour    = 4
    }
    monitoring = true
    tags       = ["web", "env:prod"]
  }

  digitalocean_token = var.digitalocean_token
}
```

## Outputs

Exactly the kind's stack-output contract, identical to the Pulumi module:

| Output | Description |
|--------|-------------|
| `droplet_id` | The droplet's integer id (import id for `digitalocean_droplet`) |
| `ipv4_address` | Public IPv4 address |
| `ipv6_address` | Public IPv6 address (empty unless IPv6 is enabled) |
| `ipv4_address_private` | Private IPv4 address inside the droplet's VPC |
| `urn` | `do:droplet:<id>` |
| `vpc_uuid` | The VPC the droplet landed in (the region's default when `vpc` was omitted) |

## Behavior notes

- `region` is omitted when unset — DigitalOcean chooses a region with available capacity. The unspecified enum name is never sent as a slug, and there is no silent fallback region.
- `resize_disk`, `droplet_agent`, and `public_networking` pass through as null when unset, so a provider default is never silently flipped (`resize_disk` defaults ON provider-side).
- `ssh_keys` and `volume_ids` are sent as null when empty so computed sets never diff.
- `ssh_keys`, `user_data`, and `droplet_agent` are under `lifecycle.ignore_changes`: all three are ForceNew and never read back by the API, so without the guard a manifest edit — or adopting an existing droplet whose manifest carries them — would plan a destroy-and-recreate of a running machine. They mean nothing after first boot; the Pulumi module ignores the same trio.
- Tags are `spec.tags` plus the standard Planton labels rendered as `key:value` — the exact set the Pulumi module applies. See the kind [GUIDE](../../GUIDE.md).
