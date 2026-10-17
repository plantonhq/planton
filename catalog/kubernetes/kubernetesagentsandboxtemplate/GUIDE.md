# KubernetesAgentSandboxTemplate Guide

The judgment this guide carries: the template is where a sandbox's safety is decided, once, for every sandbox a product will ever claim. Isolation, the network fence and what a claim may change all live here, and each one trades against warm starts or reachability. Make those trades deliberately.

## One template per kind of sandbox, owned by the platform

A template is a blueprint and a fence; it runs nothing. Products never claim from a template directly. They claim from a warm pool that references it (`KubernetesAgentSandboxWarmPool`, `KubernetesAgentSandboxClaim`), so the template is the node a platform or security team owns, and pools are the nodes a product team sizes. Keep one template per distinct sandbox shape (a coding agent, a browser, a notebook), not one per tenant: tenants get their own pools, namespaces or both.

## Isolation for every sandbox

Everything the [KubernetesAgentSandbox guide](../kubernetesagentsandbox/GUIDE.md) says about isolation applies here, multiplied by every sandbox the template produces:

- a sandboxing runtime class (`gvisor` on GKE Sandbox nodes, a Kata class elsewhere) with its node selector and toleration;
- a non-root, capability-free container with no privilege escalation and the runtime's seccomp profile;
- no Kubernetes credentials. The controller sets `automountServiceAccountToken` to false for a template's sandboxes when the template leaves it unset, so a sandbox holds credentials only if the template says so.

Resources matter more here than on a single sandbox: a warm pool reserves them for every idle member.

## The fence: secure default, your rules, or someone else's

With `networkPolicyManagement` at its default (`Managed`):

- **No `networkPolicy`** gets the controller's secure default. Ingress comes only from the router (pods labelled `app: sandbox-router` in `agent-sandbox-system`), which is how callers and the SDKs reach a sandbox. Egress goes only to the public internet, with private ranges, link-local and the cloud metadata server blocked. That also blocks cluster DNS, so a sandbox under the default resolves nothing in-cluster, and reaches public hosts only by IP or through public resolvers it is configured to use.
- **A `networkPolicy`** replaces the default with exactly its rules. An empty block denies everything both ways, and so does a direction with no rules, so an egress list with only DNS lets a sandbox resolve and nothing else. Keep the router's ingress (namespace `kubernetes.io/metadata.name: agent-sandbox-system`, pods `app: sandbox-router`) unless nothing should reach the sandboxes. A sidecar that listens on its own port (a mesh proxy, a metrics agent) needs its own ingress rule, or its health checks fail.

The policy is one object per template, so editing the rules re-fences every existing sandbox at once.

`Unmanaged` creates no policy and ignores `networkPolicy`. Choose it when Cilium, Calico or a mesh owns network policy on the cluster (and for host-name egress rules, which a Kubernetes NetworkPolicy cannot express). Deploy that system's policy before the pool, because an Unmanaged template with nothing behind it leaves its sandboxes open.

A NetworkPolicy only takes effect on a network that enforces it. Most managed clusters do; kind's default network does not.

## Claim policies: flexibility costs warm starts

A warm pool's members are running pods. Environment variables and volumes are fixed when a pod starts, so the only way to honor a claim's `env` or `volumeClaimTemplates` is a fresh sandbox, a cold start that takes the full pod start-up time. The two policies decide whether claims may do that at all:

- **`Disallowed`** (the default for both) refuses such a claim; it reports `Ready=False` with reason `EnvVarsInjectionRejected` or `VolumeClaimTemplatesError`. Every claim binds warm.
- **`Allowed`** lets a claim add names or templates the template doesn't define, never change what it does.
- **`Overrides`** also lets a claim change the template's own values, or replace a volume template by name (a larger workspace).

Open them only for what products genuinely pass per sandbox, and size the pool for the claims that pass nothing.

## Changing a template

The controller compares each warm pool member's blueprint with the template's. A change to the pod template, the volume claim templates or `service` makes members stale, and the pool replaces them by its update strategy: at once with `Recreate`, or as members are claimed with `OnReplenish`. Sandboxes already claimed keep running what they started with. Network policy changes apply to every sandbox immediately, claimed or not.

## Parity accounting

Built for 100% parity with the upstream SandboxTemplate (`extensions.agents.x-k8s.io/v1beta1`) at agent-sandbox v1.0.7. Every key of its served schema is modeled under the same name, and the mirror test holds the spec to the pinned definition in both directions.

| Upstream field | Spec field |
|----------------|------------|
| `metadata.namespace`, `metadata.labels`, `metadata.annotations` | `namespace`, `labels`, `annotations` (the envelope) |
| `spec.podTemplate.spec`, `.metadata` | `podTemplate`: Kubernetes 1.37's full PodSpec, from `catalog/kubernetes/core_v1_api.proto`, and the pod's labels and annotations |
| `spec.volumeClaimTemplates[]` | `volumeClaimTemplates[]`, with the full PersistentVolumeClaimSpec |
| `spec.service` | `service` |
| `spec.networkPolicy.ingress[]`, `.egress[]` | `networkPolicy.ingress[]`, `.egress[]`: Kubernetes networking/v1 rules, from `catalog/kubernetes/networking_v1_api.proto` |
| `spec.networkPolicyManagement` | `networkPolicyManagement` |
| `spec.envVarsInjectionPolicy` | `envVarsInjectionPolicy` |
| `spec.volumeClaimTemplatesPolicy` | `volumeClaimTemplatesPolicy` |

### What the spec adds beyond upstream's schema

Each addition is a rule the cluster enforces later, moved earlier:

- **The pod's rules**: closed sets, quantity grammar, number-or-name ports, one source per volume and handler, selector operator value counts (see the KubernetesAgentSandbox guide).
- **Network rules**: protocols as a closed set, ports bounded, an IP block standing alone in its peer, an end port only with a numeric port and never below it (the API server refuses each when it applies the policy).
- **A name on every volume claim template**, as a DNS label, unique within the template.
- **References** for the Secrets, ConfigMaps, ServiceAccount, PriorityClass, PersistentVolumeClaims and StorageClasses the pod names.

### Not modeled

- **Status**: a SandboxTemplate has none.
- **The NetworkPolicy's `podSelector` and `policyTypes`**: upstream's restricted subset omits both on purpose; the controller selects the template's sandboxes and always enforces both directions.

## Pairs well with

- **[KubernetesAgentSandboxController](../kubernetesagentsandboxcontroller/README.md)** — installs the definition and fences the template's sandboxes.
- **[KubernetesAgentSandboxWarmPool](../kubernetesagentsandboxwarmpool/README.md)** — keeps sandboxes from the template ready to claim.
- **[KubernetesServiceAccount](../kubernetesserviceaccount/README.md)** — the identity every sandbox runs as, with workload identity for cloud access.
- **[KubernetesStorageClass](../kubernetesstorageclass/README.md)** — an encrypted or faster class for every sandbox's workspace.
