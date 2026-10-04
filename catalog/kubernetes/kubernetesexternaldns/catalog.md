# ExternalDNS

Installs the ExternalDNS controller from the official Helm chart, watching your cluster's Services, Ingresses, and Gateway API routes and publishing their hostnames as records in a real DNS provider — so a hostname exists the moment the workload does, and disappears with it. The spec is deliberately two-sided, because the cluster often runs in one environment while the zone lives in another (an EKS cluster publishing into Cloudflare is an ordinary case, not an exception): one arm selects WHERE records are written (Route 53, Cloud DNS, Azure DNS, Cloudflare, a webhook provider, or the in-memory sandbox), and workload identity plus per-provider credentials select HOW the controller authenticates. One installation manages one provider; clusters publishing to several deploy several instances of this kind.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; otherwise installs into an existing namespace
- **Helm Release** -- the `external-dns` chart from kubernetes-sigs, rendering the controller Deployment, its ServiceAccount (annotated for workload identity when configured), and RBAC (a ClusterRole, or a namespace-scoped Role when `namespaced` is true)
- **Kubernetes Secrets** -- created only when static credentials are declared (a Cloudflare API token, AWS access keys, a GCP service-account key, or the Azure `azure.json`); the module materializes them and wires them into the controller, so they never appear in chart values or pod specs
- **DNS records in your zone** -- the controller's runtime output: for every hostname on a watched resource, the record plus (under the `txt` registry) an ownership TXT record carrying this instance's owner ID

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **A DNS zone in the target provider** the controller can write to -- a Route 53 hosted zone, Cloud DNS zone, Azure DNS zone, or Cloudflare zone.
- **A cloud identity for keyless authentication** (only for the cloud-provider arms): an IRSA role on EKS, a GCP service account with a Workload Identity binding on GKE, or an Azure federated credential on AKS, granted DNS write permissions on the zones. The trust policy references exactly the ServiceAccount name and namespace this component exports.
- **A scoped API token** (only for Cloudflare): Zone:Read + DNS:Edit on the zones this instance manages.
- **Prometheus operator CRDs** (only for `prometheus.serviceMonitor`): the release fails to install without them.

## Deploy

### Console

Open the deployment store, find **ExternalDNS**, and click **Deploy**. The creation wizard walks you through placement, the DNS provider arm, authentication, sources, and the sync policy and ownership settings. Start from the **AWS Route 53 on EKS (Keyless via IRSA)** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesExternalDns
metadata:
  name: external-dns-route53
  org: acme-corp
  env: prod
spec:
  namespace:
    value: external-dns
  createNamespace: true
  awsRoute53:
    region: us-east-1
    zoneIdFilters:
      - value: Z104533312EOZ72FQZ4TT
    zoneType: public
  workloadIdentity:
    eks:
      roleArn:
        value: arn:aws:iam::123456789012:role/external-dns
  sources:
    - service
    - ingress
  policy: sync
  txtOwnerId: prod-eks-cluster
  domainFilters:
    - acme-corp.com
```

```shell
planton apply -f external-dns.yaml
```

This installs the controller keylessly via IRSA, watching Services and Ingresses and fully reconciling (including deletes of records it owns) one hosted zone scoped to `acme-corp.com`. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the zone filter and the IAM identity to resources managed by other Infra Components:

```yaml
spec:
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: external-dns-namespace
      fieldPath: spec.name
  awsRoute53:
    region: us-east-1
    zoneIdFilters:
      - valueFrom:
          kind: AwsRoute53Zone
          name: acme-prod-zone
          fieldPath: status.outputs.zone_id
  workloadIdentity:
    eks:
      roleArn:
        valueFrom:
          kind: AwsIamRole
          name: external-dns-irsa-role
          fieldPath: status.outputs.role_arn
```

The InfraPipeline deploys the zone and the IAM role first, then installs the controller with the resolved zone ID and role ARN.

## Key Configuration

These are the most important decisions when configuring an External DNS installation. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Distinct owner IDs, or instances destroy each other's records** -- the registry decides which records belong to this installation, and under the `sync` policy the controller DELETES records it believes it owns. Two instances sharing a `txtOwnerId` will each clean up the other's work, and the damage lands in the DNS zone, far from either manifest. Every instance sharing a zone needs its own owner ID -- set it FIRST when adding a second instance.

**Policy decides how far reconciliation goes** -- `upsert-only` (the default) creates and updates but never deletes: the safe posture for zones shared with records managed elsewhere, at the cost of stale records surviving workload deletion. `sync` is full reconciliation including deletes of owned records -- the right choice for dedicated zones. `create-only` writes once and never touches again.

**Filter, or the controller sees every zone its credentials can** -- with no `domainFilters` and no zone ID filters, the controller may act on every zone the credentials reach. In a shared cloud account that is how one cluster starts rewriting another team's records. Set the provider arm's `zoneIdFilters`, `domainFilters`, or both; `excludeDomains` carves exceptions back out.

**Pick the provider arm for where the zone lives, not where the cluster runs** -- exactly one of `awsRoute53`, `googleCloudDns`, `azureDns`, `cloudflare`, `webhook`, or `inMemory`. Cloudflare is the canonical cross-cloud arm (token-authenticated from any cluster); the `webhook` arm runs any out-of-tree provider's image as a sidecar; `inMemory` is the sandbox -- records live in the pod and vanish with it, never production.

**Keyless beats keys** -- on EKS/GKE/AKS, bind the controller's ServiceAccount to a cloud identity through `workloadIdentity` and leave the static credential fields empty. Static keys (materialized as Secrets by the module) are the fallback for clusters with no ambient cloud identity. Route 53's `assumeRole` adds the cross-account pattern: authenticate in the cluster's account, then assume a role in the account that owns the zones.

**The Cloudflare token fails late** -- the controller validates the token at first zone sync, not at startup: a revoked or mis-scoped token surfaces as a crash-looping pod with a Cloudflare 4xx in its logs, not as an install-time error.

**Sources are what the controller watches** -- the chart default is `service` and `ingress`. Add Gateway API route sources (`gateway-httproute`, ...) when hostnames live on routes, or `crd` to manage records declaratively via DNSEndpoint objects. Every type in `managedRecordTypes` is subject to the sync policy including deletes -- extend it deliberately.

**Pin the chart version from the served index** -- chart releases are cut separately from the controller (`chart 1.21.x` ships controller `v0.21.x`), and the served chart is the contract: the upstream source tree can claim a version at a tag that was never published. `imageTag` exists only to hold the controller back or roll it forward independently.

**TXT prefix sidesteps CNAME collisions** -- a TXT record cannot coexist with a CNAME of the same name; `txtPrefix` (e.g. `edns-`) moves the ownership records aside. Mutually exclusive with `txtSuffix`.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **AwsRoute53Zone** | `awsRoute53.zoneIdFilters[]` | `status.outputs.zone_id` |
| **AwsIamRole** | `awsRoute53.assumeRole`, `workloadIdentity.eks.roleArn` | `status.outputs.role_arn` |
| **GcpProject** | `googleCloudDns.project` | `status.outputs.project_id` |
| **GcpDnsZone** | `googleCloudDns.zoneIdFilters[]` | `status.outputs.zone_id` |
| **GcpServiceAccount** | `workloadIdentity.gke.serviceAccountEmail` | `status.outputs.email` |
| **AzureDnsZone** | `azureDns.zoneIdFilters[]` | `status.outputs.zone_id` |
| **AzureUserAssignedIdentity** | `workloadIdentity.aks.clientId` | `status.outputs.client_id` |
| **CloudflareDnsZone** | `cloudflare.zoneIdFilters[]` | `status.outputs.zone_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | The namespace the controller runs in | The other half of every workload-identity trust reference |
| `release_name` | The Helm release name | Operational tooling and release inspection |
| `service_account_name` | The controller's ServiceAccount | Exactly what an IAM role trust policy, a Google service-account binding, or an Azure federated credential references when wiring keyless authentication |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Route 53 from EKS, keyless** -- IRSA authentication, a zone ID filter, and `sync` on a dedicated zone: the production posture on AWS. Start from the **AWS Route 53 on EKS (Keyless via IRSA)** preset.

**Cloud DNS from GKE, keyless** -- Workload Identity binding to a GCP service account with DNS admin on the project's zones. Start from the **Google Cloud DNS on GKE (Keyless via Workload Identity)** preset.

**Azure DNS from AKS, keyless** -- Azure AD Workload Identity via a user-assigned identity's client ID. Start from the **Azure DNS on AKS (Keyless via Workload Identity)** preset.

**Cloudflare from anywhere** -- the cross-cloud shape: any cluster, records in Cloudflare, authenticated by a scoped API token, with per-record proxying (orange cloud) via `proxied`. Start from the **Cloudflare DNS from Any Cluster** preset.

**An out-of-tree provider via webhook** -- run the provider's own webhook image as a sidecar for every provider not built into the controller. Start from the **Webhook Provider (Out-of-Tree DNS)** preset.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- where the controller runs
- [**Ingress NGINX**](/infra-catalog/kubernetes-ingress-nginx) -- the entry point whose Ingress hostnames this controller publishes
- [**Cert Manager**](/infra-catalog/kubernetes-cert-manager) -- the HTTPS half of the same completeness story: a public endpoint needs its name resolvable AND its certificate signed
- [**AWS Route 53 Zone**](/infra-catalog/aws-route53-zone) -- the Route 53 zone referenced by `zoneIdFilters`
- [**GCP DNS Zone**](/infra-catalog/gcp-dns-zone) -- the Cloud DNS zone referenced by `zoneIdFilters`
- [**Azure DNS Zone**](/infra-catalog/azure-dns-zone) -- the Azure zone referenced by `zoneIdFilters`
- [**Cloudflare DNS Zone**](/infra-catalog/cloudflare-dns-zone) -- the Cloudflare zone referenced by `zoneIdFilters`
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- the IRSA identity and the cross-account `assumeRole` target
- [**GCP Service Account**](/infra-catalog/gcp-service-account) -- the identity GKE Workload Identity federates with
