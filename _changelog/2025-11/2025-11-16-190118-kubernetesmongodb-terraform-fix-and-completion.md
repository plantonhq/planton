# KubernetesMongodb Terraform Module Fix and Kind Completion

**Date**: November 16, 2025  
**Type**: Bug Fix + Enhancement  
**Components**: Kubernetes Provider, Terraform Module, Percona MongoDB Operator, Documentation

## Summary

Fixed a critical bug in the KubernetesMongodb Terraform module where `mongodb.tf` contained incorrect OpenFGA code instead of MongoDB implementation, updated stale examples to match current API, and completed the kind with comprehensive Terraform documentation. This brings the kind from 94.5% to 100% completion and makes it fully functional for Terraform users.

## Problem Statement / Motivation

The KubernetesMongodb kind had a **critical blocking issue** that prevented any Terraform deployments:

### Critical Bug: Wrong Implementation in mongodb.tf

```hcl
# WRONG - This was deploying OpenFGA instead of MongoDB!
resource "helm_release" "this" {
  repository = "https://openfga.github.io/helm-charts"
  chart      = "openfga"
  version    = "0.2.12"
  # ...
}
```

This was clearly a copy-paste error from another kind that was never corrected. **Any Terraform deployment would fail or deploy the wrong application.**

### Pain Points

- **Terraform module broken**: Users attempting to use Terraform would deploy OpenFGA instead of MongoDB
- **API mismatch in examples**: Examples used wrong kind name (`MongodbKubernetes` vs `KubernetesMongodb`)
- **Non-existent fields**: Examples referenced `kubernetesProviderConfigId` which doesn't exist in the spec
- **Missing documentation**: No Terraform examples for users

## Solution / What's New

### 1. Fixed Critical Terraform Bug

Completely rewrote `iac/tf/mongodb.tf` to properly implement Percona MongoDB Operator deployment:

**New Implementation**:
```hcl
# Generate random password for MongoDB
resource "random_password" "mongodb_root_password" {
  length      = 12
  special     = true
  min_special = 3
  min_numeric = 2
}

# Create Kubernetes secret for MongoDB credentials
resource "kubernetes_secret_v1" "mongodb_password" {
  data = {
    "MONGODB_DATABASE_ADMIN_PASSWORD" = base64encode(random_password.mongodb_root_password.result)
  }
}

# Deploy MongoDB using Percona Operator CRD
resource "kubernetes_manifest" "percona_server_mongodb" {
  manifest = {
    apiVersion = "psmdb.percona.com/v1"
    kind       = "PerconaServerMongoDB"
    spec = {
      crVersion = "1.20.1"
      image     = "percona/percona-server-mongodb:8.0.12-4"
      replsets = [{
        name = "rs0"
        size = var.spec.container.replicas
        resources = { /* ... */ }
        volumeSpec = var.spec.container.persistence_enabled ? {
          persistentVolumeClaim = {
            resources = { requests = { storage = var.spec.container.disk_size }}
          }
        } : null
      }]
      secrets = { users = kubernetes_secret_v1.mongodb_password.metadata[0].name }
      unsafeFlags = { replsetSize = true }  # Allow <3 replicas for dev
    }
  }
}

# LoadBalancer service for external access (if ingress enabled)
resource "kubernetes_service_v1" "mongodb_external_lb" {
  count = local.ingress_is_enabled ? 1 : 0
  spec {
    type = "LoadBalancer"
    annotations = {
      "external-dns.alpha.kubernetes.io/hostname" = local.ingress_external_hostname
    }
  }
}
```

### 2. Fixed User-Facing Examples

Updated `examples.md` to use correct API structure:

**Before (Broken)**:
```yaml
kind: MongodbKubernetes  # WRONG!
spec:
  kubernetesProviderConfigId: my-cluster  # Doesn't exist!
```

**After (Correct)**:
```yaml
kind: KubernetesMongodb  # Correct kind name
spec:
  container:
    persistenceEnabled: true  # Correct field names
    diskSize: 10Gi
```

### 3. Fixed Test Manifest

Updated `iac/tf/hack/manifest.yaml` to use correct kind name.

### 4. Created Terraform Examples

Added comprehensive `iac/tf/examples.md` with 6 complete examples:
1. Basic MongoDB deployment
2. MongoDB with persistence enabled
3. MongoDB with ingress and external access
4. Production deployment with full configuration
5. Development deployment with minimal resources
6. Custom Helm values configuration

Each example includes:
- Complete Terraform code
- Output retrieval
- Connection instructions
- Password retrieval commands

## Implementation Details

### Percona MongoDB Operator Integration

The Terraform implementation now correctly uses the Percona Server for MongoDB Operator:

**Key Features**:
- **CRD-based deployment**: Uses `PerconaServerMongoDB` custom resource
- **Automatic replica sets**: Configures `rs0` replica set based on replica count
- **Resource mapping**: Maps container resources to Percona format
- **Persistence support**: Conditional PVC creation based on `persistence_enabled`
- **Secure credentials**: Auto-generated passwords stored in Kubernetes secrets
- **Development-friendly**: `unsafeFlags.replsetSize = true` allows <3 replicas for testing

**MongoDB Versions**:
- MongoDB: 8.0.12-4
- Percona Operator CRD: 1.20.1

### Password Management

```hcl
# Auto-generated secure password
resource "random_password" "mongodb_root_password" {
  length           = 12
  special          = true
  override_special = "!@#$%^&*()-_=+[]{}:?"
  min_special      = 3
  min_numeric      = 2
  min_upper        = 2
  min_lower        = 2
}

# Stored in Kubernetes secret
data = {
  "MONGODB_DATABASE_ADMIN_PASSWORD" = base64encode(password.result)
}
```

### External Access Pattern

```hcl
# Conditional LoadBalancer service
resource "kubernetes_service_v1" "mongodb_external_lb" {
  count = local.ingress_is_enabled ? 1 : 0
  
  spec {
    type = "LoadBalancer"
    selector = {
      "app.kubernetes.io/name"       = "percona-server-mongodb"
      "app.kubernetes.io/instance"   = var.metadata.name
      "app.kubernetes.io/managed-by" = "percona-server-mongodb-operator"
    }
  }
}
```

## Benefits

1. **Unblocked Terraform users**: Module now works correctly - was completely broken before
2. **Production-ready**: Percona Operator provides enterprise-grade MongoDB deployment
3. **Secure by default**: Auto-generated passwords, Kubernetes secret integration
4. **Feature-complete**: External access, persistence, resource management
5. **Developer-friendly**: Allows single-replica deployments for testing
6. **Documentation parity**: Terraform users now have examples like Pulumi users

## Impact

### Critical Fix
- **Before**: Terraform would deploy OpenFGA (wrong application) - **100% failure rate**
- **After**: Terraform correctly deploys MongoDB using Percona Operator

### Users Affected
- **All Terraform users**: Previously couldn't use this kind at all
- **Production deployments**: Now have reliable, operator-based MongoDB
- **Development teams**: Can deploy with flexible replica counts

### Completion Metrics
- **Score**: 94.5% → 100%
- **Terraform module**: 80% → 100% (was broken, now complete)
- **User-facing docs**: 75% → 100% (examples fixed)
- **Nice to Have**: 66.7% → 100% (added Terraform examples)

## Spec Changes

**None** - No changes to protobuf specifications. The spec was correct; only the Terraform implementation and examples were wrong.

**API Definition** (unchanged):
```protobuf
message KubernetesMongodbSpec {
  KubernetesMongodbContainer container = 1;
  KubernetesMongodbIngress ingress = 2;
  map<string, string> helm_values = 3;
}

message KubernetesMongodbContainer {
  int32 replicas = 1;
  ContainerResources resources = 2;
  bool persistence_enabled = 3;
  string disk_size = 4;
}
```

The examples were updated to match this existing spec correctly.

## Testing

To verify the fix:

```bash
# Create a test MongoDB instance
terraform apply -var-file=test.tfvars

# Verify MongoDB pods are running
kubectl get pods -n <namespace> | grep percona-server-mongodb

# Check the CRD was created
kubectl get perconaservermongodb -n <namespace>

# Retrieve password
kubectl get secret <password-secret-name> -n <namespace> \
  -o jsonpath='{.data.MONGODB_DATABASE_ADMIN_PASSWORD}' | base64 -d
```

## Related Work

- Pulumi implementation (`iac/pulumi/module/mongodb.go`) was correct and used as reference
- Audit report: `2025-11-15-120155.md`
- Similar Percona operator pattern used in KubernetesMySQL kind

## Code Metrics

- **Files modified**: 4 (mongodb.tf, examples.md, hack/manifest.yaml, outputs.tf)
- **File created**: 1 (iac/tf/examples.md)
- **Lines in mongodb.tf**: 0 → 142 (was OpenFGA code, now MongoDB)
- **Critical bugs fixed**: 1 (blocking issue)
- **Examples corrected**: 5 (all examples had wrong kind name)

## Backward Compatibility

**Breaking Change Note**: If anyone was somehow using the broken Terraform module (which would have deployed OpenFGA), they'll need to destroy and recreate. However, given the module was completely non-functional for MongoDB, this is unlikely to affect real users.

---

**Status**: ✅ Production Ready  
**Completion**: 100% (from 94.5%)  
**Critical Fix**: Terraform module now functional (was broken)  
**Timeline**: Bug fix required complete rewrite of core Terraform resource

