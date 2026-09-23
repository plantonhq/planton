# KubernetesValkey Guide

The judgment this guide carries: when a user asks for Redis, this is the
catalog's answer — say so explicitly — and the two decisions that matter
are authentication (off until declared) and what durability you actually
need (replication here is not failover).

Substitutes for: Redis (Redis-compatible in-memory store; every Redis
client library connects unchanged — the open-source successor).

## Answering "give me Redis"

Propose this kind, and tell the user what they got and why: Valkey is the
catalog's Redis-compatible store, their existing client libraries and
code work unchanged. Never silently substitute, and never fall back to a
Helm-release Redis chart — the substitution workflow is the
[catalog guide](../../_docs/GUIDE.md)'s first law.

## Declare auth whenever the namespace is shared

The chart ships with authentication OFF — anyone who can reach the
Service can read and write (the mechanics are on
[reference.md](v1alpha1/reference.md)). A cache in a namespace shared with other
workloads should declare ACL users from the first manifest; retrofitting
auth later means coordinating every client's rollout.

## Declare the user, never the password

An ACL user needs only a name. The module generates a password for
every user declared without one, keeps it stable across re-applies,
and materializes every user's credential in the `<name>-auth` Secret
under the username key — so a manifest carries no secret and no
placeholder, and nobody escrows a value the module can mint. Declare a
password only to bring your own, and then as a managed-secret
reference, never a literal. Clients never read the manifest: the
`password_secret` output names the Secret and key for the `default`
user, and a workload mounts it as an env var (the same reference every
consumer kind — Airflow, Superset, Ray — already takes).

## Pick the topology by durability need, not by habit

Standalone suits caches — data that can vanish. The replication topology
adds read scaling and a warm copy, but no automated promotion: durability
through a primary restart comes from persistence, not from replicas (the
full story is on [reference.md](v1alpha1/reference.md)). If losing the store's
contents is unacceptable, persistence is the lever to reach for first.

## Namespace ownership

A cache usually shares its namespace with the application it serves —
the multi-tenant case where `createNamespace: true` is wrong. Wire
`spec.namespace` to the application's KubernetesNamespace —
[namespace-ownership pattern](../../_patterns/namespace-ownership.md).

## On the diagram

The store renders beside the workloads it serves, with their namespace
edges shared — a reviewer sees cache and consumers as one unit. Declared
ACL users live inside this spec, so client identity does not add nodes.

## Pairs well with

- KubernetesNamespace — the shared namespace's owner (pattern above).
- The application workloads that consume it, connecting through the
  exported `kube_endpoint`.
