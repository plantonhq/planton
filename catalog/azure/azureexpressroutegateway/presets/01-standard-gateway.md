# Standard Gateway

This preset creates the gateway alone: one scale unit (~2 Gbps, the guaranteed floor ARM auto-scales above) in the hub, with no connections yet. Connections join circuit peerings later -- typically once the carrier has provisioned the circuit.

## When to Use

- The circuit is ordered and provisioning; the WAN side should be ready when it lands
- Any hub expecting ExpressRoute connectivity (a hub holds at most one of these gateways)

## Key Configuration Choices

- **One scale unit** -- the right floor unless committed circuit bandwidth already exceeds ~2 Gbps; raising it later is an in-place update
- **Cost honesty** -- the gateway bills hourly PER SCALE UNIT from creation, connections or not, and takes ~30 minutes to provision; the verified figure lives in the kind's generated estimate at `catalog/_pricing/estimates/azureexpressroutegateway.yaml`
- **Region must match the hub's** -- the gateway deploys INTO the hub

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<azure-region>` | The gateway's region (must match the hub's) | The hub's configuration |
| `<your-resource-group-name>` | Name of the resource group | Azure portal or `AzureResourceGroup` status outputs |
| `<your-virtual-hub-arm-id>` | ARM ID of the hub | `AzureVirtualHub` status outputs (`virtual_hub_id`), or reference it with valueFrom |
