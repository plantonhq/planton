# KubernetesStigmer Terraform Module

Installs Stigmer: the `stigmer-operator` Helm chart (`oci://ghcr.io/stigmer/charts`) as one real Helm release, the Secrets the spec gives values for, and the one Stigmer resource the operator reconciles. The Pulumi module in `../pulumi` applies the same documents; its `module/parity_test.go` holds them together against a real `tofu plan` of this module.

## Files

- `locals.tf` — the fixed names and the documents as data: `local.chart_values`, `local.secrets` and `local.stigmer_document`. A key renders only when the spec declares it; each object is built with null for an undeclared key and then filtered, because HCL refuses a conditional whose branches are objects of different shapes.
- `main.tf` — the resources, in order: the operator's namespace, the chart release, the Secrets, the restore check, the cleanup wait, the Stigmer.
- `preflight.tf` — the refusals, as preconditions on `terraform_data.preflight`, read at plan time and skipped on destroy.
- `scripts.tf` — the two shell scripts both engines run, byte-identical to the Pulumi module's copies.
- `outputs.tf` — the outputs, derived from the declaration.

## Module Behavior

- **The Stigmer is a `kubectl_manifest`** (alekc/kubectl, server-side apply), which needs no definition at plan time: on a clean cluster the definition arrives with the release the Stigmer depends on.
- **Stigmers are read only when their definition exists**: a read of a kind the cluster does not serve fails the plan instead of returning nothing. Reads are in the Stigmer's own namespace; the data source has no cluster-wide list.
- **The cleanup wait is a state-only `terraform_data`** whose input holds the script and its inputs, because a destroy-time step can read only its own state. Nobody with access to the cluster can change what the runner executes. A fix to the script reaches a destroy after one apply.
- **The restore check is a `kubectl_manifest` ConfigMap** with `force_new` and a create-time `local-exec`, so it runs again whenever the namespace or the archive moves.
- **Both scripts need `kubectl` on PATH** and read the kubeconfig from the process environment (`KUBECONFIG`, else `KUBE_CONFIG_PATH`, and `KUBE_CTX`), never from state.

## Usage

```shell
planton tofu apply --manifest stigmer.yaml
```
