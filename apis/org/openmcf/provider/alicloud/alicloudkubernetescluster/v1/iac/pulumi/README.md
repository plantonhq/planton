# Pulumi Module to Deploy AliCloudKubernetesCluster

## CLI usage (OpenMCF pulumi)

```bash
# Preview
openmcf pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Update (apply)
openmcf pulumi update \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir . \
  --yes

# Refresh
openmcf pulumi refresh \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .

# Destroy
openmcf pulumi destroy \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

## Debugging

This module includes a `debug.sh` helper. To enable debugging, edit `Pulumi.yaml` and uncomment the `runtime.options.binary` line so Pulumi runs the program via the script:

```yaml
name: alicloud-module-test-pulumi-project
runtime:
  name: go
#  options:
#    binary: ./debug.sh
```

Then make the script executable and run your command (e.g., `preview` or `update`). See `docs/pages/docs/guide/debug-pulumi-modules.mdx` for full instructions.

```bash
chmod +x debug.sh
openmcf pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

## Module Overview

This Pulumi module deploys an Alibaba Cloud ACK Managed Kubernetes cluster using a single `cs.ManagedKubernetes` resource. The module reads an `AliCloudKubernetesClusterStackInput` protobuf message, resolves defaults for optional fields, and provisions the cluster with the specified networking, security, addon, logging, maintenance, and auto-upgrade configuration.

The module supports both Flannel (overlay) and Terway (ENI-based) CNI modes through mutually exclusive spec fields (`podCidr` for Flannel, `podVswitchIds` for Terway). Addons, maintenance windows, and auto-upgrade policies are configured via structured spec fields that map directly to the provider resource arguments.

All 11 stack outputs (cluster ID, name, API server endpoints, VPC ID, security group ID, NAT gateway ID, worker RAM role name, and RRSA OIDC metadata) are exported for use by downstream components such as AliCloudKubernetesNodePool.

---

## Further Reading

- **[examples.md](./examples.md)**: Runnable manifests for common cluster configurations.
- **[overview.md](./overview.md)**: Module architecture, file organization, and design decisions.
- **[hack/manifest.yaml](../hack/manifest.yaml)**: Minimal test manifest.
