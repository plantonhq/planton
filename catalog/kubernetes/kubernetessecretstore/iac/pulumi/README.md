# KubernetesSecretStore Pulumi Module

## Module Behavior

- **Shared spec builder**: the CR spec renders through the
  `externalsecretsstore` package — the SAME builder the
  KubernetesClusterSecretStore module uses, because upstream gives the two
  kinds an identical spec. One builder means the twins can never drift.
- **Untyped CR apply**: the store applies as an untyped CustomResource
  (`external-secrets.io/v1`, kind SecretStore). ESO's validating webhook
  checks the applied spec strictly and the kind-cluster E2E lanes exercise
  the machinery live, so shape errors fail loudly without typed args.
- **Credential Secret materialization**: static credentials declared in the
  spec land in a `<resource-name>-credentials` Secret in the store's own
  namespace, created BEFORE the CR (the CR depends on it), so ESO never
  observes a store whose secretRefs dangle. Namespaced stores read Secrets
  from their own namespace, so refs carry no explicit namespace.
- **Never waits for Ready**: store readiness depends on external
  reachability (the cloud secrets API, Vault) that is not part of applying
  the resource — the same never-block-on-a-controller posture as the
  cert-manager issuers.

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
