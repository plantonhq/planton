---
title: "Concepts"
description: "Core concepts and architecture of OpenMCF"
icon: "lightbulb"
order: 3
---

# Core Concepts

Understanding the core concepts of OpenMCF will help you make the most of the framework.

## Overview

OpenMCF is built on three foundational pillars:

1. **APIs** - Standardized configuration schema using Protocol Buffers
2. **IaC Modules** - Pre-built Pulumi and Terraform modules
3. **CLI** - Orchestration layer that brings everything together

## Key Concepts

### Kubernetes Resource Model (KRM)

Every deployment component follows the same structure:

```yaml
apiVersion: <provider>.openmcf.org/<version>
kind: <ComponentType>
metadata:
  name: <resource-name>
  org: <organization>
  env: <environment>
spec:
  # Provider-specific configuration
status:
  # System-managed status (read-only)
```

This consistent structure makes it easy to understand and work with any resource type.

### Protocol Buffers

Unlike Kubernetes (which uses Go structs), OpenMCF uses Protocol Buffers to enable:

- **Language Neutrality**: Auto-generate SDKs in Go, Java, Python, TypeScript
- **Beautiful Documentation**: Publish to Buf Schema Registry
- **Field-Level Validations**: Define validation rules directly in the API schema
- **Early Error Detection**: Catch configuration errors before deployment

### Deployment Components

A deployment component is a self-contained unit that includes:

- **API Definition**: Protocol Buffer schema defining configuration
- **Pulumi Module**: Infrastructure-as-code in a real programming language
- **Terraform Module**: Alternative IaC implementation
- **Documentation**: Auto-generated from protobuf definitions

Examples:
- `PostgresKubernetes` - Deploy PostgreSQL to any Kubernetes cluster
- `AwsRdsInstance` - Deploy to AWS RDS
- `GcpCloudSql` - Deploy to Google Cloud SQL

### Provider-Specific vs. Abstract

OpenMCF **does NOT** abstract away cloud provider differences. This is intentional.

Each cloud provider has different capabilities, pricing models, and operational characteristics. Attempting to abstract these would either:

1. Force a "lowest common denominator" approach (losing provider-specific capabilities)
2. Create a leaky abstraction that's harder to understand

Instead, OpenMCF provides:
- ✅ **Consistent structure**: Every resource uses KRM
- ✅ **Consistent workflow**: Same CLI commands, same validation
- ✅ **Consistent developer experience**: Same documentation approach
- ✅ **Provider-specific manifests**: Each deployment target has its own manifest

## Learn More

- **[Architecture](architecture)** - Deep dive into the technical architecture
- **[Deployment Components](/docs/catalog)** - Explore available components
- **[Getting Started](../getting-started)** - Deploy your first resource

