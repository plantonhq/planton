# Pulumi Module to Deploy KubernetesGoFeatureFlag

This module installs the GO Feature Flag relay proxy from the official `relay-proxy` chart (`charts.gofeatureflag.org`) as a real Helm release (`helm.v3 Release`: Helm's own lifecycle, wait, atomic rollback), plus the objects the chart does not provide:

- `namespace.go` -- the namespace, only when `create_namespace` is true;
- `env_secret.go` -- `<name>-env`, every secret value of the spec as an environment variable (created only when the spec holds one);
- `rbac.go` -- `<name>-flag-reader`, a Role and RoleBinding in each namespace a `config_map` retriever reads, granting `get` on exactly the named ConfigMaps;
- `helm_release.go` -- the release, its values rendered from the typed spec, with `helm_values` merged last and `fullnameOverride` re-pinned;
- `service_monitor.go` -- the optional ServiceMonitor on the `monitoring` port.

`relay_config.go` builds the relay configuration document: a `# server.monitoringPort: <port>` line (the chart scans the config string for that text to expose the monitoring port) followed by one JSON object with every `{{` escaped for the chart's tpl pass, and the indexed secret variables (`RETRIEVERS_<i>_...`, `FLAGSETS_<i>_...`, `AUTHORIZEDKEYS_*`, all prefixed with `runtime.env_variable_prefix`). Keys the relay pre-loads keep their documented camelCase; maps a secret variable also writes into (`headers`, `redisoptions`, a Kafka `config`) are lowercased. `relay_config_test.go` pins these rules. The Terraform module renders the same document byte for byte.

It exports the same outputs as the Terraform module: `namespace`, `service`, `api_endpoint`, `monitoring_endpoint` and `port_forward_command`.

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
