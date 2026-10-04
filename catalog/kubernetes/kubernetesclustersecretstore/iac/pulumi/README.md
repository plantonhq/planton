# KubernetesClusterSecretStore Pulumi Module

## Module Behavior

- **Shared spec builder**: the CR spec renders through the
  `externalsecretsstore` package — the SAME builder the KubernetesSecretStore
  module uses, because upstream gives the two kinds an identical spec. One
  builder means the twins can never drift. The cluster kind adds only the
  `conditions` namespace fence on top.
- **Untyped CR apply**: the store applies as an untyped CustomResource
  (`external-secrets.io/v1`, kind ClusterSecretStore). ESO's validating
  webhook checks the applied spec strictly and the kind-cluster E2E lanes
  exercise the machinery live, so shape errors fail loudly without typed
  args.
- **Credential Secret materialization**: static credentials declared in the
  spec land in a `<resource-name>-credentials` Secret in the spec's secrets
  namespace, created BEFORE the CR (the CR depends on it), so ESO never
  observes a store whose secretRefs dangle.
- **Never waits for Ready**: store readiness depends on external
  reachability (the cloud secrets API, Vault) that is not part of applying
  the resource — the same never-block-on-a-controller posture as the
  cert-manager issuers.

## Usage

Run from this directory (it holds `Pulumi.yaml`, so the planton CLI runs this module). With no Kubernetes provider config, the module uses your kubeconfig; `--kube-context` picks the context:

```bash
planton pulumi init --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack>
planton pulumi update --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack> --kube-context <context>
```

The CLI wraps the manifest into a `KubernetesClusterSecretStoreIacInput` (under `target`) and hands it to the module through `IAC_INPUT_YAML_FILE`. The module also reads that input from the Pulumi config key `planton:iac-input` or from `IAC_INPUT_YAML` (YAML content).

## Local Development

```bash
go build .
```

## Debug

```bash
planton pulumi preview --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack> --kube-context <context> --diff
```
