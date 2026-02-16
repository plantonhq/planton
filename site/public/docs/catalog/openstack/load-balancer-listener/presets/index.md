---
title: "Presets"
description: "Ready-to-deploy configuration presets for Load Balancer Listener"
type: "preset-list"
componentSlug: "load-balancer-listener"
componentTitle: "Load Balancer Listener"
provider: "openstack"
icon: "package"
order: 200
presets:
  - slug: "01-http"
    rank: "01"
    title: "HTTP Listener"
    excerpt: "This preset creates an HTTP listener on port 80. It accepts unencrypted HTTP traffic and forwards it to a backend pool. This is the simplest and most common listener configuration -- suitable for..."
  - slug: "02-https-terminated"
    rank: "02"
    title: "HTTPS Listener with TLS Termination"
    excerpt: "This preset creates an HTTPS listener on port 443 that terminates TLS at the load balancer. Traffic is decrypted at the Octavia amphora and forwarded to backend pools as plain HTTP. The..."
  - slug: "03-tcp-passthrough"
    rank: "03"
    title: "TCP Passthrough Listener"
    excerpt: "This preset creates a TCP listener that passes raw TCP traffic to backend pools without any protocol-level processing. Use this for databases, message queues, gRPC, or any non-HTTP service that needs..."
---

# Load Balancer Listener Presets

Ready-to-deploy configuration presets for Load Balancer Listener. Each preset is a complete manifest you can copy, customize, and deploy.
