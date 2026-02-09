# OpenStackFloatingIp Pulumi Module -- Architecture Overview

## Module Flow

```
OpenStackFloatingIpStackInput
  ├── target: OpenStackFloatingIp (api.proto)
  │   ├── metadata.name → OpenMCF identity (FIPs have no OS name)
  │   └── spec: OpenStackFloatingIpSpec
  │       ├── floating_network_id (required StringValueOrRef FK → OpenStackNetwork)
  │       ├── port_id (optional StringValueOrRef FK → OpenStackNetworkPort)
  │       ├── fixed_ip (plain string, requires port_id)
  │       ├── subnet_id (plain string, external subnet)
  │       ├── address (specific IP request, ForceNew)
  │       ├── description
  │       ├── tags[]
  │       └── region
  └── provider_config: OpenStackProviderConfig

         │
         ▼

   initializeLocals()
  ├── FloatingNetworkId = spec.FloatingNetworkId.GetValue()    [always]
  └── PortId = spec.PortId.GetValue()                          [if present]

         │
         ▼

   networking.NewFloatingIp()
  ├── Pool = FloatingNetworkId
  ├── PortId = PortId (if non-empty)
  ├── FixedIp = spec.FixedIp (if non-empty)
  ├── SubnetId = spec.SubnetId (if non-empty)
  ├── Address = spec.Address (if non-empty)
  ├── Description, Tags, Region
  └── Provider = openstackProvider

         │
         ▼

   Exports → stack_outputs.proto
  ├── floating_ip_id = resource ID
  ├── address = allocated IP (FK target for FloatingIpAssociate)
  ├── floating_network_id = external network
  ├── port_id = associated port (or empty)
  ├── fixed_ip = mapped fixed IP (or empty)
  └── region = deployment region
```

## FK Resolution

| Field | Type | Resolution |
|-------|------|------------|
| `floating_network_id` | Required FK | `spec.FloatingNetworkId.GetValue()` -- always present |
| `port_id` | Optional FK | `spec.PortId.GetValue()` -- nil-guarded, empty when not set |

## Resource Mapping

| Pulumi Resource | TF Equivalent | Count |
|-----------------|---------------|-------|
| `networking.FloatingIp` | `openstack_networking_floatingip_v2` | 1 |

Single-resource module. No multi-resource pattern.
