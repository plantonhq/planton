# KubernetesIssuer Pulumi Module

## Usage

```bash
export STACK_INPUT=$(cat ../hack/manifest.yaml | base64)
pulumi up
```

## Local Development

```bash
make deps
make build
```

## Debug

```bash
bash debug.sh ../hack/manifest.yaml
```
