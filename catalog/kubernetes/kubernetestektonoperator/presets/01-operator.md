# Tekton Operator preset

The complete install, which is deliberately tiny: the operator is a
lifecycle manager, not the product. It brings the Tekton CRDs and the
controller that turns a `TektonConfig` declaration into running
components (Pipelines, Triggers, Dashboard, Chains) — and nothing
else. Automatic component installation is disabled by design, so this
resource alone deploys no Tekton workloads: the KubernetesTekton
resource is the single place the cluster's Tekton shape is declared.

Know the destroy contract: the operator's CRDs delete with it, which
cascade-deletes any TektonConfig — always destroy the KubernetesTekton
resource FIRST (its teardown needs the operator running to process the
InstallerSet finalizers; the modules block until that completes).

Change first: nothing, usually. Set `image_registry` on clusters that
pull Tekton from a mirror or pull-through cache of ghcr.io (every image
Tekton publishes follows it, at its digest), and the resource blocks on
clusters with namespace quotas.

See [01-operator.yaml](./01-operator.yaml) for the manifest.
