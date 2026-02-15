---
title: "Internal Ingress NGINX"
description: "This preset deploys the ingress-nginx controller with an internal load balancer. Traffic is only reachable from within the VPC or connected networks (VPN, peering), not from the public internet."
type: "preset"
rank: "02"
presetSlug: "02-internal"
componentSlug: "ingress-nginx"
componentTitle: "Ingress Nginx"
provider: "kubernetes"
icon: "package"
order: 2
---

# Internal Ingress NGINX

This preset deploys the ingress-nginx controller with an internal load balancer. Traffic is only reachable from within the VPC or connected networks (VPN, peering), not from the public internet.

## When to Use

- Internal microservice communication that needs HTTP routing
- Private APIs or dashboards that should not be internet-accessible
- Clusters where all public traffic enters through a separate edge layer (CDN, WAF, external ALB)

## Key Configuration Choices

- **Internal load balancer** (`internal: true`) -- the controller service gets a private IP within the VPC; cloud providers annotate the service accordingly
- **No provider config** -- works on any cloud provider; add a `gke`, `eks`, or `aks` block for provider-specific options (subnet selection, managed identity)

## Placeholders to Replace

No placeholders -- this preset is directly deployable with sensible defaults.

## Related Presets

- **01-internet-facing** -- Use when the ingress controller should be accessible from the public internet
