# KubernetesTempo Guide

The judgment this guide carries: Tempo's scaling floor is a storage
decision — one replica runs honestly on a local volume, but the moment
you need two, object storage stops being optional. Decide by trace
volume up front, not mid-incident.

## The replica/storage floor

`local` storage (the default) keeps trace blocks on a PersistentVolume —
honest for a single replica. MORE than one replica REQUIRES an object
storage backend; on-cluster that means composing a
[KubernetesSeaweedFs](../kubernetesseaweedfs/v1alpha1/reference.md) and
pointing `storage.s3` at it (the validated shape is in this kind's own
example). Proposing `replicas: 2` with local storage is a manifest that
will not deploy. This kind deliberately models single-binary Tempo — by
the time per-component microservices are needed, that is its own design
conversation (the reference page says so).

## How traces arrive, how they are read

Applications send OTLP straight to the exported `otlp_grpc_endpoint` /
`otlp_http_endpoint`, or a
[KubernetesOtelCollector](../kubernetesotelcollector/GUIDE.md)
gateway sits in between (sampling, enrichment, fan-out). Grafana reads
traces back through a `tempo` datasource on the exported `http_endpoint`.
The full wired composition:
[observability-stack pattern](../../_patterns/observability-stack.md).

## Namespace ownership

Shares the observability namespace with its siblings — wire
`spec.namespace` through that KubernetesNamespace
([namespace-ownership pattern](../../_patterns/namespace-ownership.md)).

## Traces outside the cluster: R2

When the traces must outlive the cluster that wrote them (a rebuilt
cluster, a moved region, an incident review a month later), store them in
Cloudflare R2 through the `r2` arm. It names the bucket, its account and
its jurisdiction by reference to a `CloudflareR2Bucket`, and the key pair
by reference (`$secret/` for a token minted in the dashboard, or a
`CloudflareAccountApiToken`'s outputs). The modules compose the S3 host,
region `auto` and path-style addressing, and write the pair into their own
`<name>-r2-credentials` Secret; a rotated key rolls the pods on the next
apply. Three things to get right:

- **Scope the token to the one bucket** with Object Read & Write. Nothing
  else needs it.
- **The bucket's expiry runs later than `retention`**, or objects vanish
  under the index. Declare the bucket's lifecycle rule with the retention,
  never shorter.
- **Pick the location hint where the cluster runs.** R2 honours it only
  at creation; a bucket can't move later. `jurisdiction` is a different,
  legal setting that changes the bucket's host; leave it unset unless
  data residency requires it.

## Pairs well with

- KubernetesOtelCollector — the optional gateway in front of ingestion.
- KubernetesGrafana — the reader (`tempo` datasource).
- KubernetesSeaweedFs — the storage floor above one replica.
- CloudflareR2Bucket — trace storage outside every cluster (the `r2` arm).
