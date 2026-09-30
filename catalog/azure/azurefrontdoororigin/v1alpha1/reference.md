# AzureFrontDoorOrigin

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `azure.planton.dev/v1alpha1`

**AzureFrontDoorOriginSpec** defines the configuration for creating an
origin (one backend) inside an Azure Front Door origin group: where the
backend is (host name and ports), how Front Door authenticates its TLS
certificate, how traffic is weighted against sibling origins, and --
on PREMIUM profiles -- whether Front Door reaches it over Private Link
instead of the public internet.

Origins are many-per-group with independent lifecycles: a regional
stamp adds its backend to a shared group without touching other
regions' origins, a blue/green cutover swaps origins one at a time, and
each Private Link origin carries its own connection-approval workflow.
That is why the origin is a first-class kind referencing the group
rather than a list folded into the group's spec.

**ForceNew fields**: `origin_group_id`, `origin_name` -- both fix the
origin's ARM identity at creation.

## Example

```yaml
apiVersion: azure.planton.dev/v1alpha1
kind: AzureFrontDoorOrigin
metadata:
  name: test-front-door-origin
spec:
  originGroupId:
    value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Cdn/profiles/test-frontdoor/originGroups/api-backends
  originName: primary-app
  # host_name and origin_host_header are references-or-literals; the
  # hack manifest exercises the literal (value:) form -- the E2E
  # reference scenario proves the valueFrom path.
  hostName:
    value: myapp.azurewebsites.net
  certificateNameCheckEnabled: true
  # Exercises the multi-tenant host-header seam (App Service routes by
  # Host header).
  originHostHeader:
    value: myapp.azurewebsites.net
  priority: 1
  weight: 700
  # Exercises the Private Link seam (Premium profiles; target_type
  # "sites" for App Service).
  privateLink:
    location: eastus
    privateLinkTargetId:
      value: /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/test-rg/providers/Microsoft.Web/sites/myapp
    targetType: SITES
    requestMessage: Front Door origin connection for myapp
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.originGroupId` | `string \| valueFrom` | yes |  | AzureFrontDoorOriginGroup (`status.outputs.origin_group_id`) |
| `spec.originName` | `string` | yes |  |  |
| `spec.hostName` | `string \| valueFrom` | yes |  |  |
| `spec.certificateNameCheckEnabled` | `bool` |  | `true` |  |
| `spec.originHostHeader` | `string \| valueFrom` |  |  |  |
| `spec.httpPort` | `int32` |  | `80` |  |
| `spec.httpsPort` | `int32` |  | `443` |  |
| `spec.priority` | `int32` |  | `1` |  |
| `spec.weight` | `int32` |  | `500` |  |
| `spec.enabled` | `bool` |  | `true` |  |
| `spec.privateLink` | `AzureFrontDoorOriginPrivateLink` |  |  |  |
| `spec.privateLink.location` | `string` | yes |  |  |
| `spec.privateLink.privateLinkTargetId` | `string \| valueFrom` | yes |  | AzureLinuxWebApp (`status.outputs.web_app_id`), AzureFunctionApp (`status.outputs.function_app_id`), AzureFunctionAppFlexConsumption (`status.outputs.function_app_id`), AzureStorageAccount (`status.outputs.storage_account_id`), AzureContainerAppEnvironment (`status.outputs.environment_id`), AzureApplicationGateway (`status.outputs.application_gateway_id`), AzurePrivateLinkService (`status.outputs.private_link_service_id`) |
| `spec.privateLink.targetType` | `enum` |  |  |  |
| `spec.privateLink.requestMessage` | `string` |  | `Access request for CDN FrontDoor Private Link Origin` |  |

## Field Details

### spec.originGroupId

`string | valueFrom` · required

The origin group the origin belongs to, by ARM ID. References an
AzureFrontDoorOriginGroup's origin_group_id output so the group and
its origins compose in one manifest set. Fixed at creation.

- references: AzureFrontDoorOriginGroup (`status.outputs.origin_group_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AzureFrontDoorOriginGroup, name: <that resource's name>, fieldPath: status.outputs.origin_group_id}} -- a bare string does not parse

### spec.originName

`string` · required

The origin's name -- unique within the origin group.

2-90 characters; letters, digits, and hyphens; must start and end
with a letter or digit.

**ForceNew**: changing the name replaces the origin.

- rule: origin_name must be 2-90 characters, start and end with a letter or digit, and contain only letters, digits, and hyphens
- rule: {"required":true,"string":{"minLen":"2","maxLen":"90"}}

### spec.hostName

`string | valueFrom` · required

The address Front Door connects to for content: a DNS hostname, an
IPv4, or an IPv6 address. A reference or a literal: reference the
backend's hostname output when the backend is part of the same
deployment (an AzureLinuxWebApp's default_hostname, an
AzureStorageAccount's primary_web_host), pass a literal for anything
reachable outside it ("api.example.com"). No kind dominates origin
backends, so references declare their kind explicitly.

- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.certificateNameCheckEnabled

`bool` · optional (explicit presence)

Whether Front Door validates that the origin's TLS certificate
matches the host name it connects with. Default true -- keep it on;
disabling it accepts ANY valid certificate from the origin, opening
the door to man-in-the-middle. Azure requires it to be enabled when
private_link is configured.

- default: `true`

### spec.originHostHeader

`string | valueFrom`

The Host header Front Door sends to the origin. When unset, Azure
uses the origin's own host_name -- correct for most backends.
Multi-tenant Azure services (App Service, Container Apps, Functions,
Storage static sites) route BY Host header, so for them the default
is exactly right; override only when the backend expects the
client-facing domain instead (and then make sure it can serve it).
A hostname, IPv4, or IPv6 address -- as a reference to another
resource's hostname output or a literal.

- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.httpPort

`int32` · optional (explicit presence)

The port Front Door uses when connecting to the origin over HTTP,
1-65535. Default 80.

- default: `80`
- rule: {"int32":{"lte":65535,"gte":1}}

### spec.httpsPort

`int32` · optional (explicit presence)

The port Front Door uses when connecting to the origin over HTTPS,
1-65535. Default 443.

- default: `443`
- rule: {"int32":{"lte":65535,"gte":1}}

### spec.priority

`int32` · optional (explicit presence)

The origin's failover tier within its group, 1-5 (lower serves
first). Traffic only reaches priority-2 origins when every
priority-1 origin is unhealthy -- so equal priorities load-balance,
distinct priorities express active/passive failover. Default 1.

- default: `1`
- rule: {"int32":{"lte":5,"gte":1}}

### spec.weight

`int32` · optional (explicit presence)

The origin's traffic share relative to siblings at the SAME
priority, 1-1000. Equal weights split evenly; a 950/50 split is the
classic canary. Default 500.

- default: `500`
- rule: {"int32":{"lte":1000,"gte":1}}

### spec.enabled

`bool` · optional (explicit presence)

Whether the origin receives traffic. Disabling drains it (health
probes stop, load balancing skips it) without deleting it -- the
maintenance/cutover switch. Default true.

- default: `true`

### spec.privateLink

`AzureFrontDoorOriginPrivateLink`

Reach the origin over Azure Private Link instead of the public
internet, so the backend can disable public access entirely.
PREMIUM-profile only (Azure rejects it at apply on STANDARD), and
requires certificate_name_check_enabled. After deploy, the target
resource's owner must approve the pending private-endpoint
connection (portal: the resource's Networking > Private endpoint
connections blade) before traffic flows.

- rule: target_type is required unless private_link_target_id is a Private Link Service (a /privateLinkServices/ ID or an AzurePrivateLinkService reference) -- Azure needs the sub-resource to attach to

### spec.privateLink.location

`string` · required

The Azure region of the private-link target -- must match the
target resource's own region (private-link connections are
regional even though Front Door is global). Examples: "eastus",
"westeurope".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.privateLink.privateLinkTargetId

`string | valueFrom` · required

The ARM ID of the resource Front Door connects to privately -- an
App Service site or function app, a storage account, a Container Apps
environment, an Application Gateway, or a Private Link Service
fronting an internal load balancer. A literal ARM ID, or a reference
to the target's resource ID output: the catalog kinds a target can be
are declared as candidates. Set target_type to the sub-resource the
target exposes (every target but a Private Link Service needs one).

Containment-exempt: Front Door reaches the target privately; the
origin does not live inside it.

- references: AzureLinuxWebApp (`status.outputs.web_app_id`), AzureFunctionApp (`status.outputs.function_app_id`), AzureFunctionAppFlexConsumption (`status.outputs.function_app_id`), AzureStorageAccount (`status.outputs.storage_account_id`), AzureContainerAppEnvironment (`status.outputs.environment_id`), AzureApplicationGateway (`status.outputs.application_gateway_id`), AzurePrivateLinkService (`status.outputs.private_link_service_id`)
- rule: private_link_target_id must be an ARM resource ID (starting with /subscriptions/)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: <Kind>, name: <that resource's name>, fieldPath: status.outputs.<output>}} -- a bare string does not parse

### spec.privateLink.targetType

`enum`

Which sub-resource of the target the private endpoint attaches to.
Required for every target EXCEPT a Private Link Service (whose ARM
ID is itself the attachment point). Azure rejects the apply when
this is unset for a non-PLS target.

- rule: {"enum":{"definedOnly":true}}

Allowed values (use exactly as shown):

- `azure_front_door_origin_private_link_target_type_unspecified` -- Not specified -- valid ONLY when the target is a Private Link Service (its ARM ID is the attachment point).
- `SITES` -- App Service / Function App (wire value "sites").
- `BLOB` -- Storage account blob endpoint (wire value "blob").
- `BLOB_SECONDARY` -- Storage account secondary blob endpoint (wire value "blob_secondary").
- `WEB` -- Storage static website endpoint (wire value "web").
- `WEB_SECONDARY` -- Storage static website secondary endpoint (wire value "web_secondary").
- `MANAGED_ENVIRONMENTS` -- Container Apps environment (wire value "managedEnvironments").
- `GATEWAY` -- Application Gateway (wire value "Gateway").

### spec.privateLink.requestMessage

`string` · optional (explicit presence)

The message shown to the target resource's owner on the pending
connection they approve, at most 140 characters. Default "Access
request for CDN FrontDoor Private Link Origin".

- default: `Access request for CDN FrontDoor Private Link Origin`
- rule: {"string":{"maxLen":"140"}}

## Validation Rules

- `front_door_origin_private_link_requires_cert_check`: private_link requires certificate_name_check_enabled to be true (Azure rejects the combination at apply)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AzureFrontDoorOrigin, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.origin_id` | `string` | The Azure Resource Manager ID of the origin -- what AzureFrontDoorRoute's origin_ids list references to sequence route deployment after the backends exist. Format: /subscriptions/{sub}/resourceGroups/{rg}/providers/Microsoft.Cdn/profiles/{profile}/originGroups/{group}/origins/{name} |
| `status.outputs.origin_name` | `string` | The origin's name -- unique within its origin group. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.originGroupId` | AzureFrontDoorOriginGroup | `status.outputs.origin_group_id` |
| `spec.privateLink.privateLinkTargetId` | AzureLinuxWebApp | `status.outputs.web_app_id` |
| `spec.privateLink.privateLinkTargetId` | AzureFunctionApp | `status.outputs.function_app_id` |
| `spec.privateLink.privateLinkTargetId` | AzureFunctionAppFlexConsumption | `status.outputs.function_app_id` |
| `spec.privateLink.privateLinkTargetId` | AzureStorageAccount | `status.outputs.storage_account_id` |
| `spec.privateLink.privateLinkTargetId` | AzureContainerAppEnvironment | `status.outputs.environment_id` |
| `spec.privateLink.privateLinkTargetId` | AzureApplicationGateway | `status.outputs.application_gateway_id` |
| `spec.privateLink.privateLinkTargetId` | AzurePrivateLinkService | `status.outputs.private_link_service_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AzureFrontDoorRoute | `spec.originIds` | `status.outputs.origin_id` |

## See Also

- [Overview](../README.md)
