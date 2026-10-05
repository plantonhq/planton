# AzureBastionHost

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**AzureBastionHostSpec** defines an Azure Bastion host -- the managed
jump service that opens browser-based (and, on higher SKUs,
native-client) RDP/SSH sessions to virtual machines over their
PRIVATE addresses, so the machines themselves never expose a public
IP or an inbound NSG rule.

**Four SKUs, two deployment shapes**:
- **DEVELOPER**: Azure-hosted SHARED infrastructure -- free, no
  dedicated subnet, no public IP; the host simply attaches to a
  virtual network (virtual_network_id). One connection per user, no
  feature knobs, offered in a limited set of regions. Dev/test only.
- **BASIC** (the default), **STANDARD**, and **PREMIUM**: dedicated
  infrastructure deployed into the network's "AzureBastionSubnet"
  with a Standard-SKU static public IP (ip_configuration). STANDARD
  unlocks scaling (2-50 scale units) and the feature knobs (file
  copy, native-client tunneling, IP connect, shareable links,
  Kerberos); PREMIUM adds session recording and private-only
  deployment (omit the public IP to keep the host off the internet
  entirely).

**The AzureBastionSubnet contract**: dedicated-infrastructure hosts
deploy ONLY into a subnet whose ARM name is EXACTLY
"AzureBastionSubnet", sized /26 or larger. ARM validates the name at
deploy time (the subnet reference resolves after validation, so it
cannot be checked here).

**SKU changes**: upgrading (Basic -> Standard -> Premium) is
in-place; DOWNGRADING replaces the host (the provider forces a new
resource -- Azure has no downgrade path). Choose the long-term tier
deliberately.

The host bills hourly per SKU (plus per-scale-unit on
Standard/Premium) from the moment it provisions; creates take ~10
minutes.

## Example

```yaml
# Offline-plan test manifest. Exercises the full dedicated-
# infrastructure surface: STANDARD with the feature knobs, explicit
# scale units, zone redundancy, clipboard disabled, and user tags
# merged over the derived ones. (The Developer shape is exercised by
# its own offline plan -- its virtual_network_id arm is exclusive with
# this one's ip_configuration by SKU.)
apiVersion: azure.planton.dev/v1alpha1
kind: AzureBastionHost
metadata:
  name: test-bastion-host
  org: test-org
  env: dev
spec:
  region: eastus
  resourceGroup:
    value: platform-rg
  name: test-bastion
  sku: STANDARD
  ipConfiguration:
    name: bastion-ip-config
    subnetId:
      value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.Network/virtualNetworks/platform-vnet/subnets/AzureBastionSubnet
    publicIpAddressId:
      value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/platform-rg/providers/Microsoft.Network/publicIPAddresses/bastion-pip
  scaleUnits: 4
  copyPasteEnabled: false
  fileCopyEnabled: true
  ipConnectEnabled: true
  tunnelingEnabled: true
  zones:
    - "1"
    - "2"
    - "3"
  tags:
    cost-center: platform
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.resourceGroup` | `string \| valueFrom` | yes |  | AzureResourceGroup (`status.outputs.resource_group_name`) |
| `spec.name` | `string` | yes |  |  |
| `spec.sku` | `enum` |  |  |  |
| `spec.ipConfiguration` | `AzureBastionHostIpConfiguration` |  |  |  |
| `spec.ipConfiguration.name` | `string` | yes |  |  |
| `spec.ipConfiguration.subnetId` | `string \| valueFrom` | yes |  | AzureSubnet (`status.outputs.subnet_id`) |
| `spec.ipConfiguration.publicIpAddressId` | `string \| valueFrom` |  |  | AzurePublicIp (`status.outputs.public_ip_id`) |
| `spec.virtualNetworkId` | `string \| valueFrom` |  |  | AzureVirtualNetwork (`status.outputs.virtual_network_id`) |
| `spec.scaleUnits` | `int32` |  | `2` |  |
| `spec.copyPasteEnabled` | `bool` |  | `true` |  |
| `spec.fileCopyEnabled` | `bool` |  |  |  |
| `spec.ipConnectEnabled` | `bool` |  |  |  |
| `spec.kerberosEnabled` | `bool` |  |  |  |
| `spec.shareableLinkEnabled` | `bool` |  |  |  |
| `spec.tunnelingEnabled` | `bool` |  |  |  |
| `spec.sessionRecordingEnabled` | `bool` |  |  |  |
| `spec.zones` | `[]string` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |

## Field Details

### spec.region

`string` · required

The Azure region the host lives in, e.g. "eastus". Must match the
virtual network it serves. Changing the region replaces the host.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.resourceGroup

`string | valueFrom` · required

The Azure resource group the host is created in. Can be a literal
resource-group name or a reference to an AzureResourceGroup's name
output. Changing it replaces the host.

- references: AzureResourceGroup (`status.outputs.resource_group_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureResourceGroup, name: <that resource's name>, fieldPath: status.outputs.resource_group_name}} -- a bare string does not parse

### spec.name

`string` · required

The host's name, unique within the resource group. 3-80
characters; starts with a letter or number, ends with a letter,
number, or underscore, and may contain letters, numbers,
underscores, periods, and hyphens. Changing the name replaces the
host.

- rule: Bastion host names start with a letter or number, end with a letter, number, or underscore, and may contain alphanumerics, underscores, periods, and hyphens (3-80 characters)
- rule: {"required":true,"string":{"maxLen":"80"}}

### spec.sku

`enum`

The host's SKU -- picks the deployment shape, the feature set, and
the hourly cost. Unspecified deploys BASIC (dedicated
infrastructure at fixed capacity), the provider's own default.
Upgrades are in-place; a DOWNGRADE replaces the host.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_bastion_host_sku_unspecified` -- Not specified -- deploys BASIC, the provider's default.
- `DEVELOPER` -- Shared Azure-hosted infrastructure: free, no dedicated subnet or public IP, one connection per user, no feature knobs. Offered in a limited set of regions. Dev/test only.
- `BASIC` -- Dedicated infrastructure at fixed capacity (2 scale units). Browser sessions with copy/paste; no scaling or feature knobs.
- `STANDARD` -- Dedicated infrastructure with scaling (2-50 units) and the feature set: file copy, native-client tunneling, IP connect, shareable links, Kerberos.
- `PREMIUM` -- Everything in STANDARD plus session recording and private-only deployment (omit ip_configuration's public IP).

### spec.ipConfiguration

`AzureBastionHostIpConfiguration`

The dedicated-infrastructure binding: the "AzureBastionSubnet" the
host deploys into and the Standard static public IP sessions enter
through. REQUIRED (with its public IP) on BASIC and STANDARD;
optional public IP on PREMIUM (omit it for a private-only host);
not used by DEVELOPER (shared infrastructure -- set
virtual_network_id instead). FIXED AT CREATION -- any change
replaces the host.

### spec.ipConfiguration.name

`string` · required

The configuration's name (e.g. "bastion-ip-config"), unique on the
host. Same character rules as the host name.

- rule: IP configuration names start with a letter or number, end with a letter, number, or underscore, and may contain alphanumerics, underscores, periods, and hyphens (3-80 characters)
- rule: {"required":true,"string":{"maxLen":"80"}}

### spec.ipConfiguration.subnetId

`string | valueFrom` · required

The subnet the host deploys into -- references an AzureSubnet's
ARM id. ARM requires the subnet's name to be EXACTLY
"AzureBastionSubnet" and its prefix /26 or larger; the subnet
carries nothing but the Bastion host. (The name lives on the
referenced subnet, so it is validated at deploy time, not here.)

- references: AzureSubnet (`status.outputs.subnet_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.ipConfiguration.publicIpAddressId

`string | valueFrom`

The public IP sessions enter through -- references an
AzurePublicIp's ARM id (Standard SKU, static allocation; the host
binds it exclusively). REQUIRED on BASIC and STANDARD. On PREMIUM,
omit it to deploy the host PRIVATE-ONLY (reachable only from
connected networks; surfaced in the private_only_enabled output).

- references: AzurePublicIp (`status.outputs.public_ip_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzurePublicIp, name: <that resource's name>, fieldPath: status.outputs.public_ip_id}} -- a bare string does not parse

### spec.virtualNetworkId

`string | valueFrom`

DEVELOPER SKU only: the virtual network the shared-infrastructure
host attaches to -- references an AzureVirtualNetwork's ARM id. No
dedicated subnet or public IP is involved. Required on DEVELOPER;
forbidden on every other SKU. Fixed at creation.

- references: AzureVirtualNetwork (`status.outputs.virtual_network_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureVirtualNetwork, name: <that resource's name>, fieldPath: status.outputs.virtual_network_id}} -- a bare string does not parse

### spec.scaleUnits

`int32` · optional (explicit presence)

The host's capacity, 2-50 scale units (each unit carries ~20
concurrent sessions). Adjustable only on STANDARD and PREMIUM --
BASIC and DEVELOPER are fixed at 2. Unspecified applies 2, the
provider's default. Updatable in place.

- default: `2`
- rule: {"int32":{"lte":50,"gte":2}}

### spec.copyPasteEnabled

`bool` · optional (explicit presence)

Copy/paste between the local clipboard and the remote session.
Available on every SKU. Unspecified applies true, the provider's
default (set false to block clipboard exfiltration). Updatable in
place.

- default: `true`

### spec.fileCopyEnabled

`bool`

File upload/download through the browser session. STANDARD or
PREMIUM only. Updatable in place.

### spec.ipConnectEnabled

`bool`

Connect to a VM by its private IP ADDRESS (instead of picking the
VM resource) -- reaches machines Bastion cannot enumerate, e.g.
across network peerings. STANDARD or PREMIUM only. Updatable in
place.

### spec.kerberosEnabled

`bool`

Kerberos authentication for sessions (domain-joined VM sign-in
against a domain controller reachable from the Bastion subnet).
STANDARD or PREMIUM only. The provider applies this at CREATE only
-- changing it later is silently ignored (a provider gap): plan
Kerberos up front or replace the host to change it.

### spec.shareableLinkEnabled

`bool`

Shareable links: time-boxed URLs that open a session to one
specific VM without the holder needing Azure RBAC on the host.
STANDARD or PREMIUM only. Updatable in place.

### spec.tunnelingEnabled

`bool`

Native-client tunneling: az network bastion tunnel / ssh / rdp
from a local terminal instead of the browser. STANDARD or PREMIUM
only. Updatable in place.

### spec.sessionRecordingEnabled

`bool`

Record sessions for compliance/audit review. PREMIUM only.
Updatable in place.

### spec.zones

`[]string`

Availability zones the host's instances are pinned to (e.g.
["1","2","3"] for zone redundancy where the region supports it).
Empty deploys regionally. FIXED AT CREATION -- any change replaces
the host.

- rule: {"repeated":{"items":{"string":{"minLen":"1"}}}}

### spec.tags

`map<string, string>`

Free-form tags applied to the host, merged over the
Planton-derived resource tags (organization, environment, resource
id); a user tag with the same key wins.

## Validation Rules

- `basic_standard_require_ip_configuration`: BASIC and STANDARD hosts deploy dedicated infrastructure: set ip_configuration with its subnet (named exactly AzureBastionSubnet) AND public_ip_address_id
- `developer_requires_virtual_network`: DEVELOPER hosts attach to a virtual network: set virtual_network_id (no subnet or public IP is involved)
- `virtual_network_is_developer_only`: virtual_network_id applies only to the DEVELOPER SKU -- dedicated-infrastructure hosts bind their network through ip_configuration's subnet
- `scale_units_require_standard_or_premium`: scale_units is adjustable only on STANDARD or PREMIUM -- BASIC and DEVELOPER hosts are fixed at 2
- `file_copy_requires_standard_or_premium`: file_copy_enabled requires the STANDARD or PREMIUM SKU
- `ip_connect_requires_standard_or_premium`: ip_connect_enabled requires the STANDARD or PREMIUM SKU
- `kerberos_requires_standard_or_premium`: kerberos_enabled requires the STANDARD or PREMIUM SKU
- `shareable_link_requires_standard_or_premium`: shareable_link_enabled requires the STANDARD or PREMIUM SKU
- `tunneling_requires_standard_or_premium`: tunneling_enabled requires the STANDARD or PREMIUM SKU
- `session_recording_requires_premium`: session_recording_enabled requires the PREMIUM SKU

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureBastionHost, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.bastion_host_id` | `string` | The host's ARM resource ID (.../providers/Microsoft.Network/bastionHosts/{name}). |
| `status.outputs.bastion_host_name` | `string` | The host's name within its resource group. |
| `status.outputs.dns_name` | `string` | The DNS name sessions connect through (e.g. "bst-{guid}.bastion.azure.com"). Empty until Azure assigns it. |
| `status.outputs.private_only_enabled` | `bool` | Whether the host deployed PRIVATE-ONLY (a PREMIUM host whose ip_configuration carries no public IP) -- reachable only from connected networks. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.resourceGroup` | AzureResourceGroup | `status.outputs.resource_group_name` |
| `spec.ipConfiguration.subnetId` | AzureSubnet | `status.outputs.subnet_id` |
| `spec.ipConfiguration.publicIpAddressId` | AzurePublicIp | `status.outputs.public_ip_id` |
| `spec.virtualNetworkId` | AzureVirtualNetwork | `status.outputs.virtual_network_id` |

## See Also

- [Overview](../README.md)
