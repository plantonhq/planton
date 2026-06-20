---
title: "Presets"
description: "Ready-to-deploy configuration presets for HTTP Route"
type: "preset-list"
componentSlug: "http-route"
componentTitle: "HTTP Route"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-host-path-routing"
    rank: "01"
    title: "Host + Path Routing"
    excerpt: "The most common HTTPRoute: match a public hostname and a path prefix, then forward to a backend Service. This is the standard pattern for exposing a web application behind a Gateway."
  - slug: "02-weighted-canary"
    rank: "02"
    title: "Weighted Canary Split"
    excerpt: "Send most traffic to a stable backend and a small slice to a canary, using backend weights. The Gateway distributes requests in proportion to each backend's `weight` (here 90/10), which is the..."
---

# HTTP Route Presets

Ready-to-deploy configuration presets for HTTP Route. Each preset is a complete manifest you can copy, customize, and deploy.
