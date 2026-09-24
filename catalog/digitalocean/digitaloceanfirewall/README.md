# DigitalOcean Firewall

A DigitalOcean Cloud Firewall described once in a Planton manifest: a named, stateful, default-deny rule set applied to Droplets directly by reference or dynamically by tag, with rule sources and destinations that can name other Planton resources — Droplets, load balancers, Kubernetes clusters — instead of hand-copied IDs.

## What this component models

The spec maps one-to-one onto DigitalOcean's `digitalocean_firewall`:

| Spec field | What it controls |
|---|---|
| `firewallName` | The firewall's name (unique per account; letters, numbers, colons, dashes, underscores) |
| `inboundRules` | Traffic allowed *to* the protected Droplets; anything unmatched is dropped |
| `outboundRules` | Traffic allowed *from* the protected Droplets |
| `dropletIds` | Droplets to protect, as literal numeric IDs or references to `DigitalOceanDroplet` resources (API max 10) |
| `tags` | Droplet tags to protect: membership follows the tag automatically (API max 5; every tag must already exist — Droplets create tags, firewalls never do) |

Each rule carries a `protocol` (`tcp`, `udp`, `icmp`), a `portRange` (`"443"`, `"8000-9000"`, or `"all"`; omitted for icmp), and any mix of sources/destinations: CIDR `addresses` (IPv4 and IPv6), Droplet `tags`, `dropletIds`, `kubernetesIds`, and `loadBalancerUids` — the last three as references or literals.

At least one rule (in either direction) is required — DigitalOcean rejects an empty firewall, and validation catches it before any provisioner runs.

## Quick start

A web-server firewall — HTTPS from everywhere, SSH from a management network, all outbound open:

```yaml
apiVersion: digital-ocean.planton.dev/v1alpha1
kind: DigitalOceanFirewall
metadata:
  name: web-firewall
spec:
  firewallName: web-firewall
  tags:
    - web
  inboundRules:
    - protocol: tcp
      portRange: "443"
      sourceAddresses: ["0.0.0.0/0", "::/0"]
    - protocol: tcp
      portRange: "22"
      sourceAddresses: ["203.0.113.0/24"]
  outboundRules:
    - protocol: tcp
      portRange: all
      destinationAddresses: ["0.0.0.0/0", "::/0"]
    - protocol: udp
      portRange: all
      destinationAddresses: ["0.0.0.0/0", "::/0"]
```

Protect a specific Droplet by reference instead of (or alongside) tags:

```yaml
spec:
  firewallName: app-firewall
  dropletIds:
    - valueFrom:
        kind: DigitalOceanDroplet
        name: my-app-droplet
        fieldPath: status.outputs.droplet_id
  inboundRules:
    - protocol: tcp
      portRange: "8080"
      sourceLoadBalancerUids:
        - valueFrom:
            kind: DigitalOceanLoadBalancer
            name: my-lb
            fieldPath: status.outputs.load_balancer_id
```

## Behavior worth knowing

- **Stateful rules** — return traffic for an allowed connection is allowed automatically; you never write the reverse rule.
- **`portRange: all`, not `1-65535`** — the API reports "all ports" as `all`, so authoring `1-65535` reads back differently and diffs forever.
- **icmp takes no port range** — the provider drops one if set.
- **Tag targeting scales; ID targeting doesn't** — tags follow Droplets automatically and have no practical churn cost; the ID list caps at 10 and must be edited on every change. Prefer tags for anything long-lived.
- **Everything updates in place** — no firewall field forces recreation.

## Outputs

| Output | Meaning |
|---|---|
| `firewall_id` | The firewall's UUID — what other resources and imports address it by |

## See also

- `GUIDE.md` — operational judgment calls (tier layouts, egress policy, the double-firewall trap)
- `presets/` — ready-to-adapt web-tier and database-tier rule sets
- `v1alpha1/reference.md` — the generated field-by-field contract

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
