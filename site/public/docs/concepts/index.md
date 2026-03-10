---
title: "Concepts"
description: "The core ideas behind OpenMCF: how a multi-cloud deployment framework brings Kubernetes-style consistency to infrastructure across 17 cloud providers"
icon: "lightbulb"
order: 10
---

# Concepts

OpenMCF is a multi-cloud deployment framework that brings Kubernetes-style consistency to infrastructure provisioning across any cloud provider. It is built on three foundational ideas: Protocol Buffer APIs define the resource model, dual IaC engines (Pulumi and OpenTofu/Terraform) implement the deployments, and a Go CLI orchestrates the entire workflow.

This section explains the core concepts that make the framework work.

## The Problem

Deploying infrastructure across multiple cloud providers is fragmented. AWS has its own resource model, its own CLI, its own state management. GCP has a different one. Azure, another. Kubernetes, yet another. Teams that operate across providers end up with different tools, different patterns, different validation models, and different expertise requirements for each platform.

The common response is to build an abstraction layer -- a `GenericDatabase` that works on every cloud. But abstractions leak. A generic database component either exposes the lowest common denominator (losing provider-specific capabilities like AWS RDS read replicas, GCP Cloud SQL private IP allocation, or Azure Managed Identity integration) or creates a fragile mapping layer that breaks when providers change.

## The OpenMCF Approach

OpenMCF takes a different path: consistency of structure and workflow, not abstraction of capability.

Every resource across every provider follows the same manifest format (the Kubernetes Resource Model), uses the same validation framework (Protocol Buffers with buf-validate), is deployed with the same CLI commands, and is managed through the same module and state systems. But the spec -- the actual configuration surface -- is provider-specific. An `AwsS3Bucket` exposes the full S3 feature set. A `GcpGcsBucket` exposes the full GCS feature set. Neither pretends to be the other.

The result: you learn one set of tools and one workflow pattern, then apply it to 360+ deployment components across 17 cloud providers. The learning curve is the framework itself, not each provider's toolchain.

## Key Concepts

### Deployment Components

A deployment component is the atomic unit of OpenMCF -- a self-contained package combining a Protocol Buffer API definition, dual IaC module implementations (Pulumi and Terraform), and auto-generated documentation. OpenMCF ships with 360+ components spanning 17 providers.

Every component follows the same four-file protobuf contract: `api.proto` (resource envelope), `spec.proto` (configuration surface), `stack_input.proto` (IaC input), and `stack_outputs.proto` (IaC output).

**[Read more: Deployment Components](deployment-components)**

### Manifests

OpenMCF manifests use the Kubernetes Resource Model: `apiVersion`, `kind`, `metadata`, `spec`, `status`. The manifest is the single source of truth for what you want to deploy. Metadata labels configure the IaC engine and state backend. The spec holds provider-specific configuration, with every field defined by protobuf and validated before deployment.

**[Read more: Manifests](manifests)**

### Cloud Resource Kinds

The `CloudResourceKind` enum is the canonical registry of everything OpenMCF can deploy -- 360+ entries mapped to 17 providers. Each kind maps to a provider, an API version, a module path, and a validation schema. The kind name in your manifest is the key that drives the entire deployment pipeline.

**[Read more: Cloud Resource Kinds](cloud-resource-kinds)**

### Validation

OpenMCF validates manifests at three layers: schema-level rules defined in protobuf (constant enforcement, required fields, patterns, CEL expressions), CLI-side validation using the protovalidate library, and cloud provider API validation during deployment. The first two layers catch the vast majority of errors before any cloud API call is made.

**[Read more: Validation](validation)**

### Dual IaC Engines

Every component ships with both a Pulumi module (Go) and an OpenTofu/Terraform module (HCL). Both receive the same stack input and produce the same outputs. You choose your engine based on team preference -- Pulumi for programmatic Go workflows, OpenTofu/Terraform for HCL familiarity. The CLI handles the orchestration for either path.

**[Read more: Dual IaC Engines](dual-iac-engines)**

### Module System

The CLI resolves IaC modules through a priority chain: local directory, pre-built binary, zip download, or staging repository. Modules are cached locally, workspace-isolated per deployment, and version-pinnable. The `checkout`, `pull`, and `modules-version` commands manage the local module cache.

**[Read more: Module System](module-system)**

### State Management

Deployment state is tracked by the IaC engine's backend. Pulumi supports Pulumi Cloud, S3, GCS, Azure Blob, and local backends -- configured via manifest labels. OpenTofu/Terraform supports S3, GCS, Azure Storage, and local backends, with S3-compatible backends (R2, MinIO) also supported. State backend configuration is embedded in the manifest, not in separate configuration files.

**[Read more: State Management](state-management)**

## Architecture Overview

For a visual guide to how these concepts connect -- the deployment flow, the component anatomy, and the three-layer architecture -- see the **[Architecture](architecture)** page.

## Getting Started

If you are ready to deploy your first resource, head to the **[Getting Started](/docs/getting-started)** guide. If you want to browse what OpenMCF can deploy, explore the **[Component Catalog](/docs/catalog)**.
