# DigitalOcean Load Balancer -- Terraform Module

Deploys a `digitalocean_loadbalancer` from a `DigitalOceanLoadBalancer` spec: regional and global types, sizing, forwarding rules with TLS termination or passthrough, health checks, sticky sessions, backend targeting by Droplet IDs or tag, VPC/subnet placement, firewall, HTTPS redirect, PROXY protocol, keepalive, idle-timeout, TLS cipher policy, project placement, BYOIP, and the global balancer's domains, targets, CDN, and failover. Provider pin is `~> 2.99`.

`variables.tf` is generated (`planton tofu generate-variables DigitalOceanLoadBalancer`). Do not hand-edit it. The API token lives in `credentials.tf`.

## Prerequisites

- OpenTofu or Terraform 1.5+
- DigitalOcean API token (`digitalocean_token`)

## Usage

```hcl
module "load_balancer" {
  source = "./path/to/module"

  metadata = {
    name = "web-lb"
  }

  spec = {
    load_balancer_name = "web-lb"
    region             = "nyc3"
    vpc                = "b5648f9e-a28a-4760-bb87-b2fad07ae295"
    droplet_tag        = "web"
    forwarding_rules = [
      {
        entry_port      = 80
        entry_protocol  = "http"
        target_port     = 8080
        target_protocol = "http"
      }
    ]
    health_check = {
      port     = 8080
      protocol = "http"
      path     = "/health"
    }
  }

  digitalocean_token = var.digitalocean_token
}
```

## Outputs

Exactly the kind's output contract, identical to the Pulumi module:

| Output | Description |
|--------|-------------|
| `load_balancer_id` | The balancer UUID (import id for `digitalocean_loadbalancer`) |
| `ip` | Public IPv4 address |
| `urn` | `do:loadbalancer:<id>` |
| `ipv6` | IPv6 address when `network_stack` is `DUALSTACK` |

## Behavior notes

- `droplet_ids` is sent as null when empty so tag-managed membership never diffs against an empty computed set. There is no `ignore_changes` on `droplet_ids` — listed Droplets are reconciled after create.
- `size` and `size_unit` are strictly either/or; the provider prefers `size_unit` when both would be sent. Spec CEL forbids carrying both.
- `region` is omitted for GLOBAL balancers. The unspecified enum name is never sent as a slug.
- `network`, `network_stack`, and `tls_cipher_policy` are never read back by the provider. See the kind [GUIDE](../../GUIDE.md).
