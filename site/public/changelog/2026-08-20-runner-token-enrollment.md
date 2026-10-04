---
title: "Runner Tokens: Enroll Runners with One Command"
date: 2026-08-20
category: feature
tags:
  - runner
  - security
  - console
  - cli
excerpt: "Create a named runner token once and start runners anywhere with it — each runner registers itself on arrival and receives its own individually revocable identity. No credential files to generate, download, or mount."
author:
  - name: Swarup Donepudi
    title: Founder
---

Getting a runner online is now two commands. Create a named runner token once, then start runners with it — on a laptop, a VM, or any of the supported cloud targets. Each runner enrolls itself the moment it starts: it registers with Planton on arrival and receives its own identity, minted for it alone and delivered exactly once, to the runner itself. There is no credential file to generate, download, or move between machines.

```bash
planton runner token create prod-fleet
planton-runner start --token prt_... --name prod-a
```

The runner appears in your console (Organization Settings → Runners) the moment it joins.

## The Token Is Never an Identity

A runner token authorizes joining your organization — nothing more. Every runner admitted by a token still receives its own separate identity, so the security model stays honest under real operations:

- **Revoke a token** and the door closes for new joins — runners it already admitted keep their own identities and keep working until you revoke them individually
- **Revoke one runner** without touching its siblings or the token that admitted them
- **A leaked token cannot impersonate an existing runner** — each runner's registration records which token admitted it, and only that token can re-admit it (after a lost disk, for example); anything else is refused
- **Attribution is built in** — every arrival is tied to the named token that authorized it, so you always know how a runner got in

## Manage Tokens from the Console

Organization Settings has a new RUNNERS section on both the web console and the desktop app: **Runners** lists your fleet with each runner's capabilities and arrival details, and **Runner Tokens** is where you create tokens (revealed exactly once, with expiry options) and revoke them.

## Deploy Runners into Your Own Cloud

The Runners page funnels straight into deployment: pick AWS (ECS Fargate), GCP (Cloud Run), Azure (Container Apps), or Kubernetes (Helm), and the matching wizard walks you through deploying a runner as a first-class Infra Component. On Planton, the platform mints the join token server-side into the deployment's secret store — declaring a runner is genuinely one step, and the deployed runner enrolls itself on first boot.

The same is available from the terminal: `planton runner deploy` ships a token into the target's secret store and deploys the runner container for any of the four targets.

## Self-Hosted Installations Go Keyless

A self-hosted installation's own in-cluster runner needs no token at all: your cluster itself vouches for it on every call, with rotation handled by Kubernetes. There is no stored runner credential to steal — and nothing to rotate manually, ever.

## Why This Matters

- **One secret, clearly scoped** — the only thing you handle is a named, revocable join token; identities are born where the runners run and never transit your machine
- **Fleet-ready** — start any number of runners with one token, each individually visible, attributable, and revocable
- **Recovery without ceremony** — a runner that loses its disk simply rejoins with its token; a runner that loses everything is recovered with one reset command
- **Every path covered** — interactive starts, container deployments, managed deploys, and declarative infrastructure all speak the same token contract
