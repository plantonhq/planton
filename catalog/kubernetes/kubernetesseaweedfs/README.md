# Kubernetes SeaweedFS

## When NOT to Use This

**One resource is ONE SeaweedFS store.** The chart installs the whole
topology — masters, volume servers, filer, S3 gateway — as a single
release, and each tier scales through its own `replicas` field on this
one resource. You grow a store by sizing its tiers, never by
installing more resources.

Also not the right kind when:

- **You want a managed object store** — use a cloud bucket service;
  this kind is for serving the S3 API from ON the Kubernetes
  cluster itself.
- **You expect a public S3 endpoint out of the box** — everything
  stays ClusterIP by design. Exposure composes from first-class kinds
  (KubernetesIngress, Gateway API kinds) over the exported service
  handles.
- **You expect IaC to delete buckets** — buckets declared in
  `s3.buckets` are created by the chart's install hook, but removing
  an entry later does NOT delete the bucket or its data. Bucket
  deletion is a data operation, never an IaC one.
- **You need multiple filers on the default metadata store** — the
  embedded leveldb store is per-pod. Wire a shared external store
  (Postgres/MySQL via `helm_values`) before raising filer replicas.

## Overview

**KubernetesSeaweedFs** deploys SeaweedFS — the catalog's in-cluster
S3-compatible object store (Apache-2.0) — from the official
`seaweedfs` Helm chart (https://seaweedfs.github.io/seaweedfs/helm).
SeaweedFS separates metadata from data: `master` servers coordinate
the cluster and assign file ids, `volume` servers store the actual
object bytes, and the `filer` provides the file/bucket namespace plus
the S3 API. Each tier scales independently; the defaults (1/1/1 on
PersistentVolumeClaims — master 5Gi, volume 30Gi, filer 10Gi) give a
working single-node store, and 3 masters + N volume servers is the HA
shape.

**The credential contract**: the S3 gateway is ON by default with auth
ON — the chart materializes an admin and a read-only credential pair
in the `<name>-s3-secret` Secret (keys `admin_access_key_id` /
`admin_secret_access_key` / `read_access_key_id` /
`read_secret_access_key`; generated once, stable across upgrades, kept
on uninstall). The outputs point at it — credentials ride the
Secret, never the manifest. To own every identity yourself, reference
a Secret carrying the chart's `seaweedfs_s3_config` contract via
`s3.existing_config_secret`; the chart then generates nothing.

**Key design points:**

- **S3 is the point.** `s3.enabled` and `s3.enable_auth` default to
  TRUE — unset renders as enabled. The gateway runs embedded on the
  filer pods unless `s3.dedicated` is declared (then a separate
  Deployment scales the API independently of metadata); both shapes
  expose the same `<name>-s3` Service on port 8333.
- **PVCs, not hostPath.** The chart's out-of-the-box storage is
  hostPath (bare-metal grain); this component deliberately maps every
  data volume to a PersistentVolumeClaim and every logs volume to
  emptyDir — portable across every managed cloud and kind cluster.
  Declare sizes and (optionally) a StorageClass per tier.
- **Replication is one code.** `replication` is the SeaweedFS XYZ
  placement code (`"001"` = one extra copy on another server,
  requiring at least 2 volume servers; `"010"` needs rack topology).
  Setting it flips the chart's `enableReplication` and overrides
  master and filer placement together — the copies must fit the
  declared volume topology or writes fail.
- **The admin console is never open.** `admin.enabled` installs the
  SeaweedFS management console; unless `existing_auth_secret` is
  given, the modules generate the `<name>-admin-auth` Secret (keys
  `user`/`password`) — no unauthenticated install path exists.
- **ClusterIP by design.** In-cluster clients use the exported
  endpoints; external exposure composes from KubernetesIngress or
  Gateway API kinds.
- **`helm_values` is the escape hatch** — additional chart values
  merged LAST over everything the typed fields render (Helm `-f`
  semantics, identical on both engines); for the chart surface beyond
  the typed fields (per-tier scheduling, SFTP, the maintenance
  worker, mTLS via `global.seaweedfs.enableSecurity` with
  cert-manager, external filer stores, COSI, the all-in-one dev
  mode), never a substitute for them — and never for secrets.

## Essential Configuration Fields

### Required

- **`spec.namespace`**: namespace to install into — literal or a
  KubernetesNamespace reference (`create_namespace` to own it)

### Common

- **`spec.chart_version`**: chart pin (default `4.40.0` — chart
  versions track SeaweedFS releases; 4.40.0 ships appVersion 4.40)
- **`spec.master`**: replicas (1 dev, 3 HA — masters form a Raft
  quorum, use an odd count), `data_volume` (default 5Gi),
  `volume_size_limit_mb` (default 1000), resources
- **`spec.volume`**: replicas, `data_volume` per pod (default 30Gi —
  the only tier that grows with stored bytes), `max_volumes` (0 =
  auto-size from free disk), `index_mode` (memory / leveldb /
  leveldbMedium / leveldbLarge), `min_free_space_percent`, resources
- **`spec.filer`**: replicas (keep 1 on the embedded leveldb store),
  `data_volume` (default 10Gi), `encrypt_volume_data`,
  `extra_environment_vars` (`WEED_*` keys — upstream's configuration
  surface), resources
- **`spec.s3`**: `enabled`/`enable_auth` (both default true),
  `buckets` (name, `anonymous_read`, `ttl` like `"7d"`,
  `object_lock` — irreversible, forces versioning — and
  `versioning`), `existing_config_secret`, `domain_name`
  (virtual-hosted-style suffix; empty = path-style), `dedicated`
  (own Deployment: replicas + resources)
- **`spec.replication`**: the XYZ placement code; empty = no
  replication (`"000"`)
- **`spec.admin`**: `enabled`, `existing_auth_secret`, `data_volume`
  (empty = in-memory console state), resources
- **`spec.service_monitor_enabled`**: ServiceMonitors on every
  enabled tier (requires the Prometheus Operator CRDs)
- **`spec.image` / `spec.helm_values`**: the air-gap image path and
  the escape hatch

## Outputs

| Output | Purpose |
|---|---|
| `namespace` | Namespace the store runs in |
| `release_name` | Helm release name (= `metadata.name`) |
| `s3_endpoint` | In-cluster S3 endpoint SDKs point at (`http://<name>-s3.<ns>.svc.cluster.local:8333`); empty when the gateway is disabled |
| `s3_credentials_secret_name` | Secret holding the S3 credentials (the chart-generated `<name>-s3-secret`, or the referenced existing config secret); empty when auth is disabled |
| `filer_service_name` | The filer Service (file namespace HTTP API, port 8888) |
| `master_service_name` | The master Service (cluster coordination, port 9333) |
| `admin_endpoint` | In-cluster admin console endpoint (port 23646); empty when the console is disabled |
| `admin_auth_secret_name` | Secret holding the console credentials (keys `user`/`password`); empty when the console is disabled |
| `port_forward_command` | Port-forward command for S3 access from a workstation |

## Composing in Infra Charts

- **`spec.namespace`** is a foreign key (default kind
  KubernetesNamespace, field path `spec.name`); every
  **`data_volume.storage_class`** references a KubernetesStorageClass.
- **Applications consume the outputs**: `s3_endpoint` as the SDK
  endpoint (path-style addressing — the in-cluster norm; declare
  `s3.domain_name` for virtual-hosted-style),
  `s3_credentials_secret_name` for the access keys — credentials ride
  the Secret, never the manifest.
- **Exposure composes, never embeds**: a KubernetesIngress or Gateway
  API route over the `<name>-s3` Service for external S3 clients, or
  over `admin_endpoint`'s Service for the console.
- **Buckets are declared where the store is declared**: each entry in
  `s3.buckets` is created at install/upgrade by the chart's hook —
  applications receive a bucket name and credentials, not the power
  to create buckets.

## Examples

### Development (defaults, one bucket)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesSeaweedFs
metadata:
  name: dev-store
spec:
  namespace:
    value: dev-store
  create_namespace: true
  s3:
    buckets:
      - name: dev-data
```

### Production HA (3 masters, replicated volumes, dedicated gateway)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesSeaweedFs
metadata:
  name: object-store
spec:
  namespace:
    value: object-store
  create_namespace: true
  master:
    replicas: 3
  volume:
    replicas: 3
    data_volume:
      size: 200Gi
      storage_class:
        value: <fast-ssd-storage-class>
  replication: "001"
  s3:
    dedicated:
      replicas: 2
  admin:
    enabled: true
```

### Artifact store (typed buckets)

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesSeaweedFs
metadata:
  name: artifacts
spec:
  namespace:
    value: artifacts
  create_namespace: true
  volume:
    data_volume:
      size: 100Gi
  s3:
    buckets:
      - name: ci-artifacts
        ttl: 30d
      - name: release-artifacts
        versioning: true
      - name: public-assets
        anonymous_read: true
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
