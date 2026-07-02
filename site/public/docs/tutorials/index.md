---
title: "Tutorials"
description: "Step-by-step walkthroughs for deploying infrastructure with Planton — from your first resource to multi-provider workflows"
icon: "tutorial"
order: 40
---

# Tutorials

Tutorials are linear, end-to-end walkthroughs. Each one takes you from an empty directory to a deployed resource and back, covering the full lifecycle: write a manifest, preview the plan, deploy, verify, modify, and destroy.

If you are looking for focused reference material on a specific topic (credentials, state backends, Kustomize configuration), see [Guides](../guides). If you want to understand the concepts behind Planton's design, see [Concepts](../concepts).

## Before You Start

All tutorials use the Planton CLI; everything shown can also be done from the desktop app. If you have not set either up yet, start with [Getting Started](../getting-started).

## Tutorials

### [Deploy Your First AWS Resource](./first-aws-resource)

Deploy an S3 bucket to AWS — write a manifest, preview, apply, modify with a lifecycle rule, and destroy. Covers the `AwsS3Bucket` component and the core `plan` -> `apply` -> `destroy` workflow.

**You need**: AWS credentials, Pulumi or OpenTofu CLI

### [Deploy Your First Kubernetes Resource](./first-kubernetes-resource)

Deploy a production-oriented PostgreSQL database on Kubernetes with custom databases, named users, resource tuning, and persistent storage. Covers connecting to the database via port-forwarding and modifying deployments with `--set` overrides.

**You need**: A Kubernetes cluster, Pulumi CLI

### [Multi-Environment Deployments](./multi-environment)

Deploy the same PostgreSQL component to dev, staging, and production with different resource sizing using Kustomize overlays. Covers directory structure, strategic merge patches, and the `--kustomize-dir` / `--overlay` flags.

**You need**: A Kubernetes cluster, Pulumi CLI

### [Deploy Across Providers](./multi-provider)

Deploy object storage on both AWS (S3) and GCP (GCS) to see Planton's consistent cross-provider workflow in action. Same CLI commands, same manifest structure, different `spec` fields.

**You need**: AWS credentials, GCP credentials, Pulumi CLI

## Suggested Order

Start with **First AWS Resource** or **First Kubernetes Resource** depending on which provider you have access to. Then move to **Multi-Environment** to learn Kustomize overlays, and finish with **Deploy Across Providers** to see the cross-provider pattern.
