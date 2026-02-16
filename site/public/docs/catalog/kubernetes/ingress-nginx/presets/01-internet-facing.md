---
title: "Internet-Facing Ingress NGINX"
description: "This preset deploys the ingress-nginx controller with an internet-facing (external) load balancer. This is the most common configuration for clusters that serve public traffic."
type: "preset"
rank: "01"
presetSlug: "01-internet-facing"
componentSlug: "ingress-nginx"
componentTitle: "Ingress Nginx"
provider: "kubernetes"
icon: "package"
order: 1
---

# Internet-Facing Ingress NGINX

This preset deploys the ingress-nginx controller with an internet-facing (external) load balancer. This is the most common configuration for clusters that serve public traffic.

## When to Use

- You need an ingress controller for public-facing web applications and APIs
- Your cluster serves traffic from the internet
- You do not need provider-specific load balancer customization (static IPs, specific subnets)

## Key Configuration Choices

- **External load balancer** (`internal: false`) -- the controller service gets a public IP accessible from the internet
- **No provider config** -- works on any cloud provider with default load balancer settings; add a `gke`, `eks`, or `aks` provider config block for cloud-specific options (static IPs, subnet placement, managed identity)
- **No chart version pinned** -- uses whatever version the IaC module defaults to; set `chartVersion` explicitly for reproducible deployments

## Placeholders to Replace

No placeholders -- this preset is directly deployable with sensible defaults.

## Related Presets

- **02-internal** -- Use when the ingress controller should only be reachable within the VPC/network
