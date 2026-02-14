# Log Collector DaemonSet

This preset deploys a log collector on every node with host path mounts for `/var/log` and container log directories. Designed for log forwarders like Fluent Bit, Fluentd, or Filebeat that need to read container and system logs from the node filesystem.

## When to Use

- Cluster-wide log forwarding to a central logging system (Elasticsearch, Loki, CloudWatch, etc.)
- You need access to node-level logs (`/var/log`) and container runtime logs
- The log collector must run on all nodes including control-plane

## Key Configuration Choices

- **Host path mounts** -- mounts `/var/log` and `/var/lib/docker/containers` from the host for reading container logs
- **Higher memory limit** (`512Mi`) -- log collectors can be memory-intensive when buffering; adjust based on log volume
- **Control-plane toleration** -- collects logs from control-plane nodes as well

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-namespace>` | Target namespace for the DaemonSet | Your namespace management or `KubernetesNamespace` resource |
| `<your-container-registry>/<your-log-collector-image>` | Container image (e.g., `fluent/fluent-bit`, `elastic/filebeat`) | Vendor documentation |
| `<your-image-tag>` | Image tag or version | Vendor release page |

## Related Presets

- **01-monitoring-agent** -- Lightweight agent without host path mounts
