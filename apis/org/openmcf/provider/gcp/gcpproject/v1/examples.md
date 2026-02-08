# GCP Project Examples

This document provides comprehensive examples of `GcpProject` resource configurations, ranging from minimal setups to production-grade configurations.

---

## Example 1: Minimal Development Project

A minimal configuration for a development/sandbox project. This keeps the default network for quick iteration.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpProject
metadata:
  name: dev-sandbox
spec:
  parentType: folder
  parentId: "123456789012"  # Your "sandbox" or "development" folder ID
  billingAccountId: "ABCDEF-123456-ABCDEF"
  labels:
    env: dev
    team: research
  disableDefaultNetwork: false  # Keep default network for quick prototyping
  enabledApis:
    - compute.googleapis.com
    - storage.googleapis.com
  ownerMember: "user:alice@example.com"
```

**Use Case:** Quick sandbox for developers to experiment without complex network setup.

---

## Example 2: Standard Project with Essential APIs

A standard project configuration with commonly used APIs enabled.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpProject
metadata:
  name: staging-api-service
spec:
  parentType: folder
  parentId: "234567890123"  # Your "staging" folder ID
  billingAccountId: "ABCDEF-123456-ABCDEF"
  labels:
    env: staging
    team: backend
    component: api-service
  disableDefaultNetwork: true  # Security best practice
  enabledApis:
    - compute.googleapis.com
    - storage.googleapis.com
    - container.googleapis.com         # For GKE
    - logging.googleapis.com
    - monitoring.googleapis.com
    - iam.googleapis.com
    - iamcredentials.googleapis.com
  ownerMember: "group:devops-staging@example.com"
```

**Use Case:** Staging environment for API service with GKE and standard observability.

---

## Example 3: Production-Grade Project with Full Security Hardening

A production project with comprehensive security, governance labels, deletion protection, and all essential APIs.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpProject
metadata:
  name: prod-payment-processing
spec:
  parentType: folder
  parentId: "345678901234"  # Your "production" folder ID
  billingAccountId: "ABCDEF-123456-ABCDEF"
  labels:
    env: prod
    team: e-commerce
    component: payment-processing
    cost-center: cc-4510
    compliance: pci-dss
    criticality: high
  disableDefaultNetwork: true  # CRITICAL: Never use default network in production
  deleteProtection: true  # CRITICAL: Prevent accidental project deletion
  enabledApis:
    - compute.googleapis.com
    - storage.googleapis.com
    - container.googleapis.com          # GKE
    - logging.googleapis.com
    - monitoring.googleapis.com
    - cloudtrace.googleapis.com         # Distributed tracing
    - cloudprofiler.googleapis.com      # Performance profiling
    - iam.googleapis.com
    - iamcredentials.googleapis.com
    - servicenetworking.googleapis.com  # For VPC peering (Cloud SQL, etc.)
    - compute.googleapis.com
    - dns.googleapis.com                # Cloud DNS
    - secretmanager.googleapis.com      # Secret management
  ownerMember: "group:platform-admins@example.com"
```

**Use Case:** Production environment for critical payment processing workload with full observability, security, deletion protection, and compliance.

---

## Example 4: Project Under Organization (Not Folder)

Some organizations prefer flat hierarchies or have specific projects directly under the organization node.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpProject
metadata:
  name: shared-networking
spec:
  parentType: organization
  parentId: "987654321098"  # Your organization ID
  billingAccountId: "ABCDEF-123456-ABCDEF"
  labels:
    env: shared
    team: platform
    component: networking
  disableDefaultNetwork: true
  enabledApis:
    - compute.googleapis.com
    - dns.googleapis.com
    - servicenetworking.googleapis.com
  ownerMember: "group:network-admins@example.com"
```

**Use Case:** Shared services project (e.g., Shared VPC host) placed directly under organization.

---

## Example 5: Minimal Project with No Owner IAM Binding

Sometimes IAM is managed separately. This example omits the `ownerMember` field.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpProject
metadata:
  name: demo-project
spec:
  parentType: folder
  parentId: "456789012345"
  billingAccountId: "ABCDEF-123456-ABCDEF"
  labels:
    env: demo
    team: marketing
  disableDefaultNetwork: true
  enabledApis:
    - compute.googleapis.com
    - storage.googleapis.com
```

**Use Case:** Project where IAM bindings are managed through separate IAM resources or organization policies.

---

## Example 6: Data Science Project with BigQuery and AI APIs

A project optimized for data science and machine learning workloads.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpProject
metadata:
  name: ml-research
spec:
  parentType: folder
  parentId: "567890123456"  # Your "data-science" folder ID
  billingAccountId: "ABCDEF-123456-ABCDEF"
  labels:
    env: dev
    team: data-science
    component: ml-research
  disableDefaultNetwork: false  # May need default network for Vertex AI notebooks
  enabledApis:
    - compute.googleapis.com
    - storage.googleapis.com
    - bigquery.googleapis.com           # Data warehouse
    - bigquerystorage.googleapis.com    # BigQuery Storage API
    - aiplatform.googleapis.com         # Vertex AI
    - notebooks.googleapis.com          # Vertex AI Workbench
    - ml.googleapis.com                 # Legacy ML Engine (if needed)
    - dataflow.googleapis.com           # Data processing pipelines
  ownerMember: "group:data-scientists@example.com"
```

**Use Case:** Data science and ML research environment with BigQuery and Vertex AI.

---

## Example 7: Service Account as Owner

Using a service account (e.g., for CI/CD automation) as the project owner.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpProject
metadata:
  name: ci-automation
spec:
  parentType: folder
  parentId: "678901234567"
  billingAccountId: "ABCDEF-123456-ABCDEF"
  labels:
    env: shared
    team: platform
    component: ci-cd
  disableDefaultNetwork: true
  enabledApis:
    - compute.googleapis.com
    - storage.googleapis.com
    - cloudbuild.googleapis.com
    - containerregistry.googleapis.com
  ownerMember: "serviceAccount:ci-automation@my-seed-project.iam.gserviceaccount.com"
```

**Use Case:** Project managed by CI/CD automation where a service account needs owner permissions.

---

## Example 8: Multi-Region Project with Cloud SQL and Redis

A project configured for a multi-region application with database and caching services.

```yaml
apiVersion: gcp.openmcf.org/v1
kind: GcpProject
metadata:
  name: prod-global-app
spec:
  parentType: folder
  parentId: "789012345678"
  billingAccountId: "ABCDEF-123456-ABCDEF"
  labels:
    env: prod
    team: platform
    component: global-app
    region: multi-region
  disableDefaultNetwork: true
  enabledApis:
    - compute.googleapis.com
    - storage.googleapis.com
    - container.googleapis.com
    - sqladmin.googleapis.com           # Cloud SQL
    - redis.googleapis.com              # Memorystore for Redis
    - servicenetworking.googleapis.com  # VPC peering for private services
    - cloudloadbalancing.googleapis.com # Global load balancing
    - logging.googleapis.com
    - monitoring.googleapis.com
  ownerMember: "group:sre-team@example.com"
```

**Use Case:** Production multi-region application with managed database and caching.

---

## Field Descriptions

### Required Fields

- **`metadata.name`**: Human-readable project name. Used as the basis for generating the globally unique `project_id` (a random suffix is appended automatically).
- **`spec.parentType`**: Must be `"organization"` or `"folder"`.
- **`spec.parentId`**: Numeric string ID of the parent organization or folder.
- **`spec.billingAccountId`**: Billing account in format `"XXXXXX-XXXXXX-XXXXXX"`.

### Optional Fields

- **`spec.labels`**: Map of key-value labels for cost allocation, governance, and filtering.
- **`spec.disableDefaultNetwork`**: Boolean (default: `true`). If `true`, the insecure default VPC is deleted. **Best practice: always set to `true` in production.**
- **`spec.enabledApis`**: List of Google Cloud APIs to enable (e.g., `compute.googleapis.com`). Must end with `.googleapis.com`.
- **`spec.ownerMember`**: IAM member to grant `roles/owner` at project creation. Format: `user:email`, `group:email`, or `serviceAccount:email`.
- **`spec.deleteProtection`**: Boolean (default: `false`). If `true`, enables GCP-native deletion protection. **Best practice: always set to `true` for critical production projects.**

---

## Best Practices

### 1. Enable Deletion Protection for Critical Projects
```yaml
deleteProtection: true
```
GCP-native deletion protection prevents accidental project deletion. **Always enable for production workloads.**

### 2. Always Disable Default Network in Production
```yaml
disableDefaultNetwork: true
```
The default network is overly permissive and violates security best practices.

### 3. Use Folders for Hierarchy
```yaml
parentType: folder
parentId: "your-folder-id"
```
Flat hierarchies (all projects under organization) are an anti-pattern.

### 4. Apply Governance Labels
```yaml
labels:
  env: prod
  team: platform
  cost-center: cc-4510
  component: api-service
```
Labels are critical for cost allocation and resource filtering.

### 5. Grant Permissions to Groups, Not Users
```yaml
ownerMember: "group:devops-admins@example.com"
```
Using groups makes permission management scalable.

### 6. Enable Required APIs at Creation Time
```yaml
enabledApis:
  - compute.googleapis.com
  - logging.googleapis.com
  - monitoring.googleapis.com
```
Prevents "API not enabled" errors during deployment.

---

## Common API Sets

### Minimal (Compute + Storage)
```yaml
enabledApis:
  - compute.googleapis.com
  - storage.googleapis.com
```

### Standard (Compute + Kubernetes + Observability)
```yaml
enabledApis:
  - compute.googleapis.com
  - storage.googleapis.com
  - container.googleapis.com
  - logging.googleapis.com
  - monitoring.googleapis.com
  - iam.googleapis.com
  - iamcredentials.googleapis.com
```

### Production (Full Stack)
```yaml
enabledApis:
  - compute.googleapis.com
  - storage.googleapis.com
  - container.googleapis.com
  - logging.googleapis.com
  - monitoring.googleapis.com
  - cloudtrace.googleapis.com
  - cloudprofiler.googleapis.com
  - iam.googleapis.com
  - iamcredentials.googleapis.com
  - servicenetworking.googleapis.com
  - dns.googleapis.com
  - secretmanager.googleapis.com
```

---

## Next Steps

After creating a project:
1. **Set up billing exports** to BigQuery for cost analysis
2. **Create custom VPCs** (since default network is disabled)
3. **Configure organization policies** for security and compliance
4. **Set up IAM bindings** for fine-grained access control
5. **Enable audit logging** for security monitoring

---

## Related Resources

- **GcpVpc**: Create custom VPCs after project creation
- **GcpProjectLien**: Protect critical projects from accidental deletion
- **GcpIamBinding**: Manage complex IAM permissions
- **GcpSharedVpc**: Attach service projects to Shared VPC host projects
