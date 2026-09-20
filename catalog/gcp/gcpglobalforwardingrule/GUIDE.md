# GcpGlobalForwardingRule Guide

The judgment this guide protects: the forwarding rule is the entry point
of the whole load-balancer chain — the piece that binds a reserved IP, a
port, and a target proxy into "traffic flows". It is also the piece whose
recreation is instantly user-visible.

## Bring your own IP, always

Let the rule reference a GcpGlobalAddress instead of auto-allocating.
An auto-allocated VIP dies with the rule; a referenced reservation
survives it, which turns "recreate the forwarding rule" from a DNS
incident into a blip. The rule's own recreation only breaks the binding
— with a stable IP the new rule picks up exactly where the old one
stopped.

## One rule per port contract

A production HTTPS frontend is typically TWO rules on one IP: 443 to the
HTTPS proxy, 80 to the HTTP proxy whose URL map redirects. The spec's
`portRange` accepts a single port or range, but resist wide ranges on
EXTERNAL schemes — every open port is attack surface, and GCP's
port-range semantics differ between classic and envoy-based schemes.

## Scheme is destiny

`loadBalancingScheme` decides which targets are legal, whether a network
is required, and which extras (metadata filters, Service Directory) even
apply. The spec's `NONE` sentinel maps to the API's empty scheme for PSC
frontends — the module translates it, so never write the empty string
yourself. The Service Directory registration is PSC-only; the spec
enforces the pairing pre-deploy.

## One kind, two scopes

The kind is named for the GLOBAL forwarding rule it began as; `region`
empty still builds exactly that — the global external ALB, the
cross-region internal ALB, Traffic Director, PSC to Google APIs. `region`
set builds the REGIONAL rule, which is three products in one resource:
the front door of the regional external and internal Application Load
Balancers (`target` = a regional proxy), of the internal and external
passthrough Network Load Balancers (`backendService` = a regional backend
service, no proxy at all — set the scheme to `INTERNAL` or `EXTERNAL`
outright), and of a Private Service Connect consumer endpoint (`target` =
a producer's service attachment, scheme `NONE`). Everything the rule
points at must be regional in the same region: a regional proxy, a
regional backend service, a regional `GcpAddress` (attached with an
explicit `valueFrom.kind`, since `ipAddress` defaults to the global
address kind). A regional external ALB also needs a proxy-only subnet
(`GcpSubnetwork` with `purpose: REGIONAL_MANAGED_PROXY`) in the region
before the rule can be created, and takes `network` — the one external
scheme that does. `region` is immutable.

## Source filtering is a regional EXTERNAL lever

`sourceIpRanges` (up to 64 IPs or CIDRs) forwards only traffic from those
sources — the external passthrough Network Load Balancer's coarse
allowlist at the VIP. Google honors it only on a REGIONAL rule with
scheme `EXTERNAL`, and the spec admits it there alone. Needing source
filtering on a global frontend means Cloud Armor, not the forwarding
rule.

## Ports come in three shapes

`portRange` (one contiguous range) is the proxy-based load balancers'
form on both scopes. `ports` (up to five individual ports or ranges) and
`allPorts` (every port, plus port-less packets such as UDP fragments) are
the passthrough Network Load Balancers' forms and exist only on a
regional rule; `L3_DEFAULT` — forward every IP protocol at once — requires
`allPorts` and a backend service whose protocol is `UNSPECIFIED`. The
three are mutually exclusive and the spec rejects two at once.

## Teardown discipline

Deleting the rule stops traffic to that IP immediately — the proxies and
backends behind it stay healthy but unreachable on that VIP. `PREVENT`
suits any rule fronting production; `ABANDON` keeps traffic flowing while
dropping management (the escape hatch for handing a frontend to another
stack). The reserved address it referenced is governed by its OWN kind's
deletion policy — destroying the rule never releases the IP.
