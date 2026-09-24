# Production Droplet

This preset creates a production-ready DigitalOcean Droplet with SSH key access, automated weekly backups in a fixed window, VPC isolation, the monitoring agent, graceful shutdown, and resource tags for firewall targeting. It uses a general-purpose 2 vCPU / 4 GB instance suitable for most web applications and microservices.

## When to Use

- Production workloads requiring reliable compute with automated backups
- Web servers, API backends, or application hosts behind a load balancer
- Any Droplet that needs VPC isolation and tag-based firewall rules

## Key Configuration Choices

- **General-purpose sizing** (`size: s-2vcpu-4gb`) -- balanced CPU/RAM for typical web workloads. Scale up to `s-4vcpu-8gb` or dedicated CPU (`c-4vcpu-8gb`) as needed; resizing powers the Droplet off briefly.
- **SSH key access** (`sshKeys`) -- the standard access path to a production VM; password logins stay disabled. Keys are create-only: changing them later recreates the Droplet, so set them here.
- **Automated backups with a policy** (`enableBackups` + `backupPolicy`) -- a weekly snapshot every Sunday in the 04:00 window, so backup I/O never competes with peak traffic. Drop `backupPolicy` to let DigitalOcean pick a daily window instead.
- **VPC isolation** (`vpc`) -- references a `DigitalOceanVpc` resource by name, resolved to its UUID at deploy time. All production Droplets should be in a VPC.
- **Monitoring agent** (`monitoring: true`) -- enables enhanced graphs and DigitalOcean monitor alert policies.
- **Graceful shutdown** (`gracefulShutdown: true`) -- the OS gets an ACPI power-off (flushing writes, stopping services) before the Droplet is destroyed.
- **Tags** (`production`, `web`) -- used by DigitalOcean Cloud Firewalls and Load Balancers for tag-based targeting.

## Placeholders to Replace

- `metadata.name` / `dropletName` -- your Droplet's name.
- `vpc.valueFrom.name` -- the name of your `DigitalOceanVpc` resource (or replace the block with `value: <vpc-uuid>` for an existing VPC).
- `sshKeys` -- each entry is a reference to a `DigitalOceanSshKey` you manage with Planton (`valueFrom` its `ssh_key_id` output) or, under `value:`, the ID or fingerprint of a key already registered on your account (`doctl compute ssh-key list`).
