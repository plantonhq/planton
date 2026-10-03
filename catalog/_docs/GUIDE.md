# Catalog Guide

Authored wisdom about using this catalog as a whole. The generated companion
files ([reference-index.md](reference-index.md),
[reference-commons.md](reference-commons.md),
[reference-graph.yaml](reference-graph.yaml)) carry the facts; this page
carries the judgment.

## When the software you were asked for has no kind of its own

Do NOT jump to a Helm-release or raw-manifest workaround. Follow this order:

1. **Search the whole pack by the name the user said**, not just kind names:

```
rg -il "redis" <provider directories>
```

   Compatible alternatives document the names they substitute for in their
   own reference pages, so a full-text search finds them even when no kind
   carries the asked-for name.

2. **Propose the catalog's alternative and say what you did.** Add the
   compatible kind to the architecture, then tell the user explicitly:
   what was added instead of what they asked for, and why it serves the same
   purpose (client compatibility, protocol compatibility). Never silently
   substitute — the user asked for Redis and should hear "you got Valkey,
   and every Redis client works with it unchanged."

3. **Only if the catalog truly has nothing** — no direct kind, no compatible
   alternative — fall back to a generic mechanism (a Helm release), and say
   plainly that the catalog has no first-class kind for it yet.

## Verified alternatives

Compatible substitutes for well-known names, each verified against the
kind's own documentation:

| If asked for | Use | Compatibility |
|---|---|---|
| Redis | KubernetesValkey | Redis-compatible in-memory store; every Redis client library speaks to it natively (open-source successor) |
| Elasticsearch | KubernetesOpensearch | Apache-2.0 fork; drop-in replacement for the Elasticsearch APIs at the 7.10 fork line — existing clients and integrations connect unchanged |
| Kibana | KubernetesOpensearch | OpenSearch Dashboards — the Kibana-role console — is a section of the same spec (`dashboards`), not a separate kind |
| Vault | KubernetesOpenBao | Linux Foundation-governed secrets manager, MPL-2.0 fork of Vault; auto-unseal interoperates with the OpenBao/Vault transit engine family |
| Confluent Schema Registry | KubernetesKarapace | Apache-2.0, Confluent-API-compatible schema registry — existing Confluent SR clients work unchanged |
| MinIO / in-cluster S3 | KubernetesSeaweedFs | Not a MinIO fork — an S3-compatible object store whose S3 gateway is on by default; clients speaking the S3 API connect to it |
| PrestoSQL / Presto | KubernetesTrino | Trino is the renamed community successor of PrestoSQL (the Presto fork by Presto's creators); PrestoSQL/Presto clients, drivers, and connector vocabulary carry over |

Many well-known names need no substitution — Kafka, MongoDB, PostgreSQL,
ClickHouse, RabbitMQ, and others have kinds of their own; the per-provider
indexes are the authoritative list.

## Conventions that span the catalog

- The manifest grammar every kind shares (envelope, metadata,
  value/valueFrom, fieldPath spelling, search grammar):
  [reference-commons.md](reference-commons.md).
- Composition wisdom — how kinds wire together and the trade-offs:
  [patterns/](../_patterns/README.md).
- Per-kind judgment lives in `GUIDE.md` beside each kind's `reference.md`
  (the index tables' Guide column shows which kinds have one).
- Verified facts about money, posture, and access live beside each kind as
  machine-checked sidecars — `cost.yaml`, `controls.yaml`,
  `iac/permissions.yaml` — with the central control catalog and framework
  crosswalks in `../_compliance/` and the pinned price books, derivation
  rules, and generated estimates in `../_pricing/`. Read those files for
  any cost, control-posture, or permissions question; never restate their
  numbers or claims from memory.

## This page is contributed wisdom

Like every `GUIDE.md` and pattern, this file is authored, openly improvable
through pull requests, and checked by CI (kind names must resolve, embedded
manifests must validate). When you verify a new alternative, add its row
with the compatibility statement grounded in the kind's documentation.
