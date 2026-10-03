# KubernetesClusterIssuer Pulumi Module

## Usage

```bash
export IAC_INPUT=$(cat ../../e2e/manifest.yaml | base64)
pulumi up
```

## Local Development

```bash
make deps
make build
```

## Debug

```bash
bash debug.sh ../../e2e/manifest.yaml
```
