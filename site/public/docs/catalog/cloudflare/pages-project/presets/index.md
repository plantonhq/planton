---
title: "Presets"
description: "Ready-to-deploy configuration presets for Pages Project"
type: "preset-list"
componentSlug: "pages-project"
componentTitle: "Pages Project"
provider: "cloudflare"
icon: "package"
order: 200
presets:
  - slug: "01-git-connected-site"
    rank: "01"
    title: "Preset: Git-Connected Site"
    excerpt: "A Pages project connected to a GitHub (or GitLab) repository. Cloudflare runs your build and creates a new deployment on every push — Cloudflare is the CI."
  - slug: "02-direct-upload-site"
    rank: "02"
    title: "Preset: Direct-Upload Site"
    excerpt: "A Pages project with no git connection. You build the site in your own CI and push it with `wrangler pages deploy`, while the project, its bindings, and its domains are managed declaratively here."
---

# Pages Project Presets

Ready-to-deploy configuration presets for Pages Project. Each preset is a complete manifest you can copy, customize, and deploy.
