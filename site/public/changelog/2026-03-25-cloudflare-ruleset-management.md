---
title: "Cloudflare Ruleset Management in Infra Hub"
date: 2026-03-25
category: feature
tags:
  - infra-hub
excerpt: "You can now create and manage Cloudflare Rulesets — origin rules, cache rules, WAF rules, redirects, and more — directly from Infra Hub."
author:
  - name: Swarup Donepudi
    title: Founder
---

Cloudflare Rulesets are now a first-class infra component in Infra Hub. You can create, configure, and deploy rulesets that control how Cloudflare handles traffic for your domains — origin routing, caching, security, request transforms, and more — all managed through the same workflow you use for every other infra component on the platform.

## What You Can Do

When you create a CloudflareRuleset resource, you pick a **phase** that determines where in Cloudflare's request pipeline your rules execute. The platform supports all 24 Cloudflare phases, organized into six categories:

- **Origin & Routing** — Override origin servers, set up dynamic redirects, or manage bulk redirect lists
- **Security & Firewall** — Create custom WAF rules, enable managed rulesets (OWASP, etc.), configure rate limiting, or set up bot fight mode
- **Caching** — Control edge and browser cache behavior per request pattern
- **Transform** — Rewrite URLs, modify request or response headers, and configure compression
- **Configuration** — Set HTTPS rewrites, custom error pages, and custom log fields
- **DDoS & Network** — Layer 4 and Layer 7 DDoS protection rules, Magic Transit rules

Each phase supports specific actions (block, challenge, redirect, rewrite, route, cache settings, and others). The create wizard enforces these constraints — you only see valid actions for the phase you selected, so you cannot accidentally build an invalid configuration.

## How It Works

CloudflareRuleset follows the standard Infra Hub workflow:

1. Navigate to **Infra Hub → Create Infra Component → Cloudflare → Ruleset**
2. Select a **ruleset kind** (Zone for a single domain, Custom for reusable rules, Managed for Cloudflare-maintained rules, or Root for account-level entry points)
3. Choose the **phase** that matches what you want to control
4. Define one or more **rules** with wirefilter expressions and actions
5. Deploy — the platform provisions the ruleset through Cloudflare's API

Both Pulumi and Terraform modules are available. The platform selects the appropriate module based on your organization's IaC provisioner configuration, or you can override it per resource.

## Why This Matters

- **Manage edge logic alongside your infrastructure.** Origin rules, cache policies, and WAF rules live in the same platform as your clusters, databases, and services — with the same version history, access control, and deployment pipeline.
- **Guardrails built in.** The phase-to-action validation prevents misconfiguration before deployment, not after.
- **Full Cloudflare coverage.** All 24 phases and 13 action types are supported, from simple origin routing to DDoS protection and Magic Transit rules.
