---
kinds:
  - KubernetesKafka
  - KubernetesStrimziKafkaOperator
  - KubernetesPostgres
  - KubernetesCloudNativePgOperator
  - KubernetesClickHouse
  - KubernetesAltinityOperator
  - KubernetesAgentSandbox
  - KubernetesAgentSandboxTemplate
  - KubernetesAgentSandboxWarmPool
  - KubernetesAgentSandboxClaim
  - KubernetesAgentSandboxController
---

# Operator Prerequisite: the Controller a Custom-Resource Kind Needs

Many catalog kinds are not standalone workloads — they are declarations a
CLUSTER OPERATOR turns into running software. KubernetesPostgres renders a
CloudNativePG cluster; KubernetesKafka renders Strimzi resources;
KubernetesClickHouse, KubernetesMongodb, KubernetesRabbitMq, and a dozen
others follow the same shape. This pattern is the composition truth every
one of those pairs shares, stated once.

## The rule

A custom-resource (CR) kind requires its operator on the cluster, and the
operator must be WATCHING the CR's namespace. Miss either half and the
deploy fails silently: the CR object is created, nothing reconciles it,
and the resource sits NotReady with no error in its own manifest.

The operator is a REGISTRY PREREQUISITE of the CR kind — the platform
records the dependency in kind metadata, not as a `valueFrom` reference in
the manifest. Two consequences follow directly:

1. **The dependency draws NO edge on the architecture diagram.** Unlike a
   `spec.namespace` reference or any other `valueFrom` wiring, the
   CR-to-operator relationship is invisible in the rendered graph. When
   reviewing an architecture that includes a CR kind, verify the operator
   node exists as a deliberate check — the picture will not flag its
   absence for you.
2. **Facts about the prerequisite live in Layer 1, not the guide.** The
   CR kind's `reference.md` already states which operator it needs (from
   the registry). Guides never restate that fact; they carry the JUDGMENT
   below.

## Watch scope: the decision the operator's guide must state

Installing the operator is necessary but not sufficient — its watch scope
decides which namespaces' CRs actually get reconciled, and operators do
NOT agree on the default. Three postures exist in the catalog, and each
operator's guide names which one it has:

| Posture | Default behavior | The judgment |
|---|---|---|
| Own-namespace-only default | Watches just its install namespace | Widen it deliberately: a CR in any other namespace sits unreconciled. The silent trap. |
| All-namespaces default | Watches every namespace | Fence it when isolation matters (a watch-namespace list); otherwise one install serves the cluster. |
| Singleton / fixed-scope | One install per cluster by construction (cluster-scoped webhooks or a single config CR) | A second install cannot coexist; placement is the whole decision. |

## Placement and namespace ownership

One operator per cluster is the normal shape; it belongs in the
shared-cluster layer, and its own namespace is the sole-tenant case of
[namespace-ownership](namespace-ownership.md) (`createNamespace: true` is
correct there). Application environments declare the CR kinds; they never
declare their own operator.

## On the diagram

The operator renders as a shared-layer node; the CR renders in its
application environment. Because no edge connects them, an architecture
diagram that shows a database or a Kafka cluster with no operator node in
the shared layer is showing an incomplete composition — the reviewer, not
the renderer, is the one who catches it.

## See also

Each operator-backed kind's `GUIDE.md` names its watch posture and its
concrete failure mode; the CR kind's `reference.md` names the specific
operator it requires.
