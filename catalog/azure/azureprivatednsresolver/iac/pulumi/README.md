# AzurePrivateDnsResolver Pulumi Module

## Overview

Creates an Azure DNS Private Resolver -- the managed DNS proxy that resolves names across the hybrid boundary -- and its inbound/outbound endpoints, on the classic Pulumi Azure SDK (`pulumi-azure/sdk/v6`), wire-identical to the Terraform module.

## Resources Created

- `privatedns.Resolver` -- the resolver, anchored to one virtual network
- `privatedns.ResolverInboundEndpoint` -- one per `spec.inbound_endpoints` entry
- `privatedns.ResolverOutboundEndpoint` -- one per `spec.outbound_endpoints` entry

## Outputs

- `dns_resolver_id` -- the resolver's ARM resource ID
- `dns_resolver_name` -- the resolver's name
- `inbound_endpoint_ip` -- the FIRST declared inbound endpoint's private IP (what on-premises forwarders point at); empty when no inbound endpoints are declared
- `inbound_endpoint_ips` -- ALL inbound endpoint IPs, keyed by endpoint name
- `outbound_endpoint_id` -- the FIRST declared outbound endpoint's ARM id (what forwarding rulesets reference by default); empty when no outbound endpoints are declared
- `outbound_endpoint_ids` -- ALL outbound endpoint ids, keyed by endpoint name

## Behavior Notes

- **One resolver per virtual network** -- Azure enforces it at deploy time.
- **Every endpoint needs its own dedicated subnet**, delegated to `Microsoft.Network/dnsResolvers`, sized /28 to /24, hosting nothing else -- ARM validates at deploy time.
- **Everything except tags is create-only**; endpoint edits replace that endpoint (Pulumi resource names key on the endpoint name, so siblings are untouched).
- **The inbound Static/Dynamic contract is spec-validated** (STATIC requires an address; DYNAMIC forbids one) -- mirroring the provider's own pre-flight check.
- **Endpoint deletes outlive the API's first answer** -- the provider polls until the endpoint is verifiably gone.
- **Endpoints bill hourly**; the resolver object itself is free.

## Engine Parity

ZERO parity exceptions: the classic SDK v6.38.0 carries the complete v5 argument surface for all three resources, verified field-by-field at planning. The module always sends the effective allocation method ("Dynamic" when unset) so both engines produce identical wire shapes.
