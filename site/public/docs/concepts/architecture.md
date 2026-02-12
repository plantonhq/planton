---
title: "Architecture"
description: "Technical architecture and design of OpenMCF"
icon: "gear"
order: 1
---

# Architecture

OpenMCF is built on three foundational components that work together seamlessly.

## The Three Pillars

### 1. APIs: Standardized Configuration Schema

**Technology**: Protocol Buffers  
**Inspiration**: Kubernetes Resource Model

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

**Why Protocol Buffers?**

Unlike Kubernetes (which uses Go structs), OpenMCF uses Protocol Buffers to enable:

- **Language Neutrality**: Auto-generate SDKs in Go, Java, Python, TypeScript
- **Beautiful Documentation**: Publish to Buf Schema Registry
- **Field-Level Validations**: Define validation rules directly in the API schema
- **Early Error Detection**: Catch configuration errors before deployment
- **Platform Engineering**: Import SDKs to build custom internal tools

Example validation in protobuf:

```protobuf
message PostgresKubernetesSpec {
  string cpu = 1 [(buf.validate.field).string.pattern = "^[0-9]+m$"];
  int32 replicas = 2 [(buf.validate.field).int32 = {gte: 1, lte: 10}];
}
```

The `openmcf validate` command checks these rules **before** calling any cloud APIs.

### 2. IaC Modules: The "Recipes"

**Technology**: Pulumi and Terraform/OpenTofu  
**Approach**: Provider-specific, deliberately simple

Every deployment component has **both** a Pulumi module and a Terraform module. You choose which IaC engine to use.

**Why Both Pulumi and Terraform?**

Different teams have different preferences:

- **Pulumi**: Real programming languages (Go, Python, TypeScript), better for complex logic, type safety
- **Terraform/OpenTofu**: Mature ecosystem, HashiCorp Configuration Language, familiar to many DevOps teams

OpenMCF doesn't force a choice—it supports both.

**Design Philosophy: Deliberately Simple**

The default modules are intentionally designed to be **Terraform-like** even when written in Pulumi:

- Simple, straightforward code
- Single directory structure (like Terraform modules)
- Familiar file names (`main.go` similar to `main.tf`)
- Minimal language features

**Why?** Because **adoption matters more than perfect code**. A Terraform engineer should be able to fork a Pulumi module and immediately understand the flow.

### 3. CLI: The Orchestration Layer

**Distribution**: Homebrew  
**Role**: The "chef" that brings everything together

Installation:

```bash
brew install plantonhq/tap/openmcf
```

**What the CLI does:**

1. Reads your manifest (local file or GitHub raw URL)
2. Validates inputs using proto-validate rules
3. Maps `kind` to IaC module
4. Clones/pulls the module from GitHub (with smart caching)
5. Sets up the environment (exports manifest for the module)
6. Delegates to IaC engine (Pulumi or Terraform/OpenTofu)
7. Streams output to the developer

**Core commands:**

```bash
# Validate a manifest
openmcf validate -f postgres.yaml

# Deploy with Pulumi
openmcf pulumi up -f postgres.yaml --stack org/project/env

# Deploy with Terraform/OpenTofu
openmcf tofu apply -f postgres.yaml

# Override specific values
openmcf pulumi up \
  -f postgres.yaml \
  --set spec.container.cpu=500m \
  --stack org/project/env
```

## The Complete Workflow

Here's how a developer deploys Redis to Kubernetes:

### Step 1: Browse Available Components

Visit the OpenMCF repository to find deployment components.

**Example**: "RedisKubernetes" - deploys Redis to any Kubernetes cluster

### Step 2: Explore the API

Visit Buf Schema Registry where APIs are published to see:
- Required fields
- Optional fields with defaults
- Field-level documentation
- Validation rules

### Step 3: Write Your Manifest

Create `my-redis.yaml`:

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: RedisKubernetes
metadata:
  name: session-store
  org: acme
  env: production
spec:
  container:
    replicas: 3
    resources:
      limits:
        cpu: 500m
        memory: 1Gi
      requests:
        cpu: 250m
        memory: 512Mi
    isPersistenceEnabled: true
    diskSize: 20Gi
```

### Step 4: Validate (Optional)

```bash
openmcf validate -f my-redis.yaml
```

### Step 5: Deploy

```bash
openmcf pulumi up -f my-redis.yaml --stack acme/platform/prod
```

**What happens under the hood:**

1. CLI validates the manifest
2. CLI identifies this is `RedisKubernetes`
3. CLI clones/pulls the Redis Kubernetes module
4. CLI exports your manifest as an environment variable
5. CLI runs `pulumi up`
6. You see progress in your terminal
7. Redis is deployed to your Kubernetes cluster

## Module Distribution

**Default modules:**
- Hosted on GitHub (open source)
- Versioned with Git tags
- Cached locally in `~/.openmcf/`
- Updated via `git pull` on demand

**Custom modules:**
- Point CLI to your own GitHub repository
- Use private repos with SSH authentication
- Override module URLs via CLI flags

## State Management

**With Pulumi:**
- Supports multiple backends: Local, S3, GCS, Azure Blob, Pulumi Cloud
- State tracks deployed resources
- Enables detecting configuration drift
- Allows previewing changes before applying

**With Terraform:**
- Supports multiple backends: Local, S3, GCS, Azure Storage
- State file management via standard Terraform workflow
- Remote state sharing for team collaboration

## Validation Architecture

**Three layers of validation:**

1. **Proto-level validation** (schema definition):
   ```protobuf
   string cpu = 1 [(buf.validate.field).string.pattern = "^[0-9]+m$"];
   ```

2. **CLI validation** (before deployment):
   ```bash
   openmcf validate -f config.yaml
   ```

3. **Cloud provider validation** (during deployment):
   - Final validation by the actual cloud provider APIs
   - Catches provider-specific constraints

This layered approach catches 90%+ of errors before making any cloud API calls.

## Learn More

- [Getting Started](../getting-started) - Deploy your first resource
- [Deployment Components](/docs/catalog) - Explore available components

