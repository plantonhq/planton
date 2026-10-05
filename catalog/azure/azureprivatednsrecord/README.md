# Overview

The **AzurePrivateDnsRecord** component deploys one DNS record set in an Azure Private DNS zone -- the name entries ("db.internal.contoso.com points at 10.0.0.5") that make services findable inside the virtual networks linked to the zone, invisible to the public internet.

## Purpose

- **Private names as configuration**: the entries that used to live in hosts files or hand-run `az` commands become declarative resources beside the services they name.
- **One kind, every record type**: A, AAAA, CNAME, MX, PTR, SRV, and TXT -- the record type is whichever typed payload the spec carries, so a record can never be declared with a shape its type cannot hold.
- **Deploy-time values wire by reference**: CNAME targets and TXT values accept references to other components' outputs, so names minted at deploy time flow in without hand-copying.

## Key Features

- Full azurerm v5 surface across all seven private record resources, folded into one union spec (shared name/zone/TTL/tags at the root, typed payloads per record type).
- Typed reference to the owning AzurePrivateDnsZone (the reference is the deploy-order edge); values and TTL update in place.
- Mirrors the provider's own bounds exactly: 20 addresses per A record set, 16-bit MX/SRV integers, the 1-65535 SRV port -- and nothing stricter.

## Use Cases

- **Database and internal-API names**: A records in the application's private zone, one per service, deployed with the service.
- **Stable aliases over moving targets**: CNAME records pointing at hostnames that change with infrastructure (a failover pair's active member, a private endpoint's FQDN).
- **Service discovery for on-premises-style workloads**: SRV records for SIP/LDAP/Kerberos endpoints living in Azure.

## Future Enhancements

- The 1,024-character TXT value cap and per-zone record-set quotas stay documentation until service quotas can be introspected offline.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
