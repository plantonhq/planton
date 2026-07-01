---
title: "Desktop App and CLI"
description: "How Planton works as both a desktop app and a CLI — a local instance on your machine, and two commands that mirror kubectl and Helm"
icon: "gear"
order: 5
---

# Desktop App and CLI

Planton is both a **desktop app** and a **CLI**. They are two front doors to the same engine: click through the app, or drive it from your terminal — either way, Planton runs proven, pre-built infrastructure-as-code modules against your own cloud, with real state and history.

## A local instance on your machine

There is no account and no hosted control plane to connect to. When you open the desktop app for the first time, Planton brings a **local instance** online on your own machine — a local control plane that stores your configuration and orchestrates deployments locally. Its data lives under `~/.planton` on your computer.

Because it runs where your credentials already live, Planton detects the cloud you're already signed into and gets out of your way — no "connections" to configure to get started.

## Two commands, never interchangeable

Kubernetes gave developers two of the best experiences in infrastructure — `kubectl apply` for a single manifest, and Helm charts for a whole stack — and locked both to Kubernetes. Planton frees both, for every cloud, as two distinct commands:

### `planton apply -f` — one component

```bash
planton apply -f bucket.yaml
```

Applies a **single** component from one manifest. This is the direct parallel to `kubectl apply -f`. The CLI resolves the module for that kind, runs it locally, and streams the live output to your terminal as the resource is created.

### `planton chart install` — a whole environment

```bash
planton chart install aws-ecs --name api --env dev --values values.yaml
```

Installs a **whole environment** from a chart — many resources wired together — the parallel to `helm install`. Charts are, honestly, "Helm charts for infrastructure."

Keep the two straight: `apply -f` is always one manifest; `chart install` is always an environment. Do not reach for `apply -f` to stand up a multi-resource stack.

## Manifests are KRM

Every manifest uses the Kubernetes Resource Model — `apiVersion`, `kind`, `metadata`, `spec` — extended from Kubernetes to every cloud (for example `aws.planton.dev/v1`). If you can read a Kubernetes manifest, you can read a Planton manifest.

## Preview, then live progress

The desktop app shows a **read-only architecture preview** built from your chart or manifest *before* you deploy — so you can see the shape of what you're about to create. During the deploy, **live progress** lights each piece up as it comes online. These are two distinct moments: a preview beforehand, and progress during.

## Proven, vetted modules — not written for you

Planton does not write your infrastructure-as-code. It **runs pre-built modules** that are already written and vetted for secure, well-architected, cost-efficient infrastructure, backed by Terraform and Pulumi. Writing IaC was never the hard part — trusting it is. Your configuration is stored and versioned, every change is a diff, and you can export it and run it yourself anytime. Nothing locks you in.

## Where it deploys

Planton deploys to **AWS, GCP, Azure, and Kubernetes**. The broader catalog spans 17 cloud providers of components you can browse and configure.

## App or CLI — same engine

Use the app when you'd rather click and see the architecture; use the CLI when you'd rather script it or keep it in your terminal. Both drive the same local instance and the same modules, so you can move between them freely.
