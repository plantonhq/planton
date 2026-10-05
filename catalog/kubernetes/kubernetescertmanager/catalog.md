# Cert Manager

Installs cert-manager — the cluster's certificate machinery — from the official Helm chart. cert-manager runs three components: the controller (watches Certificates and drives issuance), the webhook (validates cert-manager resources at admission), and the cainjector (injects CA bundles into webhook configurations). This component manages the CONTROLLER MACHINERY only: who signs certificates and what certificates exist are separate first-class resources (Cert Manager Cluster Issuer, Cert Manager Issuer, Cert Manager Certificate). One installation per cluster serves all of them.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Helm Release** -- the `cert-manager` chart from `https://charts.jetstack.io`, pinned to the chart version you choose
- **Controller, Webhook, and CA Injector** -- the three cert-manager deployments, each tunable (replicas, resources, webhook networking)
- **CRDs** -- Certificate, Issuer, ClusterIssuer, CertificateRequest, Order, Challenge (installed with the release by default, kept on uninstall by default)
- **ServiceAccount** -- the controller identity, optionally annotated for keyless cloud DNS authentication (GKE Workload Identity, EKS IRSA, AKS Workload Identity)
- **Namespace** (optional) -- created with standard governance labels when `createNamespace` is true
- **startupapicheck Job** -- a post-install hook that verifies the webhook is actually serving before the release reports success

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### Kubernetes Cluster

- Any conformant cluster (GKE, EKS, AKS, or self-managed).
- When enabling the Prometheus ServiceMonitor, the Prometheus operator CRDs (e.g. kube-prometheus-stack) must already be on the cluster — the release fails to install without them.
- For keyless DNS-01, the cluster's workload-identity mechanism must be active (GKE Workload Identity, EKS OIDC provider for IRSA, or AKS Workload Identity) and the cloud-side trust policy must name the controller ServiceAccount.

## Deploy

### Console

Open the deployment store, find **Cert Manager**, and click **Deploy**. The creation wizard walks you through installation, controller tuning, secrets coordination, ACME operations, workload identity, companion components, observability, and scheduling. Start from the **Basic Cert-Manager Installation** preset for a conventional install in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesCertManager
metadata:
  name: cert-manager
  org: acme-corp
  env: prod
spec:
  namespace:
    value: cert-manager
  createNamespace: true
  chartVersion: v1.20.3
```

```shell
planton apply -f cert-manager.yaml
```

This installs cert-manager into a new `cert-manager` namespace with CRDs installed and kept on uninstall — the standard single-installation path. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the namespace to a resource managed by another Infra Component:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: cert-manager-namespace
      fieldPath: spec.name
  chartVersion: v1.20.3
```

The InfraPipeline creates the namespace first, then installs cert-manager into it.

## Key Configuration

These are the most important decisions when configuring cert-manager. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The namespace is permanent** -- Kept CRDs (the default) retain the Helm release's namespace in their ownership metadata, so re-installing into a DIFFERENT namespace fails on the surviving CRDs. Moving an install means first deleting the kept CRDs — which cascades to ALL certificate data cluster-wide. `cert-manager` is the near-universal convention.

**Pin the chart version deliberately** -- `chartVersion` is also the cert-manager version (chart and app versions are aligned upstream). Pick versions from the chart repository's index (`helm search repo`): the served chart is the contract.

**CRD lifecycle** -- `crds.install` defaults to true (one component managing both halves is strictly simpler); disable only when another installation already owns the CRDs. `crds.keepOnUninstall` defaults to true because deleting CRDs cascades to every Certificate and Issuer object cluster-wide — turning it off should be an explicit, considered act.

**The cluster-resource namespace** -- `clusterResourceNamespace` is where cert-manager reads Secrets for CLUSTER-scoped resources (ClusterIssuer credentials, ACME account keys). Empty means the installation namespace. Cert Manager Cluster Issuer resources materialize their credential Secrets here.

**Split-horizon DNS needs the self-check resolvers** -- Before asking the CA to verify a DNS-01 challenge, cert-manager checks the TXT record itself. On clusters whose in-cluster DNS serves a private view, point `dns01SelfCheck.recursiveNameservers` at resolvers that see the PUBLIC view (e.g. `8.8.8.8:53`), and set `recursiveNameserversOnly` for the full fix.

**Keyless DNS-01 via workload identity** -- `workloadIdentity` binds the controller ServiceAccount to a cloud identity (Route53 via EKS IRSA, Cloud DNS via GKE Workload Identity, Azure DNS via AKS Workload Identity). Issuers whose DNS-01 providers leave static credentials empty authenticate through this identity. Token-based providers (Cloudflare, DigitalOcean) don't need it.

**Webhook reachability** -- The webhook must be reachable by the API server. On clusters whose control plane cannot reach pod IPs (EKS with custom CNI is the canonical case), set `webhook.hostNetwork` and pair it with a `webhook.securePort` that is free on the node (10250 collides with the kubelet).

**The escape hatch** -- `helmValues` carries additional chart values as a YAML document, merged LAST over everything the typed fields render. For the chart surface beyond the typed fields — never the substitute for them, and never a place for secrets.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace cert-manager was installed into | Locating the installation; the issuers' `certManagerNamespace` field |
| `service_account_name` | Controller ServiceAccount name | The identity to bind on the cloud side for keyless DNS-01 (IRSA trust policy subject, GKE Workload Identity member, Azure federated credential subject) |
| `cluster_resource_namespace` | The resolved namespace for cluster-scoped Secrets | Where ClusterIssuer credential Secrets and ACME account keys land |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard install** -- Conventional `cert-manager` namespace, pinned chart, CRDs managed by the release. Start from the **Basic Cert-Manager Installation** preset.

**GKE with Cloud DNS** -- Workload Identity binds the controller to a GCP service account so DNS-01 for Cloud DNS zones needs no static key. Start from the **Cert-Manager with GKE Workload Identity** preset.

**EKS with Route53** -- IRSA binds the controller to an IAM role so DNS-01 for Route53 zones needs no static key. Start from the **Cert-Manager with EKS IRSA** preset.

## Works With

- [**Cert Manager Cluster Issuer**](/infra-catalog/kubernetes-cluster-issuer) and [**Cert Manager Issuer**](/infra-catalog/kubernetes-issuer) -- the signing authorities; deploy them after cert-manager (they reference its namespace and CRDs).
- [**Cert Manager Certificate**](/infra-catalog/kubernetes-certificate) -- requests certificates from those issuers into workload namespaces.
- [**Ingress NGINX**](/infra-catalog/kubernetes-ingress-nginx) and [**Kubernetes Gateway**](/infra-catalog/kubernetes-gateway) -- serve HTTPS using the issued certificate Secrets.
- [**ExternalDNS**](/infra-catalog/kubernetes-external-dns) -- manages DNS records in the same zones the DNS-01 solvers write challenge records to.
