---
title: "Presets"
description: "Ready-to-deploy configuration presets for Zero Trust Access Policy"
type: "preset-list"
componentSlug: "zero-trust-access-policy"
componentTitle: "Zero Trust Access Policy"
provider: "cloudflare"
icon: "package"
order: 200
presets:
  - slug: "01-allow-staff"
    rank: "01"
    title: "Preset: Allow staff"
    excerpt: "A simple `allow` policy that grants access to anyone with a corporate email domain, connecting from an allowed country, with a 24-hour session."
  - slug: "02-admins-with-approval"
    rank: "02"
    title: "Preset: Admins with approval and MFA"
    excerpt: "An `allow` policy for privileged access: it references a reusable admins group, requires explicit approval and a purpose justification, enforces hardware-key MFA, and uses a short 1-hour session."
---

# Zero Trust Access Policy Presets

Ready-to-deploy configuration presets for Zero Trust Access Policy. Each preset is a complete manifest you can copy, customize, and deploy.
