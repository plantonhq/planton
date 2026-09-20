# GcpGlobalForwardingRule

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpGlobalForwardingRuleSpec defines a Compute Engine forwarding rule — the
VIP node of a load balancer. The forwarding rule is where traffic enters:
it binds an IP address and port to a target (a proxy, or for passthrough
load balancers a backend service directly), and everything behind it
(proxy → URL map → backend service → backends) is wiring that decides what
happens to the connection.

One kind, two scopes. The kind is named for the GLOBAL forwarding rule it
began as; with region empty it builds exactly that (the global external
ALB, the cross-region internal ALB, Traffic Director, PSC to Google APIs).
With region set it builds the REGIONAL forwarding rule — the front door of
the regional external and internal Application Load Balancers (target = a
regional proxy), of the internal and external passthrough Network Load
Balancers (backend_service = a regional backend service, no proxy at
all), and of a Private Service Connect consumer endpoint (target = a
producer's service attachment). Everything the rule points at must live in
the same scope, and for a regional rule in the same region. A rule cannot
move between scopes.

One frontend commonly runs a PAIR of rules sharing a single static IP: a
port-80 rule pointing at a target HTTP proxy (serving an http→https
redirect URL map) and a port-443 rule pointing at the target HTTPS proxy
that serves the application.

Beyond load balancing, the forwarding rule is also the entry point for
Private Service Connect: with the load-balancing scheme set to NONE it can
forward a VPC's traffic privately to Google APIs (a global rule with target
"all-apis" / "vpc-sc") or to a producer's published service attachment (a
regional rule whose target is the attachment).

target, labels, and allow_global_access update in place; everything else —
name, IP, protocol, ports, scheme, network wiring, region — is immutable
and forces destroy-and-recreate. Because target is mutable, the standard
blue/green frontend move is to repoint the rule at a new proxy with zero
VIP churn.

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
| `spec.region` | `string` |  |  |  |
| `spec.target` | `string \| valueFrom` |  |  | GcpTargetHttpsProxy (`status.outputs.self_link`) |
| `spec.backendService` | `string \| valueFrom` |  |  | GcpBackendService (`status.outputs.self_link`) |
| `spec.ipAddress` | `string \| valueFrom` |  |  | GcpGlobalAddress (`status.outputs.address`) |
| `spec.ipProtocol` | `string` |  | `TCP` |  |
| `spec.ipVersion` | `string` |  |  |  |
| `spec.loadBalancingScheme` | `string` |  | `EXTERNAL` |  |
| `spec.portRange` | `string` |  |  |  |
| `spec.ports` | `[]string` |  |  |  |
| `spec.allPorts` | `bool` |  |  |  |
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
| `spec.serviceDirectoryRegistration.service` | `string` |  |  |  |
| `spec.allowGlobalAccess` | `bool` |  |  |  |
| `spec.allowPscGlobalAccess` | `bool` |  |  |  |
| `spec.serviceLabel` | `string` |  |  |  |
| `spec.isMirroringCollector` | `bool` |  |  |  |
| `spec.ipCollection` | `string` |  |  |  |
| `spec.recreateClosedPsc` | `bool` |  |  |  |
| `spec.sourceIpRanges` | `[]string` |  |  |  |
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

### spec.region

`string`

The scope selector. Empty builds a GLOBAL forwarding rule (the global
external ALB, the cross-region internal ALB, Traffic Director, PSC to
Google APIs); a region name such as us-central1 builds a REGIONAL one —
the front door of the regional external and internal Application Load
Balancers (target = a regional proxy), of the internal and external
passthrough Network Load Balancers (backend_service instead of target),
and of a Private Service Connect consumer endpoint (target = a service
attachment). Everything the rule points at must be in the same scope and
region: a regional proxy, a regional backend service, a regional
address (GcpAddress) rather than a GcpGlobalAddress. The regional-only
levers (backend_service, ports, all_ports, allow_global_access,
allow_psc_global_access, service_label, is_mirroring_collector,
ip_collection, recreate_closed_psc, source_ip_ranges, the L3_DEFAULT
protocol, the INTERNAL scheme, the STANDARD tier) are rejected when
region is empty; the global-only levers (metadata_filters, the
INTERNAL_SELF_MANAGED scheme, the backend-bucket migration canary, the
Service Directory region) are rejected when it is set. Immutable: a
rule cannot move between scopes or regions.

- rule: region must be a valid GCP region name such as us-central1, or empty for a global forwarding rule

### spec.target

`string | valueFrom`

The target that receives matched traffic — every proxy-based load
balancer's form. Reference a GcpTargetHttpsProxy (the default) or a
GcpTargetHttpProxy resource (a regional rule takes the regional arm of
the same kinds, in its own region), or provide a target URI directly —
other targets (target SSL/TCP proxies, target gRPC proxies, target
instances, target pools) attach by self-link until they exist as
Planton kinds. For Private Service Connect, pass the literal bundle name
"all-apis" or "vpc-sc" (Google APIs, global rule) or a service
attachment URI (a producer's service, regional rule). Exactly one of
target and backend_service is set: a passthrough Network Load Balancer
has no proxy and names its backend service instead. Mutable: GCP
repoints it in place (a dedicated setTarget call), enabling
zero-downtime frontend swaps.

- references: GcpTargetHttpsProxy (`status.outputs.self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpTargetHttpsProxy, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.backendService

`string | valueFrom`

The regional backend service that receives matched traffic directly,
with no proxy in between — the form of the internal passthrough Network
Load Balancer (scheme INTERNAL) and of the backend-service-based
external passthrough Network Load Balancer (scheme EXTERNAL). Reference
a GcpBackendService declared with the same region, or provide its
self-link. Regional rules only, and exactly one of backend_service and
target: Google requires the backend service for the passthrough load
balancers and rejects it for every proxy-based one. Immutable.

- references: GcpBackendService (`status.outputs.self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpBackendService, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.ipAddress

`string | valueFrom`

The IP address this rule accepts traffic on. Reference a
GcpGlobalAddress resource (the default kind, for a global rule) or a
regional GcpAddress resource in the rule's region (for a regional rule,
with valueFrom.kind: GcpAddress), provide a literal IP ("34.120.1.2"),
or an address resource URL. When omitted, Google Cloud assigns an
ephemeral IP — fine for testing, but production frontends should
reserve a static address so DNS never has to chase a new VIP. Required
for Private Service Connect rules to Google APIs. Immutable.

- references: GcpGlobalAddress (`status.outputs.address`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGlobalAddress, name: <that resource's name>, fieldPath: status.outputs.address}} -- a bare string does not parse

### spec.ipProtocol

`string` · optional (explicit presence)

The IP protocol this rule matches (default TCP). All proxy-based load
balancers and Private Service Connect use TCP; UDP, ESP, AH, SCTP, and
ICMP exist for passthrough load balancing and protocol forwarding.
L3_DEFAULT — regional rules only — forwards every IP protocol at once
(the passthrough Network Load Balancer's multi-protocol form); it
requires all_ports and a backend service whose protocol is UNSPECIFIED.
Immutable.

- default: `TCP`
- rule: ip_protocol must be one of TCP, UDP, ESP, AH, SCTP, ICMP, or L3_DEFAULT (L3_DEFAULT is a regional-rule protocol)

### spec.ipVersion

`string`

IP version for the auto-assigned ephemeral address (IPV4 or IPV6; GCP
default IPV4). Only meaningful when ip_address is omitted — a referenced
static address already fixes the version. Immutable.

- rule: ip_version must be IPV4 or IPV6

### spec.loadBalancingScheme

`string` · optional (explicit presence)

Which load balancer family this frontend belongs to (default EXTERNAL
on both scopes: the classic global external ALB, or the external
passthrough Network Load Balancer on a regional rule). EXTERNAL_MANAGED
is the envoy-based external ALB (global, or regional with region set);
INTERNAL_MANAGED is the internal ALB (cross-region on a global rule,
regional with region set); INTERNAL — regional rules only — is the
internal passthrough Network Load Balancer, which names a
backend_service instead of a target; INTERNAL_SELF_MANAGED — global
rules only — is Traffic Director / service mesh; NONE (sent to GCP as
an empty scheme) is Private Service Connect: to Google APIs on a global
rule, to a producer's service attachment on a regional one. The scheme
must match the family the target's backend services were created for.
Both engines send EXTERNAL explicitly when this is left empty, on both
scopes, so an unset scheme means the same thing wherever the rule
lives. Immutable — except the EXTERNAL → EXTERNAL_MANAGED canary
migration driven by external_managed_backend_bucket_migration_state.

- default: `EXTERNAL`
- rule: load_balancing_scheme must be one of EXTERNAL, EXTERNAL_MANAGED, INTERNAL, INTERNAL_MANAGED, INTERNAL_SELF_MANAGED, or NONE (NONE is the Private Service Connect form; INTERNAL is regional-only, INTERNAL_SELF_MANAGED global-only)

### spec.portRange

`string`

The port or contiguous port range ("443" or "8080-8090") this rule
matches. Requires a TCP/UDP/SCTP protocol. Proxy-based load balancers
accept only specific ports (80/8080/443 for HTTP(S)); two external
rules on the same IP+protocol cannot overlap ranges — which is exactly
how the port-80 redirect rule and the port-443 serving rule share one
VIP. Not used by Private Service Connect rules. On a regional rule, at
most one of port_range, ports, and all_ports is set. Immutable.

- rule: port_range must be a port ("443") or a contiguous range ("8080-8090")

### spec.ports

`[]string`

Up to five individual ports or ranges ("80", "443", "8080-8090") this
rule matches — the passthrough Network Load Balancer's form (internal
passthrough, backend-service-based external passthrough, internal
protocol forwarding), where the ports need not be contiguous. Requires
a TCP, UDP, or SCTP protocol; regional rules only; mutually exclusive
with port_range and all_ports. Immutable.

- rule: {"repeated":{"maxItems":"5","items":{"cel":[{"id":"valid_port_entry","message":"each entry in ports must be a port (\"443\") or a contiguous range (\"8080-8090\")","expression":"this.matches('^[0-9]+(-[0-9]+)?$')"}]}}}

### spec.allPorts

`bool`

Forward packets addressed to ANY port — and packets lacking a
destination port, such as UDP fragments after the first — to the
backends. The passthrough Network Load Balancer's form for services
that listen on many ports or for protocol forwarding; required by the
L3_DEFAULT protocol. Requires TCP, UDP, SCTP, or L3_DEFAULT; regional
rules only; mutually exclusive with port_range and ports. Immutable.

### spec.network

`string | valueFrom`

The VPC network this frontend belongs to. Reference a GcpVpcNetwork
resource or provide a network self-link. Used by the internal-facing
schemes (INTERNAL, INTERNAL_MANAGED, INTERNAL_SELF_MANAGED), by Private
Service Connect (NONE — required, on both scopes), and by the REGIONAL
external ALB (EXTERNAL_MANAGED with region set, whose proxy-only subnet
lives in this network); the global external load balancers and the
external passthrough Network Load Balancer live on Google's edge and
reject it. If omitted where applicable, GCP uses the default network.
Immutable.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.subnetwork

`string | valueFrom`

The subnetwork the load-balanced IP belongs to, for internal load
balancing and for IPv6 external passthrough Network Load Balancers.
Optional when the network is auto-mode; required when it is custom-mode
(and for an IPv6 external passthrough rule). Reference a GcpSubnetwork
resource or provide a subnetwork self-link. Immutable.

- references: GcpSubnetwork (`status.outputs.subnetwork_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_self_link}} -- a bare string does not parse

### spec.networkTier

`string`

Networking tier. PREMIUM (Google's global backbone; the default when
empty) on both scopes; STANDARD (regional ISP transit, cheaper egress)
only on a regional rule — a global rule is PREMIUM by definition. If
ip_address references a reserved address, the tiers must match.
Immutable.

- rule: network_tier must be PREMIUM or STANDARD (STANDARD is a regional-rule tier)

### spec.metadataFilters

`[]GcpGlobalForwardingRuleMetadataFilter`

Traffic Director metadata filters: restrict which xDS clients receive
this forwarding rule's configuration, by matching labels the clients
present in their node metadata. Only applies to INTERNAL_SELF_MANAGED
frontends, which are global rules. Filters set here can be overridden
by the URL map's own metadata filters. Immutable.

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
VPC workloads can discover the private endpoint by name. Used by PSC
rules (scheme NONE): a Google-APIs bundle rule names the registration
region, a regional consumer-endpoint rule names the service. Immutable.

### spec.serviceDirectoryRegistration.namespace

`string`

Service Directory namespace to register the forwarding rule under. If
omitted, GCP registers it under a Google-managed namespace.

- rule: {"string":{"maxLen":"255"}}

### spec.serviceDirectoryRegistration.serviceDirectoryRegion

`string`

Service Directory region to register this global rule under (GCP
default us-central1). All PSC-for-Google-APIs rules on one network
should use the same region. Global rules only — a regional rule is
registered in its own region.

- rule: {"string":{"maxLen":"63"}}

### spec.serviceDirectoryRegistration.service

`string`

Service Directory service to register this regional consumer endpoint
under, inside the namespace. Regional rules only.

- rule: {"string":{"maxLen":"63"}}

### spec.allowGlobalAccess

`bool`

Let clients in EVERY region reach this internal load balancer, instead
of only clients in the rule's own region (Google's default). For the
internal passthrough Network Load Balancer (scheme INTERNAL) and for
internal target-instance forwarding. Regional rules only. Mutable.

### spec.allowPscGlobalAccess

`bool`

Let clients in every region reach this Private Service Connect consumer
endpoint (a regional rule with scheme NONE whose target is a service
attachment), instead of only clients in the endpoint's region. Regional
rules only. Mutable.

### spec.serviceLabel

`string`

A DNS label (RFC 1035, 1-63 characters) prepended to the internal
passthrough Network Load Balancer's service name, giving the VIP a
stable internal name like
<service_label>.<name>.il4.<region>.lb.<project>.internal (exposed as
the service_name output). INTERNAL scheme, regional rules only.
Immutable.

- rule: service_label must be RFC1035-compliant: 1-63 lowercase letters, digits, or hyphens; must start with a letter and end with a letter or digit

### spec.isMirroringCollector

`bool`

Mark this internal passthrough Network Load Balancer as a Packet
Mirroring collector: mirrored traffic is delivered to its backends, and
to prevent mirroring loops those backends are never mirrored themselves
even when a PacketMirroring rule applies to them. INTERNAL scheme,
regional rules only. Immutable.

### spec.ipCollection

`string`

Bring your own IPv6 range: the PublicDelegatedPrefix (a sub-PDP in
EXTERNAL_IPV6_FORWARDING_RULE_CREATION mode) the external passthrough
Network Load Balancer's IPv6 address is drawn from, as a resource URL or
partial path (projects/{p}/regions/{r}/publicDelegatedPrefixes/{name}).
Regional rules only. Immutable.

- rule: {"string":{"maxLen":"1024"}}

### spec.recreateClosedPsc

`bool`

Recreate this Private Service Connect consumer endpoint when Google
reports its connection CLOSED (the producer removed or rejected it) —
the engines otherwise leave a closed endpoint in place until it is
changed by hand. Default false. Regional PSC rules only. Mutable.

### spec.sourceIpRanges

`[]string`

Forward only traffic whose SOURCE address matches one of these IP
addresses ("1.2.3.4") or CIDR ranges ("1.2.3.0/24"), up to 64 — a
coarse allowlist at the VIP for the external passthrough Network Load
Balancer. Regional rules with scheme EXTERNAL only. Immutable.

- rule: {"repeated":{"maxItems":"64","items":{"cel":[{"id":"valid_source_ip_range","message":"each entry in source_ip_ranges must be an IP address (1.2.3.4) or a CIDR range (1.2.3.0/24)","expression":"this.isIp() || this.isIpPrefix()"}]}}}

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

- `exactly_one_of_target_or_backend_service`: set exactly one of target (proxy-based load balancers and Private Service Connect) or backend_service (internal and external passthrough Network Load Balancers, regional rules only)
- `backend_service_regional_only`: backend_service is the passthrough Network Load Balancer's form and exists only on a regional forwarding rule — set region, or point a global rule at a target proxy
- `ports_regional_only`: ports, all_ports, source_ip_ranges, service_label, allow_global_access, allow_psc_global_access, is_mirroring_collector, ip_collection, and recreate_closed_psc exist only on a regional forwarding rule — set region or remove them (a global rule uses port_range)
- `ports_port_range_all_ports_exclusive`: port_range, ports, and all_ports are mutually exclusive — a rule matches one contiguous range, up to five individual ports or ranges, or every port
- `l3_default_regional_only_and_all_ports`: the L3_DEFAULT protocol exists only on a regional forwarding rule and forwards every port — set region and all_ports (Google rejects L3_DEFAULT with port_range or ports)
- `internal_scheme_regional_only`: the INTERNAL scheme (the internal passthrough Network Load Balancer) exists only on a regional forwarding rule — set region, or use INTERNAL_MANAGED for the cross-region internal Application Load Balancer
- `standard_tier_regional_only`: the STANDARD network tier exists only on a regional forwarding rule — a global rule rides Google's PREMIUM backbone by definition; set region or clear network_tier
- `self_managed_scheme_global_only`: the INTERNAL_SELF_MANAGED scheme (Traffic Director) exists only on a global forwarding rule — clear region
- `backend_bucket_migration_global_only`: external_managed_backend_bucket_migration_state and its testing percentage belong to the global external Application Load Balancer — a regional rule has no backend buckets; clear region or remove them
- `service_directory_arm_fields`: service_directory_registration.service_directory_region exists only on a global rule and service_directory_registration.service only on a regional rule — use the field for the rule's scope
- `network_requires_internal_psc_or_regional_managed_scheme`: network applies to internal and Private Service Connect frontends and to the regional external Application Load Balancer (EXTERNAL_MANAGED with region set) — the global external load balancers and the external passthrough Network Load Balancer live on Google's edge, not in a VPC; change load_balancing_scheme or remove network
- `metadata_filters_require_traffic_director`: metadata_filters only apply to Traffic Director frontends — set load_balancing_scheme INTERNAL_SELF_MANAGED (a global rule) or remove them
- `service_directory_requires_psc`: service_directory_registration only applies to Private Service Connect frontends — set load_balancing_scheme NONE or remove it
- `internal_only_levers_require_internal_scheme`: allow_global_access, service_label, and is_mirroring_collector belong to the internal passthrough Network Load Balancer — set load_balancing_scheme INTERNAL or remove them
- `psc_consumer_levers_require_psc_scheme`: allow_psc_global_access and recreate_closed_psc belong to a Private Service Connect consumer endpoint — set load_balancing_scheme NONE or remove them
- `source_ip_ranges_require_external_scheme`: source_ip_ranges filters the external passthrough Network Load Balancer only — set load_balancing_scheme EXTERNAL (with region) or remove it
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
| `status.outputs.region` | `string` | Region of a regional forwarding rule; empty for a global one. Downstream blocks read it to confirm scope compatibility (a regional link must point at a regional target in the same region), and the E2E verifier picks the regional or global API by it. |
| `status.outputs.service_name` | `string` | The internal DNS name of an internal passthrough Network Load Balancer that set service_label, in the form <service_label>.<name>.il4.<region>.lb.<project>.internal — the stable name VPC clients resolve instead of the VIP. Empty otherwise. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.target` | GcpTargetHttpsProxy | `status.outputs.self_link` |
| `spec.backendService` | GcpBackendService | `status.outputs.self_link` |
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
