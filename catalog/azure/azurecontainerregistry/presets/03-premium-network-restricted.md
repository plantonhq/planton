# Premium Network-Restricted Registry

This preset creates a Premium registry that stays publicly addressable but denies every connection not on an explicit CIDR allowlist, with dedicated data endpoints so downstream egress firewalls can allowlist the registry by exact hostname. It is the middle ground between an open registry and a fully private (private-endpoints-only) one.

## When to Use

- Registries that must be reachable from known networks only (office egress, CI runners) without the operational weight of private endpoints
- Security postures that require documented, reviewable network allowlists
- Environments whose egress firewalls allowlist by hostname and need the exact per-region data endpoints

## Key Configuration Choices

- **`networkRuleSet.defaultAction: DENY`** -- the allowlist posture; without DENY the rule set is a no-op because Azure's default action allows everything
- **`ipRules`** -- public IPv4 CIDR ranges only; ARM supports allow rules exclusively, so entries carry no per-rule action. Trusted Azure services (ACR Tasks, Microsoft Defender) still get through by default (`networkRuleBypassOption` unset = AzureServices) -- set `NONE` to close even that
- **`dataEndpointEnabled: true`** -- blob pulls move from shared regional storage hosts to `{name}.{region}.data.azurecr.io`, making firewall allowlists exact instead of broad
- **For fully private registries** -- set `publicNetworkAccessEnabled: false` instead of a rule set and reach the registry through private endpoints; disabling `exportPolicyEnabled` then also becomes available as a data-exfiltration control

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | The registry's region | Your regional deployment strategy |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `yourregistry123` (`registryName`) | Globally unique registry name (5-50 lowercase alphanumerics) | Becomes `{name}.azurecr.io` |
| `<office-egress-cidr>` | Your office/VPN egress range, e.g. `203.0.113.0/24` | Network team |
| `<ci-runner-cidr>` | Your CI runners' egress range | CI provider docs or NAT gateway outputs |
