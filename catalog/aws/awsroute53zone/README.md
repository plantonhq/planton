# Overview

The AWS Route53 Zone API resource provisions Amazon Route 53 hosted zones — the DNS containers domains live in. The zone's domain name comes from `metadata.name`, and the surface covers both zone flavors at full depth: public zones (with reusable delegation sets, DNSSEC signing, query logging, and accelerated recovery) and private zones (split-horizon DNS across one or more VPCs).

Individual DNS records are deliberately NOT part of the zone: each record is its own `AwsRoute53DnsRecord` resource composing onto the zone's `zone_id` output. Records have independent lifecycles — they are created, repointed, and deleted without touching the zone — are many-per-zone, and carry their own routing surface (weighted, latency, failover, geolocation, geoproximity, CIDR, multivalue). Folding them into the zone would bury that composability inside one document.

## Why We Created This API Resource

DNS is foundational infrastructure. When DNS works, nobody notices; when it breaks, everything breaks. We created this resource to:

- **Make the zone a first-class graph node**: certificates validate against it, load balancers register alias records through it, and every record composes onto its `zone_id` output.
- **Keep the public/private contract honest**: private zones require VPC associations; delegation sets, DNSSEC, query logging, and accelerated recovery are public-zone-only — all enforced at authoring time instead of failing at apply.
- **Expose the real operational surface**: DNSSEC signing with its KMS key requirements, query logging with its us-east-1 and resource-policy prerequisites, safe teardown via `force_destroy` — documented where you configure them.

## Key Features

### Public Hosted Zones

- **Global resolution** with the four authoritative name servers exported as outputs for registrar delegation (plus the primary name server, the SOA MNAME)
- **Reusable delegation sets** (`delegation_set_id`) for white-label DNS — many zones sharing the same name servers
- **DNSSEC signing**: a key-signing key created from a referenced KMS key plus the zone's signing toggle, protecting resolvers from spoofed responses (the KMS key must live in us-east-1 with key spec ECC_NIST_P256 and the dnssec-route53 service key policy). The KSK's operational status is configurable (`ACTIVE` by default; `INACTIVE` is the documented diagnostics lever), and the signed zone exports its chain-of-trust payload — `ds_record`, `dnskey_record`, and `key_signing_key_tag` — so the registrar DS registration can be completed straight from outputs
- **Query logging** to a CloudWatch Logs log group (must live in us-east-1; delivery requires an account-level CloudWatch Logs resource policy allowing route53.amazonaws.com — account-scoped and deliberately not created per zone)
- **Accelerated recovery** for faster control-plane propagation during regional recovery events. The field is tri-state on purpose: AWS keeps the feature's current state when it is omitted and requires an explicit `false` to switch it back off — so unset means "leave as-is", never "disable"

### Private Hosted Zones (Split-Horizon DNS)

- **VPC associations** define the zone: names resolve only inside the associated VPCs
- **Cross-region associations** within the same account (`vpc_region` defaults to the zone's region)
- Associated VPCs need DNS support and DNS hostnames enabled

### Safety

- **`force_destroy`** purges remaining records (and disables DNSSEC) before deletion so teardown cannot wedge on "zone not empty" — leave it off to protect zones carrying live records

## Composition

- `AwsRoute53DnsRecord.zone_id` → this zone's `status.outputs.zone_id`
- `AwsCertManagerCert.route53_hosted_zone_id` → DNS-validated certificates place their validation records here
- `AwsAlb.dns.route53_zone_id` / `AwsNlb.dns.route53_zone_id` → load balancers register their alias records here
- `AwsRoute53ZoneDnssec.kms_key_arn` → `AwsKmsKey.status.outputs.key_arn`
- `AwsRoute53ZoneQueryLogging.cloudwatch_log_group_arn` → `AwsCloudwatchLogGroup.status.outputs.log_group_arn`

## What Is Deliberately Elsewhere

- **DNS records** — first-class `AwsRoute53DnsRecord` resources
- **Health checks** — first-class `AwsRoute53HealthCheck` resources referenced by records
- **Cross-account VPC associations** — require an authorization handshake between accounts (a separate surface)
- **Traffic policies, CIDR collections, reusable delegation-set management, domain registration** — separate Route 53 surfaces that compose by ID where records need them

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
