# Premium Enterprise CDN

A Premium-tier Front Door profile with Private Link connectivity to an Azure App Service backend. The origin is reached exclusively through Azure's backbone network without any public internet exposure.

## What This Preset Provides

- **Premium SKU**: All Standard features plus Private Link, enhanced WAF support, and Bot Manager eligibility
- **Private Link origin**: Front Door connects to App Service through Azure Private Link -- the backend can have public access completely disabled
- **HTTPS health probes**: GET-based probes on `/health` every 30 seconds for comprehensive health evaluation
- **Edge caching**: Responses cached at edge locations with compression for common web content types
- **Certificate validation**: Enforced (`certificateNameCheckEnabled: true`), which is required for Private Link origins
- **Session affinity**: Enabled for stateful web application backends
- **HTTPS redirect**: All HTTP traffic automatically redirected to HTTPS

## When to Use

- Enterprise workloads requiring zero public internet exposure for backends
- Compliance-sensitive applications (healthcare, finance) where traffic must stay on Azure backbone
- App Service, Function App, or Container Apps backends that should only be reachable via Front Door
- Organizations already using or planning to use Front Door WAF policies (Premium required)

## What to Customize

| Field | What to Set |
|-------|-------------|
| `resourceGroup.value` | Your Azure resource group name |
| `name` | Globally unique profile name |
| `origins[].hostName` | Your App Service hostname |
| `origins[].originHostHeader` | Usually same as `hostName` for App Service |
| `privateLink.location` | Azure region of your App Service (e.g., `eastus`) |
| `privateLink.privateLinkTargetId` | Full ARM resource ID of your App Service |
| `privateLink.targetType` | `sites` for App Service, `blob` for Storage, `managedEnvironments` for Container Apps |

## Post-Deployment Steps

After provisioning, you must **approve the Private Link connection** on the target resource:

1. Go to the Azure Portal > Your App Service > Networking > Private Endpoint Connections
2. Find the pending connection from Front Door
3. Click **Approve**

Alternatively, approve via Azure CLI:

```bash
az network private-endpoint-connection approve \
  --resource-name <your-app> \
  --resource-group <resource-group> \
  --type Microsoft.Web/sites \
  --name <connection-name>
```

Traffic will not flow until the Private Link connection is approved.
