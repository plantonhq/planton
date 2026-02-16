---
title: "Presets"
description: "Ready-to-deploy configuration presets for GHA Runner Scale Set"
type: "preset-list"
componentSlug: "gha-runner-scale-set"
componentTitle: "GHA Runner Scale Set"
provider: "kubernetes"
icon: "package"
order: 200
presets:
  - slug: "01-github-cloud"
    rank: "01"
    title: "GitHub Actions Runner Scale Set (GitHub Cloud)"
    excerpt: "This preset deploys a self-hosted GitHub Actions runner scale set that connects to GitHub.com. Runners scale from 0 to 10 based on queued workflow jobs, using ephemeral pods that are created per job..."
---

# GHA Runner Scale Set Presets

Ready-to-deploy configuration presets for GHA Runner Scale Set. Each preset is a complete manifest you can copy, customize, and deploy.
