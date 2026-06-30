# Production Droplet

This preset creates a production-ready DigitalOcean Droplet with automated backups enabled, VPC isolation, and resource tags for firewall targeting. It uses a general-purpose 2 vCPU / 4 GB instance suitable for most web applications and microservices.

## When to Use

- Production workloads requiring reliable compute with automated backups
- Web servers, API backends, or application hosts behind a load balancer
- Any Droplet that needs VPC isolation and tag-based firewall rules

## Key Configuration Choices

- **General-purpose sizing** (`size: s-2vcpu-4gb`) -- balanced CPU/RAM for typical web workloads. Scale up to `s-4vcpu-8gb` or dedicated CPU (`c-4vcpu-8gb`) as needed.
- **Automated backups** (`enableBackups: true`) -- DigitalOcean takes weekly snapshots with 4-week retention. Critical for production disaster recovery.
- **VPC isolation** (`vpc`) -- places the Droplet in a private network. All production Droplets should be in a VPC.
- **Tags** (`production`, `web`) -- used by DigitalOcean Cloud Firewalls for tag-based targeting. The `web` tag enables attaching a web-tier firewall.
- **Monitoring enabled** -- `disableMonitoring` is omitted (defaults to `false`), so the DigitalOcean monitoring agent is active.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<vpc-id>` | UUID of the target VPC | DigitalOcean VPC console or `DigitalOceanVpc` status outputs |
| `nyc1` | Target DigitalOcean region slug | Must match the VPC's region |
| `s-2vcpu-4gb` | Droplet size slug | [DigitalOcean Sizes API](https://docs.digitalocean.com/reference/api/api-reference/#tag/Sizes) |
| `ubuntu-24-04-x64` | Base OS image slug | [DigitalOcean Images API](https://docs.digitalocean.com/reference/api/api-reference/#tag/Images) |

## Related Presets

- **02-development** -- Use instead for dev/test workloads where backups are unnecessary and a smaller instance suffices
