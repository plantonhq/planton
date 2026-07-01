---
title: "Getting Started"
description: "Download Planton, open it, and deploy real infrastructure to your own cloud — from the app or the CLI"
icon: "rocket"
order: 2
---

# Getting Started

Planton is a free desktop app and CLI for your cloud infrastructure. You download it, open it, and deploy to your own cloud — with clean, auditable infrastructure-as-code running underneath. No account, no sign-up, no ceremony.

This page gets you from zero to a first deployment with the **desktop app** — the primary way to use Planton — and then shows how the **CLI** drives the same engine when you'd rather stay in the terminal.

## Deploy with the desktop app

The fastest way to see Planton work.

1. **Download Planton** for your platform from [planton.dev/download](/download), and open it. On first launch, Planton starts a local instance on your machine — a local control plane that runs entirely on your own hardware. There is nothing to sign up for.
2. **It finds the cloud you're already signed into.** Planton detects your existing local credentials (for example, the AWS profile you already use) — no connection to configure.
3. **Pick a stack and fill a short form.** Choose a ready-made stack or a single component, fill in only what's truly required (smart defaults handle the rest), and review a plain summary plus the exact manifest that will be applied.
4. **Click deploy, and watch it come online.** See the architecture before you deploy, then watch each resource light up as it is created — real infrastructure-as-code, stored and versioned, every change a diff.

The desktop app deploys to **AWS, GCP, Azure, and Kubernetes**.

## Or drive it from the CLI

The `planton` CLI is a companion to the desktop app — it drives the same engine and the same proven modules. The managed backend (state, ready-made charts, and history) comes from Planton itself, so most people run the app and reach for the CLI when they'd rather stay in the terminal.

It mirrors the two gestures Kubernetes made great — and frees them for every cloud. There are two commands, and they are never interchangeable:

- **`planton apply -f <manifest>`** applies a **single** component from one manifest — the `kubectl apply -f` parallel.
- **`planton chart install <chart> …`** installs a **whole environment** from a chart — the `helm install` parallel.

### Install the CLI

```bash
go install github.com/plantonhq/planton@latest
```

Verify the installation:

```bash
planton version
```

### Apply a single manifest

Every Planton manifest follows the Kubernetes Resource Model (KRM) — the same `apiVersion`, `kind`, `metadata`, `spec` shape you already know, extended to every cloud. Create a file named `bucket.yaml`:

```yaml
apiVersion: aws.planton.dev/v1
kind: AwsS3Bucket
metadata:
  name: my-first-bucket
spec:
  awsRegion: us-east-1
```

Deploy it:

```bash
planton apply -f bucket.yaml
```

The CLI resolves the proven, pre-built module for `AwsS3Bucket`, runs it locally, and **streams the live output to your terminal** as the resource is created — genuinely live, not fire-and-forget.

### Install a whole environment

A chart installs a whole environment in one command — many resources, wired together:

```bash
planton chart install aws-ecs --name api --env dev --values values.yaml
```

You'll watch every resource in the environment come up, live, as it happens.

## What's running underneath

Whichever path you take, Planton **runs proven, pre-built infrastructure-as-code modules** for you — it does not ask you to write them. Each module is backed by Terraform or Pulumi, with secure, cost-aware, well-architected defaults baked in. Your configuration is stored and versioned, every change is a diff, and you can export it and run it yourself at any time. Nothing locks you in.

## Next steps

- **Understand the model.** Read [Desktop App and CLI](/docs/concepts/desktop-and-cli) for how the two commands map to Kubernetes, and [Core Concepts](/docs/concepts) for the resource model.
- **Browse the catalog.** Explore the [components](/docs/catalog) across 17 cloud providers, each with a guided form and ready-made presets.
- **Go deeper.** See [Manifests](/docs/concepts/manifests), [Dual IaC Engines](/docs/concepts/dual-iac-engines), and [State Management](/docs/concepts/state-management).
- **Troubleshoot.** Check the [Troubleshooting Guide](/docs/troubleshooting) if you run into problems.
