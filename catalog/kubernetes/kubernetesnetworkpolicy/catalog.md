# Kubernetes NetworkPolicy

Deploys a Kubernetes NetworkPolicy — the in-cluster firewall. The policy selects pods with a label selector and ALLOWS the traffic its rules describe; everything not allowed by some policy is denied once a pod is selected in that direction. Manages network isolation declaratively, with the full audit trail and versioning every Cloud Resource carries.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Kubernetes NetworkPolicy** -- a networking/v1 NetworkPolicy in the specified namespace carrying the pod selector, governed directions, and the ingress/egress allow rules (selectors, namespace selectors, and CIDR blocks with carve-outs)
- **Kubernetes Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking

## Before You Deploy

### Planton Setup

- **Kubernetes Provider Connection** -- an active connection in the Connect module with kubeconfig credentials for the target cluster. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline kubeconfig authentication.

### Kubernetes Cluster

- **A CNI that enforces NetworkPolicy** -- Calico, Cilium, or a cloud CNI with policy enforcement enabled. On a cluster whose CNI ignores policies (including default kind clusters), the object exists but ALL traffic still flows.
- The target namespace must already exist (the module does not create it).

## Deploy

### Console

Open the deployment store, find **Kubernetes NetworkPolicy**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Default Deny All** preset for namespace lockdown or the **Allow DNS Egress** preset for its essential companion in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesNetworkPolicy
metadata:
  name: default-deny-all
  org: acme-corp
  env: prod
spec:
  name: default-deny-all
  namespace:
    value: backend-services
  podSelector: { matchAll: true }
  policyTypes:
    - ingress
    - egress
```

```shell
planton apply -f networkpolicy.yaml
```

This denies ALL traffic to and from every pod in `backend-services` — the lockdown baseline that targeted allow policies then open back up. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the policy to a namespace managed by another Cloud Resource:

```yaml
spec:
  name: default-deny-all
  namespace:
    valueFrom:
      kind: KubernetesNamespace
      name: backend-namespace
      fieldPath: spec.name
  podSelector: { matchAll: true }
  policyTypes:
    - ingress
    - egress
```

The InfraPipeline deploys the namespace first, then applies the policy into it.

## Key Configuration

These are the most important decisions when configuring a Kubernetes NetworkPolicy. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Selection is the isolation** -- A pod selector with `matchAll: true` (or none at all) selects ALL pods in the namespace (the default-deny building block); labels narrow it to one workload. Every Planton workload stamps the `app` label (set to its name) on its pods, so `matchLabels: {app: checkout}` governs exactly that workload.

**Declare directions explicitly** -- When `policyTypes` is omitted, Kubernetes infers: ingress always, egress only when egress rules exist. Every isolation intent should set it explicitly — a deny-all-egress policy MUST say egress (there is no rule to infer it from), and an egress-only policy that omits it also isolates ingress.

**Rules are additive ORs** -- Traffic is allowed when it matches ANY rule's peers AND that rule's ports. Within a rule, empty peers mean ALL sources/destinations and empty ports mean ALL ports — an entirely empty rule allows everything in its direction.

**Peers combine precisely** -- A pod selector alone matches the policy's OWN namespace; a namespace selector alone matches ALL pods in matching namespaces; both together are ONE AND'd condition (different from two separate peers). IP blocks stand alone and are for traffic outside the cluster — pod IPs are ephemeral.

**Remember DNS** -- Deny-all-egress breaks name resolution first. Pair it with an allow-DNS rule: UDP + TCP port 53 to the cluster DNS pods (`kube-system` namespace, `k8s-app: kube-dns` pods).

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **KubernetesNamespace** | `namespace` | `spec.name` |

An omitted `namespace` means the cluster's `default` namespace.

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `network_policy_name` | The name of the created NetworkPolicy | Auditing and tooling |
| `namespace` | The namespace the policy governs | Verifying scope |
| `policy_types` | The governed directions | Auditing isolation posture |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Namespace lockdown** -- Default-deny both directions, then targeted allows per flow. Start from the **Default Deny All** preset.

**Trusted namespace interior** -- All pods accept traffic from the SAME namespace, nothing else. Start from the **Allow Same Namespace** preset.

**Cross-namespace ingress** -- Admit one namespace's traffic by its automatic `kubernetes.io/metadata.name` label. Start from the **Allow From Namespace** preset.

**The DNS companion** -- The allow-DNS egress rule every locked-down namespace needs. Start from the **Allow DNS Egress** preset.

## Works With

- [**Kubernetes Namespace**](/cloud-catalog/kubernetes-namespace) -- reference the namespace so infra charts create it and this policy in dependency order
- [**Kubernetes Deployment**](/cloud-catalog/kubernetes-deployment) -- its `app` label is the selection contract; its `selector_labels` output carries the full set
- [**Cilium**](/cloud-catalog/kubernetes-cilium) -- a CNI that enforces these policies (and extends them with its own richer policy language)
