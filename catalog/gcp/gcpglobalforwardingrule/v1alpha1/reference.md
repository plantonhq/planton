# GcpGlobalForwardingRule

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpGlobalForwardingRuleSpec defines a global Compute Engine forwarding rule
— the VIP node of a global load balancer. The forwarding rule is where
traffic enters: it binds an IP address and port to a target proxy, and
everything behind it (proxy → URL map → backend service → backends) is
wiring that decides what happens to the connection.

One frontend commonly runs a PAIR of rules sharing a single static IP: a
port-80 rule pointing at a target HTTP proxy (serving an http→https
redirect URL map) and a port-443 rule pointing at the target HTTPS proxy
that serves the application.

Beyond load balancing, the global forwarding rule is also the entry point
for Private Service Connect: with the load-balancing scheme set to NONE it
can forward a VPC's traffic privately to Google APIs (target "all-apis" /
"vpc-sc") or to a producer's published service attachment.

target and labels update in place; everything else — name, IP, protocol,
port range, scheme, network wiring — is immutable and forces
destroy-and-recreate. Because target is mutable, the standard blue/green
frontend move is to repoint the rule at a new proxy with zero VIP churn.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpGlobalForwardingRule
metadata:
  name: my-sample-forwarding-rule
spec:
  # GCP project that owns the rule.
  # Omit to use the provider's default project.
  projectId:
    value: my-gcp-project-123

  # Cloud-side name; omit to default to metadata.name.
  forwardingRuleName: web-frontend-443

  description: Port-443 VIP for the global external HTTPS load balancer

  # The target proxy receiving matched traffic (reference a
  # GcpTargetHttpsProxy / GcpTargetHttpProxy or provide a self-link; PSC
  # rules pass "all-apis", "vpc-sc", or a service attachment URI).
  target:
    value: https://www.googleapis.com/compute/v1/projects/my-gcp-project-123/global/targetHttpsProxies/web-https-frontend

  # Reserved static VIP (reference a GcpGlobalAddress or provide the IP);
  # omit for an ephemeral Google-assigned IP.
  ipAddress:
    value: 34.120.1.2

  # HTTPS on the standard port; the port-80 redirect rule is a second
  # forwarding rule sharing the same ipAddress.
  portRange: "443"

  # The classic global external Application Load Balancer family.
  loadBalancingScheme: EXTERNAL

  labels:
    env: production
    team: platform

  # Ephemeral test fixture: delete the frontend on destroy.
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.forwardingRuleName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.target` | `string \| valueFrom` | yes |  | GcpTargetHttpsProxy (`status.outputs.self_link`) |
| `spec.ipAddress` | `string \| valueFrom` |  |  | GcpGlobalAddress (`status.outputs.address`) |
| `spec.ipProtocol` | `string` |  | `TCP` |  |
| `spec.ipVersion` | `string` |  |  |  |
| `spec.loadBalancingScheme` | `string` |  | `EXTERNAL` |  |
| `spec.portRange` | `string` |  |  |  |
| `spec.network` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.subnetwork` | `string \| valueFrom` |  |  | GcpSubnetwork (`status.outputs.subnetwork_self_link`) |
| `spec.networkTier` | `string` |  |  |  |
| `spec.metadataFilters` | `[]GcpGlobalForwardingRuleMetadataFilter` |  |  |  |
| `spec.metadataFilters[].filterMatchCriteria` | `string` | yes |  |  |
| `spec.metadataFilters[].filterLabels` | `[]GcpGlobalForwardingRuleMetadataFilterLabel` | yes |  |  |
| `spec.metadataFilters[].filterLabels[].name` | `string` | yes |  |  |
| `spec.metadataFilters[].filterLabels[].value` | `string` | yes |  |  |
| `spec.serviceDirectoryRegistration` | `GcpGlobalForwardingRuleServiceDirectoryRegistration` |  |  |  |
| `spec.serviceDirectoryRegistration.namespace` | `string` |  |  |  |
| `spec.serviceDirectoryRegistration.serviceDirectoryRegion` | `string` |  |  |  |
| `spec.noAutomateDnsZone` | `bool` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.externalManagedBackendBucketMigrationState` | `string` |  |  |  |
| `spec.externalManagedBackendBucketMigrationTestingPercentage` | `double` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the forwarding rule.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.
Immutable: changing it destroys and recreates the rule.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.forwardingRuleName

`string`

Name of the forwarding rule in GCP. Must be 1-63 characters: lowercase
letters, digits, and hyphens; must start with a letter and end with a
letter or digit. Private Service Connect rules that forward to Google
APIs are stricter: 1-20 characters, lowercase letters and digits only,
starting with a letter (the name doubles as the service-directory
entry). If not specified, defaults to metadata.name.
Immutable: changing it destroys and recreates the rule — the VIP
itself survives only if it is a reserved static address.

- rule: forwarding_rule_name must be RFC1035-compliant: 1-63 lowercase letters, digits, or hyphens; must start with a letter and end with a letter or digit (Private Service Connect rules for Google APIs are limited to 20 characters, letters and digits only)

### spec.description

`string`

What this frontend serves and which proxy chain sits behind it — write
it for the operator tracing an incident from the VIP inward. Immutable.

- rule: {"string":{"maxLen":"2048"}}

### spec.target

`string | valueFrom` · required

The target that receives matched traffic. Reference a
GcpTargetHttpsProxy (the default) or a GcpTargetHttpProxy resource, or
provide a target URI directly — other global targets (target SSL/TCP
proxies, target gRPC proxies) attach by self-link until they exist as
Planton kinds. For Private Service Connect, pass the literal bundle name
"all-apis" or "vpc-sc" (Google APIs) or a service attachment URI
(producer services). Required. Mutable: GCP repoints it in place (a
dedicated setTarget call), enabling zero-downtime frontend swaps.

- references: GcpTargetHttpsProxy (`status.outputs.self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpTargetHttpsProxy, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.ipAddress

`string | valueFrom`

The IP address this rule accepts traffic on. Reference a
GcpGlobalAddress resource (its reserved IP), provide a literal IP
("34.120.1.2"), or an address resource URL. When omitted, Google Cloud
assigns an ephemeral IP — fine for testing, but production frontends
should reserve a static address so DNS never has to chase a new VIP.
Required for Private Service Connect rules. Immutable.

- references: GcpGlobalAddress (`status.outputs.address`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGlobalAddress, name: <that resource's name>, fieldPath: status.outputs.address}} -- a bare string does not parse

### spec.ipProtocol

`string` · optional (explicit presence)

The IP protocol this rule matches (default TCP). All proxy-based global
load balancers and Private Service Connect use TCP; the other protocols
exist for protocol forwarding. Immutable.

- default: `TCP`
- rule: ip_protocol must be one of TCP, UDP, ESP, AH, SCTP, or ICMP

### spec.ipVersion

`string`

IP version for the auto-assigned ephemeral address (IPV4 or IPV6; GCP
default IPV4). Only meaningful when ip_address is omitted — a referenced
static address already fixes the version. Immutable.

- rule: ip_version must be IPV4 or IPV6

### spec.loadBalancingScheme

`string` · optional (explicit presence)

Which load balancer family this frontend belongs to (default EXTERNAL,
the classic global external Application LB). EXTERNAL_MANAGED is the
newer envoy-based global external ALB; INTERNAL_MANAGED is the
cross-region internal ALB; INTERNAL_SELF_MANAGED is Traffic Director /
service mesh; NONE (sent to GCP as an empty scheme) is Private Service
Connect. The scheme must match the family the target proxy's backend
services were created for. Immutable — except the EXTERNAL →
EXTERNAL_MANAGED canary migration driven by
external_managed_backend_bucket_migration_state.

- default: `EXTERNAL`
- rule: load_balancing_scheme must be one of EXTERNAL, EXTERNAL_MANAGED, INTERNAL_MANAGED, INTERNAL_SELF_MANAGED, or NONE (NONE is the Private Service Connect form)

### spec.portRange

`string`

The port or contiguous port range ("443" or "8080-8090") this rule
matches. Requires a TCP/UDP/SCTP protocol. Proxy-based global load
balancers accept only specific ports (80/8080/443 for HTTP(S)); two
external rules on the same IP+protocol cannot overlap ranges — which is
exactly how the port-80 redirect rule and the port-443 serving rule
share one VIP. Not used by Private Service Connect rules. Immutable.

- rule: port_range must be a port ("443") or a contiguous range ("8080-8090")

### spec.network

`string | valueFrom`

The VPC network this frontend belongs to. Only used by internal-facing
schemes and Private Service Connect (INTERNAL_MANAGED,
INTERNAL_SELF_MANAGED, NONE); external load balancers live on Google's
edge, not in a VPC. Reference a GcpVpcNetwork resource or provide a network
self-link. For PSC rules a network is required. If omitted where
applicable, GCP uses the default network. Immutable.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.subnetwork

`string | valueFrom`

The subnetwork the load-balanced IP belongs to, for internal load
balancing. Optional when the network is auto-mode; required when it is
custom-mode. Reference a GcpSubnetwork resource or provide a subnetwork
self-link. Immutable.

- references: GcpSubnetwork (`status.outputs.subnetwork_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_self_link}} -- a bare string does not parse

### spec.networkTier

`string`

Networking tier. Global forwarding rules only support PREMIUM (Google's
global backbone); STANDARD tier exists only on regional forwarding
rules. Empty means PREMIUM. If ip_address references a reserved
address, the tiers must match. Immutable.

- rule: global forwarding rules only support the PREMIUM network tier — STANDARD tier load balancing is a regional forwarding rule feature

### spec.metadataFilters

`[]GcpGlobalForwardingRuleMetadataFilter`

Traffic Director metadata filters: restrict which xDS clients receive
this forwarding rule's configuration, by matching labels the clients
present in their node metadata. Only applies to INTERNAL_SELF_MANAGED
frontends. Filters set here can be overridden by the URL map's own
metadata filters. Immutable.

### spec.metadataFilters[].filterMatchCriteria

`string` · required

How the labels combine: MATCH_ALL (every label must match) or MATCH_ANY
(at least one).

- rule: filter_match_criteria must be MATCH_ALL or MATCH_ANY
- rule: {"required":true}

### spec.metadataFilters[].filterLabels

`[]GcpGlobalForwardingRuleMetadataFilterLabel` · required

The xDS node metadata labels to match against (1-64 entries).

- rule: {"repeated":{"minItems":"1","maxItems":"64"}}

### spec.metadataFilters[].filterLabels[].name

`string` · required

Label name (1-1024 characters).

- rule: {"required":true,"string":{"minLen":"1","maxLen":"1024"}}

### spec.metadataFilters[].filterLabels[].value

`string` · required

The value the label must match (up to 1024 characters).

- rule: {"required":true,"string":{"minLen":"1","maxLen":"1024"}}

### spec.serviceDirectoryRegistration

`GcpGlobalForwardingRuleServiceDirectoryRegistration`

Register this Private Service Connect frontend in Service Directory so
VPC workloads can discover the private Google-APIs endpoint by name.
Only used by PSC-for-Google-APIs rules (scheme NONE with an "all-apis" /
"vpc-sc" target). Immutable.

### spec.serviceDirectoryRegistration.namespace

`string`

Service Directory namespace to register the forwarding rule under. If
omitted, GCP registers it under a Google-managed namespace.

- rule: {"string":{"maxLen":"255"}}

### spec.serviceDirectoryRegistration.serviceDirectoryRegion

`string`

Service Directory region to register this global rule under (GCP
default us-central1). All PSC-for-Google-APIs rules on one network
should use the same region.

- rule: {"string":{"maxLen":"63"}}

### spec.noAutomateDnsZone

`bool`

Skip the DNS zone Google normally auto-creates for a Private Service
Connect Google-APIs frontend (the zone that maps googleapis.com names to
the private VIP). Set true when you manage that DNS yourself. Only
meaningful for PSC rules (scheme NONE). Immutable.

### spec.labels

`map<string, string>`

Labels to organize and bill this forwarding rule (e.g. env, team,
cost-center). Keys and values follow GCP label rules. Mutable.

### spec.externalManagedBackendBucketMigrationState

`string`

Canary state for migrating this frontend's backend BUCKETS from the
classic EXTERNAL scheme to EXTERNAL_MANAGED without recreating the VIP:
PREPARE stages the migration, TEST_BY_PERCENTAGE shifts the fraction set
in external_managed_backend_bucket_migration_testing_percentage, and
TEST_ALL_TRAFFIC must be reached before flipping load_balancing_scheme
to EXTERNAL_MANAGED. Roll back by walking the states in reverse.
Mutable.

- rule: external_managed_backend_bucket_migration_state must be one of PREPARE, TEST_BY_PERCENTAGE, or TEST_ALL_TRAFFIC

### spec.externalManagedBackendBucketMigrationTestingPercentage

`double`

Percentage (0-100) of backend-bucket requests served by the envoy-based
Global external ALB during a TEST_BY_PERCENTAGE canary migration. Only
meaningful with external_managed_backend_bucket_migration_state
TEST_BY_PERCENTAGE. Mutable.

- rule: {"double":{"lte":100,"gte":0}}

### spec.deletionPolicy

`string`

Deletion policy for the global forwarding rule — what happens on
destroy:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the frontend is deleted; the reserved IP it served stays
               reserved (its own kind's policy governs it), but traffic
               to that IP stops routing the moment the rule is gone
  "PREVENT" -- destroy FAILS; protects the entry point of a live load
               balancer chain
  "ABANDON" -- the rule is removed from management but keeps serving
               traffic in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `network_requires_internal_or_psc_scheme`: network only applies to internal or Private Service Connect frontends — external load balancers live on Google's edge, not in a VPC; set load_balancing_scheme to INTERNAL_MANAGED, INTERNAL_SELF_MANAGED, or NONE, or remove network
- `metadata_filters_require_traffic_director`: metadata_filters only apply to Traffic Director frontends — set load_balancing_scheme INTERNAL_SELF_MANAGED or remove them
- `service_directory_requires_psc`: service_directory_registration only applies to Private Service Connect frontends — set load_balancing_scheme NONE or remove it
- `no_automate_dns_zone_requires_psc`: no_automate_dns_zone only applies to Private Service Connect frontends — set load_balancing_scheme NONE or remove it
- `migration_percentage_requires_test_by_percentage`: external_managed_backend_bucket_migration_testing_percentage only applies while external_managed_backend_bucket_migration_state is TEST_BY_PERCENTAGE

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpGlobalForwardingRule, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.ip_address` | `string` | The IP address this frontend accepts traffic on — the load balancer's VIP, and the value DNS records point at. The API always reports the literal IP number here, even when the spec referenced an address resource. |
| `status.outputs.self_link` | `string` | Self-link URI of the forwarding rule. Format: https://www.googleapis.com/compute/v1/projects/{project}/global/forwardingRules/{name} |
| `status.outputs.forwarding_rule_name` | `string` | Name of the forwarding rule as it exists in GCP. |
| `status.outputs.forwarding_rule_id` | `string` | Server-assigned numeric ID of the forwarding rule. |
| `status.outputs.psc_connection_id` | `string` | The Private Service Connect connection id, populated only for PSC frontends (load_balancing_scheme NONE). |
| `status.outputs.psc_connection_status` | `string` | The Private Service Connect connection status (PENDING, ACCEPTED, REJECTED, or CLOSED), populated only for PSC frontends. ACCEPTED means the producer side admitted this consumer connection. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.target` | GcpTargetHttpsProxy | `status.outputs.self_link` |
| `spec.ipAddress` | GcpGlobalAddress | `status.outputs.address` |
| `spec.network` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.subnetwork` | GcpSubnetwork | `status.outputs.subnetwork_self_link` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpNetworkFirewallPolicy | `spec.rules[].targetForwardingRules` | `status.outputs.self_link` |

## See Also

- [Overview](../README.md)
