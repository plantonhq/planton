# Terraform Module to Deploy KubernetesGoFeatureFlag

This module installs the GO Feature Flag relay proxy from the official `relay-proxy` chart as a `helm_release` (wait, atomic, cleanup on fail), with `kubernetes_namespace_v1.relay` (only when `create_namespace` is true), `kubernetes_secret_v1.env` (the secret environment, only when the spec holds a secret), `kubernetes_role_v1.flag_reader` and `kubernetes_role_binding_v1.flag_reader` (one per namespace a `config_map` retriever reads, `get` on exactly the named ConfigMaps) and `kubectl_manifest.service_monitor` (only when the ServiceMonitor is enabled).

`locals.tf` renders the relay configuration document, the secret environment and the ConfigMap grants exactly as the Pulumi module does -- the document is byte-identical (Go `json.Marshal` and OpenTofu `jsonencode` sort keys and escape alike). Lists that hold free-form values (exporters, flag sets) arrive untyped, so their fields are read with `try()` and numeric fields with `tonumber()`; a Kafka `config` is lowercased to a depth of six. Lifecycle preconditions on the release refuse a name over 63 characters, a key containing a comma, a key shared between flag sets, a sensitive header name containing `_`, a header declared plain and sensitive, and an extra variable colliding with a generated secret variable.

Outputs: `namespace`, `service`, `api_endpoint`, `monitoring_endpoint`, `port_forward_command` -- identical to the Pulumi module.

Generated `variables.tf` reflects the proto schema for `KubernetesGoFeatureFlag`.

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
