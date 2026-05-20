# Deploying cert-manager on Kubernetes: From Manual Manifests to Production Automation

## The TLS Certificate Challenge

Every production Kubernetes cluster needs TLS certificates. Without automation, certificate management becomes a manual, error-prone process that leads to expired certificates, outages, and security vulnerabilities. cert-manager solves this by automating the entire certificate lifecycle -- request, validation, issuance, and renewal.

## What This Component Does

KubernetesCertManager installs the cert-manager controller on a Kubernetes cluster. It handles:

1. **Helm chart deployment** -- installs cert-manager with CRDs, controller, and webhook
2. **ServiceAccount configuration** -- creates a dedicated SA with optional workload identity annotations for cloud DNS authentication
3. **DNS resolver configuration** -- configures recursive nameservers for reliable DNS-01 challenge propagation

**What this component does NOT do:** It does not create ClusterIssuers. ClusterIssuer management is handled by the separate **KubernetesClusterIssuer** component. This decoupling allows independent lifecycle management of the controller and the issuers.

## Deployment Maturity Spectrum

### Level 0: Manual kubectl

Install cert-manager CRDs and controller via `kubectl apply`:

```bash
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.19.1/cert-manager.yaml
```

**Pros:** Simple, quick start.
**Cons:** No version tracking, no declarative management, manual upgrades, no workload identity integration.
**Verdict:** Fine for learning, not for production.

### Level 1: Helm CLI

```bash
helm repo add jetstack https://charts.jetstack.io
helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager --create-namespace \
  --set installCRDs=true
```

**Pros:** Version pinning, upgrade path, Helm values for customization.
**Cons:** Imperative, state lives in the cluster, no GitOps, manual SA annotation.
**Verdict:** Good for development clusters.

### Level 2: Terraform

```hcl
resource "helm_release" "cert_manager" {
  name       = "cert-manager"
  repository = "https://charts.jetstack.io"
  chart      = "cert-manager"
  namespace  = "cert-manager"
  values     = [yamlencode({ installCRDs = true })]
}
```

**Pros:** Declarative, state tracked, reproducible.
**Cons:** Must manage provider configuration, SA annotations, and workload identity manually.
**Verdict:** Production-ready for Terraform shops.

### Level 3: OpenMCF (This Component)

```yaml
apiVersion: kubernetes.openmcf.org/v1
kind: KubernetesCertManager
metadata:
  name: cert-manager
spec:
  namespace:
    value: cert-manager
  createNamespace: true
  workloadIdentity:
    gke:
      serviceAccountEmail: "cert-manager@my-project.iam.gserviceaccount.com"
```

**Pros:** Consistent KRM structure, workload identity as a first-class config, dual IaC (Pulumi + Terraform), validation before deployment, presets for common patterns.
**Cons:** Requires OpenMCF CLI or Planton platform.
**Verdict:** Production-ready with the best developer experience.

## Architectural Decision: Decoupled ClusterIssuers

Previously, this component created ClusterIssuers as part of the cert-manager installation. This was changed to a decoupled model:

- **KubernetesCertManager** -- installs the controller, configures workload identity
- **KubernetesClusterIssuer** -- creates one ClusterIssuer per DNS domain

**Rationale:**
1. Adding/removing DNS domains shouldn't require redeploying the controller
2. Different teams can manage their own ClusterIssuers independently
3. ClusterIssuer misconfiguration shouldn't affect the controller lifecycle
4. Single Responsibility Principle -- installation and issuer configuration are different concerns

## Workload Identity Architecture

For cloud DNS providers (GCP Cloud DNS, AWS Route53, Azure DNS), cert-manager needs cloud API credentials. The modern approach uses workload identity -- binding the Kubernetes ServiceAccount to a cloud identity:

| Cloud | Mechanism | SA Annotation |
|-------|-----------|---------------|
| GKE | Workload Identity | `iam.gke.io/gcp-service-account` |
| EKS | IRSA | `eks.amazonaws.com/role-arn` |
| AKS | Managed Identity | `azure.workload.identity/client-id` |

This component configures the annotation on the cert-manager controller SA. The KubernetesClusterIssuer component then creates ClusterIssuers that leverage this identity for DNS-01 challenges.

For Cloudflare, no workload identity is needed -- the KubernetesClusterIssuer creates a Kubernetes Secret with the API token directly.

## Production Best Practices

### Version Pinning

Always pin both `kubernetesCertManagerVersion` (image tag) and `helmChartVersion`. Mismatched versions between the image and chart can cause CRD incompatibilities.

### DNS Resolver Configuration

This component configures `--dns01-recursive-nameservers=1.1.1.1:53,8.8.8.8:53` and `--dns01-recursive-nameservers-only` by default. This ensures DNS-01 challenge verification uses public DNS resolvers rather than the cluster's internal DNS, which prevents false negatives from DNS caching.

### Monitoring

Monitor cert-manager health via:
- `cert-manager_controller_sync_call_count` -- certificate sync operations
- `cert-manager_certificate_expiration_timestamp_seconds` -- certificate expiry times
- `cert-manager_http_acme_client_request_count` -- ACME server interactions

## Conclusion

KubernetesCertManager provides a clean, single-responsibility abstraction for installing cert-manager. By delegating ClusterIssuer management to the KubernetesClusterIssuer component, it enables independent lifecycle management and cleaner multi-team workflows.
