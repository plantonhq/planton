# DigitalOcean Container Registry Examples

This document provides copy-paste ready examples for common `DigitalOceanContainerRegistry` use cases.

## Table of Contents

- [Quick Start Examples](#quick-start-examples)
  - [Minimal Configuration](#minimal-configuration)
  - [Production Configuration](#production-configuration)
- [Development and Testing](#development-and-testing)
  - [Personal Development Registry](#personal-development-registry)
  - [Team Development Registry](#team-development-registry)
- [Staging Environments](#staging-environments)
  - [Basic Staging Registry](#basic-staging-registry)
  - [Staging with Garbage Collection](#staging-with-garbage-collection)
- [Production Environments](#production-environments)
  - [Single-Region Production](#single-region-production)
  - [Multi-Region Production](#multi-region-production)
  - [Production with Cost Optimization](#production-with-cost-optimization)
- [Regional Deployment Examples](#regional-deployment-examples)
  - [US East (NYC)](#us-east-nyc)
  - [US West (San Francisco)](#us-west-san-francisco)
  - [Europe (London)](#europe-london)
  - [Europe (Frankfurt)](#europe-frankfurt)
  - [Asia (Singapore)](#asia-singapore)
- [Multi-Environment Setup](#multi-environment-setup)
  - [Complete Dev/Staging/Prod Stack](#complete-devstagingprod-stack)
- [Advanced Patterns](#advanced-patterns)
  - [Microservices Architecture](#microservices-architecture)
  - [Global Application with DR](#global-application-with-dr)
  - [Cost-Optimized Multi-Team Setup](#cost-optimized-multi-team-setup)
- [Integration Examples](#integration-examples)
  - [Using Images in Kubernetes](#using-images-in-kubernetes)
  - [CI/CD Pipeline Integration](#cicd-pipeline-integration)

---

## Quick Start Examples

### Minimal Configuration

The absolute minimum configuration for a development registry:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: my-registry
spec:
  name: my-registry
  subscriptionTier: STARTER
  region: NYC3
```

**Use Case:** Learning DOCR, personal projects, quick experiments.

**Cost:** $0/month (Starter tier)

**Limitations:** 500 MiB storage, 1 repository

---

### Production Configuration

Production-ready configuration with garbage collection:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: prod-registry
spec:
  name: prod-registry
  subscriptionTier: PROFESSIONAL
  region: NYC3
  garbageCollectionEnabled: true
```

**Use Case:** Production workloads requiring reliability and cost control.

**Cost:** $20/month (Professional tier)

**Benefits:** 100 GiB storage, unlimited repositories, automated garbage collection, multi-registry support (up to 10)

---

## Development and Testing

### Personal Development Registry

Free tier for individual developers:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: john-dev-registry
spec:
  name: john-dev-registry
  subscriptionTier: STARTER
  region: NYC3
  garbageCollectionEnabled: true
```

**Rationale:**
- Starter tier keeps costs at $0
- Garbage collection prevents hitting the 500 MiB limit
- Personal namespace avoids conflicts with team registries

**Workflow:**
```bash
# Build and push
docker build -t registry.digitalocean.com/john-dev-registry/my-app:latest .
docker push registry.digitalocean.com/john-dev-registry/my-app:latest
```

---

### Team Development Registry

Shared registry for a small development team:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: team-dev-registry
spec:
  name: team-dev-registry
  subscriptionTier: BASIC
  region: NYC3
  garbageCollectionEnabled: true
```

**Rationale:**
- Basic tier ($5/month) provides 5 GiB storage and 5 repositories
- Sufficient for a team of 5-10 developers
- Garbage collection runs automatically (scheduled by OpenMCF)

**Cost:** $5/month

---

## Staging Environments

### Basic Staging Registry

Minimal staging environment:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: staging-registry
spec:
  name: app-staging
  subscriptionTier: BASIC
  region: NYC3
  garbageCollectionEnabled: false
```

**Use Case:** Staging environment that closely mirrors production configuration but with manual cleanup.

**Cost:** $5/month

---

### Staging with Garbage Collection

Staging with automated cleanup:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: staging-registry
spec:
  name: myapp-staging
  subscriptionTier: BASIC
  region: SFO3
  garbageCollectionEnabled: true
```

**Rationale:**
- Region `SFO3` co-located with staging DOKS cluster in San Francisco
- Garbage collection prevents storage bloat from frequent staging deployments
- Basic tier is cost-effective for pre-production

**Cost:** $5/month

---

## Production Environments

### Single-Region Production

Production registry co-located with a DOKS cluster:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: prod-registry
spec:
  name: myapp-prod
  subscriptionTier: PROFESSIONAL
  region: NYC3
  garbageCollectionEnabled: true
```

**Rationale:**
- Professional tier ($20/month) for production reliability
- 100 GiB storage handles growth
- Garbage collection scheduled for low-traffic windows (e.g., Sunday 3 AM)
- Region matches DOKS cluster for fastest image pulls

**Cost:** $20/month

**Storage Capacity:** 100 GiB included, $0.02/GiB overage

---

### Multi-Region Production

Production registries in multiple regions for global applications:

**US East Registry:**
```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: prod-us-east
spec:
  name: myapp-prod-us-east
  subscriptionTier: PROFESSIONAL
  region: NYC3
  garbageCollectionEnabled: true
```

**US West Registry:**
```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: prod-us-west
spec:
  name: myapp-prod-us-west
  subscriptionTier: PROFESSIONAL
  region: SFO3
  garbageCollectionEnabled: true
```

**Europe Registry:**
```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: prod-europe
spec:
  name: myapp-prod-europe
  subscriptionTier: PROFESSIONAL
  region: LON1
  garbageCollectionEnabled: true
```

**Rationale:**
- Each region has a dedicated registry co-located with local DOKS clusters
- Minimizes image pull latency and future bandwidth costs
- Professional tier supports up to 10 registries on a single subscription

**Total Cost:** $20/month (all registries share the Professional subscription)

**CI/CD Strategy:** Push production images to all three registries; each regional DOKS cluster pulls from its local registry.

---

### Production with Cost Optimization

Production registry with aggressive cost controls:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: cost-optimized-prod
spec:
  name: myapp-prod-optimized
  subscriptionTier: PROFESSIONAL
  region: NYC3
  garbageCollectionEnabled: true
```

**Cost Optimization Practices:**
1. **Enable Garbage Collection:** Removes untagged images automatically (can save 50%+ on storage)
2. **Use Immutable Tags:** Push images with Git SHAs (`v1.2.3-abc123f`) instead of `latest` to make GC predictable
3. **Monitor Storage Usage:** Track consumption via DigitalOcean dashboard or API
4. **Prune Old Images:** Set retention policies in CI/CD (e.g., keep last 30 builds)

**Expected Savings:** With GC enabled, typical production registries see 40-60% reduction in storage usage compared to no GC.

---

## Regional Deployment Examples

### US East (NYC)

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: us-east-registry
spec:
  name: myapp-us-east
  subscriptionTier: PROFESSIONAL
  region: NYC3
  garbageCollectionEnabled: true
```

**Data Center:** New York City, New York

**Use Case:** Primary US region for applications serving North American east coast users.

---

### US West (San Francisco)

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: us-west-registry
spec:
  name: myapp-us-west
  subscriptionTier: PROFESSIONAL
  region: SFO3
  garbageCollectionEnabled: true
```

**Data Center:** San Francisco, California

**Use Case:** US west coast applications, co-located with SFO-based DOKS clusters.

---

### Europe (London)

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: europe-uk-registry
spec:
  name: myapp-eu-uk
  subscriptionTier: PROFESSIONAL
  region: LON1
  garbageCollectionEnabled: true
```

**Data Center:** London, United Kingdom

**Use Case:** GDPR-compliant storage for EU users, UK-based operations.

---

### Europe (Frankfurt)

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: europe-de-registry
spec:
  name: myapp-eu-de
  subscriptionTier: PROFESSIONAL
  region: FRA1
  garbageCollectionEnabled: true
```

**Data Center:** Frankfurt, Germany

**Use Case:** EU data residency requirements, central European operations.

---

### Asia (Singapore)

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: asia-sg-registry
spec:
  name: myapp-asia-sg
  subscriptionTier: PROFESSIONAL
  region: SGP1
  garbageCollectionEnabled: true
```

**Data Center:** Singapore

**Use Case:** Asia-Pacific applications, low-latency image pulls for APAC DOKS clusters.

---

## Multi-Environment Setup

### Complete Dev/Staging/Prod Stack

Deploy all three environments with a single Professional subscription:

**Development:**
```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: myapp-dev
spec:
  name: myapp-dev
  subscriptionTier: STARTER
  region: NYC3
  garbageCollectionEnabled: true
```

**Staging:**
```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: myapp-staging
spec:
  name: myapp-staging
  subscriptionTier: BASIC
  region: NYC3
  garbageCollectionEnabled: true
```

**Production:**
```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: myapp-prod
spec:
  name: myapp-prod
  subscriptionTier: PROFESSIONAL
  region: NYC3
  garbageCollectionEnabled: true
```

**Total Cost:**
- Dev: $0/month (Starter)
- Staging: $5/month (Basic)
- Prod: $20/month (Professional)
- **Total:** $25/month

**Alternative (Cost-Optimized):** Use a single Professional subscription and create dev/staging/prod as separate registries (instead of separate subscriptions), reducing total cost to $20/month.

---

## Advanced Patterns

### Microservices Architecture

Production setup for a microservices platform with multiple teams:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: microservices-prod
spec:
  name: platform-prod
  subscriptionTier: PROFESSIONAL
  region: NYC3
  garbageCollectionEnabled: true
```

**Repository Organization:**
```
registry.digitalocean.com/platform-prod/
  ├── frontend/web-app:v1.2.3
  ├── frontend/mobile-api:v2.0.1
  ├── backend/auth-service:v1.5.0
  ├── backend/payment-service:v1.8.2
  ├── backend/notification-service:v1.3.1
  ├── data/etl-pipeline:v2.1.0
  └── infra/monitoring-agent:v1.0.0
```

**Rationale:**
- Professional tier supports unlimited repositories
- Single registry simplifies credential management across all services
- Garbage collection essential (microservices generate many images)
- Repository naming convention: `{layer}/{service}:{version}`

**Cost:** $20/month (Professional tier, ~100 repositories)

---

### Global Application with DR

Production setup with disaster recovery across three continents:

**Primary (US):**
```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: global-primary
spec:
  name: myapp-primary-us
  subscriptionTier: PROFESSIONAL
  region: NYC3
  garbageCollectionEnabled: true
```

**DR - Europe:**
```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: global-dr-eu
spec:
  name: myapp-dr-europe
  subscriptionTier: PROFESSIONAL
  region: FRA1
  garbageCollectionEnabled: true
```

**DR - Asia:**
```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: global-dr-asia
spec:
  name: myapp-dr-asia
  subscriptionTier: PROFESSIONAL
  region: SGP1
  garbageCollectionEnabled: true
```

**DR Strategy:**
- CI/CD pipelines push all critical images to all three registries
- Each regional DOKS cluster pulls from its local registry (primary)
- Failover: If a regional registry is unavailable, DOKS can pull from alternate regions (slower, but functional)

**Cost:** $20/month (all three registries on one Professional subscription)

---

### Cost-Optimized Multi-Team Setup

Single Professional registry shared by multiple teams with repository-level organization:

```yaml
apiVersion: digital-ocean.openmcf.org/v1
kind: DigitalOceanContainerRegistry
metadata:
  name: multi-team-prod
spec:
  name: company-prod
  subscriptionTier: PROFESSIONAL
  region: NYC3
  garbageCollectionEnabled: true
```

**Repository Organization by Team:**
```
registry.digitalocean.com/company-prod/
  ├── team-alpha/service-a:v1.0.0
  ├── team-alpha/service-b:v2.1.3
  ├── team-bravo/api-gateway:v1.5.0
  ├── team-bravo/worker-queue:v1.2.8
  ├── team-charlie/ml-inference:v3.0.1
  └── shared/base-images/python:3.11-slim
```

**Access Control:**
- All teams share a single DigitalOcean API token (registry-level auth only)
- Repository-level access control managed via naming convention and CI/CD pipeline permissions
- Shared base images in `shared/` namespace reduce storage (image layer deduplication)

**Cost Savings:**
- Single Professional subscription ($20/month) vs. 3 separate subscriptions ($60/month)
- **Savings:** $40/month (67% reduction)

---

## Integration Examples

### Using Images in Kubernetes

After deploying a `DigitalOceanContainerRegistry`, use images in Kubernetes deployments:

**Deployment Manifest:**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-app
  namespace: production
spec:
  replicas: 3
  selector:
    matchLabels:
      app: my-app
  template:
    metadata:
      labels:
        app: my-app
    spec:
      containers:
      - name: app
        image: registry.digitalocean.com/myapp-prod/my-app:v1.2.3
        ports:
        - containerPort: 8080
        env:
        - name: ENV
          value: production
```

**Key Points:**
- No `imagePullSecrets` needed if DOKS integration is enabled (OpenMCF handles this automatically)
- Use immutable tags (semantic versions or Git SHAs) instead of `latest`
- Image format: `registry.digitalocean.com/{registry-name}/{repository}:{tag}`

---

### CI/CD Pipeline Integration

**GitHub Actions Example:**

```yaml
name: Build and Push to DOCR

on:
  push:
    branches: [main]

env:
  REGISTRY: registry.digitalocean.com/myapp-prod
  IMAGE_NAME: my-app

jobs:
  build-and-push:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      - name: Install doctl
        uses: digitalocean/action-doctl@v2
        with:
          token: ${{ secrets.DIGITALOCEAN_ACCESS_TOKEN }}

      - name: Log in to DOCR
        run: doctl registry login --expiry-seconds 3600

      - name: Build and tag image
        run: |
          docker build -t ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }} .
          docker build -t ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:latest .

      - name: Push image
        run: |
          docker push ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ github.sha }}
          docker push ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:latest
```

**GitLab CI Example:**

```yaml
stages:
  - build
  - push

variables:
  REGISTRY: registry.digitalocean.com/myapp-prod
  IMAGE_NAME: my-app

build-image:
  stage: build
  image: docker:latest
  services:
    - docker:dind
  before_script:
    - apk add --no-cache curl
    - curl -sL https://github.com/digitalocean/doctl/releases/download/v1.98.0/doctl-1.98.0-linux-amd64.tar.gz | tar -xzv
    - mv doctl /usr/local/bin
    - doctl auth init -t $DIGITALOCEAN_TOKEN
    - doctl registry login
  script:
    - docker build -t $REGISTRY/$IMAGE_NAME:$CI_COMMIT_SHA .
    - docker push $REGISTRY/$IMAGE_NAME:$CI_COMMIT_SHA
  only:
    - main
```

---

## Best Practices Summary

### Naming Conventions

**Registry Names:**
- Use descriptive, environment-aware names: `myapp-prod`, `myapp-staging`, `team-dev`
- Include region for multi-region setups: `myapp-prod-nyc3`, `myapp-prod-sfo3`
- Avoid generic names (they're globally unique): `app` ❌ → `company-app-prod` ✅

**Repository Names:**
- Use namespaces: `{team}/{service}`, `{layer}/{component}`
- Examples: `backend/auth-service`, `frontend/web-app`, `infra/monitoring`

**Image Tags:**
- Use immutable tags: Git SHA (`abc123f`), semantic version (`v1.2.3`), or combined (`v1.2.3-abc123f`)
- Avoid `latest` in production (makes rollback difficult and causes storage bloat)

---

### Tier Selection Guide

| Scenario | Recommended Tier | Cost | Rationale |
|----------|------------------|------|-----------|
| Personal learning | STARTER | $0/month | Free, sufficient for experimentation |
| Team dev (< 5 people) | BASIC | $5/month | 5 GiB storage, 5 repositories |
| Staging environment | BASIC | $5/month | Cost-effective pre-production |
| Production (single region) | PROFESSIONAL | $20/month | 100 GiB, unlimited repos, multi-registry support |
| Production (multi-region) | PROFESSIONAL | $20/month | Supports up to 10 registries on one subscription |
| Enterprise (many teams) | PROFESSIONAL | $20/month | Unlimited repos enable shared registry model |

---

### Garbage Collection Strategy

**Always enable for:**
- Any environment that pushes images frequently (dev, staging, prod)
- Any registry using `latest` tags (even if you shouldn't use them)
- Cost-sensitive deployments (GC can reduce storage by 40-60%)

**Disable only for:**
- Static registries with infrequent updates
- Registries where you need to preserve all historical images for compliance

**Schedule:** OpenMCF automatically schedules GC during low-traffic windows (e.g., Sunday 3 AM) to avoid disrupting CI/CD pipelines.

---

## Troubleshooting Examples

### Example: Registry Name Conflict

**Error:**
```
Error: registry name "my-app" already exists in another DigitalOcean account
```

**Solution:** Choose a more specific name:

```yaml
# Instead of:
spec:
  name: my-app  # ❌ Too generic

# Use:
spec:
  name: acme-corp-my-app  # ✅ Company-specific
```

---

### Example: Storage Limit Exceeded

**Error:**
```
Error: storage limit exceeded (500 MiB used, 500 MiB limit)
```

**Solution:** Enable garbage collection or upgrade tier:

```yaml
spec:
  name: my-registry
  subscriptionTier: BASIC  # Upgrade from STARTER
  region: NYC3
  garbageCollectionEnabled: true  # Enable automated cleanup
```

---

## Additional Resources

- **Main Documentation:** [README.md](./README.md)
- **Deep Dive:** [docs/README.md](./docs/README.md)
- **Pulumi Module:** [iac/pulumi/](./iac/pulumi/)
- **Terraform Module:** [iac/tf/](./iac/tf/)
- **DigitalOcean Documentation:** https://docs.digitalocean.com/products/container-registry/

---

**Need Help?** File an issue in the OpenMCF repository or refer to the comprehensive guides linked above.

