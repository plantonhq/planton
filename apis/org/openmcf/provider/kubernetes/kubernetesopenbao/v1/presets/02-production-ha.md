# Production OpenBao with HA

This preset deploys OpenBao in high-availability mode with 3 replicas, TLS encryption, and the sidecar injector for automatic secrets injection into pods.

## When to Use

- Production secrets management requiring fault tolerance
- Environments where TLS encryption is mandatory for all communication
- Applications that need automatic secrets injection via sidecar (Agent Injector)

## Key Configuration Choices

- **HA mode** with 3 replicas -- uses Raft consensus for leader election and data replication; tolerates 1 node failure
- **TLS enabled** -- encrypts client-to-server and server-to-server communication
- **Agent Injector enabled** -- automatically injects secrets into pods via annotations; no application code changes needed
- **UI enabled** with ingress -- web interface for managing secrets accessible at the specified hostname

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-openbao.example.com>` | Hostname for the OpenBao UI and API | Your DNS provider |

## Related Presets

- **01-dev-mode** -- Simple single-server deployment for development
- **03-production-ha-gcp-auto-unseal** -- HA mode with GCP Cloud KMS auto-unseal and Workload Identity
