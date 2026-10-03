# Kubernetes queue messaging: RabbitMQ Cluster Operator + RabbitMQ cluster — two kinds at full depth

## What changed

- **KubernetesRabbitMqOperator (925, new)** — installs the RabbitMQ
  Cluster Operator (MPL-2.0, maintained by the RabbitMQ team) from its
  released single-file manifest, pinned to v2.22.3. The operator has no
  Helm chart: the released `cluster-operator.yml` is the official
  distribution, and the modules fetch it tag-pinned and apply it per
  document — Terraform through one `kubectl_manifest` keyed by each
  document's own composed identity (`apiVersion//kind//name[//namespace]`,
  exactly the kubectl importer's ID form), Pulumi through a classic-yaml
  ConfigFile with a transformation that patches the typed overrides onto
  the operator Deployment. Typed surface: watch-namespace fencing (empty =
  all namespaces, the upstream default), fleet-wide default images for
  air-gapped clusters, operator image/resources/placement overrides.
  cert-manager is a hard prerequisite (the manifest ships fail-closed
  admission webhooks whose serving certificate is a cert-manager
  Certificate with CA injection) and is declared as a registry
  prerequisite. The namespace is fixed at `rabbitmq-system` (baked into
  the manifest's own cross-references; the webhooks are cluster-scoped
  singletons, so exactly one install per cluster). The spec deliberately
  has no version field — the installed CRD schema stays in lockstep with
  the typed SDK generated from it. Destroying the operator deletes the
  CRD and cascade-deletes every RabbitmqCluster; the spec carries the
  warning prominently. Both engines apply with server-side apply — the
  CRD document (~342 KB) exceeds the client-side annotation cap.

- **KubernetesRabbitMq (926, new)** — declares one RabbitMQ cluster as a
  `RabbitmqCluster` custom resource at the full meaningful surface:
  quorum-oriented replica counts (odd numbers taught on the field; the
  operator does not support scale-down), persistent storage with a
  storage-class reference and an explicit ephemeral (emptyDir) arm,
  same-memory requests/limits teaching (RabbitMQ derives its memory
  watermark from the container limit), the full configuration layer
  (additional plugins, rabbitmq.conf, advanced.config, rabbitmq-env.conf
  with the CRD's own no-shell-substitution rule mirrored, Erlang inet),
  TLS from a certificate Secret (the cert-manager seam; mutual TLS via a
  CA Secret; optional closing of every plain listener), client-Service
  type/annotations as the cloud-exposure surface (no ingress resources
  by design), tolerations + node-selector placement (rendered as
  required node affinity — the CR has no nodeSelector field) + a
  spread-across-nodes anti-affinity switch, post-deploy and feature-flag
  knobs, and Vault or external-Secret credential backends. The operator
  generates the admin credentials; the `<name>-default-user` Secret is
  exported as an output, along with the client/headless Services and the
  effective AMQP/management endpoints. The CR declares background
  deletion propagation on both engines — the operator's finalizer is the
  cascade. Pulumi renders through a typed SDK generated from the pinned
  CRD; Terraform through a hand-authored `kubectl_manifest` twin.

- **Import machinery: `from_address_key_segment`** — a new derivation arm
  in the catalog kind import map vocabulary for modules that apply
  multi-GVK manifest bundles keyed by composed identity: each composed
  import-ID placeholder derives from one `//`-delimited segment of the
  address key (an out-of-range index resolves empty, so the namespace
  group drops for cluster-scoped documents). Unit-tested; the RabbitMQ
  operator's import map is its first user.

- **E2E surface (authored and compiled; live proving runs separately)**:
  dedicated verifiers — the operator install (Deployment Available, CRD
  Established, both webhook configurations present, CRD asserted GONE on
  destroy) and the cluster (ClusterAvailable/AllReplicasReady conditions,
  the naming contract, a live message round-trip through the management
  API as the generated credentials, and a quorum-queue durability arm
  that deletes a broker pod mid-proof) — plus five scenarios across the
  two kinds and import maps for both. Profiles ship as `pending_proof`.

## Validation

Spec tests for both kinds (every CEL rule accept+reject locked); offline
`tofu` plan and `pulumi preview` proofs across full-surface AND minimal
shapes for all four modules with type-fidelity and patch spot-checks;
secret-coverage, reference, containment, import-map and outputs
conformance gates; repo-wide Bazel build; e2e-build/e2e-vet; license
footers; all presets and scenario manifests CLI-validated. The offline
gates caught and fixed in-session: three HCL conditional
type-unification defects, a CEL exclusion rule that the platform's
defaulting middleware would have made unexpressible, a pointless
sensitive-exemption annotation, and a cross-engine divergence in
pull-secret folding.
