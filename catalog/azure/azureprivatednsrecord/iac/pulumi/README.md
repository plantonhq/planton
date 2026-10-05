# AzurePrivateDnsRecord Pulumi Module

## Overview

Creates one DNS record set (A, AAAA, CNAME, MX, PTR, SRV, or TXT) in an Azure PRIVATE DNS zone -- name resolution visible only inside the virtual networks linked to the zone. The record type is whichever typed payload the spec carries (validation guarantees exactly one), so exactly one branch of the module runs.

## Resources Created

Exactly one of the SDK's private DNS record resources: `privatedns.ARecord`, `privatedns.AAAARecord`, `privatedns.CnameRecord`, `privatedns.MxRecord`, `privatedns.PTRRecord`, `privatedns.SRVRecord`, or `privatedns.TxtRecord`.

## Outputs

- `record_id` -- the record set's ARM resource ID
- `fqdn` -- the record set's fully qualified name with a trailing dot, resolvable only from networks linked to the zone

## Behavior Notes

- **Zone addressing differs across engines by SDK shape, not by result**: the spec carries the zone's ARM id (`private_dns_zone_id`); this SDK's record resources address the zone by (resource group, zone name), so the module splits the resolved id into its segments. The Terraform provider passes the id directly. Both write the same ARM object.
- **Private DNS has no alias records and no CAA/NS types** -- alias is a public-DNS concept; private zones cannot delegate subdomains and never face public certificate authorities.
- **The TTL is always sent explicitly** -- 300 seconds (5 minutes) when the spec leaves it unset -- so both engines send identical wire shapes.
- **Azure caps an A record set at 20 addresses** (spec-validated) and each TXT value at 1,024 characters (documented; ARM enforces).
- **Tags land as ARM record-set METADATA** -- record sets carry no ARM tags proper.
- **Name, zone, and type are fixed at creation** -- changing any replaces the record set; values and TTL update in place.

## Required Permissions

Least-privilege runner permissions for this kind are declared in [`../permissions.yaml`](../permissions.yaml).
