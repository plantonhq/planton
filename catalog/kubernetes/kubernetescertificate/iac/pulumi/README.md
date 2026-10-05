# KubernetesCertificate Pulumi Module

## Usage

Run from this directory (it holds `Pulumi.yaml`, so the planton CLI runs this module). With no Kubernetes provider config, the module uses your kubeconfig; `--kube-context` picks the context:

```bash
planton pulumi init --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack>
planton pulumi update --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack> --kube-context <context>
```

The CLI wraps the manifest into a `KubernetesCertificateIacInput` (under `target`) and hands it to the module through `IAC_INPUT_YAML_FILE`. The module also reads that input from the Pulumi config key `planton:iac-input` or from `IAC_INPUT_YAML` (YAML content).

## Local Development

```bash
go build .
```

## Debug

```bash
planton pulumi preview --manifest ../../e2e/manifest.yaml --stack <org>/<project>/<stack> --kube-context <context> --diff
```
