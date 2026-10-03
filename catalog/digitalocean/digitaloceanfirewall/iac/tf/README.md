# DigitalOcean Firewall -- Terraform Module

Deploys a `digitalocean_firewall` from a `DigitalOceanFirewall` spec: the named rule set (both directions, all five source/destination classes), Droplet targeting by resolved reference, and tag targeting. Provider pin is `~> 2.99`.

`variables.tf` is generated (`planton tofu generate-variables DigitalOceanFirewall`). Do not hand-edit it. The API token lives in `credentials.tf`.

## Prerequisites

- OpenTofu or Terraform 1.5+
- DigitalOcean API token (`digitalocean_token`)

## Usage

```hcl
module "firewall" {
  source = "./path/to/module"

  metadata = {
    name = "web-firewall"
  }

  spec = {
    firewall_name = "web-firewall"
    tags          = ["web"]
    inbound_rules = [
      {
        protocol         = "tcp"
        port_range       = "443"
        source_addresses = ["0.0.0.0/0", "::/0"]
      }
    ]
    outbound_rules = [
      {
        protocol              = "tcp"
        port_range            = "all"
        destination_addresses = ["0.0.0.0/0", "::/0"]
      }
    ]
  }

  digitalocean_token = var.digitalocean_token
}
```

## Behavior notes

- Reference fields (`droplet_ids` and the rule-level droplet/kubernetes/load-balancer lists) arrive flattened as plain strings — the Planton orchestrator resolves `valueFrom` references before Terraform runs. Droplet IDs convert with `tonumber` (the API wants integers).
- Empty collections are sent as `null`, never `[]`: the provider omits absent collections on read-back, and the rule blocks are set-hashed, so an empty list would diff forever.
- Write `port_range = "all"` (not `"1-65535"`) and omit it on icmp rules — the provider normalizes both on read.

## Outputs

Exactly the kind's output contract, identical to the Pulumi module: `firewall_id`.
