# KubernetesNamespace

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

**KubernetesNamespaceSpec** defines the configuration for creating and managing a Kubernetes namespace.
This spec implements a "Namespace-as-a-Service" pattern, abstracting the complexity of namespace
configuration, resource quotas, network policies, and access control into a batteries-included,
production-ready primitive.

Based on extensive research into namespace deployment patterns, this spec focuses on the 80/20 rule:
exposing the 20% of configuration options that deliver 80% of the value while maintaining
flexibility for advanced use cases.

## Example

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesNamespace
metadata:
  name: test-namespace
spec:
  name: test-namespace
  labels:
    team: platform-engineering
    environment: test
    cost-center: engineering
    created-by: planton
  annotations:
    description: "Test namespace created by Planton"
  resource_profile:
    preset: small
  network_config:
    isolate_ingress: true
    restrict_egress: true
    allowed_ingress_namespaces:
      - "kube-system"
      - "istio-system"
    allowed_egress_cidrs:
      - "10.0.0.0/8"
  pod_security_standard: baseline
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.name` | `string` | yes |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.resourceProfile` | `KubernetesNamespaceResourceProfile` |  |  |  |
| `spec.resourceProfile.preset` | `enum` |  |  |  |
| `spec.resourceProfile.custom` | `KubernetesNamespaceCustomQuotas` |  |  |  |
| `spec.resourceProfile.custom.cpu` | `KubernetesNamespaceCpuQuota` |  |  |  |
| `spec.resourceProfile.custom.cpu.requests` | `string` | yes |  |  |
| `spec.resourceProfile.custom.cpu.limits` | `string` | yes |  |  |
| `spec.resourceProfile.custom.memory` | `KubernetesNamespaceMemoryQuota` |  |  |  |
| `spec.resourceProfile.custom.memory.requests` | `string` | yes |  |  |
| `spec.resourceProfile.custom.memory.limits` | `string` | yes |  |  |
| `spec.resourceProfile.custom.objectCounts` | `KubernetesNamespaceObjectCountQuotas` |  |  |  |
| `spec.resourceProfile.custom.objectCounts.pods` | `int32` |  |  |  |
| `spec.resourceProfile.custom.objectCounts.services` | `int32` |  |  |  |
| `spec.resourceProfile.custom.objectCounts.configmaps` | `int32` |  |  |  |
| `spec.resourceProfile.custom.objectCounts.secrets` | `int32` |  |  |  |
| `spec.resourceProfile.custom.objectCounts.persistentVolumeClaims` | `int32` |  |  |  |
| `spec.resourceProfile.custom.objectCounts.loadBalancers` | `int32` |  |  |  |
| `spec.resourceProfile.custom.defaultLimits` | `KubernetesNamespaceDefaultLimits` |  |  |  |
| `spec.resourceProfile.custom.defaultLimits.defaultCpuRequest` | `string` | yes |  |  |
| `spec.resourceProfile.custom.defaultLimits.defaultCpuLimit` | `string` | yes |  |  |
| `spec.resourceProfile.custom.defaultLimits.defaultMemoryRequest` | `string` | yes |  |  |
| `spec.resourceProfile.custom.defaultLimits.defaultMemoryLimit` | `string` | yes |  |  |
| `spec.resourceProfile.custom.additionalHardLimits` | `map<string, string>` |  |  |  |
| `spec.networkConfig` | `KubernetesNamespaceNetworkConfig` |  |  |  |
| `spec.networkConfig.isolateIngress` | `bool` |  |  |  |
| `spec.networkConfig.restrictEgress` | `bool` |  |  |  |
| `spec.networkConfig.allowedIngressNamespaces` | `[]string` |  |  |  |
| `spec.networkConfig.allowedEgressCidrs` | `[]string` |  |  |  |
| `spec.networkConfig.allowedEgressDomains` | `[]string` |  |  |  |
| `spec.serviceMeshConfig` | `KubernetesNamespaceServiceMeshConfig` |  |  |  |
| `spec.serviceMeshConfig.enabled` | `bool` |  |  |  |
| `spec.serviceMeshConfig.meshType` | `enum` |  |  |  |
| `spec.serviceMeshConfig.revisionTag` | `string` |  |  |  |
| `spec.podSecurityStandard` | `enum` |  |  |  |

## Field Details

### spec.name

`string` · required

The unique name of the namespace.
This will be used as the Kubernetes namespace metadata.name.
Must be a valid DNS label (lowercase alphanumeric and hyphens).

- rule: Name must be a valid DNS label (lowercase alphanumeric and hyphens, no leading/trailing hyphens)
- rule: {"string":{"minLen":"1","maxLen":"63"}}

### spec.labels

`map<string, string>`

Additional labels to be applied to the namespace.
These are merged with standard labels (environment, team, cost-center) for
cost allocation, monitoring, and governance.

### spec.annotations

`map<string, string>`

Additional annotations to be applied to the namespace.
Common use cases:
- linkerd.io/inject: "enabled" for service mesh injection
- janitor/ttl: "24h" for ephemeral namespace cleanup
- scheduler.alpha.kubernetes.io/node-selector: for node affinity

### spec.resourceProfile

`KubernetesNamespaceResourceProfile`

Resource allocation profile for the namespace.
This abstracts ResourceQuota and LimitRange configuration into T-shirt sizes
or allows custom specifications for advanced users.

### spec.resourceProfile.preset

`enum`

Pre-defined T-shirt size profiles (small, medium, large, xlarge).
These provide opinionated defaults for different namespace sizes.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `built_in_profile_unspecified` -- Unspecified profile. System will apply a default minimal profile.
- `small` -- Small profile for development and testing. CPU: 2 cores request / 4 cores limit Memory: 4Gi request / 8Gi limit Pods: 20, Services: 10, ConfigMaps: 50, Secrets: 50
- `medium` -- Medium profile for staging and small production workloads. CPU: 4 cores request / 8 cores limit Memory: 8Gi request / 16Gi limit Pods: 50, Services: 20, ConfigMaps: 100, Secrets: 100
- `large` -- Large profile for production workloads. CPU: 8 cores request / 16 cores limit Memory: 16Gi request / 32Gi limit Pods: 100, Services: 40, ConfigMaps: 200, Secrets: 200
- `xlarge` -- Extra large profile for high-scale production workloads. CPU: 16 cores request / 32 cores limit Memory: 32Gi request / 64Gi limit Pods: 200, Services: 80, ConfigMaps: 400, Secrets: 400

### spec.resourceProfile.custom

`KubernetesNamespaceCustomQuotas`

Custom resource limits for advanced users who need precise control.

### spec.resourceProfile.custom.cpu

`KubernetesNamespaceCpuQuota`

CPU resource limits

### spec.resourceProfile.custom.cpu.requests

`string` · required

Total CPU requests allowed (e.g., "4", "4000m")

- rule: {"string":{"minLen":"1"}}

### spec.resourceProfile.custom.cpu.limits

`string` · required

Total CPU limits allowed (e.g., "8", "8000m")

- rule: {"string":{"minLen":"1"}}

### spec.resourceProfile.custom.memory

`KubernetesNamespaceMemoryQuota`

Memory resource limits

### spec.resourceProfile.custom.memory.requests

`string` · required

Total memory requests allowed (e.g., "8Gi", "8192Mi")

- rule: {"string":{"minLen":"1"}}

### spec.resourceProfile.custom.memory.limits

`string` · required

Total memory limits allowed (e.g., "16Gi", "16384Mi")

- rule: {"string":{"minLen":"1"}}

### spec.resourceProfile.custom.objectCounts

`KubernetesNamespaceObjectCountQuotas`

Object count limits (pods, services, configmaps, secrets)

### spec.resourceProfile.custom.objectCounts.pods

`int32`

Maximum number of pods

- rule: {"int32":{"gte":1}}

### spec.resourceProfile.custom.objectCounts.services

`int32`

Maximum number of services

- rule: {"int32":{"gte":1}}

### spec.resourceProfile.custom.objectCounts.configmaps

`int32`

Maximum number of configmaps

- rule: {"int32":{"gte":1}}

### spec.resourceProfile.custom.objectCounts.secrets

`int32`

Maximum number of secrets

- rule: {"int32":{"gte":1}}

### spec.resourceProfile.custom.objectCounts.persistentVolumeClaims

`int32`

Maximum number of persistent volume claims

- rule: {"int32":{"gte":0}}

### spec.resourceProfile.custom.objectCounts.loadBalancers

`int32`

Maximum number of load balancers

- rule: {"int32":{"gte":0}}

### spec.resourceProfile.custom.defaultLimits

`KubernetesNamespaceDefaultLimits`

Default resource requests/limits for containers without explicit values

### spec.resourceProfile.custom.defaultLimits.defaultCpuRequest

`string` · required

Default CPU request (e.g., "100m")

- rule: {"string":{"minLen":"1"}}

### spec.resourceProfile.custom.defaultLimits.defaultCpuLimit

`string` · required

Default CPU limit (e.g., "1000m")

- rule: {"string":{"minLen":"1"}}

### spec.resourceProfile.custom.defaultLimits.defaultMemoryRequest

`string` · required

Default memory request (e.g., "128Mi")

- rule: {"string":{"minLen":"1"}}

### spec.resourceProfile.custom.defaultLimits.defaultMemoryLimit

`string` · required

Default memory limit (e.g., "512Mi")

- rule: {"string":{"minLen":"1"}}

### spec.resourceProfile.custom.additionalHardLimits

`map<string, string>`

Additional ResourceQuota hard limits beyond the typed CPU/memory/object-count fields,
as quota resource name → quantity. Kubernetes models quota as an open map, so anything
it accepts is expressible here: storage ("requests.storage": "500Gi",
"<class>.storageclass.storage.k8s.io/requests.storage": "100Gi"), extended resources
("requests.nvidia.com/gpu": "4"), object counts for any resource
("count/jobs.batch": "30"), and scoped ephemeral storage. Entries here are merged into
the same ResourceQuota; keys that collide with the typed fields are rejected by the
cluster, so keep the typed fields authoritative for CPU, memory, and the standard
object counts.

- rule: {"map":{"keys":{"string":{"minLen":"1"}},"values":{"string":{"minLen":"1"}}}}

### spec.networkConfig

`KubernetesNamespaceNetworkConfig`

Network isolation and security configuration.
Controls ingress/egress network policies to enforce zero-trust networking.

### spec.networkConfig.isolateIngress

`bool`

Enable ingress network isolation.
When true, creates a NetworkPolicy that denies all ingress traffic by default.
Traffic must be explicitly allowed.

### spec.networkConfig.restrictEgress

`bool`

Enable egress network restriction.
When true, blocks all egress except to kube-system (DNS) and Kubernetes API.
This prevents pods from accessing external networks without explicit rules.

### spec.networkConfig.allowedIngressNamespaces

`[]string`

Allow ingress from specific namespaces.
List of namespace names that are allowed to send traffic to this namespace.

### spec.networkConfig.allowedEgressCidrs

`[]string`

Allow egress to specific external CIDR blocks.
List of CIDR blocks (e.g., "10.0.0.0/8", "192.168.1.0/24") that pods can access.

### spec.networkConfig.allowedEgressDomains

`[]string`

Allow egress to specific DNS domains.
List of domains (e.g., "api.stripe.com", "*.github.com") that pods can access.
Requires a CNI that supports DNS-based policies (like Calico or Cilium).

### spec.serviceMeshConfig

`KubernetesNamespaceServiceMeshConfig`

Service mesh integration configuration.
Enables automatic sidecar injection and mesh-specific features.

### spec.serviceMeshConfig.enabled

`bool`

Enable service mesh sidecar injection.
When true, service mesh proxies are automatically injected into pods.

### spec.serviceMeshConfig.meshType

`enum`

Service mesh type.
Determines which mesh to use (Istio, Linkerd, etc.).

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `service_mesh_type_unspecified` -- No service mesh
- `istio` -- Istio service mesh
- `linkerd` -- Linkerd service mesh
- `consul` -- Consul Connect service mesh

### spec.serviceMeshConfig.revisionTag

`string`

Mesh revision tag (Istio-specific).
Allows pointing to a specific control plane version without hardcoding the version.
Example: "prod-stable", "canary", "1-19-5"
This enables safe, granular mesh upgrades.

- rule: {"string":{"maxLen":"63"}}

### spec.podSecurityStandard

`enum`

Pod security standards enforcement level.
Defines the security posture for pods running in the namespace.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `pod_security_standard_unspecified` -- No enforcement (permissive)
- `privileged` -- Privileged: Unrestricted policy, allows known privilege escalations. Use only for system-level workloads (monitoring, logging, CNI).
- `baseline` -- Baseline: Minimally restrictive, prevents known privilege escalations. Suitable for most applications that don't need special permissions.
- `restricted` -- Restricted: Heavily restricted, follows pod hardening best practices. Use for security-sensitive workloads. May require significant configuration.

## Validation Rules

- `service_mesh_requires_mesh_type`: mesh_type must be set when service mesh is enabled

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesNamespace, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.namespace` | `string` | The name of the created Kubernetes namespace. This is the primary identifier for referencing the namespace in other resources. |
| `status.outputs.namespace_id` | `string` | The fully qualified namespace identifier. Format: <namespace> This is the same as namespace but provided for consistency with other components. |
| `status.outputs.resource_quotas_applied` | `string` | Indicates whether resource quotas were applied to the namespace. "true" if ResourceQuota objects were created, "false" otherwise. |
| `status.outputs.limit_ranges_applied` | `string` | Indicates whether LimitRanges were applied to the namespace. "true" if LimitRange objects were created, "false" otherwise. |
| `status.outputs.network_policies_applied` | `string` | Indicates whether network policies were applied to the namespace. "true" if NetworkPolicy objects were created, "false" otherwise. |
| `status.outputs.service_mesh_enabled` | `string` | Indicates whether service mesh sidecar injection is enabled. "true" if the namespace is configured for automatic sidecar injection, "false" otherwise. |
| `status.outputs.service_mesh_type` | `string` | The service mesh type that is configured for this namespace. Empty string if no service mesh is configured. Possible values: "istio", "linkerd", "consul", "" |
| `status.outputs.pod_security_standard` | `string` | The pod security standard enforcement level applied to this namespace. Possible values: "privileged", "baseline", "restricted", "" |
| `status.outputs.labels_json` | `string` | List of Kubernetes labels applied to the namespace. This is a JSON string representation of the labels map for easy reference. |
| `status.outputs.annotations_json` | `string` | List of Kubernetes annotations applied to the namespace. This is a JSON string representation of the annotations map for easy reference. |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpDataprocCluster | `spec.virtualClusterConfig.kubernetesClusterConfig.kubernetesNamespace` | `spec.name` |
| KubernetesAirflow | `spec.namespace` | `spec.name` |
| KubernetesAltinityOperator | `spec.namespace` | `spec.name` |
| KubernetesArgoWorkflows | `spec.namespace` | `spec.name` |
| KubernetesArgocd | `spec.namespace` | `spec.name` |
| KubernetesAuthorizationPolicy | `spec.namespace` | `spec.name` |
| KubernetesBackendTlsPolicy | `spec.namespace` | `spec.name` |
| KubernetesCertManager | `spec.namespace` | `spec.name` |
| KubernetesCertificate | `spec.namespace` | `spec.name` |
| KubernetesCilium | `spec.namespace` | `spec.name` |
| KubernetesClickHouse | `spec.namespace` | `spec.name` |
| KubernetesCloudNativePgOperator | `spec.namespace` | `spec.name` |
| KubernetesClusterAutoscaler | `spec.namespace` | `spec.name` |
| KubernetesConfigMap | `spec.namespace` | `spec.name` |
| KubernetesCronJob | `spec.namespace` | `spec.name` |
| KubernetesDaemonSet | `spec.namespace` | `spec.name` |
| KubernetesDeployment | `spec.namespace` | `spec.name` |
| KubernetesDestinationRule | `spec.namespace` | `spec.name` |
| KubernetesEnvoyFilter | `spec.namespace` | `spec.name` |
| KubernetesExternalDns | `spec.namespace` | `spec.name` |
| KubernetesExternalSecret | `spec.namespace` | `spec.name` |
| KubernetesExternalSecretsOperator | `spec.namespace` | `spec.name` |
| KubernetesFlinkDeployment | `spec.namespace` | `spec.name` |
| KubernetesFlinkOperator | `spec.namespace` | `spec.name` |
| KubernetesGatekeeper | `spec.namespace` | `spec.name` |
| KubernetesGateway | `spec.namespace` | `spec.name` |
| KubernetesGhaRunnerScaleSet | `spec.namespace` | `spec.name` |
| KubernetesGhaRunnerScaleSetController | `spec.namespace` | `spec.name` |
| KubernetesGrafana | `spec.namespace` | `spec.name` |
| KubernetesGrpcRoute | `spec.namespace` | `spec.name` |
| KubernetesHarbor | `spec.namespace` | `spec.name` |
| KubernetesHelmRelease | `spec.namespace` | `spec.name` |
| KubernetesHorizontalPodAutoscaler | `spec.namespace` | `spec.name` |
| KubernetesHttpRoute | `spec.namespace` | `spec.name` |
| KubernetesIngress | `spec.namespace` | `spec.name` |
| KubernetesIngressNginx | `spec.namespace` | `spec.name` |
| KubernetesIssuer | `spec.namespace` | `spec.name` |
| KubernetesIstio | `spec.namespace` | `spec.name` |
| KubernetesJenkins | `spec.namespace` | `spec.name` |
| KubernetesJob | `spec.namespace` | `spec.name` |
| KubernetesJupyterHub | `spec.namespace` | `spec.name` |
| KubernetesKafka | `spec.namespace` | `spec.name` |
| KubernetesKafkaConnect | `spec.namespace` | `spec.name` |
| KubernetesKafkaConnector | `spec.namespace` | `spec.name` |
| KubernetesKafkaMirrorMaker2 | `spec.namespace` | `spec.name` |
| KubernetesKafkaTopic | `spec.namespace` | `spec.name` |
| KubernetesKafkaUi | `spec.namespace` | `spec.name` |
| KubernetesKafkaUser | `spec.namespace` | `spec.name` |
| KubernetesKarapace | `spec.namespace` | `spec.name` |
| KubernetesKarpenter | `spec.namespace` | `spec.name` |
| KubernetesKeda | `spec.namespace` | `spec.name` |
| KubernetesKeycloak | `spec.namespace` | `spec.name` |
| KubernetesKeycloakOperator | `spec.namespace` | `spec.name` |
| KubernetesKubePrometheusStack | `spec.namespace` | `spec.name` |
| KubernetesKubeRayOperator | `spec.namespace` | `spec.name` |
| KubernetesKyverno | `spec.namespace` | `spec.name` |
| KubernetesListenerSet | `spec.namespace` | `spec.name` |
| KubernetesLocust | `spec.namespace` | `spec.name` |
| KubernetesLoki | `spec.namespace` | `spec.name` |
| KubernetesManifest | `spec.namespace` | `spec.name` |
| KubernetesMetricsServer | `spec.namespace` | `spec.name` |
| KubernetesMlflow | `spec.namespace` | `spec.name` |
| KubernetesMongodb | `spec.namespace` | `spec.name` |
| KubernetesMysql | `spec.namespace` | `spec.name` |
| KubernetesNats | `spec.namespace` | `spec.name` |
| KubernetesNeo4j | `spec.namespace` | `spec.name` |
| KubernetesNetworkPolicy | `spec.namespace` | `spec.name` |
| KubernetesOpenBao | `spec.namespace` | `spec.name` |
| KubernetesOpenFga | `spec.namespace` | `spec.name` |
| KubernetesOpenSearch | `spec.namespace` | `spec.name` |
| KubernetesOpenSearchOperator | `spec.namespace` | `spec.name` |
| KubernetesOtelCollector | `spec.namespace` | `spec.name` |
| KubernetesOtelOperator | `spec.namespace` | `spec.name` |
| KubernetesPeerAuthentication | `spec.namespace` | `spec.name` |
| KubernetesPerconaMongoOperator | `spec.namespace` | `spec.name` |
| KubernetesPerconaMysqlOperator | `spec.namespace` | `spec.name` |
| KubernetesPersistentVolumeClaim | `spec.namespace` | `spec.name` |
| KubernetesPlantonOperator | `spec.namespace` | `spec.name` |
| KubernetesPlantonPlatform | `spec.namespace` | `spec.name` |
| KubernetesPlantonRunner | `spec.namespace` | `spec.name` |
| KubernetesPodDisruptionBudget | `spec.namespace` | `spec.name` |
| KubernetesPostgres | `spec.namespace` | `spec.name` |
| KubernetesPrometheusRule | `spec.namespace` | `spec.name` |
| KubernetesQdrant | `spec.namespace` | `spec.name` |
| KubernetesRabbitMq | `spec.namespace` | `spec.name` |
| KubernetesRayCluster | `spec.namespace` | `spec.name` |
| KubernetesRbac | `spec.namespaceScope.namespace` | `spec.name` |
| KubernetesRbac | `spec.subjects[].serviceAccount.namespace` | `spec.name` |
| KubernetesReferenceGrant | `spec.namespace` | `spec.name` |
| KubernetesRequestAuthentication | `spec.namespace` | `spec.name` |
| KubernetesResourceQuota | `spec.namespace` | `spec.name` |
| KubernetesSeaweedFs | `spec.namespace` | `spec.name` |
| KubernetesSecret | `spec.namespace` | `spec.name` |
| KubernetesSecretStore | `spec.namespace` | `spec.name` |
| KubernetesService | `spec.namespace` | `spec.name` |
| KubernetesServiceAccount | `spec.namespace` | `spec.name` |
| KubernetesServiceEntry | `spec.namespace` | `spec.name` |
| KubernetesSignoz | `spec.namespace` | `spec.name` |
| KubernetesSolr | `spec.namespace` | `spec.name` |
| KubernetesSolrOperator | `spec.namespace` | `spec.name` |
| KubernetesSparkOperator | `spec.namespace` | `spec.name` |
| KubernetesStatefulSet | `spec.namespace` | `spec.name` |
| KubernetesStrimziKafkaOperator | `spec.namespace` | `spec.name` |
| KubernetesSuperset | `spec.namespace` | `spec.name` |
| KubernetesTcpRoute | `spec.namespace` | `spec.name` |
| KubernetesTelemetry | `spec.namespace` | `spec.name` |
| KubernetesTempo | `spec.namespace` | `spec.name` |
| KubernetesTemporal | `spec.namespace` | `spec.name` |
| KubernetesTlsRoute | `spec.namespace` | `spec.name` |
| KubernetesTrino | `spec.namespace` | `spec.name` |
| KubernetesUdpRoute | `spec.namespace` | `spec.name` |
| KubernetesValkey | `spec.namespace` | `spec.name` |
| KubernetesVelero | `spec.namespace` | `spec.name` |

## See Also

- [Overview](../README.md)
