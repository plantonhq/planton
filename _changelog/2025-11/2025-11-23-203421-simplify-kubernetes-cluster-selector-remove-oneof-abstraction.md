# Simplify Kubernetes Cluster Selector - Remove Unnecessary oneof Abstraction

**Date**: November 23, 2025  
**Type**: Refactoring | Breaking Change  
**Components**: API Definitions, Kubernetes Provider, Protobuf Schemas

## Summary

Removed the `KubernetesAddonTargetCluster` message wrapper with its `oneof` abstraction, simplifying target cluster selection for Kubernetes addons. The `KubernetesClusterSelector` message now serves as the direct type for selecting target clusters, eliminating an unnecessary layer of indirection. This change also corrects enum value validations and adds support for Civo Kubernetes clusters.

## Problem Statement / Motivation

The original design used a `KubernetesAddonTargetCluster` wrapper message with a `oneof` field to support two ways of specifying a target cluster:

```protobuf
// Old structure - unnecessary abstraction
message KubernetesAddonTargetCluster {
  oneof credential_source {
    string kubernetes_credential_id = 1;
    KubernetesClusterCloudResourceSelector kubernetes_cluster_selector = 2;
  }
}
```

### Pain Points

- **Over-engineered**: The `oneof` abstraction added complexity without providing meaningful value
- **Confusing naming**: Two similar names (`KubernetesClusterCloudResourceSelector` vs `KubernetesClusterSelector`) caused confusion
- **Direct selection sufficient**: In practice, direct cluster selection via kind and name is all that's needed
- **Incorrect validations**: The enum validation contained wrong values (615, 218) that didn't match actual cluster resource kinds
- **Missing support**: Civo Kubernetes clusters were not included in supported cluster types

## Solution / What's New

Simplified the API by using `KubernetesClusterSelector` directly as the target cluster type across all Kubernetes addon specs.

### New Structure

```protobuf
// Simplified - no wrapper, direct usage
message KubernetesClusterSelector {
  // Can be one of the supported kubernetes cluster kinds
  dev.planton.shared.cloudresourcekind.CloudResourceKind cluster_kind = 1 [(buf.validate.field).enum = {
    in: [
      400,  //AzureAksCluster
      207,  //AwsEksCluster
      607,  //GcpGkeCluster
      1208, //DigitalOceanKubernetesCluster
      1507  //CivoKubernetesCluster
    ]
  }];
  
  // Name of the kubernetes cluster in the same environment as the addon
  string cluster_name = 2;
}
```

### Key Changes

1. **Direct usage**: All addon specs now use `KubernetesClusterSelector target_cluster = 1` directly
2. **Corrected validations**: Fixed enum values to match actual cluster resource kinds:
   - AWS: 207 (`AwsEksCluster`) instead of 218 (`AwsIamUser`)
   - GCP: 607 (`GcpGkeCluster`) instead of 615 (`GcpGkeWorkloadIdentityBinding`)
3. **Added Civo**: Included 1507 (`CivoKubernetesCluster`) in supported cluster kinds
4. **Cleaner naming**: Renamed `KubernetesClusterCloudResourceSelector` → `KubernetesClusterSelector`

## Implementation Details

### Proto Changes

**File**: `apis/dev/planton/provider/kubernetes/target_cluster.proto`

Removed the `KubernetesAddonTargetCluster` message entirely and updated `KubernetesClusterSelector`:

```diff
-// **KubernetesAddonTargetCluster** defines the target cluster for a Kubernetes addon.
-message KubernetesAddonTargetCluster {
-  oneof credential_source {
-    string kubernetes_credential_id = 1;
-    KubernetesClusterCloudResourceSelector kubernetes_cluster_selector = 2;
-  }
-}

 // **KubernetesClusterSelector** defines a selector for a Kubernetes cluster in the same environment as the addon.
 message KubernetesClusterSelector {
-  //can be either gcp-gke-cluster-core
   dev.planton.shared.cloudresourcekind.CloudResourceKind cluster_kind = 1 [(buf.validate.field).enum = {
     in: [
-      400, //AzureAksCluster
-      615, //GcpGkeClusterCore
-      218, //AwsEksClusterCore
-      1208 //DigitalOceanKubernetesCluster
+      400,  //AzureAksCluster
+      207,  //AwsEksCluster
+      607,  //GcpGkeCluster
+      1208, //DigitalOceanKubernetesCluster
+      1507  //CivoKubernetesCluster
     ]
   }];
   string cluster_name = 2;
 }
```

### Updated Addon Specs

All Kubernetes addon specs were updated to use the direct type:

```diff
 message KubernetesZalandoPostgresOperatorSpec {
-  dev.planton.provider.kubernetes.KubernetesAddonTargetCluster target_cluster = 1;
+  dev.planton.provider.kubernetes.KubernetesClusterSelector target_cluster = 1;
   KubernetesZalandoPostgresOperatorSpecContainer container = 2;
   KubernetesZalandoPostgresOperatorBackupConfig backup_config = 3;
 }
```

**Affected addon specs** (9 total):
- `kubernetesaltinityoperator/v1/spec.proto`
- `kuberneteselasticoperator/v1/spec.proto`
- `kubernetesingressnginx/v1/spec.proto`
- `kubernetesistio/v1/spec.proto`
- `kubernetesperconamongooperator/v1/spec.proto`
- `kubernetesperconamysqloperator/v1/spec.proto`
- `kubernetesperconapostgresoperator/v1/spec.proto`
- `kubernetessolroperator/v1/spec.proto`
- `kubernetesstrimzikafkaoperator/v1/spec.proto`

### Test Updates

Updated all test files to use enum constants instead of magic numbers for better readability:

**Before**:
```go
spec = &KubernetesAltinityOperatorSpec{
    TargetCluster: &kubernetes.KubernetesAddonTargetCluster{
        CredentialSource: &kubernetes.KubernetesAddonTargetCluster_KubernetesCredentialId{
            KubernetesCredentialId: "my-k8s-cluster",
        },
    },
}
```

**After**:
```go
spec = &KubernetesAltinityOperatorSpec{
    TargetCluster: &kubernetes.KubernetesClusterSelector{
        ClusterKind: cloudresourcekind.CloudResourceKind_GcpGkeCluster,
        ClusterName: "my-k8s-cluster",
    },
}
```

All 9 test files now import and use `cloudresourcekind.CloudResourceKind_GcpGkeCluster` instead of hardcoded `615`.

## Benefits

### Simpler API Surface

- **Reduced complexity**: Removed 1 message type and 1 oneof abstraction
- **Clearer intent**: Direct cluster selection is more intuitive
- **Less cognitive load**: Developers don't need to understand the oneof pattern for simple cluster selection

### Correct Validations

- **Accurate constraints**: Enum validation now matches actual Kubernetes cluster resource kinds
- **Prevents errors**: Users can't accidentally specify non-cluster resource kinds (like IAM users or workload identity bindings)

### Better Test Code

- **More readable**: `CloudResourceKind_GcpGkeCluster` is self-documenting vs magic number `615`
- **Type-safe**: IDE autocomplete and compile-time checking for enum values
- **Less error-prone**: No risk of typos in numeric values

### Extended Support

- **Civo clusters**: Added support for Civo Kubernetes clusters (enum value 1507)
- **Future-ready**: Easier to add new cluster types without wrapper modifications

## Impact

### Breaking Change

This is a **breaking change** for the API schema. Any existing code using `KubernetesAddonTargetCluster` will need to migrate:

```diff
-target_cluster: &kubernetes.KubernetesAddonTargetCluster{
-    CredentialSource: &kubernetes.KubernetesAddonTargetCluster_KubernetesClusterSelector{
-        KubernetesClusterSelector: &kubernetes.KubernetesClusterSelector{
-            ClusterKind: 615,
-            ClusterName: "prod-cluster",
-        },
-    },
-}
+target_cluster: &kubernetes.KubernetesClusterSelector{
+    ClusterKind: 607,  // Corrected value for GcpGkeCluster
+    ClusterName: "prod-cluster",
+}
```

### Affected Components

**API Layer**:
- All Kubernetes addon proto definitions
- Generated protobuf stubs (Go, TypeScript, Java, Python)

**Test Layer**:
- 9 test files updated with new structure and enum constants

**Validation Layer**:
- Proto validation now enforces correct cluster kind values

### Migration Required

For any existing manifests or code using the old structure:
1. Remove the `KubernetesAddonTargetCluster` wrapper
2. Use `KubernetesClusterSelector` directly
3. Update cluster kind values to correct enum constants
4. Add Civo support if needed (optional)

## Code Metrics

- **Proto files changed**: 10 (1 core + 9 addon specs)
- **Test files updated**: 9
- **Lines removed**: ~30 (wrapper message + old oneof logic)
- **Lines added**: ~25 (corrected validations + test imports)
- **Net reduction**: ~5 lines (but significant complexity reduction)

## Related Work

### Enum Value Corrections

The incorrect enum values were discovered during this refactoring:

| Purpose | Old Value | Old Enum | Correct Value | Correct Enum |
|---------|-----------|----------|---------------|--------------|
| GCP GKE | 615 | `GcpGkeWorkloadIdentityBinding` | 607 | `GcpGkeCluster` |
| AWS EKS | 218 | `AwsIamUser` | 207 | `AwsEksCluster` |

This explains why validation failures may have occurred with the old values.

### Future Enhancements

With this simplified structure, future cluster provider additions will be straightforward:
- Add the cluster resource kind to `CloudResourceKind` enum
- Add the enum value to `KubernetesClusterSelector` validation list
- No wrapper modifications needed

---

**Status**: ✅ Production Ready  
**Impact**: Breaking API change - requires migration for existing usage  
**Testing**: All 9 addon test suites passing with new structure

