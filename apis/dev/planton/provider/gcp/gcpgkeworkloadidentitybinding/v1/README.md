# GCP GKE Workload Identity Binding

Atomically bind a Kubernetes Service Account (KSA) to a Google Service Account (GSA) to enable secure, keyless authentication for GKE workloads accessing Google Cloud APIs.

## Purpose

GKE Workload Identity is the **recommended way** for applications running in GKE to authenticate to Google Cloud services—eliminating the need for distributable service account keys. However, provisioning a Workload Identity binding manually is error-prone and requires a precise "dual-write" operation across two systems (GCP IAM and Kubernetes).

This component provides a **pattern-level abstraction** that:

1. **Atomically manages the dual-write**: Synchronizes the IAM policy binding on the GSA with the annotation on the KSA
2. **Constructs the brittle principal string automatically**: Builds `serviceAccount:{project}.svc.id.goog[{namespace}/{ksa}]` from simple inputs
3. **Prevents common failure modes**: Eliminates typos in the principal identifier and synchronization issues between IAM and Kubernetes

## Features

- ✅ **Zero-credential architecture**: No service account keys to rotate or leak
- ✅ **Pod-level IAM identity**: Each workload can have distinct GCP permissions
- ✅ **Simplified provisioning**: One resource definition instead of two synchronized resources
- ✅ **Atomic updates**: Changes to bindings are applied consistently
- ✅ **Clear audit trail**: Cloud Audit Logs show exactly which KSA impersonated which GSA

## Prerequisites

- GKE cluster with Workload Identity enabled
- Node pools configured with `workloadMetadataConfig.mode = GKE_METADATA`
- A Google Service Account (GSA) with the necessary IAM roles for your workload
- A Kubernetes Service Account (KSA) that your pods will use

## Basic Example

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpGkeWorkloadIdentityBinding
metadata:
  name: cert-manager-dns-binding
spec:
  projectId: "prod-project-123"
  serviceAccountEmail: "dns01-solver@prod-project-123.iam.gserviceaccount.com"
  ksaNamespace: "cert-manager"
  ksaName: "cert-manager"
```

This binding allows pods using the `cert-manager` ServiceAccount in the `cert-manager` namespace to impersonate the `dns01-solver@...` Google Service Account.

## What Happens Behind the Scenes

When you create this resource, Planton:

1. **Creates an IAM policy binding** on the GSA:
   ```
   roles/iam.workloadIdentityUser → serviceAccount:prod-project-123.svc.id.goog[cert-manager/cert-manager]
   ```

2. **Annotates the Kubernetes ServiceAccount** (if it exists):
   ```yaml
   metadata:
     annotations:
       iam.gke.io/gcp-service-account: "dns01-solver@prod-project-123.iam.gserviceaccount.com"
   ```

You never construct the complex principal string manually, and you never worry about synchronization issues.

## Common Use Cases

### cert-manager with Cloud DNS

Bind cert-manager to a GSA with `roles/dns.admin` to solve DNS-01 ACME challenges:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpGkeWorkloadIdentityBinding
metadata:
  name: cert-manager-dns-binding
spec:
  projectId: "prod-project"
  serviceAccountEmail: "dns01-solver@prod-project.iam.gserviceaccount.com"
  ksaNamespace: "cert-manager"
  ksaName: "cert-manager"
```

### Application Accessing Cloud Storage

Bind your app to a GSA with `roles/storage.objectViewer`:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpGkeWorkloadIdentityBinding
metadata:
  name: app-gcs-binding
spec:
  projectId: "prod-project"
  serviceAccountEmail: "my-app-gsa@prod-project.iam.gserviceaccount.com"
  ksaNamespace: "my-app"
  ksaName: "my-app-ksa"
```

### ExternalDNS with Cloud DNS

Bind external-dns to a GSA with `roles/dns.admin`:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpGkeWorkloadIdentityBinding
metadata:
  name: external-dns-binding
spec:
  projectId: "prod-project"
  serviceAccountEmail: "external-dns@prod-project.iam.gserviceaccount.com"
  ksaNamespace: "external-dns"
  ksaName: "external-dns"
```

## Security Best Practices

### One GSA Per Application

Never share a Google Service Account between different applications. Each app should have its own GSA with only the IAM roles it needs:

- **cert-manager**: `roles/dns.admin` (to solve DNS-01 challenges)
- **external-dns**: `roles/dns.admin` (to manage DNS records)
- **billing-api**: `roles/storage.objectViewer` (to read a specific bucket)

The role for the binding itself is always `roles/iam.workloadIdentityUser`.

### Namespace Isolation

Workload Identity provides perfect identity isolation at the namespace level. The principal identifier explicitly includes the namespace:

```
serviceAccount:{project}.svc.id.goog[{namespace}/{ksa-name}]
```

A KSA named `default` in `namespace-a` cannot be confused with an identically named KSA in `namespace-b`.

### Multi-Cluster Considerations

**Critical security consideration**: The Workload Identity Pool (`{project}.svc.id.goog`) is a **per-project resource, not per-cluster**.

All GKE clusters in the same GCP project share a single trust domain. GCP IAM cannot distinguish between identically named KSAs from different clusters.

**Best practice**: Use separate GCP projects for each environment (dev, staging, prod) to create fully isolated Workload Identity Pools.

## Troubleshooting

### Permission Denied Errors

1. **Check the KSA annotation**:
   ```bash
   kubectl get serviceaccount <ksa-name> -n <namespace> -o yaml
   ```
   Look for: `iam.gke.io/gcp-service-account: "<gsa-email>"`

2. **Check the IAM binding**:
   ```bash
   gcloud iam service-accounts get-iam-policy <gsa-email>
   ```
   Look for a member like: `serviceAccount:{project}.svc.id.goog[{namespace}/{ksa}]` with role `roles/iam.workloadIdentityUser`

3. **Check node pool configuration**:
   ```bash
   gcloud container node-pools describe <node-pool> --cluster <cluster>
   ```
   Verify: `workloadMetadataConfig.mode: GKE_METADATA`

### Testing the Binding

Deploy a test pod to verify authentication:

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: workload-identity-test
  namespace: <your-namespace>
spec:
  serviceAccountName: <your-ksa>
  containers:
  - name: test-sdk
    image: google/cloud-sdk:slim
    command: ["sleep", "3600"]
```

Then run:

```bash
kubectl exec -it workload-identity-test -n <namespace> -- gcloud auth list
```

**Success**: Shows the GSA email (`...iam.gserviceaccount.com`)  
**Failure**: Shows the Compute Engine default service account (Workload Identity not enabled on node pool)

## API Reference

### Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `projectId` | string | Yes | GCP project hosting the GKE cluster |
| `serviceAccountEmail` | string | Yes | Email of the Google Service Account to impersonate |
| `ksaNamespace` | string | Yes | Kubernetes namespace of the ServiceAccount |
| `ksaName` | string | Yes | Name of the Kubernetes ServiceAccount |

### Stack Outputs

After creation, the following outputs are available:

| Output | Description |
|--------|-------------|
| `member` | The IAM member string added to the policy |
| `service_account_email` | The bound GSA email (echoed from spec) |

## Related Documentation

- **[Deep Dive Research](docs/README.md)**: Comprehensive guide on Workload Identity levels, failure modes, and design decisions
- **[Examples](examples.md)**: Additional copy-paste ready examples for various scenarios
- **[Pulumi Module](iac/pulumi/README.md)**: Pulumi-specific deployment guide
- **[Terraform Module](iac/tf/README.md)**: Terraform-specific deployment guide

## Why This Component Exists

Existing IaC tools (Terraform, Pulumi, Config Connector, Crossplane) all provide **low-level primitives** that mirror the manual provisioning process. They force operators to:

- Manage two separate resources (IAM binding + KSA annotation)
- Manually construct the brittle principal string
- Debug synchronization issues
- Handle the four common failure modes

This component provides a **pattern-level abstraction** that matches user intent: "Bind this KSA to this GSA." The complexity is handled internally.

## Support

For issues, questions, or contributions, see the main [Planton documentation](https://planton.dev).


