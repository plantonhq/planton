# Terraform Module to Deploy KubernetesLangfuse

This module installs Langfuse from the official `langfuse` chart as a `helm_release` (wait, atomic, cleanup on fail) with the chart's four bundled stores held off, plus `kubernetes_namespace_v1.langfuse` (only when `create_namespace` is true) and `kubernetes_secret_v1.instance` (`<name>-secrets`, always): every declared secret value, and the credentials `secret.tf` generates when the spec leaves them empty -- `random_password` for the salt, NextAuth secret, Slack state secret, seeded admin password and the random part of the seeded project key pair, `random_bytes` for the 64-hex-character encryption key. Where the value Langfuse reads differs from the value the importer takes, the Secret carries both: `encryption-key-std-b64` holds the key's bytes in standard base64, and `init-project-public-key-suffix` and `init-project-secret-key-suffix` hold each key without its `pk-lf-`/`sk-lf-` prefix. Every generator ignores its generation shape after creation, so an imported credential never regenerates.

`values.tf` renders the chart's values and `env.tf` the three `additionalEnv` channels (both workloads, web only, worker only), exactly as the Pulumi module does. Unset keys are pruned, never rendered null: a null in a Helm values document deletes the chart's default for that key. A lifecycle precondition on the release refuses a name over 52 characters (the worker PodDisruptionBudget, `<name>-worker-pdb`, must fit a 63-character object name).

Outputs: `namespace`, `service`, `endpoint`, `external_url`, `otel_endpoint`, `port_forward_command`, `instance_secret_name`, `organization_id`, `project_id`, `project_public_key`, `project_secret_key_secret`, `admin_email`, `admin_password_secret` -- identical to the Pulumi module.

Generated `variables.tf` reflects the proto schema for `KubernetesLangfuse`.

## Usage

Use the Planton CLI (tofu) with the default local backend:

```shell
planton tofu init --manifest ../../e2e/manifest.yaml
planton tofu plan --manifest ../../e2e/manifest.yaml
planton tofu apply --manifest ../../e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --auto-approve
```

**Note**: Credentials are provided via IaC input (CLI), not in the manifest `spec`.

For a complete manifest, see [`e2e/manifest.yaml`](../../e2e/manifest.yaml).
