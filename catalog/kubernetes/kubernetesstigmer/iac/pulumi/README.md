# KubernetesStigmer Pulumi Module

Installs Stigmer: the `stigmer-operator` Helm chart (`oci://ghcr.io/stigmer/charts`) as one real Helm release, the Secrets the spec gives values for, and the one Stigmer resource the operator reconciles. The Terraform module in `../tf` applies the same documents; `module/parity_test.go` holds them together against a real `tofu plan`.

## Files

- `module/main.go` — the resources, in order: the operator's namespace (with `operator.create_namespace`), the chart release, the Secrets, the restore check, the cleanup marker, the Stigmer, the outputs.
- `module/render.go` — the documents as plain data: the chart's values (`helmValues`), the Secrets (`materializedSecrets`) and the Stigmer's spec (`stigmerSpec`). A key renders only when the spec declares it.
- `module/preflight.go` — the refusals, read from the cluster before anything registers, skipped on destroy.
- `module/gates.go` — the restore check (a `BeforeCreate` hook on a ConfigMap, re-created whenever the namespace or the archive moves) and the cleanup wait (a `BeforeDelete` hook on a marker ConfigMap the Stigmer depends on).
- `module/scripts.go` — the two shell scripts both engines run, byte-identical to the Terraform module's copies (`hack/guards/ensure_cross_engine_script_parity.sh`).
- `module/outputs.go` — the outputs, derived from the declaration.

## Module Behavior

- **The Stigmer is untyped** (`apiextensions.NewCustomResource` with a plain map). Its definition arrives inside the chart's release, and Pulumi previews a custom resource of a kind the cluster does not serve yet without asking the cluster, so a first deploy plans on a clean cluster.
- **The release name is fixed to `stigmer-operator`**, as Stigmer's own guide installs it, and the Stigmer to `stigmer`.
- **Delete hooks need `pulumi destroy --run-program`**, which Planton's stack runner passes. The cleanup wait streams its progress to the engine's log.
- **Both hooks run `kubectl`** with the deploy's kubeconfig environment (`KUBECONFIG`, else `KUBE_CONFIG_PATH`, and `KUBE_CTX`); without kubectl on PATH the restore check refuses and the cleanup wait says so and lets the destroy go on.

## Tests

`go test ./catalog/kubernetes/kubernetesstigmer/...` runs the render tests. The cross-engine test plans the Terraform module against a cluster: `STIGMER_TF_PARITY=1 KUBE_CONFIG_PATH=<kubeconfig>` with `tofu` on PATH, on a cluster where namespace `stigmer` exists and no other Stigmer operator runs.
