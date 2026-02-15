---
title: "Enterprise API"
description: "This preset deploys a production-grade API with User Assigned managed identity, Key Vault secrets, ACR authentication via identity, IP security restrictions, full health probe coverage (liveness,..."
type: "preset"
rank: "03"
presetSlug: "03-enterprise-api"
componentSlug: "azurecontainerapp-research-design-documentation"
componentTitle: "AzureContainerApp: Research & Design Documentation"
provider: "azure"
icon: "package"
order: 3
---

# Enterprise API

This preset deploys a production-grade API with User Assigned managed identity, Key Vault secrets, ACR authentication via identity, IP security restrictions, full health probe coverage (liveness, readiness, startup), and HTTP auto-scaling. This is the standard pattern for enterprise APIs that require security controls, credential-free authentication, and production resilience.

## When to Use

- Production APIs that must comply with enterprise security policies
- Services that need Key Vault-managed secrets (no plain-text credentials in manifests)
- APIs restricted to corporate IP ranges (office, VPN)
- Workloads pulling from private ACR registries using managed identity (no registry password)
- High-availability services that need at least 2 replicas and all three probe types

## Key Configuration Choices

- **User Assigned Identity** -- Enables credential-free access to Key Vault and ACR; shared identity lifecycle independent of the app
- **Key Vault secret** (`keyVaultSecretId`) -- DB connection string stored in Key Vault, not in the manifest
- **ACR with identity** -- Registry authentication via managed identity instead of username/password
- **2 min replicas** (`minReplicas: 2`) -- High availability; never scale below 2
- **20 max replicas** (`maxReplicas: 20`) -- Production ceiling; adjust based on load testing
- **1.0 vCPU / 2 GiB memory** -- Suitable for production API workloads; right-size after profiling
- **All three probes** -- Liveness (restart on failure), readiness (remove from LB until ready), startup (tolerate slow starts)
- **IP restrictions** -- Allow corporate office CIDR, deny all other traffic
- **Graceful shutdown** (`terminationGracePeriodSeconds: 30`) -- Allows in-flight requests to complete before termination

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<container-app-environment-id>` | ARM ID of the Container App Environment | Azure portal or `AzureContainerAppEnvironment` status outputs |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-user-assigned-identity-id>` | User Assigned Identity ARM resource ID (used for secrets, registries, and identity) | Azure portal or `AzureUserAssignedIdentity` status outputs |
| `keyVaultSecretId: https://...` | Key Vault secret URI for your DB connection | Azure portal -> Key Vault -> Secrets |
| `ipAddressRange: 203.0.113.0/24` | Your corporate office or VPN CIDR | Network administrator |

## Related Presets

- **01-web-service** -- Use instead for simpler web services without identity or IP restrictions
- **02-background-worker** -- Use instead for queue-processing workers with no ingress
