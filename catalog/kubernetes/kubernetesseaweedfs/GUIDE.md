# KubernetesSeaweedFs Guide

The judgment this guide carries: this is the catalog's in-cluster answer
whenever another kind says "requires object storage" — Loki's
scalable mode, Tempo beyond one replica, Flink's stateful upgrades,
MLflow's multi-replica artifacts, Velero's backup target. When one of
those guides sends you here, this kind plus one credentials hop
completes the composition.

## The composition role: serve S3 to the rest of the architecture

The S3 gateway is ON by default with auth ON — the chart materializes
admin and read-only credential pairs in the `<name>-s3-secret` Secret
(stable across upgrades, kept on uninstall), and the outputs point
at it. Consumers wire the endpoint from the exported outputs and read
credentials from that Secret by reference — never copy values. Declare
the consumers' buckets in `s3.buckets` so they exist from first boot.

## Size the tiers by role, not uniformly

Masters coordinate, volume servers hold the blobs, filers serve the
namespace + S3 API — each tier scales independently (the topology is on
[reference.md](v1alpha1/reference.md)). The 1/1/1 default is a working single-node
store; 3 masters + N volumes + 2 filers is the HA shape. When the S3 API
itself is hot (many small requests), `s3.dedicated` splits the gateway
into its own Deployment so API capacity scales without touching metadata.

## Namespace ownership

An object store exists to be consumed — commonly from several
namespaces. Give it a dedicated namespace wired through a
KubernetesNamespace
([namespace-ownership pattern](../../_patterns/namespace-ownership.md));
its consumers reach it by service endpoint, not by sharing its namespace.

## On the diagram

SeaweedFS renders as the storage node its consumers point at — every
"requires object storage" composition (Loki, Tempo, Flink, MLflow,
Velero) becomes a visible endpoint-and-credentials relationship instead
of an unstated assumption.

## Pairs well with

- KubernetesLoki / KubernetesTempo / KubernetesFlinkDeployment /
  KubernetesMlflow — the guides that send their storage requirement
  here.
- KubernetesVelero — an in-cluster backup target (its guide covers when
  that is and isn't wise).
