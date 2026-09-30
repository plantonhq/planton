# KubernetesExternalSecret Guide

The judgment this guide carries: an ExternalSecret keeps the external
system as the single source of truth — the cluster copy is a projection.
The decisions that matter are where it lives, which sync form to use,
and what happens to the projected Secret when the declaration goes away.

## It lives with its consumer

Declare the ExternalSecret in the namespace of the workload that consumes
the materialized Secret — that is where the Secret lands, and workloads
mount it like any other. The backend connection is someone else's node:
`storeRef.secretStore` points at a same-namespace KubernetesSecretStore,
`storeRef.clusterSecretStore` at a platform
[KubernetesClusterSecretStore](../kubernetesclustersecretstore/GUIDE.md).

## Explicit entries vs bulk pulls

`data` maps one backend entry (or one property of it) to one Secret key —
the precise, reviewable form; prefer it for application credentials.
`dataFrom` pulls whole structured entries or pattern-matched fleets —
right for a JSON document of related credentials, at the cost of the
Secret's contents no longer being enumerable from the manifest. When a
reviewer cannot tell what keys will exist, prefer `data`.

## The projection outlives the declaration — know that

The default lifecycle (`creationPolicy: Owner`,
`deletionPolicy: Retain` — full vocabulary on
[reference.md](v1alpha1/reference.md)) means deleting the ExternalSecret KEEPS the
materialized Secret, which then never refreshes again: consumers keep
working against silently aging credentials. Retain is the safe default
against accidental outage; the judgment is to pair teardown of the
declaration with teardown of the projection when a credential is being
retired, not just rotated.

## Namespace ownership

`spec.namespace` follows the shared-namespace rule: the consumer's
namespace is multi-tenant by definition, so wire it to the application's
KubernetesNamespace via `valueFrom` —
[namespace-ownership pattern](../../_patterns/namespace-ownership.md).

## On the diagram

Each ExternalSecret draws a labeled "reads from" edge into its store —
the architecture shows exactly which external backend feeds which
application, per secret. Credentials copied into manifests by hand show
nothing.

## Pairs well with

- KubernetesSecretStore / KubernetesClusterSecretStore — the backend
  connection it reads from.
- The consuming workload — mounts the materialized Secret via env or
  volumes.
- KubernetesExternalSecretsOperator — the machinery, once per cluster.
