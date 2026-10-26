# KubernetesAgentSubstrate Guide

The judgment this guide carries: this kind is the one Agent Substrate control plane on a cluster. It installs upstream's own release the way upstream's installer composes it, generates the secrets upstream's installer generates, and refuses the things that do irreversible damage: taking over someone else's installation, starting a new minor release on an older one's data, and installing a release whose shape this catalog does not know.

## One per cluster, in fixed namespaces

Everything installs into `ate-system`, and the pod-certificate controller into `podcertificate-controller-system`, with upstream's own resource names. The namespace is not a field because upstream bakes it in three times: in the manifests, in the API's server name (`api.ate-system.svc`) and in its token audience. Who owns `ate-system` is a choice, as for the agent-sandbox controller:

- **`createNamespace: true`** — this resource creates it and deletes it on destroy.
- **A KubernetesNamespace named `ate-system`**, referenced from `namespace` — right when a Secret this kind reads must be there first: the S3 access key, an external database's connection strings, image pull secrets.

`podcertificate-controller-system` holds only the pod-certificate controller and three of the CA pools, so it is always this resource's. The deploy refuses a namespace that is missing while another resource owns it, one someone else created while this resource would create it, and turning `createNamespace` off on a live installation.

## The cluster it needs

Substrate's components authenticate each other with certificates that Kubernetes itself issues, through the PodCertificateRequest and ClusterTrustBundle APIs, signed by Substrate's own pod-certificate controller. A cluster without those APIs cannot run any of it.

- **GKE** — version 1.36 created with `k8sBetaApis` enabling `certificates.k8s.io/v1beta1/podcertificaterequests` and `clustertrustbundles`, or 1.37. They can only be enabled at cluster creation.
- **Other clusters** — Kubernetes 1.37 with the `ClusterTrustBundle`, `ClusterTrustBundleProjection` and `PodCertificateRequest` feature gates and the `certificates.k8s.io/v1beta1` API, as Substrate's own kind clusters run.

The network must enforce NetworkPolicy: every pod in `ate-system` is behind a default-deny ingress policy, with allowances for exactly the paths the components use. A caller reaches the API only from a namespace labelled `ate.dev/api-access=true` (the `api_access_namespace_label` output).

## Who may call

Two doors, each closed to anyone the spec does not name:

- **The API** authorizes every call. A caller authenticates with its service-account token or its pod certificate, and reaches an atespace only through an access policy. `apiServer.authorization.bootstrapOwners` names the principals that always own everything, which is how the first owner gets in: name the service account your product calls the API as. Left empty, it is upstream's `system:serviceaccount:ate-system:ate-client`, which only someone who can create that service account in `ate-system` can act as.
- **The router** carries requests to agents only from the workloads `router.allowedClientSpiffeIds` names, by the SPIFFE ID of the pod certificate they present (`spiffe://cluster.local/ns/<namespace>/sa/<service-account>`, from the signer the `router_client_signer_name` output names) at the `router_tls_endpoint` output. The Envoy router serves nothing else, so it needs at least one client. agentgateway checks no client certificate yet and still serves plaintext at `router_endpoint` to any caller that reaches it.

The identity server publishes the keys agent tokens are signed with at the actor-JWT issuer (`https://idp.ate-system.svc` unless `apiServer.actorJwt.issuer` names your own host), so a service verifies a token an agent presents without calling the API. The issuer is a host, with no path, because the identity server serves the discovery document at its root.

## The platform

`platform.gke` and `platform.generic` decide three things.

**Whose tokens the API trusts.** The API server authenticates a caller by its Kubernetes service-account token, so it must know the cluster's token issuer. On GKE that is the cluster's own (`https://container.googleapis.com/v1/projects/<p>/locations/<l>/clusters/<c>`), derived from `clusterId`. Elsewhere, `serviceAccountIssuer` names it; read it with `kubectl get --raw /.well-known/openid-configuration`. For the two in-cluster forms (`https://kubernetes.default.svc` and its `.cluster.local` sibling, which kind and kubeadm clusters use), the API server verifies tokens against the cluster itself, as upstream's installer arranges.

**The GKE-only pieces.** GKE keeps the router's Managed Service for Prometheus `PodMonitoring`, and the node agent's mount of GKE's image credential provider, which lets agents pull private Artifact Registry images with the node's identity. Elsewhere both are left out, exactly as upstream's kind overlay leaves them out.

**Keyless access.** GCS storage and Cloud SQL authenticate through GKE Workload Identity, so they need `platform.gke`; the spec refuses them elsewhere. `apiServerServiceAccount` and `nodeAgentServiceAccount` name the Google service accounts the two components act as (the annotation on their Kubernetes service accounts). Leave them empty to grant the roles to each Kubernetes service account's own Workload Identity principal instead, which is upstream's own setup. With GCS, `snapshotTransferServiceAccount` names a third account, which every snapshot transfer runs as (below).

## The database

The API server keeps every atespace, template, actor and policy in PostgreSQL, through two pools: an owner pool that runs schema migrations at every start, and a read/write pool that serves requests. Each assumes a NOLOGIN role, so the login users can be anything that may assume them.

- **Bundled** — upstream's PostgreSQL 18 StatefulSet. The API server connects with client certificates the pod-certificate controller issues; no password exists. Upstream's installer runs the setup script (roles, users, schema) by exec after every deploy, which no IaC engine can do, so the database's own start hook runs it instead, with upstream's exact variables, in one transaction (the script is idempotent). The volume (`data-postgres-0`, 500Gi by default) is created by the StatefulSet and outlives a destroy while another resource owns `ate-system`; with `createNamespace: true` the destroy deletes the namespace, and the volume with it.
- **External** — your PostgreSQL. Run upstream's `pkg/postgressetup/setup.sql` on it once, then give the connection strings by Secret reference; they reach the API server as environment entries read from your Secret, never copied. A CA certificate for `sslmode=verify-full` goes in `serverCaCertificate`, mounted at `/run/postgres-server-ca/server-ca.pem`.
- **Cloud SQL** — through upstream's Cloud SQL Auth Proxy sidecar. With `iamAuthentication` on (the default), the proxy logs in as the API server's Google service account, keylessly: create the Cloud SQL IAM user named after it, and run the setup script once. With it off, give a password connection string by Secret reference.

Changing a database setting rolls the API server: its pod template carries upstream's own environment digest.

## Storage

Snapshots of suspended agents, and the sandbox runtime files every node downloads, move through an object-storage sidecar beside the API server and on every node.

- **GCS** — the bucket, keyless through Workload Identity. Every snapshot moves as one transfer service account (`platform.gke.snapshotTransferServiceAccount`): the API server impersonates it and hands each node agent a token narrowed to the one agent it moves, so a node can never read another agent's snapshots. Grants:
  - the API server's account: `storage.objectAdmin` and `storage.bucketViewer` on the bucket, and `roles/iam.serviceAccountTokenCreator` on the transfer account (a GcpServiceAccountIamMember);
  - the transfer account: `storage.objectAdmin` and `storage.bucketViewer` on the bucket;
  - the node agent's account: nothing on the bucket; `artifactregistry.reader` on the project for private agent images, and, for microVM workers, `storage.objectViewer` on the asset bucket holding their guest kernels and images (upstream names it `<snapshot-bucket>-assets`). A gVisor-only install needs no asset bucket: gVisor's releases come from the public `gs://gvisor`.
- **S3-compatible** — the endpoint, bucket, region and path style, and a Secret in `ate-system` holding the access key (Kubernetes reads a Secret only from the pod's own namespace). SeaweedFS and MinIO need path-style addressing, the default.

Snapshots land under the bucket's root, or under `storage.snapshotPrefix` for a bucket shared with other data or installations. That location, the `storage_location` output, is the API server's: every agent's snapshots go there, and an actor template names another only with the `PerTemplateSnapshotLocation` preview gate.

## The generated secrets

Upstream's installer creates five certificate authorities, and a signing key for the JWTs the API server mints for agents:

| Pool | Namespace | Key | What it signs |
|------|-----------|-----|---------------|
| `service-dns-ca-pool` | `podcertificate-controller-system` | Ed25519 | every in-cluster Substrate service's serving certificate |
| `pod-identity-ca-pool` | `podcertificate-controller-system` | Ed25519 | every component's SPIFFE client certificate |
| `postgres-ca-pool` | `podcertificate-controller-system` | Ed25519 | the bundled database's client certificates |
| `actor-id-ca-pool` | `ate-system` | Ed25519 | agent identity certificates |
| `egress-mitm-ca-pool` | `ate-system` | ECDSA P-256 | the short-lived certificates that intercept agents' https traffic |
| `actor-id-jwt-pool` | `ate-system` | ES256 or RS256 | agent identity tokens |

The module generates each with the TLS provider, both engines alike, and writes them in upstream's exact pool format. Both engines' pools are parsed by upstream's own parsers in a test. They live in the deployment's state: a re-apply never replaces them, and a destroy does, like an upstream reinstall.

Two choices differ from upstream's installer, both on purpose:

- **Ten-year authorities, not one.** Every connection that chains to an expired root fails, so an upstream install stops working on day 366. Worse, the TLS provider would replace an expired root on the next apply, rotating trust out from under every certificate already issued. Ten years is the kubeadm convention for self-managed roots; the certificates under them last 24 hours (pod certificates) or 15 minutes (interception). Rotating an authority with overlap, which the pool format allows, is not offered by upstream yet.
- **The token key's ID.** Upstream names the key by hashing its public key's DER; Terraform hashes only text, so both engines name it by hashing the public key's PEM, in the same base64url form. Upstream accepts any ID, and the published key set carries whatever ID the pool names. Changing `actorJwt.algorithm` replaces the key, so tokens signed with the old one stop verifying.

Beside the token pool sits its public key set (`jwks` in the same Secret), which the API server's credential exchange verifies agent tokens with. Both engines write it byte for byte as upstream derives it from the pool, which the same test proves.

## Settings that take effect

The API server reads its authentication file once, the credential provider reads its namespace policy once, and every component reads the OpenTelemetry settings once. Upstream's installer also writes the authentication file and the policy only on first install, so later edits silently do nothing. This kind reconciles both from the spec, and each workload's pod template carries a digest of the configuration it reads, so a change rolls exactly the workloads that read it. The API server's log level lives in its configuration file (`ate-api-config`), which it reads as it changes, so a new level applies without a restart. Edits made by hand to these ConfigMaps are replaced on the next apply.

## The version

`version` picks the Agent Substrate release to install, from those Planton publishes, and defaults to the release this catalog composes (v0.6.0).

The modules compose one **install shape**: the files, objects, flags and settings of one release, v0.6.0's. Planton publishes a later upstream release with the shape it shares: one whose install differs from v0.6.0's only in its images keeps v0.6.0's shape, which Planton proves before it publishes (the manifests, the definitions and the database script compared file for file, and upstream's installer code compared at both tags). So an upstream patch that changes only its images reaches you by setting `version`, with no catalog update. A release of another shape is refused at plan time, before anything is applied:

- **Observed:** Agent Substrate v0.7.0 has the install shape of v0.7.0, and this KubernetesAgentSubstrate composes v0.6.0's.
- **Meaning:** v0.7.0 installs differently from what this Planton catalog release knows how to compose, so nothing was applied.
- **Next step:** set spec.version to v0.6.0 or a release of its shape, or deploy this resource from a Planton catalog release that composes v0.7.0's shape.

A release Planton has not published is refused by name, and every file the modules read is checked against the release's checksums. The resource definitions belong to the shape, so they stay the same across its releases, and moving between them never reads as a definition downgrade.

**Adopting a new patch:**
1. Set `version` and apply: the control plane moves in place, either way between releases of a minor.
2. Redeploy your worker pools: a pool takes its worker image from this resource's outputs when it deploys, so its workers run the new release's image only after its next deploy. Upstream's order for upgrades (workers before the API server) is a rule from 1.0 on.
3. If you pull images from your own mirror (`imageRegistry`), copy the new release's images there first. Planton's own mirror (`us-central1-docker.pkg.dev`) carries every release it publishes.

## Upgrades

Upstream's policy: before 1.0 every minor release is a reinstall, because the API server migrates its database schema irreversibly on its first start; from 1.0, one minor at a time, and a rollback of at most one. Patch releases are always free. Upstream writes no version marker and checks nothing, so the module stamps two:

- the installed release, on the API server Deployment;
- the minor that wrote the bundled database, on its volume claim template, so the volume carries it past a destroy.

An apply that breaks the policy stops at plan time, before anything changes:

- **Observed:** Deployment ate-system/ate-api-server records Agent Substrate v0.5.0 on this cluster, and this KubernetesAgentSubstrate installs v0.6.0.
- **Meaning:** Agent Substrate does not move from v0.5.0 to v0.6.0 in place: before 1.0 every minor release is a reinstall.
- **Next step:** back up what you need, destroy this resource, clear the database it used (`kubectl -n ate-system delete pvc data-postgres-0` for the bundled one, or drop the schema of an external one) and the node agent's state on every node that ran v0.5.0 (`/var/lib/ate`, or replace those nodes), and apply again; or set `version` back to the installed release, or deploy from the Planton release that installed it when that release has another install shape.

When this catalog later composes a new shape, an installation of the old one cannot be re-applied from it until it is reinstalled: the old shape's releases are refused, and the new shape's are another minor.

**Nodes keep state of their own.** Each node agent keeps its image cache, its sandbox runtime files and its actors' working state under `/var/lib/ate` on the node's disk, and that outlives a destroy. A new minor can change its layout: v0.6.0's node agent refuses v0.5.0's image cache and stays down until it is gone. So a reinstall across minors also clears `/var/lib/ate` on every node the old release ran on, once no actor runs there (a privileged pod mounting the host path, or `kubectl debug node/<name> --profile=sysadmin`), or replaces those nodes. A patch keeps the layout, so a patch never needs it.

The volume stamp catches the quieter case: a destroy that left the bundled database's volume behind, followed by an apply of a newer minor. An external or Cloud SQL database's schema version lives inside the database, out of the module's sight; clear it yourself when moving across a minor.

## Destroying it

Destroy runs in reverse: the default SandboxConfig first, while the controller that clears its finalizer still runs, then the control plane, then the namespaces. The definitions are kept unless `crds.keepOnUninstall` is false, so worker pools survive, stopped until a control plane returns. Destroy worker pools before the control plane: a worker pool that still runs `gvisor-default` holds that SandboxConfig in deletion until the pool goes. A KubernetesAgentSubstrateWorkerPool that takes its worker image from this resource's outputs is ordered after it, so a combined destroy removes the pool first and never waits.

## Pulling from a mirror

`imageRegistry` replaces the registry part of every image the install runs, keeping each path, tag and digest, so a mirror must copy images byte for byte. A Docker Hub name without a registry (`postgres:18-alpine`) is read the way container runtimes read it, `docker.io/library/postgres:18-alpine`. `us-central1-docker.pkg.dev` pulls Planton's own mirror of the Substrate images. The worker-image outputs follow the mirror. Upstream's Cloud SQL proxy is its one image pinned by tag alone.

## Parity accounting

Built for 100% parity with every setting upstream's installer and its components offer at v0.6.0. A setting left unset keeps upstream's value: the module writes a flag only when it differs from upstream's default, except the three upstream's installer always writes (the API server's bootstrap owners, the identity server's issuer, the Envoy router's allowed clients).

### Installer settings

| Upstream setting | Spec field |
|------------------|------------|
| `namespace` | `namespace` (fixed: upstream's deploy refuses any other) |
| `kindCluster.enabled` | `platform.generic` (the kind overlay's GKE-only removals) |
| `gcp.projectID`, `gcp.clusterLocation`, `gcp.clusterName` | `platform.gke.clusterId` (the issuer derivation; the installer's `gcloud` credential fetch is the provider connection's job) |
| `ateapi.expectedJWTIssuer` | `platform.generic.serviceAccountIssuer`, or derived on GKE |
| `actorJWT.algorithm`, `actorJWT.issuer` | `apiServer.actorJwt.algorithm`, `.issuer` (the identity server serves it) |
| `ateapi.authz.bootstrapOwners`, `ateapi.logLevel` | `apiServer.authorization.bootstrapOwners`, `apiServer.logLevel` (the configuration file's) |
| `ateapi.postgres.readWrite.connectionString`, `owner.connectionString` | `database.external` / `database.cloudSql` `readWriteConnectionString`, `ownerConnectionString` (Secret references) |
| `ateapi.postgres.readWrite.role`, `owner.role`, `schema` | `database.external.roles` / `database.cloudSql.roles` |
| `ateapi.postgres.poolMaxConns` | `database.poolMaxConnections` |
| `ateapi.postgres.serverCAFile` | `database.external.serverCaCertificate` |
| `ateapi.postgres.storageClass` | `database.bundled.storageClass` |
| `ateapi.postgres.cloudsql.instance`, `.iamAuth`, `.ipType`, `.gsa` | `database.cloudSql.instanceConnectionName`, `.iamAuthentication`, `.ipType`, and `platform.gke.apiServerServiceAccount` |
| `storage.bucketName`, `storage.snapshotLocation` | `storage.gcs.bucket` / `storage.s3.bucket` and `storage.snapshotPrefix` (the `storage_location` output) |
| `storage.transferGSAName` | `platform.gke.snapshotTransferServiceAccount` |
| `ingressAllowedSPIFFEIDs` | `router.allowedClientSpiffeIds` |
| `preview` | `previewGates` (written to every component, as the installer writes it) |
| `atenet.dataplane` | `router.dataplane` |
| `atenet.egress.credentialProvider` | `egress.credentialProvider` |
| `atenet.egress.additionalExtprocService` | `egress.additionalExtProcService` |
| `clusterSize` | `clusterSize` |
| `cordonControlPlane` | `cordonControlPlane` |
| `podcert.workersPerSigner` | `podCertificates.workersPerSigner` |
| `e2e.envoyConcurrency` | `egress.envoyConcurrency` |
| `otlpEndpoint` | `telemetry.otlpEndpoint` |
| `images.repo`, `images.tag` | `version` (a published release, its digests pinned) and `imageRegistry` |
| `rolloutTimeout` | the engines' own rollout waits |

### Component settings

| Component | Upstream flag or variable | Spec field |
|-----------|---------------------------|------------|
| API server | `--authz-bootstrap-owners`, the configuration file's log level | `apiServer.authorization.bootstrapOwners`, `apiServer.logLevel` |
| API server | `--drain-delay`, `--drain-timeout`, `--template-resync-interval` | `apiServer.drainDelay`, `.drainTimeout`, `.templateResyncInterval` |
| API server | replicas, resources | `apiServer.replicas`, `.resources` |
| Identity server | `--issuer`, `--log-level`, `--drain-grace`, replicas, resources | `apiServer.actorJwt.issuer`, `identityServer.logLevel`, `.drainGrace`, `.replicas`, `.resources` |
| Controller | `--log-level`, `--worker-sync-concurrency` | `controller.logLevel`, `.workerSyncConcurrency` |
| Node agent | `--log-level`, `--actor-stats-poll-interval`, `--drain-delay`, `--drain-timeout` | `nodeAgent.logLevel`, `.actorStatsPollInterval`, `.drainDelay`, `.drainTimeout` |
| Node agent | `--image-cache-gc-period`, `-high-percent`, `-low-percent`, `-max-bytes`, `-min-age`, `-gc-dry-run` | `nodeAgent.imageCache` |
| Node agent | `--snapshot-upload-delay`, `-detach-delay`, `-upload-timeout`, `-upload-est-throughput-mib`, `-upload-workers`, `-detach-workers`, `-report-workers`, `-report-give-up-after` | `nodeAgent.snapshots` |
| Node agent | `--sandbox-asset-cache-min-age`, `--sandbox-asset-fetch-timeout`, `--sandbox-asset-cache-gc-period`, `-max-bytes`, `-gc-dry-run` | `nodeAgent.sandboxAssetCache` |
| Node agent | `--snapshot-cache-dir` (empty turns it off), `--snapshot-cache-min-age`, `-fetch-timeout`, `-gc-period`, `-max-bytes`, `-gc-dry-run` | `nodeAgent.snapshotCache` |
| Node agent | objectstore `--asset-staging-capacity` (and the staging volume), `--upload-memory-budget` | `nodeAgent.assetStagingCapacityGib`, `nodeAgent.snapshots.uploadMemoryBudget` |
| Node agent | node selector, tolerations, resources | `nodeAgent.nodeSelector`, `.tolerations`, `.resources` |
| Router | `--log-level`, `--ingress-allowed-spiffe-ids`, `--route-timeout` | `router.logLevel`, `.allowedClientSpiffeIds`, `.routeTimeout` |
| Router | `--parked-request-budget`, `-max`, `-retry-interval`, `-retry-factor`, `-retry-jitter` | `router.parkedRequests` |
| Router | `--extproc-max-requests`, `--drain-delay`, `--drain-timeout` | `router.extProcMaxRequests`, `.drainDelay`, `.drainTimeout` |
| Router | replicas, resources | `router.replicas`, `.resources` |
| Egress | `--log-level`, `--egress-policy-cache-ttl`, `--drain-delay` | `egress.logLevel`, `.policyCacheTtl`, `.drainDelay` |
| Egress | sdsmint `--leaf-cert-ttl`, `--log-level` | `egress.interceptCertificateTtl`, `egress.logLevel` |
| Egress | replicas, resources | `egress.replicas`, `.resources` |
| Credential provider | `--log-level`, `--kube-api-qps`, `--kube-api-burst`, `--drain-grace`, the namespace policy | `egress.credentialProvider.kubernetesSecrets` `logLevel`, `kubeApiQps`, `kubeApiBurst`, `drainGrace`, `grants` |
| Pod certificates | `--kube-api-qps`, `--kube-api-burst`, resources | `podCertificates.kubeApiQps`, `.kubeApiBurst`, `.resources` |
| Bundled database | volume size, resources | `database.bundled.volumeSize`, `.resources` |
| All | `OTEL_TRACES_EXPORTER`, `OTEL_METRICS_EXPORTER`, `OTEL_LOGS_EXPORTER` | `telemetry.tracesExporters`, `.metricsExporters`, `.logsExporters` |
| All | `OTEL_TRACES_SAMPLER`, `OTEL_TRACES_SAMPLER_ARG`, `OTEL_METRIC_EXPORT_INTERVAL`, `OTEL_METRIC_EXPORT_TIMEOUT` | `telemetry.tracesSampler`, `.tracesSamplerRatio`, `.metricExportIntervalMs`, `.metricExportTimeoutMs` |
| Workers | `ATE_WORKER_OTEL_TRACES_EXPORTER`, `_METRICS_EXPORTER`, `_LOGS_EXPORTER` (the controller hands them to worker pods) | `telemetry.workerTracesExporters`, `.workerMetricsExporters`, `.workerLogsExporters` |
| All but Envoy | `--preview` | `previewGates` |
| All | image pull secrets | `imagePullSecrets` |

With agentgateway, upstream's component replaces the router and egress binaries with agentgateway's static configuration, so their flags above no longer apply; resources still do.

### What the module adds

- The generated secrets and their ten-year validity (above).
- The database setup in the bundled database's own start hook (above).
- The rollout digests on each workload's pod template (above).
- The version stamps and the upgrade refusal (above).
- The release's seal and install shape, checked before anything is applied (above).
- On a generic cluster with no OTLP endpoint, traces and logs off and metrics left on each component's Prometheus endpoint, so nothing dials a collector that is not there.

### Not modeled, and why

- **Wiring flags** — the components' socket paths, mounted file paths, listener ports, the API address they dial, the configuration file's path, the node agent's cache directories, and the SPIFFE IDs and service accounts they expect of each other (the API server's sidecar identity, the node agent's allowed caller, the controller's view of the node agent's and router's accounts). Each is matched by a Service, a NetworkPolicy, a volume or another component's flag in upstream's manifests, so changing one alone breaks the install.
- **`--dns-egress-policy`** — upstream's controller accepts only the value its manifest sets.
- **`--localhost-registry-replacement`** — the kind overlay's pointer at a local development registry.
- **`csi.setup`** — installs third-party test CSI drivers from upstream's `hack/` tree. Register real drivers with a [KubernetesAgentSubstrateCsiDriverConfig](../kubernetesagentsubstratecsidriverconfig/README.md).
- **`fakeDataPlane.enabled`, `benchmark.*`, `demo.*`** — upstream's benchmark and demo machinery.
- **`kubeconfig`, `context`, `ko.*`, `docker.buildFlags`, `record.dir`, `kindCluster.name`** — the installer's own machine settings; the provider connection and Planton's image line take their place.
- **Worker settings** (`--max-actors` and the like) — a worker pool's, not the control plane's ([KubernetesAgentSubstrateWorkerPool](../kubernetesagentsubstrateworkerpool/README.md)).

## Pairs well with

- **[KubernetesAgentSubstrateWorkerPool](../kubernetesagentsubstrateworkerpool/README.md)** — the workers agents run on, from this release's `gvisor_worker_image` or `microvm_worker_image` output.
- **[KubernetesAgentSubstrateSandboxConfig](../kubernetesagentsubstratesandboxconfig/README.md)** — a runtime of your own beside `gvisor-default`: a pinned gVisor release, a rollout with a way back, or microVM.
- **[KubernetesAgentSubstrateCsiDriverConfig](../kubernetesagentsubstratecsidriverconfig/README.md)** — your CSI driver registered for agents' external volumes.
- **KubernetesNamespace** — own `ate-system`, with the S3 Secret or the database Secret placed first.
- **KubernetesSecret** — the S3 access key, or an external database's connection strings.
- **KubernetesSeaweedFs** — an in-cluster S3 store.
- **GcpGkeCluster**, **GcpServiceAccount**, **GcpCloudSql**, **GcpGcsBucket** — the keyless Google Cloud setup.
