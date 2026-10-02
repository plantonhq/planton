# KubernetesExternalDns

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**KubernetesExternalDnsSpec** installs ExternalDNS — the controller that
watches Kubernetes resources (Services, Ingresses, Gateway API routes, ...)
and publishes matching DNS records into a DNS provider — from the official
Helm chart (`external-dns` at https://kubernetes-sigs.github.io/external-dns/).

Kubernetes is a dual provider: the cluster runs IN one environment while the
DNS zone often lives in ANOTHER (an EKS cluster publishing to Cloudflare, a
GKE cluster publishing to Route 53). This spec therefore separates the two
halves cleanly: `dns_provider` selects WHERE records are written, and
`workload_identity` (plus per-provider static credentials) selects HOW the
controller authenticates — any host/provider combination is expressible.

One installation manages one DNS provider. Clusters that publish to several
providers (or with different TXT owner IDs) deploy MULTIPLE instances of
this component — the release is named after `metadata.name`, so instances
coexist naturally; give each its own `txt_owner_id` so their registries
never fight over record ownership.

The typed fields cover the chart's meaningful configuration surface;
`helm_values` remains as the escape hatch for values beyond them (merged
last, Helm `-f` semantics, identical on both engines) — a safety valve,
never the primary interface.

## Example

```yaml
# Full-surface offline-proof manifest: exercises the Cloudflare arm
# (declared credential + provider flags + zone references), the complete
# watching/sync/registry surface, workload identity, scheduling, sizing,
# observability, image override, and the helm_values escape hatch — so the
# offline tofu plan and pulumi preview proofs cover arms the live kind
# lanes exclude. Placeholder values; never applied to a real cluster.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesExternalDns
metadata:
  name: hack-external-dns
spec:
  namespace:
    value: external-dns
  createNamespace: true
  chartVersion: 1.21.1
  cloudflare:
    apiToken: hack-placeholder-token
    zoneIdFilters:
      - value: 023e105f4ecef8ad9ca31a8372d0c353
    proxied: true
    dnsRecordsPerPage: 5000
  workloadIdentity:
    eks:
      roleArn:
        value: arn:aws:iam::123456789012:role/external-dns
  sources:
    - service
    - ingress
    - gateway-httproute
    - crd
  gatewayListenerSets: true
  policy: sync
  registry: txt
  txtOwnerId: hack-cluster
  txtPrefix: "edns-"
  domainFilters:
    - example.org
  excludeDomains:
    - internal.example.org
  annotationFilter: "external-dns.alpha.kubernetes.io/include=true"
  labelFilter: "team=platform"
  managedRecordTypes:
    - A
    - AAAA
    - CNAME
  interval: 2m
  triggerLoopOnEvent: true
  namespaced: false
  logLevel: debug
  logFormat: json
  resources:
    requests:
      cpu: 50m
      memory: 64Mi
    limits:
      cpu: 200m
      memory: 256Mi
  nodeSelector:
    kubernetes.io/os: linux
  tolerations:
    - key: dedicated
      operator: Equal
      value: platform
      effect: NoSchedule
  priorityClassName: system-cluster-critical
  prometheus:
    serviceMonitor: true
    serviceMonitorInterval: 1m
    serviceMonitorLabels:
      release: kube-prometheus-stack
  imageRepository: registry.example.com/external-dns/external-dns
  imageTag: v0.21.0
  helmValues: |
    revisionHistoryLimit: 3
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.createNamespace` | `bool` |  |  |  |
| `spec.chartVersion` | `string` |  | `1.21.1` |  |
| `spec.awsRoute53` | `KubernetesExternalDnsAwsRoute53` |  |  |  |
| `spec.awsRoute53.region` | `string` |  |  |  |
| `spec.awsRoute53.zoneIdFilters` | `[]string \| valueFrom` |  |  | AwsRoute53Zone (`status.outputs.zone_id`) |
| `spec.awsRoute53.zoneType` | `string` |  |  |  |
| `spec.awsRoute53.assumeRole` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.awsRoute53.assumeRoleExternalId` | `string` |  |  |  |
| `spec.awsRoute53.accessKeyId` | `string` |  |  |  |
| `spec.awsRoute53.secretAccessKey` | `string` (sensitive) |  |  |  |
| `spec.googleCloudDns` | `KubernetesExternalDnsGoogleCloudDns` |  |  |  |
| `spec.googleCloudDns.project` | `string \| valueFrom` | yes |  | GcpProject (`status.outputs.project_id`) |
| `spec.googleCloudDns.zoneIdFilters` | `[]string \| valueFrom` |  |  | GcpDnsZone (`status.outputs.zone_id`) |
| `spec.googleCloudDns.zoneVisibility` | `string` |  |  |  |
| `spec.googleCloudDns.serviceAccountKeyJson` | `string` (sensitive) |  |  |  |
| `spec.azureDns` | `KubernetesExternalDnsAzureDns` |  |  |  |
| `spec.azureDns.resourceGroup` | `string` | yes |  |  |
| `spec.azureDns.subscriptionId` | `string` | yes |  |  |
| `spec.azureDns.tenantId` | `string` |  |  |  |
| `spec.azureDns.privateZones` | `bool` |  |  |  |
| `spec.azureDns.zoneIdFilters` | `[]string \| valueFrom` |  |  | AzureDnsZone (`status.outputs.zone_id`) |
| `spec.azureDns.managedIdentityClientId` | `string` |  |  |  |
| `spec.azureDns.clientId` | `string` |  |  |  |
| `spec.azureDns.clientSecret` | `string` (sensitive) |  |  |  |
| `spec.cloudflare` | `KubernetesExternalDnsCloudflare` |  |  |  |
| `spec.cloudflare.apiToken` | `string` (sensitive) | yes |  |  |
| `spec.cloudflare.zoneIdFilters` | `[]string \| valueFrom` |  |  | CloudflareDnsZone (`status.outputs.zone_id`) |
| `spec.cloudflare.proxied` | `bool` |  |  |  |
| `spec.cloudflare.dnsRecordsPerPage` | `uint32` |  |  |  |
| `spec.webhook` | `KubernetesExternalDnsWebhook` |  |  |  |
| `spec.webhook.imageRepository` | `string` | yes |  |  |
| `spec.webhook.imageTag` | `string` |  |  |  |
| `spec.webhook.args` | `[]string` |  |  |  |
| `spec.webhook.env` | `map<string, string>` |  |  |  |
| `spec.webhook.resources` | `ContainerResources` |  |  |  |
| `spec.webhook.resources.limits` | `CpuMemory` |  |  |  |
| `spec.webhook.resources.limits.cpu` | `string` |  |  |  |
| `spec.webhook.resources.limits.memory` | `string` |  |  |  |
| `spec.webhook.resources.requests` | `CpuMemory` |  |  |  |
| `spec.webhook.resources.requests.cpu` | `string` |  |  |  |
| `spec.webhook.resources.requests.memory` | `string` |  |  |  |
| `spec.inMemory` | `KubernetesExternalDnsInMemory` |  |  |  |
| `spec.inMemory.zones` | `[]string` |  |  |  |
| `spec.workloadIdentity` | `KubernetesWorkloadIdentity` |  |  |  |
| `spec.workloadIdentity.gke` | `KubernetesWorkloadIdentityGke` |  |  |  |
| `spec.workloadIdentity.gke.serviceAccountEmail` | `string \| valueFrom` | yes |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.workloadIdentity.eks` | `KubernetesWorkloadIdentityEksIrsa` |  |  |  |
| `spec.workloadIdentity.eks.roleArn` | `string \| valueFrom` | yes |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.workloadIdentity.aks` | `KubernetesWorkloadIdentityAks` |  |  |  |
| `spec.workloadIdentity.aks.clientId` | `string \| valueFrom` | yes |  | AzureUserAssignedIdentity (`status.outputs.client_id`) |
| `spec.workloadIdentity.aks.tenantId` | `string` |  |  |  |
| `spec.sources` | `[]string` |  |  |  |
| `spec.gatewayListenerSets` | `bool` |  |  |  |
| `spec.policy` | `string` |  | `upsert-only` |  |
| `spec.registry` | `string` |  | `txt` |  |
| `spec.txtOwnerId` | `string` |  |  |  |
| `spec.txtPrefix` | `string` |  |  |  |
| `spec.txtSuffix` | `string` |  |  |  |
| `spec.dynamodbTable` | `string` |  |  |  |
| `spec.dynamodbRegion` | `string` |  |  |  |
| `spec.domainFilters` | `[]string` |  |  |  |
| `spec.excludeDomains` | `[]string` |  |  |  |
| `spec.annotationFilter` | `string` |  |  |  |
| `spec.labelFilter` | `string` |  |  |  |
| `spec.managedRecordTypes` | `[]string` |  |  |  |
| `spec.interval` | `string` |  | `1m` |  |
| `spec.triggerLoopOnEvent` | `bool` |  |  |  |
| `spec.namespaced` | `bool` |  |  |  |
| `spec.logLevel` | `string` |  | `info` |  |
| `spec.logFormat` | `string` |  | `text` |  |
| `spec.resources` | `ContainerResources` |  |  |  |
| `spec.resources.limits` | `CpuMemory` |  |  |  |
| `spec.resources.limits.cpu` | `string` |  |  |  |
| `spec.resources.limits.memory` | `string` |  |  |  |
| `spec.resources.requests` | `CpuMemory` |  |  |  |
| `spec.resources.requests.cpu` | `string` |  |  |  |
| `spec.resources.requests.memory` | `string` |  |  |  |
| `spec.nodeSelector` | `map<string, string>` |  |  |  |
| `spec.tolerations` | `[]WorkloadToleration` |  |  |  |
| `spec.tolerations[].key` | `string` |  |  |  |
| `spec.tolerations[].operator` | `string` |  |  |  |
| `spec.tolerations[].value` | `string` |  |  |  |
| `spec.tolerations[].effect` | `string` |  |  |  |
| `spec.tolerations[].tolerationSeconds` | `int64` |  |  |  |
| `spec.priorityClassName` | `string` |  |  |  |
| `spec.prometheus` | `KubernetesExternalDnsPrometheus` |  |  |  |
| `spec.prometheus.serviceMonitor` | `bool` |  |  |  |
| `spec.prometheus.serviceMonitorInterval` | `string` |  |  |  |
| `spec.prometheus.serviceMonitorLabels` | `map<string, string>` |  |  |  |
| `spec.imageRepository` | `string` |  |  |  |
| `spec.imageTag` | `string` |  |  |  |
| `spec.helmValues` | `string` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace to install ExternalDNS into ("external-dns" by convention).
Accepts a literal namespace name or a reference to a KubernetesNamespace
resource.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.createNamespace

`bool`

When true, the namespace is created (with the standard Planton governance
labels) before installing and deleted with the resource. When false, the
namespace must already exist.

### spec.chartVersion

`string` · optional (explicit presence)

Helm chart version to install (chart releases are cut separately from the
controller: chart 1.21.x ships controller v0.21.x — the chart's
appVersion decides the image tag unless image_tag overrides it). Pin
deliberately; upgrades re-run the release with the new chart. Pick
versions from the chart repository's index (`helm search repo`): the
served chart is the contract — the upstream source tree's Chart.yaml
can claim a version at a tag that was never served.

- default: `1.21.1`

### spec.awsRoute53

`KubernetesExternalDnsAwsRoute53`

AWS Route 53. Keyless on EKS via `workload_identity.eks` (IRSA) —
the production posture; static keys are the fallback for non-EKS hosts
without ambient AWS credentials.

- rule: access_key_id and secret_access_key form one credential — set both or neither

### spec.awsRoute53.region

`string`

AWS region for the Route 53 API client (Route 53 itself is global, but
the SDK requires a region, and the DynamoDB registry uses it). E.g.
"us-east-1".

### spec.awsRoute53.zoneIdFilters

`[]string | valueFrom`

Restrict management to these hosted zones. Accepts literal zone IDs
(e.g. "Z104533312EOZ72FQZ4TT") or references to AwsRoute53Zone
resources — in an infra chart the zone and this component deploy in one
run with the ID flowing in as a reference. Empty = every zone the
credentials can see (filter with domain_filters at minimum).

containment_exempt: a watch-scope filter — the controller manages
records IN the zone; it does not deploy into it.

- references: AwsRoute53Zone (`status.outputs.zone_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsRoute53Zone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.awsRoute53.zoneType

`string` · optional (explicit presence)

Filter for zones of this type. Empty = both.

- rule: {"string":{"in":["","public","private"]}}

### spec.awsRoute53.assumeRole

`string | valueFrom`

IAM role to assume for Route 53 calls (full ARN) — the cross-account
pattern: the controller authenticates in the cluster's account, then
assumes this role in the account that owns the zones. Accepts a literal
ARN or a reference to an AwsIamRole resource's output.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.awsRoute53.assumeRoleExternalId

`string`

External ID presented when assuming assume_role (the confused-deputy
guard some role trust policies require).

### spec.awsRoute53.accessKeyId

`string`

Static AWS access key ID. Prefer keyless (workload_identity.eks / node
role) — set static keys only for clusters with no ambient AWS identity.
Both halves of the key pair must be set together; the module
materializes them as a Kubernetes Secret wired into the controller
environment.

### spec.awsRoute53.secretAccessKey

`string` · sensitive

Static AWS secret access key (the secret half of the key pair).

### spec.googleCloudDns

`KubernetesExternalDnsGoogleCloudDns`

Google Cloud DNS. Keyless on GKE via `workload_identity.gke`
(Workload Identity); a service-account key is the fallback for
non-GKE hosts.

### spec.googleCloudDns.project

`string | valueFrom` · required

GCP project that owns the Cloud DNS zones. Accepts a literal project ID
or a reference to a GcpProject resource's output.

containment_exempt: names the project whose zones the controller
manages — the controller itself runs in the cluster, not the project.

- references: GcpProject (`status.outputs.project_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.googleCloudDns.zoneIdFilters

`[]string | valueFrom`

Restrict management to these Cloud DNS zones (zone names/IDs). Accepts
literals or references to GcpDnsZone resources. Empty = every zone in
the project.

containment_exempt: a watch-scope filter — the controller manages
records IN the zone; it does not deploy into it.

- references: GcpDnsZone (`status.outputs.zone_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.googleCloudDns.zoneVisibility

`string` · optional (explicit presence)

Filter for zones with this visibility. Empty = both.

- rule: {"string":{"in":["","public","private"]}}

### spec.googleCloudDns.serviceAccountKeyJson

`string` · sensitive

Static GCP service-account key (JSON). Prefer keyless
(workload_identity.gke) — set a key only for clusters with no ambient
GCP identity. The module materializes it as a Kubernetes Secret mounted
into the controller with GOOGLE_APPLICATION_CREDENTIALS pointed at it.

### spec.azureDns

`KubernetesExternalDnsAzureDns`

Azure DNS (public or private zones). Keyless on AKS via
`workload_identity.aks` (Azure AD Workload Identity) or a node
managed identity; a service-principal secret is the fallback.

- rule: client_id and client_secret form one service-principal credential — set both or neither
- rule: Service-principal authentication needs tenant_id — the Entra tenant the principal lives in

### spec.azureDns.resourceGroup

`string` · required

Resource group that contains the DNS zones.

- rule: {"required":true}

### spec.azureDns.subscriptionId

`string` · required

Azure subscription that owns the zones.

- rule: {"required":true}

### spec.azureDns.tenantId

`string`

Entra (Azure AD) tenant. Required for service-principal auth; optional
otherwise.

### spec.azureDns.privateZones

`bool`

When true, manage Azure Private DNS zones (the "azure-private-dns"
upstream provider) instead of public zones.

### spec.azureDns.zoneIdFilters

`[]string | valueFrom`

Restrict management to these DNS zones. Accepts literal zone IDs or
references to AzureDnsZone resources. Empty = every zone in the
resource group.

containment_exempt: a watch-scope filter — the controller manages
records IN the zone; it does not deploy into it.

- references: AzureDnsZone (`status.outputs.zone_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.azureDns.managedIdentityClientId

`string`

Client ID of a user-assigned managed identity to authenticate with
(azure.json useManagedIdentityExtension). For AKS clusters using
kubelet/user-assigned identities rather than Workload Identity.

### spec.azureDns.clientId

`string`

Service-principal application (client) ID. Both halves of the
service-principal credential must be set together; the module wires
them into azure.json.

### spec.azureDns.clientSecret

`string` · sensitive

Service-principal client secret (the secret half).

### spec.cloudflare

`KubernetesExternalDnsCloudflare`

Cloudflare DNS. Always token-authenticated (Cloudflare has no
workload-identity federation with Kubernetes clusters) — the canonical
cross-cloud arm: any cluster, records in Cloudflare.

### spec.cloudflare.apiToken

`string` · required · sensitive

Cloudflare API token. The module materializes it as a Kubernetes Secret
wired into the controller environment (CF_API_TOKEN) — it never appears
in chart values or pod specs. Note: the controller validates the token
at first zone sync, not at startup — a revoked/invalid token surfaces as
a crash-looping pod with a Cloudflare 4xx in its logs.

- rule: {"required":true}

### spec.cloudflare.zoneIdFilters

`[]string | valueFrom`

Restrict management to these Cloudflare zones. Accepts literal zone IDs
or references to CloudflareDnsZone resources. Empty = every zone the
token can see.

containment_exempt: a watch-scope filter — the controller manages
records IN the zone; it does not deploy into it.

- references: CloudflareDnsZone (`status.outputs.zone_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: CloudflareDnsZone, name: <that resource's name>, fieldPath: status.outputs.zone_id}} -- a bare string does not parse

### spec.cloudflare.proxied

`bool`

When true, created records are proxied through Cloudflare (orange
cloud): CDN/WAF in front of the record. Per-resource override via the
external-dns.alpha.kubernetes.io/cloudflare-proxied annotation.

### spec.cloudflare.dnsRecordsPerPage

`uint32` · optional (explicit presence)

DNS records fetched per Cloudflare API page. Upstream default: 100;
raise (max 5000) for zones with thousands of records to cut sync-loop
API calls.

- rule: {"uint32":{"lte":5000}}

### spec.webhook

`KubernetesExternalDnsWebhook`

Webhook provider — upstream's extension architecture for every DNS
provider that is not in-tree. Runs the provider's webhook implementation
as a sidecar container next to the controller.

### spec.webhook.imageRepository

`string` · required

Webhook provider image repository (e.g. a provider's published
external-dns webhook image).

- rule: {"required":true}

### spec.webhook.imageTag

`string`

Webhook provider image tag.

### spec.webhook.args

`[]string`

Arguments passed to the webhook container (provider-specific
configuration).

### spec.webhook.env

`map<string, string>`

Environment variables for the webhook container. Values land in chart
values as-is — put secrets in Kubernetes Secrets and reference them via
helm_values env entries with valueFrom instead of inlining them here.

### spec.webhook.resources

`ContainerResources`

Webhook container CPU/memory requests and limits.

### spec.webhook.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.webhook.resources.limits.cpu

`string`

### spec.webhook.resources.limits.memory

`string`

### spec.webhook.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.webhook.resources.requests.cpu

`string`

### spec.webhook.resources.requests.memory

`string`

### spec.inMemory

`KubernetesExternalDnsInMemory`

In-memory provider — upstream's built-in sandbox. Records live only in
the controller pod's memory (lost on restart, visible in logs/metrics).
For evaluating source/filter/policy behavior without touching any real
DNS zone; never for production.

### spec.inMemory.zones

`[]string`

Zones to pre-create in the in-memory store (e.g. "example.org").
Records for names outside these zones are skipped, mirroring real
provider zone matching.

### spec.workloadIdentity

`KubernetesWorkloadIdentity`

Binds the controller ServiceAccount to a cloud identity for KEYLESS
provider authentication (Route 53 via EKS IRSA, Cloud DNS via GKE
Workload Identity, Azure DNS via AKS Workload Identity). Provider arms
whose static credentials are left empty authenticate through this
identity (or the node's ambient identity). Not used by token-based
providers (Cloudflare) or the webhook/in-memory arms.

### spec.workloadIdentity.gke

`KubernetesWorkloadIdentityGke`

GKE Workload Identity: annotate the ServiceAccount with a GCP service account email.

### spec.workloadIdentity.gke.serviceAccountEmail

`string | valueFrom` · required

GCP service account email, e.g. "dns-manager@my-project.iam.gserviceaccount.com".
Applied as the `iam.gke.io/gcp-service-account` annotation. Accepts a literal
email or a reference to a GcpServiceAccount resource's output.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.workloadIdentity.eks

`KubernetesWorkloadIdentityEksIrsa`

EKS IRSA: annotate the ServiceAccount with an AWS IAM role ARN.

### spec.workloadIdentity.eks.roleArn

`string | valueFrom` · required

AWS IAM role ARN, e.g. "arn:aws:iam::123456789012:role/dns-manager".
Applied as the `eks.amazonaws.com/role-arn` annotation. Accepts a literal ARN
or a reference to an AwsIamRole resource's output.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.workloadIdentity.aks

`KubernetesWorkloadIdentityAks`

Azure AD Workload Identity: annotate the ServiceAccount with an Entra application
(or user-assigned managed identity) client ID.

### spec.workloadIdentity.aks.clientId

`string | valueFrom` · required

Client ID (GUID) of the user-assigned managed identity or Entra application.
Applied as the `azure.workload.identity/client-id` annotation. Accepts a literal
GUID or a reference to an AzureUserAssignedIdentity resource's output.

- references: AzureUserAssignedIdentity (`status.outputs.client_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureUserAssignedIdentity, name: <that resource's name>, fieldPath: status.outputs.client_id}} -- a bare string does not parse

### spec.workloadIdentity.aks.tenantId

`string` · optional (explicit presence)

Entra tenant ID (GUID). Optional: only needed for cross-tenant scenarios; when
omitted the azure-workload-identity webhook uses its default tenant. Applied as
the `azure.workload.identity/tenant-id` annotation when set.

- rule: {"string":{"uuid":true}}

### spec.sources

`[]string`

Kubernetes resource types the controller watches for DNS names. Chart
default: ["service", "ingress"]. Add Gateway API route sources
("gateway-httproute", ...) when routes carry the hostnames, or "crd" to
manage records declaratively via DNSEndpoint objects.

- rule: {"repeated":{"items":{"string":{"in":["service","ingress","node","pod","gateway-httproute","gateway-grpcroute","gateway-tlsroute","gateway-tcproute","gateway-udproute","istio-gateway","istio-virtualservice","contour-httpproxy","gloo-proxy","fake","connector","crd","empty","skipper-routegroup","openshift-route","ambassador-host","kong-tcpingress","f5-virtualserver","f5-transportserver","traefik-proxy","unstructured"]}}}}

### spec.gatewayListenerSets

`bool`

When true, the controller also follows Gateway API ListenerSets: a route
attached to a listener that a KubernetesListenerSet adds to a shared
Gateway gets its record, not only a route on the Gateway's own
listeners. Turn it on wherever components bring their own hostname and
certificate as a listener set (a monitoring hub beside a platform on one
Gateway is the common case); the Gateway must also admit listener sets
(`allowed_listeners` on KubernetesGateway). Renders the chart's
`enableGatewayListenerSets`, which adds the `--gateway-listener-sets`
flag and the permission to read ListenerSets. The chart grants that
permission only alongside a Gateway API route source, so at least one
`gateway-*` entry in `sources` is required. Default false.

### spec.policy

`string` · optional (explicit presence)

How aggressively cluster state is pushed to the DNS zone.
"upsert-only" (default): create and update records, never delete —
the safe default for zones shared with records managed elsewhere.
"sync": full reconciliation including deletes of records this instance
owns (per the TXT registry) — the right choice for dedicated zones.
"create-only": create and never touch again.

- default: `upsert-only`
- rule: {"string":{"in":["create-only","sync","upsert-only"]}}

### spec.registry

`string` · optional (explicit presence)

How record ownership is tracked, so multiple ExternalDNS instances (and
humans) can share a zone without overwriting each other.
"txt" (default): a TXT record per managed name carries the owner ID.
"noop": no ownership tracking — the instance considers every record in
the zone fair game; only for zones this instance exclusively owns.
"dynamodb": ownership tracked in a DynamoDB table (AWS deployments that
must keep zones free of TXT metadata).
"aws-sd": AWS Cloud Map service discovery instead of Route 53 zones.

- default: `txt`
- rule: {"string":{"in":["txt","noop","dynamodb","aws-sd"]}}

### spec.txtOwnerId

`string`

Identifier this instance writes into the ownership registry (TXT or
DynamoDB). Every instance sharing a zone MUST have a distinct owner ID —
it is what stops one instance from deleting another's records under the
"sync" policy. Upstream default: "default".

### spec.txtPrefix

`string`

Prefix for ownership TXT record names (e.g. "edns-"), keeping them
visually grouped and avoiding CNAME collisions (a TXT record cannot
coexist with a CNAME of the same name — prefixing sidesteps that).
May contain the "%{record_type}" template. Mutually exclusive with
txt_suffix.

### spec.txtSuffix

`string`

Suffix appended to the host portion of ownership TXT record names.
May contain the "%{record_type}" template. Mutually exclusive with
txt_prefix.

### spec.dynamodbTable

`string`

DynamoDB table name for the "dynamodb" registry. Upstream default:
"external-dns". Only meaningful when registry is "dynamodb".

### spec.dynamodbRegion

`string`

AWS region of the DynamoDB registry table. Only meaningful when registry
is "dynamodb".

### spec.domainFilters

`[]string`

Restrict management to zones/records whose domain ends with one of these
suffixes (e.g. "example.com"). Empty = every zone the provider
credentials can see. The first guardrail against touching unrelated
zones in a shared account.

### spec.excludeDomains

`[]string`

Domains to explicitly exclude, carving exceptions out of domain_filters
(e.g. filter "example.com" but exclude "internal.example.com").

### spec.annotationFilter

`string`

Only manage resources whose annotations match this selector (e.g.
"external-dns.alpha.kubernetes.io/include=true"). Empty = no annotation
gating. Label-selector syntax.

### spec.labelFilter

`string`

Only manage resources matching this label selector (more efficient than
annotation_filter — filtering happens API-side). Empty = no label gating.

### spec.managedRecordTypes

`[]string`

DNS record types the controller manages. Upstream default: A, AAAA,
CNAME. Extend deliberately (e.g. add "NS") — every listed type is
subject to the sync policy including deletes.

### spec.interval

`string` · optional (explicit presence)

Reconciliation interval (Go duration, e.g. "1m", "5m"). Chart default:
"1m". Raise it for providers with tight API rate limits.

- default: `1m`
- rule: {"string":{"pattern":"^([0-9]+(\\.[0-9]+)?(ns|us|µs|ms|s|m|h))+$"}}

### spec.triggerLoopOnEvent

`bool`

When true, a source create/update/delete triggers reconciliation
immediately instead of waiting for the next interval — snappier record
propagation at the cost of more provider API calls.

### spec.namespaced

`bool`

When true, the controller runs namespace-scoped: it watches only its own
namespace and gets a Role instead of a ClusterRole. For multi-tenant
clusters where each team runs its own instance.

### spec.logLevel

`string` · optional (explicit presence)

Log verbosity. Chart default: "info".

- default: `info`
- rule: {"string":{"in":["panic","debug","info","warning","error","fatal"]}}

### spec.logFormat

`string` · optional (explicit presence)

Log output format. Chart default: "text"; "json" for log pipelines.

- default: `text`
- rule: {"string":{"in":["text","json"]}}

### spec.resources

`ContainerResources`

Controller container CPU/memory requests and limits. Empty = chart
defaults (no explicit resources).

### spec.resources.limits

`CpuMemory`

The resource limits for the container.
Specify the maximum amount of CPU and memory that the container can use.

### spec.resources.limits.cpu

`string`

### spec.resources.limits.memory

`string`

### spec.resources.requests

`CpuMemory`

The resource requests for the container.
Specify the minimum amount of CPU and memory that the container is guaranteed.

### spec.resources.requests.cpu

`string`

### spec.resources.requests.memory

`string`

### spec.nodeSelector

`map<string, string>`

Node selector for the controller pod. Empty = chart default.

### spec.tolerations

`[]WorkloadToleration`

Tolerations for the controller pod.

### spec.tolerations[].key

`string`

Taint key to tolerate. Empty key with operator "Exists" tolerates every taint.

### spec.tolerations[].operator

`string`

How key/value match: "Equal" (default — value must match too) or "Exists"
(key presence alone matches).

- rule: Toleration operator must be either "Equal" or "Exists"

### spec.tolerations[].value

`string`

Taint value to match when operator is "Equal".

### spec.tolerations[].effect

`string`

Which taint effect is tolerated: "NoSchedule", "PreferNoSchedule", or
"NoExecute". Empty tolerates all effects for the key.

- rule: Toleration effect must be one of "NoSchedule", "PreferNoSchedule", or "NoExecute"

### spec.tolerations[].tolerationSeconds

`int64` · optional (explicit presence)

For "NoExecute" taints only: how many seconds already-running pods stay bound
after the taint appears. Unset means tolerate forever.

### spec.priorityClassName

`string`

PriorityClass for the controller pod (e.g. "system-cluster-critical" on
clusters where DNS publication is infrastructure-critical).

### spec.prometheus

`KubernetesExternalDnsPrometheus`

Prometheus metrics exposure. The metrics port (7979) is always served;
the ServiceMonitor is opt-in and requires the Prometheus operator CRDs
on the cluster.

### spec.prometheus.serviceMonitor

`bool`

Create a ServiceMonitor for scrape discovery. Requires the Prometheus
operator CRDs (e.g. kube-prometheus-stack) on the cluster — the release
FAILS to install without them.

### spec.prometheus.serviceMonitorInterval

`string`

Scrape interval for the ServiceMonitor (e.g. "1m"). Empty = Prometheus
default.

### spec.prometheus.serviceMonitorLabels

`map<string, string>`

Extra labels on the ServiceMonitor — how a Prometheus instance's
serviceMonitorSelector finds it (e.g. {"release": "kube-prometheus-stack"}).

### spec.imageRepository

`string`

Full image repository override (registry + path) for the controller
image — the air-gapped/mirror knob. Empty = the chart default
(registry.k8s.io/external-dns/external-dns).

### spec.imageTag

`string`

Controller image tag override. Empty = the chart's appVersion (the
controller release the pinned chart was cut for) — override only to hold
the controller back or roll it forward independently of the chart.

### spec.helmValues

`string`

Escape hatch: additional chart values as a YAML document, merged LAST
over everything the typed fields render (Helm `-f` semantics, identical
on both engines). For the chart surface beyond the typed fields — never
the substitute for them. Do not put secrets here.

## Validation Rules

- `externaldns.provider_required`: Select the DNS provider records are written to — set exactly one of aws_route53, google_cloud_dns, azure_dns, cloudflare, webhook, or in_memory
- `externaldns.txt_prefix_xor_suffix`: txt_prefix and txt_suffix are mutually exclusive — ownership TXT names can be prefixed or suffixed, not both
- `externaldns.dynamodb_requires_registry`: dynamodb_table and dynamodb_region configure the "dynamodb" registry — set registry to "dynamodb" or clear these fields
- `externaldns.listener_sets_require_gateway_source`: gateway_listener_sets follows routes on listener sets, so it needs a Gateway API route source — add a gateway-* entry (e.g. "gateway-httproute") to sources

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesExternalDns, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | Namespace ExternalDNS is installed in. |
| `status.outputs.release_name` | `string` | Helm release name (equals metadata.name — multiple instances of this component coexist in one cluster, each with its own release). |
| `status.outputs.service_account_name` | `string` | Name of the controller ServiceAccount. The cloud-side half of a keyless binding (IAM role trust policy, Workload Identity binding, federated credential) references this name together with the namespace. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |
| `spec.awsRoute53.zoneIdFilters` | AwsRoute53Zone | `status.outputs.zone_id` |
| `spec.awsRoute53.assumeRole` | AwsIamRole | `status.outputs.role_arn` |
| `spec.googleCloudDns.project` | GcpProject | `status.outputs.project_id` |
| `spec.googleCloudDns.zoneIdFilters` | GcpDnsZone | `status.outputs.zone_id` |
| `spec.azureDns.zoneIdFilters` | AzureDnsZone | `status.outputs.zone_id` |
| `spec.cloudflare.zoneIdFilters` | CloudflareDnsZone | `status.outputs.zone_id` |
| `spec.workloadIdentity.gke.serviceAccountEmail` | GcpServiceAccount | `status.outputs.email` |
| `spec.workloadIdentity.eks.roleArn` | AwsIamRole | `status.outputs.role_arn` |
| `spec.workloadIdentity.aks.clientId` | AzureUserAssignedIdentity | `status.outputs.client_id` |

## See Also

- [Overview](../README.md)
