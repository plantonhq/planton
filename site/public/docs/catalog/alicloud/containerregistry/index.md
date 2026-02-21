---
title: "ContainerRegistry"
description: "ContainerRegistry deployment documentation"
icon: "package"
order: 100
componentName: "alicloudcontainerregistry"
---

# AliCloud ContainerRegistry

Deploy an Alibaba Cloud Container Registry (ACR) Enterprise Edition instance with namespaces for organizing container images.

## What It Does

AliCloudContainerRegistry provisions a managed container image registry on Alibaba Cloud with enterprise-grade security, scalable storage, and optional VPC-internal access for fast, cost-free image pulls from within your network.

## When to Use

- You need a private container image registry on Alibaba Cloud
- Your Kubernetes clusters (ACK) need a registry to pull images from
- You want to organize images by team or application using namespaces
- You need both internet and VPC-internal endpoints for flexible access

## Configuration Highlights

| Field | Required | Description |
|-------|----------|-------------|
| `region` | Yes | Alibaba Cloud region |
| `instanceName` | Yes | Registry instance name |
| `instanceType` | Yes | Tier: Basic, Standard, or Advanced |
| `paymentType` | No | Subscription (default) or PayAsYouGo |
| `period` | No | Subscription period in months |
| `password` | No | Registry login password |
| `namespaces` | No | List of namespace configurations |

## Namespace Configuration

Each namespace supports:
- `name` -- Namespace identifier (2-120 characters)
- `autoCreate` -- Auto-create repos on image push (default: false)
- `defaultVisibility` -- PUBLIC or PRIVATE (default: PRIVATE)

## Outputs

After deployment, the following outputs are available:
- **instance_id** -- For referencing in other resources
- **public_endpoint** -- Internet-facing domain for `docker login`
- **vpc_endpoint** -- VPC-internal domain for in-VPC image pulls
- **namespace_ids** -- Map of namespace names to IDs

## Instance Tiers

| Tier | Use Case | Namespaces | Repos |
|------|----------|------------|-------|
| Basic | Individual developers, small teams | 10 | 3,000 |
| Standard | Small and medium enterprises | 50 | 10,000 |
| Advanced | Large enterprises, geo-replication | 100 | 100,000 |

## Cross-Cloud Comparison

| Feature | Alibaba Cloud ACR | Azure ACR | AWS ECR |
|---------|-------------------|-----------|---------|
| Component | AliCloudContainerRegistry | AzureContainerRegistry | AwsEcrRepo |
| Tiers | Basic/Standard/Advanced | Basic/Standard/Premium | N/A (per-repo) |
| Namespaces | Bundled | N/A (flat) | N/A (per-repo) |
| Geo-Replication | Advanced tier | Premium tier | Cross-region replication |

## Related Components

- **AliCloudAckManagedCluster** -- Kubernetes cluster that pulls images from this registry
- **AliCloudAckNodePool** -- Node pools that need access to container images
