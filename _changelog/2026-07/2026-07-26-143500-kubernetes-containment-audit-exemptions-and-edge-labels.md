# Kubernetes Containment Audit: Kafka Rooms, Access Exemptions, and Human Edge Labels

**Date**: July 26, 2026
**Type**: Enhancement
**Components**: Kubernetes Provider, API Definitions, Infra Catalog

## Summary

The rooms-vs-furniture containment audit — already applied to the other
twelve providers — now covers the Kubernetes provider. Two kinds become
containers (`KubernetesKafka` and `KubernetesKafkaConnect`: Strimzi topics,
users, and connectors deploy INTO their cluster via the strimzi.io/cluster
label, so engineers draw them inside the cluster's box), seventeen
access-style references gain `containment_exempt` so diagrams stop nesting
kinds inside resources they merely talk to, and the flagship Kubernetes
edges gain human-authored `diagram_label` wording ("issued by", "manages
records in", "mirrors from", "schema registry for").

## Problem Statement / Motivation

Without verdicts, diagram containment treats every reference into a
container kind as placement. That drew false pictures: external-dns nested
INSIDE the DNS zone it manages records in, OpenBao nested INSIDE the KMS
keyring that unseals it, an RBAC grant nested INSIDE its subject's home
namespace. Meanwhile the genuinely room-like Kubernetes shapes — a Kafka
cluster enclosing its topics and users — carried no container mark at all,
so those diagrams read flat.

## What Changed

### Container marks (rooms)

- `KubernetesKafka`: KafkaTopic and KafkaUser declarations belong to one
  cluster in Strimzi's own model — placement, not access.
- `KubernetesKafkaConnect`: KafkaConnector declarations deploy into their
  Connect cluster the same way.
- The doctrine rationale is recorded on both enum entries; every other
  Kubernetes kind was audited and stays furniture (gateways are
  flow-through, issuers are trust anchors, storage classes and service
  accounts are access targets, node pools and node classes are compute).

### Access exemptions (`containment_exempt`)

- external-dns: four per-provider `zone_id_filters` (watch scope) plus the
  Cloud DNS `project` — the controller manages records in zones; it never
  deploys into them.
- external-secrets: the GCP store's `project_id` — read access.
- OpenBao: the GCP KMS seal's `project` and `key_ring` — unseal-key source.
- RBAC: the ServiceAccount subject's `namespace` — locates the SUBJECT of a
  grant, not the grant itself.
- Kafka clients: Connect's bootstrap, MirrorMaker2's source and target
  bootstraps, the UI console's bootstrap/CA-trust/Connect-management
  references, and Karapace's bootstrap/CA-trust — all talk TO a cluster
  without living inside it.

### Edge labels (`diagram_label`)

Authored only on edges that actually render. Placement references
deliberately carry no label: nesting already expresses containment on the
diagram, and those edges are suppressed by design. Labels shipped:
"issued by" (certificate → issuer), "attached to" / "extends gateway"
(routes and listener sets → gateway), "machine template" (Karpenter pool →
node class), "reads from" (external secret → store), "grants to" (RBAC →
subject), "manages records in" (external-dns → zones), "unseal key ring" /
"unseal key" (OpenBao → KMS), "reads and writes", "mirrors from",
"mirrors to", "console for", "manages", "trusts", "schema registry for"
(the Kafka ecosystem's client edges).

## Validation

- The containment-decision registry regenerated from the compiled
  descriptors; the diff shows exactly the authored verdicts (new contained
  lines for topic/user/connector placement, exempt lines for every
  annotated access reference).
- Full `catalogkind` test package green, including the option-number
  pins and the annotation-placement guards.
- Stubs regenerated for every edited proto directory.
