# Kubernetes Network Policy

## Overview

**KubernetesNetworkPolicy** is a Planton component that creates and manages Kubernetes NetworkPolicies — the in-cluster firewall — as first-class, declaratively managed resources. A NetworkPolicy selects a set of pods with `pod_selector` and declares which network traffic is allowed to (`ingress_rules`) and from (`egress_rules`) those pods.

The component covers the complete `networking/v1` NetworkPolicySpec surface: pod and namespace label selectors (exact-match and set-based expressions), IP blocks with exceptions, TCP/UDP/SCTP ports, named ports, and port ranges. There is nothing an upstream NetworkPolicy can express that this spec cannot.

## Purpose

By default, every pod in a Kubernetes cluster can talk to every other pod. NetworkPolicies are the standard mechanism for changing that: they let you segment the cluster network so a compromised or misbehaving pod cannot reach things it has no business reaching — the databases of other teams, the metadata endpoint, the internet.

**Key value over raw manifests:**

- **Schema-level validation**: Direction/rule consistency (rules in an ungoverned direction are rejected instead of silently ignored), CIDR format checks, port-range sanity, selector operator contracts, and empty-peer rejection — all caught before anything reaches the cluster
- **Namespace by value or reference**: `spec.namespace` accepts a literal name or a reference to a `KubernetesNamespace` resource, so an infra chart can create the namespace and its policies in one run
- **Deterministic policy types**: Both IaC modules always submit the governed directions explicitly, applying the Kubernetes inference rule when the spec omits `policy_types`, so the deployed object never depends on which engine applied it
- **Dual IaC support**: Both Pulumi and Terraform implementations with feature parity
- **Lifecycle management**: Integrated with Planton's deployment lifecycle for status tracking and outputs

## The Allow-Only Model

The single most important thing to understand about NetworkPolicies: **policies are additive allows — there is no deny rule.**

- A pod that is not selected by any policy accepts and sends all traffic (the Kubernetes default).
- Once a pod is selected by any policy in a direction, only traffic allowed by *some* policy in that direction gets through. Everything else is denied.
- Multiple policies selecting the same pods combine by union. You can never write a policy that *removes* an allow granted by another policy.

Isolation therefore comes from **selecting pods while allowing little**. The two canonical shapes fall out of this model:

- **Default-deny**: a `pod_selector` with `match_all: true` (or no `pod_selector` at all — both select ALL pods in the namespace) with `policy_types: [ingress, egress]` and no rules — everything is denied, and every subsequent policy is a targeted exception.
- **Targeted allow**: select specific pods and enumerate the peers and ports allowed.

## Selectors and Peers

### `pod_selector` — which pods the policy governs

A selector with `match_all: true` (or no `pod_selector` at all) selects **all pods in the namespace** — the default-deny building block; a selector that names neither `match_all` nor a label criterion is refused, so selecting everything is always written down. To target one Planton workload, match on its `app` label: every Planton workload kind stamps `app: <workload-metadata-name>` on its pods as immutable selection identity, so `match_labels: {app: backend-api}` targets exactly that workload's pods.

### Peers — who traffic is allowed to/from

Each peer takes one of three forms, with precise combination semantics:

- **`pod_selector` alone**: pods matching the selector in the policy's OWN namespace
- **`namespace_selector` alone**: ALL pods in namespaces matching the selector (e.g. the automatic `kubernetes.io/metadata.name: <name>` label every namespace carries)
- **Both together in ONE peer**: pods matching `pod_selector` in namespaces matching `namespace_selector` — a single AND. This is different from listing them as two separate peers, which ORs: *any* pod in the selected namespaces, plus matching pods in the policy's own namespace
- **`ip_block`**: a CIDR (with optional `except` carve-outs) for traffic outside the cluster's pod network. Mutually exclusive with the selector forms — cluster-internal pod IPs are ephemeral and should be matched with selectors, never CIDRs

### Rules — the OR/AND grid

Each rule in `ingress_rules`/`egress_rules` is an independent OR: traffic matches if it satisfies ANY rule. Within a rule, peers OR with each other, ports OR with each other, and peers AND ports (the traffic must match a peer AND a port). An empty `from`/`to` allows all sources/destinations; an empty `ports` allows all ports.

## `policy_types` and Inference

`policy_types` declares which directions the policy governs. When omitted, Kubernetes infers: **ingress is always included, and egress is included only when egress rules are present.** Both IaC modules apply this same rule and always send the result explicitly.

Set `policy_types` explicitly whenever the intent is isolation:

- An **egress-only** policy must say `[egress]` — otherwise inference adds ingress, and an ingress-selected pod with no ingress rules is deny-all-ingress
- A **deny-all-egress** policy must say `[egress]` with no egress rules — there is no rule to infer the direction from

The schema rejects rules in an ungoverned direction (e.g. ingress rules with `policy_types: [egress]`), because Kubernetes would silently ignore them.

## Enforcement Requires a Capable CNI

NetworkPolicy objects are only **enforced** by a CNI plugin that implements them — Calico, Cilium, and the cloud CNIs with policy enforcement enabled (GKE Dataplane V2, EKS with a policy-capable CNI configuration, AKS with Azure or Calico network policy). On a cluster whose CNI ignores them — **including default kind/kindnet clusters** — the NetworkPolicy object exists and is readable, but all traffic still flows. Verify enforcement in each cluster before relying on policies for security.

## Essential Configuration Fields

### Required

- **`spec.name`**: The NetworkPolicy name (DNS subdomain: lowercase alphanumeric, hyphens, dots, 1–253 chars)

### Common

- **`spec.namespace`**: Literal namespace name or reference to a KubernetesNamespace resource. A NetworkPolicy governs only pods in its own namespace. When omitted, the policy lands in the cluster's `default` namespace
- **`spec.pod_selector`**: Which pods the policy applies to; omit it or set `match_all: true` to select all pods in the namespace
- **`spec.policy_types`**: The governed directions (`ingress`, `egress`); omit to use the Kubernetes inference rule
- **`spec.ingress_rules`** / **`spec.egress_rules`**: The allow rules
- **`spec.labels`** / **`spec.annotations`**: Merged with standard Planton labels for tracking and governance

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

- **`network_policy_name`**: The name of the NetworkPolicy object as created in the cluster
- **`namespace`**: The namespace the NetworkPolicy was created in
- **`policy_types`**: The directions the policy governs as deployed — `"Ingress"`, `"Egress"`, or `"Ingress,Egress"` — including inferred directions when the spec omitted `policy_types`

## How It Works

This component includes both **Pulumi** (Go) and **Terraform** (HCL) modules that:

1. Resolve the target namespace (literal value or resolved reference)
2. Merge user labels and annotations with standard Planton tracking labels
3. Resolve the governed directions — explicit `policy_types`, or the Kubernetes inference rule — and always submit them explicitly
4. Create the `networking/v1` NetworkPolicy with the pod selector, ingress rules, and egress rules mapped one-to-one (each spec peer maps to exactly one API peer, preserving the AND'd pod+namespace semantics)
5. Export the policy name, namespace, and governed directions for downstream composition

Both IaC implementations have feature parity and follow identical logic.

## When to Use

Use **KubernetesNetworkPolicy** when you need:

- Namespace-level default-deny as a security baseline
- Segmenting traffic between teams, tiers, or environments sharing a cluster
- Restricting a workload's inbound traffic to its known callers
- Locking down egress — allowing only DNS, specific services, or specific external CIDRs
- Meeting compliance requirements for network segmentation

**Do NOT use** when:

- Your cluster's CNI does not enforce NetworkPolicies — the objects will exist but do nothing; fix the CNI first
- You need L7 rules (HTTP paths, methods, gRPC services) — NetworkPolicy is L3/L4 only; use a service mesh or CNI-specific policy CRDs (CiliumNetworkPolicy, Calico policies)
- You need cluster-wide or cross-namespace policy definitions — NetworkPolicy is namespaced; use CNI-specific cluster-scoped policies

## Prerequisites

- **Kubernetes Cluster**: Access to a Kubernetes cluster with a **NetworkPolicy-enforcing CNI** (Calico, Cilium, or a cloud CNI with enforcement enabled)
- **Credentials**: Kubernetes cluster credentials (kubeconfig)
- **Namespace**: The target namespace must exist before creating the policy (unless deploying to `default`, or creating the namespace in the same chart via a reference)

## Best Practices

1. **Start with default-deny, then add targeted allows**: Apply the default-deny policy to a namespace first, then one small policy per workload or flow. Additive semantics make this safe — each new policy only widens
2. **Never forget DNS when denying egress**: A deny-all-egress policy blocks DNS too, and almost everything breaks in confusing ways. Pair it with an allow-DNS egress rule (UDP+TCP port 53 to the cluster DNS pods)
3. **Know your AND from your OR**: pod+namespace selectors in one peer restrict (AND); the same selectors as two peers widen (OR). This is the most common NetworkPolicy authoring mistake
4. **Select workloads by the `app` label**: `match_labels: {app: <workload-name>}` composes with every Planton workload kind's label contract
5. **Verify enforcement, then verify behavior**: Confirm the CNI enforces policies at all, then test actual connectivity (e.g. `kubectl exec ... -- nc -z <svc> <port>`) — a policy that "applies cleanly" proves nothing about traffic

## References

- [Kubernetes Network Policies Documentation](https://kubernetes.io/docs/concepts/services-networking/network-policies/)
- [Declare Network Policy Walkthrough](https://kubernetes.io/docs/tasks/administer-cluster/declare-network-policy/)
- [NetworkPolicy API Reference](https://kubernetes.io/docs/reference/kubernetes-api/policy-resources/network-policy-v1/)
- [Automatic namespace labelling (`kubernetes.io/metadata.name`)](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#automatic-labelling)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
