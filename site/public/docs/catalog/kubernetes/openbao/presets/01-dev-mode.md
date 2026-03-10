---
title: "OpenBao Dev Mode"
description: "This preset deploys OpenBao (open-source Vault fork) in a simple configuration with the UI enabled and ingress access. Suitable for development and testing secrets management workflows."
type: "preset"
rank: "01"
presetSlug: "01-dev-mode"
componentSlug: "openbao"
componentTitle: "OpenBao"
provider: "kubernetes"
icon: "package"
order: 1
---

# OpenBao Dev Mode

This preset deploys OpenBao (open-source Vault fork) in a simple configuration with the UI enabled and ingress access. Suitable for development and testing secrets management workflows.

## When to Use

- Development or staging environments for secrets management
- Evaluating OpenBao/Vault API compatibility
- Environments where HA and TLS are not yet required

## Key Configuration Choices

- **UI enabled** -- web interface for managing secrets, policies, and auth methods
- **Ingress enabled** -- exposes OpenBao at the specified hostname
- **No HA** -- single server instance; suitable for non-production use
- **No TLS** -- unencrypted communication; enable `tlsEnabled` for production

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<your-openbao.example.com>` | Hostname for the OpenBao UI and API | Your DNS provider |

## Related Presets

- **02-production-ha** -- High-availability mode with Raft storage and TLS
- **03-production-ha-gcp-auto-unseal** -- HA mode with GCP Cloud KMS auto-unseal and Workload Identity
