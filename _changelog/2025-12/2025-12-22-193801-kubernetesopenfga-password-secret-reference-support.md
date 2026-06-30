# KubernetesOpenFga Password Secret Reference Support and Datastore Configuration Split

**Date**: December 22, 2025
**Type**: Enhancement
**Components**: API Definitions, Kubernetes Provider, Pulumi CLI Integration, Terraform Module

## Summary

Added support for the database password in `KubernetesOpenFga` to be provided either as a plain string value or as a reference to an existing Kubernetes Secret. Additionally, split the single `uri` field in `KubernetesOpenFgaDataStore` into separate structured fields (`host`, `port`, `database`, `username`, `password`, `is_secure`) for better security and maintainability. The URI is now constructed in the deployment modules (Pulumi and Terraform) instead of requiring users to construct it manually.

## Problem Statement / Motivation

When configuring the OpenFGA datastore, users previously had to provide a complete database URI including the password as a plaintext string. This posed several challenges:

### Pain Points

- **Security risk**: Passwords stored in plaintext in YAML manifests could be accidentally committed to version control
- **Compliance issues**: Many organizations require secrets to be stored in dedicated secret management systems
- **Operational overhead**: Password rotation required updating manifest files rather than just rotating the Kubernetes Secret
- **Complex URI construction**: Users had to manually construct database URIs with proper escaping and formatting
- **Error-prone**: Typos in URI format led to cryptic connection errors

## Solution / What's New

### 1. Split URI into Structured Fields

Replaced the single `uri` field with structured fields:
- `host`: Database hostname/endpoint
- `port`: Database port (optional, defaults to 5432 for PostgreSQL, 3306 for MySQL)
- `database`: Database name
- `username`: Database username
- `password`: Uses `KubernetesSensitiveValue` for secure credential handling
- `is_secure`: Boolean flag for SSL/TLS connections

### 2. KubernetesSensitiveValue for Password

Reused the existing `KubernetesSensitiveValue` proto type (from the kubernetes_secret.proto) that uses `oneof` to support either:
1. A plain string value (for development/testing)
2. A reference to an existing Kubernetes Secret (recommended for production)

### Usage Examples

**Using Kubernetes Secret (Recommended for Production):**

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesOpenFga
metadata:
  name: openfga-prod
spec:
  targetCluster:
    clusterName: my-gke-cluster
  namespace:
    value: openfga-prod
  createNamespace: true
  container:
    replicas: 3
    resources:
      requests:
        cpu: 500m
        memory: 512Mi
      limits:
        cpu: 2000m
        memory: 2Gi
  datastore:
    engine: postgres
    host: prod-db-host.example.com
    port: 5432
    database: openfga
    username: openfga_user
    password:
      secretRef:
        name: openfga-db-credentials
        key: password
    isSecure: true
```

**Using Plain String (Dev/Test Only):**

```yaml
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesOpenFga
metadata:
  name: openfga-dev
spec:
  datastore:
    engine: postgres
    host: localhost
    database: openfga
    username: user
    password:
      stringValue: my-password
```

## Implementation Details

### Proto Schema Changes

**File**: `apis/dev/planton/provider/kubernetes/kubernetesopenfga/v1/spec.proto`

```protobuf
message KubernetesOpenFgaDataStore {
  string engine = 1;  // "postgres" or "mysql"
  string host = 2;
  optional int32 port = 3;  // Defaults based on engine
  string database = 4;
  string username = 5;
  dev.planton.provider.kubernetes.KubernetesSensitiveValue password = 6;
  bool is_secure = 7;  // Adds sslmode=require (Postgres) or tls=true (MySQL)
}
```

### Pulumi Module Update

**File**: `apis/dev/planton/provider/kubernetes/kubernetesopenfga/v1/iac/pulumi/module/helm_chart.go`

The Pulumi module now:
1. Constructs the URI from individual fields
2. Adds appropriate connection options (sslmode, parseTime, tls)
3. For secret refs, uses `extraEnvVars` to inject password from Kubernetes Secret

```go
if ds.Password.GetSecretRef() != nil {
    // Use environment variable placeholder in URI
    datastoreUri := fmt.Sprintf("%s://%s:$(OPENFGA_DATASTORE_PASSWORD)@%s:%d/%s%s",
        ds.Engine, ds.Username, ds.Host, port, ds.Database, connOptions)

    helmValues["datastore"] = pulumi.Map{
        "engine": pulumi.String(ds.Engine),
        "uri":    pulumi.String(datastoreUri),
    }

    // Inject password from secret via extraEnvVars
    helmValues["extraEnvVars"] = pulumi.Array{
        pulumi.Map{
            "name": pulumi.String("OPENFGA_DATASTORE_PASSWORD"),
            "valueFrom": pulumi.Map{
                "secretKeyRef": pulumi.Map{
                    "name": pulumi.String(secretRef.Name),
                    "key":  pulumi.String(secretRef.Key),
                },
            },
        },
    }
}
```

### Terraform Module Update

**File**: `apis/dev/planton/provider/kubernetes/kubernetesopenfga/v1/iac/tf/variables.tf`

```hcl
datastore = object({
  engine = string
  host = string
  port = optional(number)
  database = string
  username = string
  password = object({
    string_value = optional(string)
    secret_ref = optional(object({
      namespace = optional(string)
      name = string
      key = string
    }))
  })
  is_secure = optional(bool, false)
})
```

**File**: `apis/dev/planton/provider/kubernetes/kubernetesopenfga/v1/iac/tf/helm_chart.tf`

Similar logic to Pulumi - constructs URI from fields and uses `extraEnvVars` for secret refs.

### Test Updates

Updated all test cases in `spec_test.go` to use the new structured datastore configuration:

```go
Datastore: &KubernetesOpenFgaDataStore{
    Engine:   "postgres",
    Host:     "localhost",
    Database: "testdb",
    Username: "user",
    Password: &kubernetes.KubernetesSensitiveValue{
        Value: &kubernetes.KubernetesSensitiveValue_StringValue{
            StringValue: "pass",
        },
    },
},
```

Added tests for:
- Plain string password
- Secret reference password
- Custom port and SSL enabled
- MySQL engine
- Invalid inputs (missing fields, invalid port range)

## Benefits

- **Improved security**: Production deployments can use Kubernetes Secrets instead of plaintext passwords
- **GitOps friendly**: Manifests can be safely committed to version control without exposing credentials
- **Easier rotation**: Password changes only require updating the Kubernetes Secret, not the manifest
- **Simpler configuration**: No need to manually construct complex URIs with escaping
- **Better defaults**: Automatic port selection based on engine type
- **SSL/TLS support**: Simple boolean flag to enable secure connections
- **Reusable pattern**: Uses existing `KubernetesSensitiveValue` type for consistency across components
- **MySQL compatibility**: Automatically adds `parseTime=true` for proper MySQL time handling

## Impact

### Users
- Database credentials can now be secured properly in production
- No breaking changes for new deployments (just different YAML structure)
- Clear documentation with examples for both approaches

### Developers
- Pattern consistent with `kubernetessignoz` component
- All tests updated and passing
- Both Pulumi and Terraform modules maintain feature parity

## Files Changed

| File | Change |
|------|--------|
| `kubernetesopenfga/v1/spec.proto` | Split URI into structured fields, added KubernetesSensitiveValue for password |
| `iac/pulumi/module/helm_chart.go` | Construct URI from fields, handle password types |
| `iac/tf/variables.tf` | New structured datastore object |
| `iac/tf/helm_chart.tf` | Construct URI, handle secret refs via extraEnvVars |
| `spec_test.go` | Updated test cases for new structure |
| `examples.md` | Updated examples with new format |
| `iac/pulumi/examples.md` | Updated Pulumi examples |
| `iac/tf/examples.md` | Updated Terraform examples |

## Related Work

- Follows the pattern established in `kubernetessignoz` (2025-12-19 changelog)
- Uses existing `KubernetesSensitiveValue` from `kubernetes_secret.proto`
- Aligns with security recommendations in the component's research document (`docs/README.md`)

---

**Status**: ✅ Production Ready
**Timeline**: ~1 hour implementation

