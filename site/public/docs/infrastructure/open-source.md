---
title: "Open Source"
description: "Planton open source is the open-source foundation that provides the Protocol Buffer APIs, IaC modules, and CLI underpinning Planton's Infrastructure."
icon: cloud
order: 90
tags:
  - Infrastructure
  - Open Source
---

# Open Source

Planton open source is the open-source foundation that Planton's Infrastructure is built on. It provides three things: Protocol Buffer API definitions for infrastructure kinds, Infrastructure-as-Code modules (Pulumi, Terraform, and OpenTofu) that provision those kinds, and a standalone CLI for deploying them. The project is developed in the open at [planton.dev](https://planton.dev).

The relationship between Planton open source and the Planton platform is architectural: the open-source core defines **how** to deploy infrastructure — one manifest at a time, or a whole set of manifests in dependency order with each resource's outputs resolving the next resources' references. Planton adds **workflow and governance** — parallelized orchestration with approval gates via Infra Charts and Pipelines, credential management via Connections, team collaboration via the web console, and deployment policies via Flow Control. You can use Planton open source without the platform; the platform uses it under the hood.

## Three Pillars

### Protocol Buffer APIs

Every catalog kind in Planton open source is defined as a Protocol Buffer API following the Kubernetes Resource Model (KRM). Each catalog kind defines:

- **api.proto** — The resource structure with `apiVersion`, `kind`, `metadata`, `spec`, and `status`
- **spec.proto** — The configuration schema with provider-specific fields and validation rules
- **iac_input.proto** — The input contract for IaC modules
- **outputs.proto** — The output contract from IaC modules

These protobuf definitions are the canonical source of truth for Catalog Kinds. Planton imports them directly — the resource kind taxonomy, the provider classifications, and the cross-resource dependency mechanism all originate in the open-source core.

### IaC Modules

Each catalog kind includes one or both of:

- **Pulumi module** (Go) — A Pulumi program that provisions the infrastructure
- **Terraform module** (HCL) — A Terraform configuration that provisions the infrastructure

Modules are organized by cloud provider:

| Provider | Kinds |
|----------|-----------|
| AWS | VPC, RDS, EKS, ECS, Lambda, S3, ALB, Route 53, and more |
| GCP | GKE, Cloud SQL, Cloud Run, VPC, Cloud Functions, GCS, and more |
| Azure | AKS, SQL Database, Storage Account, and more |
| Kubernetes | PostgreSQL, Redis, Kafka, MongoDB, Neo4j, NATS, Keycloak, Istio, cert-manager, Prometheus, and many more |
| Other | Cloudflare, DigitalOcean, Auth0, OpenFGA |

### CLI

The Planton CLI (`planton`) provides standalone infrastructure deployment without requiring the platform:

- **validate** — Validate a catalog object manifest against its protobuf schema
- **pulumi** — Run Pulumi operations (up, preview, destroy, refresh)
- **tofu** — Run OpenTofu/Terraform operations (apply, plan, destroy, refresh)
- **apply** — Deploy infrastructure from a manifest, a directory of manifests, or a kustomize overlay; multi-manifest sets deploy in dependency order behind a preflight report that verifies everything verifiable before the first IaC handoff
- **destroy** — Tear down deployed infrastructure
- **plan** — Preview changes before applying

## Open source vs the platform

The boundary between Planton open source and the Planton platform is clear:

| Concern | Planton open source | Planton platform |
|---------|---------------------|------------------|
| Resource definitions | Protobuf APIs for each catalog kind | Imports the open-source APIs, adds platform metadata |
| Provisioning | IaC modules (Pulumi/Terraform/OpenTofu) | Infra Jobs that execute IaC modules via Runner |
| Dependencies | `ValueFromRef` references, resolved across a manifest set from captured outputs | Infra Charts and DAG-based Infra Pipelines |
| Credentials | Local environment variables or config files | Connect (managed credential storage and resolution) |
| Orchestration | Multi-manifest sets in dependency order, sequential, preflight-verified | Parallelized DAG orchestration with approval gates, history, and drift detection |
| Collaboration | CLI-only, single user | Web console, teams, audit trails, RBAC |
| State management | Local or configured backend; encryption is the engine's own setting (`TF_ENCRYPTION`, a Pulumi passphrase) | Managed state backends with multi-tenant isolation; every backend names the key its state is encrypted with |
| Governance | None | Flow Control policies, deployment security tiers |

## How Planton Uses the Open-Source Core

Planton imports the open-source core at the API layer:

- **Catalog Kinds** — The resource kind taxonomy defines every resource type Planton can deploy
- **Catalog Providers** — The provider taxonomy identifies supported cloud platforms
- **Cross-resource dependencies** — The dependency mechanism enables resources in Infra Charts to reference outputs of other resources
- **IaC provisioner selection** — The provisioner taxonomy determines whether Pulumi, Terraform, or OpenTofu executes the deployment
- **Kind APIs** — Each provider-specific component API defines the configuration structure stored in an Infra Component's spec

When an Infra Job executes, it uses the open-source IaC module corresponding to the Catalog Kind. The module receives an IaC input (derived from the Infra Component spec), provisions the infrastructure, and returns outputs that are stored in the resource status.

## Catalog Kind Structure

Each open-source catalog kind follows a standard directory structure:

```
<provider>/<kind>/v1/
  api.proto              # KRM resource definition
  spec.proto             # Configuration schema
  iac_input.proto      # IaC module input contract
  outputs.proto    # IaC module output contract
  README.md              # Kind documentation
  examples.md            # Usage examples
  iac/
    pulumi/              # Pulumi module (Go)
    tf/                  # Terraform module (HCL)
```

This structure ensures every kind is self-contained with its API definition, documentation, and implementation co-located.

<!-- SCREENSHOT: Planton open-source catalog on planton.dev
  Page: https://planton.dev (or equivalent catalog page)
  Action: Show the kind catalog browsing experience
  Focus: Provider categories and kind list
  Alt: Planton open-source catalog showing catalog kinds organized by provider
-->

## No Vendor Lock-In

Because Planton open source is open-source and independent of the platform:

- Infrastructure definitions (protobuf APIs) are portable and version-controlled
- IaC modules work with standard Pulumi, Terraform, and OpenTofu toolchains
- The Planton CLI can deploy resources without any platform dependency
- Organizations can migrate between Planton open source standalone and the platform at any time

Planton adds value through orchestration, governance, and collaboration — not by locking infrastructure definitions into a proprietary format.

## Related Documentation

- [Infra Components](/docs/infrastructure/infra-components) — Deployed instances of open-source components
- [Catalog Kinds](/docs/infrastructure/catalog-kinds) — The taxonomy defined by Planton open source
- [Infra Charts](/docs/infrastructure/infra-charts) — Composing open-source components into templates
- [Infra Jobs](/docs/infrastructure/infra-jobs) — How open-source IaC modules are executed
