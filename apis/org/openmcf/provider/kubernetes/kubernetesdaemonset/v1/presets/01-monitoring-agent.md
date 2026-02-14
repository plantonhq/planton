# Monitoring Agent DaemonSet

This preset deploys a monitoring agent on every node in the cluster, including control-plane nodes. Suitable for node-level metrics collection, log forwarding, or security agents that need to run on all nodes.

## When to Use

- Node-level metrics exporters (e.g., Prometheus node exporter, Datadog agent)
- Log forwarders (e.g., Fluent Bit, Filebeat)
- Security agents that need host-level access on every node

## Key Configuration Choices

- **Runs on all nodes** -- DaemonSet ensures exactly one pod per node, including new nodes added to the cluster
- **Control-plane toleration** -- tolerates `node-role.kubernetes.io/control-plane` taint so the agent also runs on control-plane nodes
- **Low resource footprint** (`50m`/`64Mi` requests, `200m`/`256Mi` limits) -- agents should be lightweight to avoid impacting application workloads

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace for the DaemonSet | Your namespace management or `KubernetesNamespace` resource |
| `<your-container-registry>/<your-agent-image>` | Container image for the monitoring agent | Your container registry |
| `<your-image-tag>` | Image tag or version | Your CI/CD pipeline or vendor documentation |

## Related Presets

- **02-log-collector** -- DaemonSet tailored for log collection with host path mounts
