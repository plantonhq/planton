---
title: "Presets"
description: "Ready-to-deploy configuration presets for Ingress Nginx"
type: "preset-list"
componentSlug: "ingress-nginx"
componentTitle: "Ingress Nginx"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-internet-facing"
    rank: "01"
    title: "Internet-Facing Ingress NGINX"
    excerpt: "This preset deploys the ingress-nginx controller with an internet-facing (external) load balancer. This is the most common configuration for clusters that serve public traffic."
  - slug: "02-internal"
    rank: "02"
    title: "Internal Ingress NGINX"
    excerpt: "This preset deploys the ingress-nginx controller with an internal load balancer. Traffic is only reachable from within the VPC or connected networks (VPN, peering), not from the public internet."
---

# Ingress Nginx Presets

Ready-to-deploy configuration presets for Ingress Nginx. Each preset is a complete manifest you can copy, customize, and deploy.
