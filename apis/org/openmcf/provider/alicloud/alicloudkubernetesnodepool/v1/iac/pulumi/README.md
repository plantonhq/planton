# Pulumi Module to Deploy AlicloudKubernetesNodePool

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

Then make the script executable and run your command (e.g., `preview` or `update`).

```bash
chmod +x debug.sh
openmcf pulumi preview \
  --manifest ../hack/manifest.yaml \
  --stack organization/<project>/<stack> \
  --module-dir .
```

## Module Overview

This Pulumi module deploys an ACK Kubernetes node pool using a single `cs.NodePool` resource. The module reads an `AlicloudKubernetesNodePoolStackInput` protobuf message, resolves foreign key references (cluster ID, VSwitch IDs, security group IDs), and provisions a node pool with configurable instance types, disk configuration, auto-scaling, managed lifecycle, spot instances, and Kubernetes scheduling properties (labels, taints).

---

## Further Reading

- **[examples.md](./examples.md)**: Runnable manifests for common node pool configurations.
- **[overview.md](./overview.md)**: Module architecture, file organization, and design decisions.
- **[hack/manifest.yaml](../hack/manifest.yaml)**: Minimal test manifest.
