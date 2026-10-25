# Agent Substrate

Installs the Agent Substrate control plane: the system that runs a large fleet of mostly idle AI agents on a small set of ready worker pods. An idle agent's memory and disk are suspended into object storage and resumed in moments when a request arrives, each agent sandboxed under gVisor or a microVM. Your product creates agents through Substrate's API; this kind runs everything that API needs.

This kind installs the **control plane**, once per cluster. Agents run on worker pools (Agent Substrate Worker Pool), declared on top of it, with the runtimes Agent Substrate Sandbox Configs pin.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions, from Planton's digest-pinned build of the Agent Substrate release `version` names (v0.6.0 by default):

- **The `ate-system` namespace** (with `createNamespace: true`) and **`podcertificate-controller-system`** — both names are fixed by upstream. Exactly one installation per cluster.
- **The resource definitions** — `WorkerPool`, `SandboxConfig` and `CSIDriverConfig`, applied as kept, ownership-checked resources, so destroying the control plane never deletes a worker pool.
- **Generated secrets** — five certificate authorities and the actor token signing key with its public key set, in upstream's format, kept in state.
- **The pod-certificate controller** — signs the certificates every component's mTLS uses, through Kubernetes' pod-certificate APIs.
- **The database** — upstream's bundled PostgreSQL 18 with its setup run for you, or your own PostgreSQL, or Cloud SQL with keyless IAM login.
- **The control plane** — the API server, identity server, controller, node agent (one per node), ingress router and egress gateway, with every typed setting applied onto upstream's own objects, and the network policies that fence them. The deploy waits until each has rolled out.
- **The credential provider** (optional) — upstream's Kubernetes Secrets provider, which injects credentials into agents' outbound calls from the namespaces each atespace is granted.
- **The default gVisor runtime** — the `gvisor-default` SandboxConfig worker pools use unless they name another.

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** — an active connection in the Connect module with credentials for the target cluster.

### Kubernetes Cluster

- **The pod-certificate APIs** — PodCertificateRequest and ClusterTrustBundle, with ClusterTrustBundle projection. On GKE: version 1.36 created with the `certificates.k8s.io/v1beta1` APIs enabled, or 1.37. They cannot be added to a running GKE cluster.
- **NetworkPolicy enforcement** — the install's isolation is a default-deny ingress policy.
- **Object storage** — a GCS bucket (GKE) or any S3-compatible bucket, and for S3 a Secret in `ate-system` holding its access key.
- **On GKE with Cloud SQL or GCS** — the Google service accounts the API server and node agent act as, with their grants and Workload Identity bindings, and with GCS the transfer service account every snapshot moves as, which the API server may impersonate.
- **Your product's identity** — the service account your product calls the API and the agents as: the API's first owner, and the router's allowed client.

## Deploy

### Console

Open the deployment store, find **Agent Substrate**, and click **Deploy**. The creation wizard walks you through the platform, the database, the storage, the egress credential provider, and each component's settings. Start from the **Local or generic preset**, or the **GKE keyless preset** on Google Cloud.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesAgentSubstrate
metadata:
  name: agent-substrate
  org: acme-corp
  env: dev
spec:
  namespace:
    value: ate-system
  createNamespace: true
  platform:
    generic:
      serviceAccountIssuer: https://kubernetes.default.svc.cluster.local
  database:
    bundled:
      volumeSize: 20Gi
  storage:
    s3:
      endpoint:
        value: http://main-s3.storage.svc.cluster.local:8333
      bucket: agent-snapshots
      credentials:
        secretName:
          value: substrate-s3
  apiServer:
    authorization:
      bootstrapOwners:
        - system:serviceaccount:my-product:agent-orchestrator
  router:
    allowedClientSpiffeIds:
      - spiffe://cluster.local/ns/my-product/sa/agent-orchestrator
  egress:
    credentialProvider:
      disabled: {}
```

```shell
planton apply -f agent-substrate.yaml
```

This installs the control plane with the bundled database, snapshots into an S3-compatible bucket, no credential injection, and your product's orchestrator as the API's owner and the router's one client. An Infra Job tracks the provisioning in real time.

### InfraChart

On GKE, keyless end to end. The cluster, the bucket, the Cloud SQL instance and the three service accounts are resources of their own, and the control plane references their outputs:

```yaml
spec:
  platform:
    gke:
      clusterId:
        valueFrom:
          kind: GcpGkeCluster
          name: agents
          fieldPath: status.outputs.cluster_id
      apiServerServiceAccount:
        valueFrom:
          kind: GcpServiceAccount
          name: ate-api-server
          fieldPath: status.outputs.email
      nodeAgentServiceAccount:
        valueFrom:
          kind: GcpServiceAccount
          name: atelet
          fieldPath: status.outputs.email
      snapshotTransferServiceAccount:
        valueFrom:
          kind: GcpServiceAccount
          name: ate-snapshot-transfer
          fieldPath: status.outputs.email
  database:
    cloudSql:
      instanceConnectionName:
        valueFrom:
          kind: GcpCloudSql
          name: substrate
          fieldPath: status.outputs.connection_name
  storage:
    gcs:
      bucket:
        valueFrom:
          kind: GcpGcsBucket
          name: agent-snapshots
          fieldPath: status.outputs.bucket_name
```

The InfraPipeline creates the cluster, the three service accounts (with a GcpServiceAccountIamMember letting the API server's mint the transfer account's tokens), the Cloud SQL instance and the bucket first, then installs the control plane with every identity and address wired by reference.

## Key Configuration

These are the decisions that matter most. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The platform** — `gke` keeps GKE's managed metrics and image credential provider, trusts the cluster's own token issuer, and unlocks keyless GCS and Cloud SQL. `generic` runs on any conformant cluster and needs the cluster's service-account issuer, which `kubectl get --raw /.well-known/openid-configuration` prints; a wrong issuer refuses every caller, so it has no default.

**The database** — `bundled` is upstream's PostgreSQL in the namespace, set up for you and reached with certificates rather than passwords; its volume outlives a destroy while another resource owns the namespace (with `createNamespace: true` the destroy deletes the namespace, and the volume with it). `external` uses your PostgreSQL through connection strings in a Secret, after you run upstream's setup script on it once. `cloudSql` uses Cloud SQL through the Auth Proxy, logging in as the API server's Google service account.

**The credential provider is chosen, never inherited** — `disabled`, `kubernetesSecrets` (upstream's provider, answering only for the namespaces each atespace is granted), or `external` (a provider you run, such as upstream's Google Secret Manager provider).

**Who may call** — the API authorizes every call: a caller in a namespace carrying `ate.dev/api-access=true` reaches an atespace only through an access policy, and `apiServer.authorization.bootstrapOwners` names the first owners (a service account calling with a token is `system:serviceaccount:<namespace>:<name>`). The router carries requests to agents only from the workloads `router.allowedClientSpiffeIds` names, by their pod certificates (`spiffe://cluster.local/ns/<namespace>/sa/<name>`).

**The version** — `version` picks the release, v0.6.0 by default. A later upstream release that changes only its images keeps v0.6.0's install shape, so you adopt it by setting `version`, with no catalog update; then redeploy your worker pools. A release of another shape, or one Planton has not published, is refused before anything is applied. Before 1.0, a new minor is a reinstall, and the deploy refuses one over an existing installation or its database with the steps that move it safely.

**Mirrors and air gaps** — `imageRegistry` moves every image the install runs to your registry, keeping each path and digest. `us-central1-docker.pkg.dev` pulls Planton's own mirror.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | Reference Path |
|------------|-------|-------------------|
| **Kubernetes Namespace** | `namespace` | `spec.name` |
| **GCP GKE Cluster** | `platform.gke.clusterId` | `status.outputs.cluster_id` |
| **GCP Service Account** | `platform.gke.apiServerServiceAccount`, `platform.gke.nodeAgentServiceAccount`, `platform.gke.snapshotTransferServiceAccount` | `status.outputs.email` |
| **GCP Cloud SQL** | `database.cloudSql.instanceConnectionName` | `status.outputs.connection_name` |
| **GCP GCS Bucket** | `storage.gcs.bucket` | `status.outputs.bucket_name` |
| **SeaweedFS** | `storage.s3.endpoint` | `status.outputs.s3_endpoint` |
| **Kubernetes Secret** | `storage.s3.credentials.secretName`, the external database's connection strings | `status.outputs.secret_name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can reference:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `api_address` | `api.ate-system.svc:443` | Your product's Substrate client |
| `api_server_name` | The name the API's certificate is issued for | TLS verification in the client |
| `api_token_audience` | The audience a caller's token must carry | A projected service-account token |
| `api_trust_bundle_name` | The ClusterTrustBundle that verifies the API | A clusterTrustBundle projected volume |
| `api_access_namespace_label` | `ate.dev/api-access=true` | The label on your product's namespace |
| `router_tls_endpoint` | Where requests reach an agent, over mutual TLS | Your product's data path |
| `router_endpoint` | agentgateway's plaintext endpoint (empty with Envoy) | A product on the agentgateway router |
| `router_client_signer_name` | The signer of the client certificates the TLS endpoint accepts | A client workload's podCertificate volume mount |
| `storage_location` | Where snapshots are stored: `gs://` or `s3://` with the bucket and prefix | Finding an agent's snapshots |
| `gvisor_worker_image`, `microvm_worker_image` | The worker images built for this release | A worker pool |
| `actor_jwt_issuer`, `version` | The actor token issuer (served by the identity server), the installed release | Verifying actor tokens, upgrades |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Local or generic** — the bundled database and an S3-compatible store, on kind or any conformant cluster. Start from the **Local or generic preset**.

**GKE keyless** — Cloud SQL with IAM login and a GCS bucket, through Workload Identity, with no key anywhere and no node able to read another agent's snapshots. Start from the **GKE keyless preset**.

**Credential injection** — the Kubernetes Secrets provider, so an agent's outbound call carries a credential it never sees. Start from the **Credential injection preset**.

**Production** — the control plane on its own nodes, three API servers, your orchestrator as owner and router client, and telemetry to your collector. Start from the **Production preset**.

## Works With

- [**Agent Substrate Worker Pool**](/infra-catalog/kubernetes-agent-substrate-worker-pool) — the workers agents run on, from this release's worker image outputs.
- [**Agent Substrate Sandbox Config**](/infra-catalog/kubernetes-agent-substrate-sandbox-config) — a runtime of your own beside `gvisor-default`.
- [**Agent Substrate CSI Driver Config**](/infra-catalog/kubernetes-agent-substrate-csi-driver-config) — your CSI driver registered for agents' external volumes.
- [**Kubernetes Namespace**](/infra-catalog/kubernetes-namespace) — own `ate-system` so the Secrets this kind reads can be placed there first.
- [**Kubernetes Secret**](/infra-catalog/kubernetes-secret) — the S3 access key, or an external database's connection strings.
- [**SeaweedFS**](/infra-catalog/kubernetes-seaweed-fs) — an in-cluster S3 store for snapshots.
- [**GKE Cluster**](/infra-catalog/gcp-gke-cluster) — a cluster created with the pod-certificate APIs.
- [**GCP Service Account**](/infra-catalog/gcp-service-account) — the identities the API server and node agent act as, and the one snapshots move as.
- [**GCP Service Account IAM Member**](/infra-catalog/gcp-service-account-iam-member) — the API server's grant to mint the transfer account's tokens.
- [**Cloud SQL**](/infra-catalog/gcp-cloud-sql) — the managed database, with IAM login.
- [**GCS Bucket**](/infra-catalog/gcp-gcs-bucket) — the snapshot store on Google Cloud.
