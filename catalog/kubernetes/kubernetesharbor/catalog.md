# Harbor

Deploys Harbor -- the CNCF-graduated container registry that stores, signs, and scans OCI artifacts -- from the official `harbor` chart at `helm.goharbor.io` (chart 1.19.x = Harbor 2.15.x). Composition-first data arms cover the whole lifecycle: the chart's in-cluster PostgreSQL and Redis for evaluation (single-node, no failover -- evaluation-grade by upstream's own position), or a composed KubernetesPostgres and KubernetesValkey for production; artifact blobs live on a PersistentVolumeClaim, or on S3-compatible (a composed KubernetesSeaweedFs works in-cluster), GCS, or Azure Blob object storage. The Trivy vulnerability scanner ships ON by chart truth.

Know the address contract before you deploy: `externalUrl` is LOAD-BEARING for pushes and pulls -- Harbor embeds it in the token-service URL returned to every OCI client, so `docker login/push/pull` fail auth when the dialed address disagrees with it. Set it to what the exposure in front of Harbor actually serves.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Kubernetes Namespace** -- created only when `createNamespace` is `true`; otherwise deploys into an existing namespace
- **Credential Secrets** -- created BEFORE the release (the chart reads several at template time): the generated admin password in `<name>-admin-auth` (unless an existing Secret is named), inter-component secrets in `<name>-internal-auth`, and -- when declared -- the external Redis credential in `<name>-redis-auth` and storage credentials in `<name>-storage-auth`. The chart's publicly documented defaults (`Harbor12345`, `changeit`, `not-a-secure-key`) NEVER ship
- **Harbor Helm Release** -- the official `harbor` chart, creating:
  - Core (API + auth; first-boot schema migrations run in its startup window -- upstream budgets 360 x 10s), Portal (web UI), Registry (the OCI distribution backend with its registryctl sidecar), Jobservice (replication, GC, and scan jobs with a job-log PVC), and the nginx front door that terminates client traffic for every modeled exposure mode
  - The Trivy scanner StatefulSet with its vulnerability-database cache volume -- unless explicitly disabled
  - The chart's single-node PostgreSQL and/or Redis StatefulSets -- only on the internal arms; their data volumes are StatefulSet-template PVCs Helm NEVER deletes
  - The front-door Service named after this resource (ClusterIP by default; NodePort or LoadBalancer with the cloud-LB annotation surface when declared)
  - The optional harbor-exporter Deployment and ServiceMonitor when metrics are enabled
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target Kubernetes cluster. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **A storage class** for persistent volumes -- the filesystem artifact backend, the internal database/Redis arms, the jobservice log volume, and Trivy's DB cache all claim PVCs.
- **Outbound internet access (or the outbound proxy, or the air-gap posture)** -- Trivy downloads its vulnerability database at scan time and refreshes it roughly every 12 hours. In an air-gapped cluster set `skipUpdate`/`skipJavaDbUpdate` and pre-load the database onto the cache volume -- with `skipUpdate` and no pre-loaded DB, every scan fails.
- **The composed data plane, when production-bound** -- a KubernetesPostgres (declare the `registry` database at bootstrap via initdb), a KubernetesValkey, and an object-storage backend. External credential Secrets are read from Harbor's OWN namespace -- co-locate or replicate them.
- **The monitoring CRDs** (a KubernetesKubePrometheusStack) -- only when enabling the ServiceMonitor.

## Deploy

### Console

Open the deployment store, find **Harbor**, and click **Deploy**. The creation wizard walks you through placement, the registry address, exposure, admin credentials, the database/cache/storage arms, the Trivy scanner, per-component sizing, and the Helm-values escape hatch. Start from the **Minimal — evaluation registry, zero dependencies** preset for evaluation or **Production — composed data plane, object storage, HA components** for the fully-composed production shape in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesHarbor
metadata:
  name: harbor
  org: acme-corp
  env: prod
spec:
  namespace:
    value: harbor
  createNamespace: true
  externalUrl: http://localhost:8080
  database:
    internal: {}
  cache:
    internal: {}
  storage:
    filesystem:
      diskSize: 20Gi
```

```shell
planton apply -f harbor.yaml
```

This creates the smallest honest Harbor: the chart's in-cluster PostgreSQL and Redis, artifact blobs on a 20Gi PersistentVolumeClaim, Trivy scanning on (the chart default), a generated admin password exported as a Secret handle, and a ClusterIP front door reached through the exported `port_forward_command`. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, wire the composed data plane through ValueFromRef -- the InfraPipeline deploys the targets first:

```yaml
spec:
  database:
    external:
      host:
        valueFrom:
          kind: KubernetesPostgres
          name: harbor-pg
          fieldPath: status.outputs.rw_service
      username: harbor
      passwordSecretName:
        valueFrom:
          kind: KubernetesPostgres
          name: harbor-pg
          fieldPath: status.outputs.password_secret.name
```

The InfraPipeline resolves the dependency graph: the PostgreSQL cluster deploys first, then Harbor is installed against it.

## Key Configuration

These are the most important decisions when configuring Harbor on Kubernetes. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The registry address** -- `externalUrl` (`protocol://host[:port]`, no trailing slash) is not a display setting: token-service auth binds to it. An `https://` value with front-door TLS off is legitimate exactly when TLS terminates in front (a composed exposure or a cloud load balancer).

**Database and cache engines** -- exactly one arm each. The internal arms are evaluation-grade by upstream's own position. For production, a KubernetesPostgres composes end to end (its read-write Service tracks the current primary, and its application Secret carries the password under the chart's contract key `password` AS-IS); a KubernetesValkey provides the Redis address -- bridge its username-keyed auth Secret to the chart's `REDIS_PASSWORD` contract key, or declare the password and let the module materialize it.

**Artifact storage** -- exactly one backend. `filesystem` is the zero-dependency arm (ReadWriteMany is REQUIRED to run more than one registry replica on it -- enforced at validation). Object storage is the production posture; for in-cluster S3 stores clients cannot reach (SeaweedFS, MinIO), `disableRedirect` is required. S3 and GCS support keyless/ambient credentials (IRSA, Workload Identity); the Azure driver has no ambient chain -- exactly one credential arm is required.

**Admin credential and anonymous access** -- unset admin auth generates a random password into `<name>-admin-auth` (the recommended posture). Harbor serves PUBLIC-project metadata unauthenticated BY DESIGN: project visibility, not credential strength, is the boundary -- hardening means auditing which projects are public.

**Volume retention** -- `keepVolumesOnUninstall` defaults TRUE (chart truth: `helm.sh/resource-policy: keep` on the registry and jobservice PVCs). The INTERNAL database and Redis volumes are StatefulSet-template PVCs Helm never deletes regardless -- sweep them explicitly when retiring an install for good.

**Name budget** -- `metadata.name` caps at 39 characters (the chart appends up to 24 characters of component suffix and Kubernetes caps names at 63) -- enforced fail-loud in both engines.

**Helm value overrides** -- `helmValues` is a YAML document merged LAST over everything the typed fields render (Helm `-f` semantics). Use it for surfaces deliberately not modeled (per-component probes/placement, the swift/oss backends, GDPR knobs, registry middleware); the module re-pins `fullnameOverride` and the exposure Service name after the merge. Never put secret material in it.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |
| **KubernetesCertificate** | `expose.tls.certSecretName` | `status.outputs.secret_name` |
| **KubernetesPostgres** | `database.external.host` | `status.outputs.rw_service` |
| **KubernetesPostgres** | `database.external.passwordSecretName` | `status.outputs.password_secret.name` |
| **KubernetesValkey** | `cache.external.addr` | `status.outputs.kube_endpoint` |
| **KubernetesSeaweedFs** | `storage.s3.endpoint` | `status.outputs.s3_endpoint` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `namespace` | Namespace Harbor is installed into | Placement for composed resources |
| `expose_service` | The front-door Service (nginx) -- the ONE handle exposure kinds point at | Composed ingress/Gateway API exposure |
| `kube_endpoint` | In-cluster URL of the front door | In-cluster clients and health checks |
| `external_url` | The declared external URL -- the address OCI clients must use | CI/CD pipeline configuration |
| `core_service` | Core (API) Service name | API and webhook integrations |
| `portal_service` | Portal (web UI) Service name | Web UI access |
| `registry_service` | Registry (OCI distribution) Service name | In-cluster push/pull paths |
| `jobservice_service` | Jobservice Service name | Job monitoring |
| `trivy_service` | Trivy Service name -- empty when the scanner is disabled | Scanner integrations |
| `database_service` | Internal database Service -- set only on the internal arm | Evaluation-install tooling |
| `redis_service` | Internal Redis Service -- set only on the internal arm | Evaluation-install tooling |
| `admin_username` | Always `admin` (a Harbor constant) | Initial login and admin API calls |
| `admin_password_secret` | The Secret holding the admin password (name/key handle) | Automation and CI/CD authentication |
| `port_forward_command` | Ready-to-run `kubectl port-forward` command | Evaluation access -- pushes/pulls through the tunnel only authenticate when `externalUrl` matches the forwarded address |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Minimal registry for evaluation** -- the chart's in-cluster PostgreSQL and Redis, filesystem blobs, Trivy on, port-forward access. Start from the **Minimal — evaluation registry, zero dependencies** preset.

**Production composed registry** -- external PostgreSQL and Valkey by reference, S3-compatible blobs on a composed SeaweedFS with `disableRedirect`, a LoadBalancer front door with TLS from a composed Certificate, two replicas per component, and metrics with a ServiceMonitor. Start from the **Production — composed data plane, object storage, HA components** preset.

## Works With

- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) -- provides the namespace for Harbor's placement
- [**PostgreSQL**](/infra-catalog/kubernetes-postgres) -- the production database arm; its application Secret composes AS-IS
- [**Valkey**](/infra-catalog/kubernetes-valkey) -- the production cache arm (Redis protocol)
- [**SeaweedFS**](/infra-catalog/kubernetes-seaweed-fs) -- in-cluster S3-compatible artifact storage
- [**Cert Manager Certificate**](/infra-catalog/kubernetes-certificate) -- front-door and internal-TLS certificates
- [**kube-prometheus-stack**](/infra-catalog/kubernetes-kube-prometheus-stack) -- provides the ServiceMonitor CRDs and scrapes the exporter
