# Pulumi Module to Deploy KubernetesLangfuse

This module installs Langfuse from the official `langfuse` chart (`langfuse.github.io/langfuse-k8s`) as a real Helm release (`helm.v3 Release`: Helm's own lifecycle, wait, atomic rollback), with the chart's four bundled stores held off and every store wired by reference, plus the objects the chart does not provide:

- `namespace.go` -- the namespace, only when `create_namespace` is true;
- `secrets.go` -- `<name>-secrets`, the instance Secret: every declared secret value, plus the salt, encryption key, NextAuth secret, seeded project key pair, seeded admin password and Slack state secret the module generates when the spec leaves them empty (generation shape ignored after creation, so an imported credential never regenerates);
- `helm_release.go` -- the release, its values rendered from the typed spec, with `helm_values` merged next and the re-pins (`fullnameOverride`, the bundled stores off) merged last.

`values.go` renders the chart's values; `blob.go` resolves each upload type's object storage target (event, media, batch export); `env.go` renders every setting the chart has no value for into the chart's three `additionalEnv` channels (both workloads, web only, worker only), sorted by name; `sso.go` lists the single sign-on providers; `bodies.go` renders the shared Kubernetes workload messages. The Terraform module renders the same values document.

It exports the same outputs as the Terraform module: `namespace`, `service`, `endpoint`, `external_url`, `otel_endpoint`, `port_forward_command`, `instance_secret_name`, `organization_id`, `project_id`, `project_public_key`, `project_secret_key_secret`, `admin_email` and `admin_password_secret`.

## CLI usage (Planton pulumi)

```bash
# Preview
planton pulumi preview \
  --manifest ../../e2e/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Update (apply)
planton pulumi update \
  --manifest ../../e2e/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes

# Destroy
planton pulumi destroy \
  --manifest ../../e2e/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

Credentials (the target cluster's kubeconfig) arrive through the IaC input's `provider_config`, never in the manifest `spec`.
