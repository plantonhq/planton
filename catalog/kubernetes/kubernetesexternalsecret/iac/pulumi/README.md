# KubernetesExternalSecret Pulumi Module

## Module Behavior

- **Typed-to-CRD rendering**: `spec_builder.go` renders the typed spec into
  the `external-secrets.io/v1` ExternalSecret CRD shape, applied as an
  untyped CustomResource — the same posture as the store kinds and the
  cert-manager family. ESO's validating webhook checks the applied spec
  strictly and the kind-cluster E2E lanes verify the full sync loop live,
  so shape errors fail loudly without typed args.
- **Pinned Secret name**: the CR's `target.name` is always rendered from
  the resolved Secret name (`target.name` when set, else `metadata.name`),
  so the exported `secret_name` output can never drift from what the
  operator creates.
- **No credential materialization**: the sync declaration carries no
  credentials — authentication lives on the store kinds.
- **Never waits for the sync**: the materialized Secret appears when the
  operator reaches the backend, which is not part of applying the resource.
  The E2E verifier (not the module) asserts synced state.

## Usage

Run from this directory (it holds `Pulumi.yaml`, so the planton CLI runs this module). With no Kubernetes provider config, the module uses your kubeconfig; `--kube-context` picks the context:

```bash
planton pulumi init --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack>
planton pulumi update --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack> --kube-context <context>
```

The CLI wraps the manifest into a `KubernetesExternalSecretIacInput` (under `target`) and hands it to the module through `IAC_INPUT_YAML_FILE`. The module also reads that input from the Pulumi config key `planton:iac-input` or from `IAC_INPUT_YAML` (YAML content).

## Local Development

```bash
go build .
```

## Debug

```bash
planton pulumi preview --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack> --kube-context <context> --diff
```
