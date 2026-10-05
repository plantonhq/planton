# AwsRoute53Zone

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsRoute53ZoneSpec defines the configuration for an Amazon Route 53 hosted
zone — the DNS container for a domain. The zone's domain name comes from
metadata.name (e.g. a zone named "example.com" hosts that domain) and is
create-time immutable (ForceNew).

A hosted zone is either PUBLIC (resolves on the internet; Route 53 assigns
four authoritative name servers, exported as outputs for registrar
delegation) or PRIVATE (resolves only inside the associated VPCs —
split-horizon DNS). The private/public choice shapes the rest of the
surface: reusable delegation sets, DNSSEC signing, accelerated recovery,
and query logging are public-zone features, while VPC associations define a
private zone.

Individual DNS records are deliberately NOT part of the zone: each record
is its own AwsRoute53DnsRecord resource referencing this zone's zone_id
output. Records have independent lifecycles (create, repoint, and delete
without touching the zone), are many-per-zone, and carry their own routing
surface — folding them here would bury that composability inside one
document.

## Example

```yaml
# AWS Route 53 hosted zone — examples
#
# The zone's domain name comes from metadata.name. Individual DNS records are
# separate AwsRoute53DnsRecord resources composing onto the zone's zone_id
# output.
#
# Usage:
#   planton apply -f manifest.yaml

apiVersion: aws.planton.dev/v1alpha1
kind: AwsRoute53Zone
metadata:
  name: example.com
spec:
  region: us-east-1
  comment: production apex zone

---
# Private split-horizon zone: resolves only inside the associated VPCs.

apiVersion: aws.planton.dev/v1alpha1
kind: AwsRoute53Zone
metadata:
  name: internal.example.com
spec:
  region: us-west-2
  isPrivate: true
  vpcAssociations:
    - vpcId:
        value: vpc-0123456789abcdef0
    # Cross-region VPCs name their region explicitly; same-region VPCs
    # inherit the zone's region.
    - vpcId:
        value: vpc-0fedcba9876543210
      vpcRegion: eu-west-1

---
# DNSSEC-signed public zone with accelerated recovery: the full
# security/resilience shape. The KMS key must live in us-east-1 with key
# spec ECC_NIST_P256 and the dnssec-route53.amazonaws.com key policy (see
# the spec docs). The ds_record output carries the value to register at
# the registrar — signing without the parent DS protects nobody.
# keySigningKeyStatus defaults to ACTIVE; INACTIVE is the documented
# diagnostics lever.

apiVersion: aws.planton.dev/v1alpha1
kind: AwsRoute53Zone
metadata:
  name: secure.example.com
spec:
  region: us-east-1
  comment: DNSSEC-signed zone
  enableAcceleratedRecovery: true
  dnssec:
    kmsKeyArn:
      value: arn:aws:kms:us-east-1:123456789012:key/1234abcd-12ab-34cd-56ef-1234567890ab
    keySigningKeyName: secure-example-ksk
    keySigningKeyStatus: ACTIVE
  queryLogging:
    cloudwatchLogGroupArn:
      value: arn:aws:logs:us-east-1:123456789012:log-group:/aws/route53/secure.example.com
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.comment` | `string` |  |  |  |
| `spec.isPrivate` | `bool` |  |  |  |
| `spec.vpcAssociations` | `[]AwsRoute53ZoneVpcAssociation` |  |  |  |
| `spec.vpcAssociations[].vpcId` | `string \| valueFrom` | yes |  | AwsVpc (`status.outputs.vpc_id`) |
| `spec.vpcAssociations[].vpcRegion` | `string` |  |  |  |
| `spec.delegationSetId` | `string` |  |  |  |
| `spec.forceDestroy` | `bool` |  |  |  |
| `spec.enableAcceleratedRecovery` | `bool` |  |  |  |
| `spec.queryLogging` | `AwsRoute53ZoneQueryLogging` |  |  |  |
| `spec.queryLogging.cloudwatchLogGroupArn` | `string \| valueFrom` | yes |  | AwsCloudwatchLogGroup (`status.outputs.log_group_arn`) |
| `spec.dnssec` | `AwsRoute53ZoneDnssec` |  |  |  |
| `spec.dnssec.kmsKeyArn` | `string \| valueFrom` | yes |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.dnssec.keySigningKeyName` | `string` |  |  |  |
| `spec.dnssec.keySigningKeyStatus` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Route 53 itself is a global service; this selects the region used for
provider API calls and for regional companions (e.g. a private zone's
default VPC region).
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.comment

`string`

Human-readable comment stored on the hosted zone, visible in the AWS
console and GetHostedZone responses. Useful for ownership or purpose
notes (e.g. "production apex zone — owned by platform team").
Maximum 256 characters.

- rule: {"string":{"maxLen":"256"}}

### spec.isPrivate

`bool`

Marks the zone as a private hosted zone that resolves only inside the
VPCs listed in vpc_associations. Private zones require at least one VPC
association at creation (AWS bakes the first VPC into CreateHostedZone).
Default: false (public zone, resolves globally).

### spec.vpcAssociations

`[]AwsRoute53ZoneVpcAssociation`

VPCs that can resolve this private zone. Required (at least one) when
is_private is true; must be empty for public zones.

Each associated VPC must have DNS support and DNS hostnames enabled.
All associations here must be same-account: associating a VPC from
ANOTHER account requires an authorization handshake between the two
accounts (a separate cross-account surface, deliberately not modeled).

### spec.vpcAssociations[].vpcId

`string | valueFrom` · required

The VPC to associate. Can reference an AwsVpc resource.

- references: AwsVpc (`status.outputs.vpc_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsVpc, name: <that resource's name>, fieldPath: status.outputs.vpc_id}} -- a bare string does not parse

### spec.vpcAssociations[].vpcRegion

`string`

The region the VPC lives in. Defaults to the zone's region when omitted
— set it only for VPCs in other regions (a private zone can be resolved
from VPCs across regions within the same account).

### spec.delegationSetId

`string`

ID of a reusable delegation set to assign this zone's name servers from.
Reusable delegation sets give many zones the same four name servers —
the white-label DNS pattern (vanity name servers) and bulk zone
migrations. Public zones only (AWS rejects it for private zones), and
create-time immutable (ForceNew).
When omitted, Route 53 assigns a fresh set of name servers to the zone.

- rule: {"string":{"maxLen":"32"}}

### spec.forceDestroy

`bool`

When true, deleting the zone first purges every record in it (except the
required NS/SOA pair) and disables DNSSEC signing, so deletion cannot
fail on "zone not empty". Leave false (default) to protect a zone that
still carries live records from accidental teardown.

### spec.enableAcceleratedRecovery

`bool` · optional (explicit presence)

Enables accelerated recovery for the hosted zone: Route 53 pre-stages
the zone's data so control-plane changes propagate faster during
regional recovery events. Public zones only.

Tri-state on purpose: AWS keeps the feature's CURRENT state when the
field is omitted, so switching it off requires an EXPLICIT false — the
provider documents that removing the argument does not disable it.
Unset = leave as-is (off for new zones), true = enable, false = disable.

### spec.queryLogging

`AwsRoute53ZoneQueryLogging`

DNS query logging to CloudWatch Logs — who is querying which names, with
response codes. Used for security monitoring, debugging resolution
issues, and understanding query patterns. Public zones only.
Warning: high-traffic domains generate large log volumes (and cost).

### spec.queryLogging.cloudwatchLogGroupArn

`string | valueFrom` · required

ARN of the destination CloudWatch Logs log group (must be in us-east-1).
Can reference an AwsCloudwatchLogGroup resource.

- references: AwsCloudwatchLogGroup (`status.outputs.log_group_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsCloudwatchLogGroup, name: <that resource's name>, fieldPath: status.outputs.log_group_arn}} -- a bare string does not parse

### spec.dnssec

`AwsRoute53ZoneDnssec`

DNSSEC signing for the zone: Route 53 signs the zone's records with a
key-signing key (KSK) backed by an asymmetric KMS key, protecting
resolvers from spoofed responses. Public zones only.

Enabling signing here is half the chain of trust — to complete it, the
DS record from the signed zone must also be registered with the parent
(the domain registrar), which is outside this resource.

### spec.dnssec.kmsKeyArn

`string | valueFrom` · required

The asymmetric KMS key backing the key-signing key. Can reference an
AwsKmsKey resource (see the message comment for the us-east-1 /
ECC_NIST_P256 / key-policy requirements).

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.dnssec.keySigningKeyName

`string`

Name of the key-signing key inside Route 53 (3–128 characters; letters,
numbers, "._-"). Defaults to a name derived from the zone when omitted.

- rule: {"string":{"pattern":"^$|^[0-9A-Za-z._-]{3,128}$"}}

### spec.dnssec.keySigningKeyStatus

`string`

Operational status of the key-signing key: "ACTIVE" (default when
omitted) or "INACTIVE". Deactivating the KSK stops Route 53 from using
it to sign — the monitoring/troubleshooting lever AWS documents for
investigating signing problems without tearing the DNSSEC config down.
Updatable in place.

Note: while signing is enabled, an INACTIVE KSK means signatures cannot
refresh — deactivate only while diagnosing, or after a replacement KSK
is active. AWS allows at most two KSKs per hosted zone (the rotation
pair); this resource models the one steady-state KSK, so a
zero-downtime rotation (create second KSK, retire the first) is
performed with AWS tooling and then reconciled here by pointing
kms_key_arn/key_signing_key_name at the new key.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["ACTIVE","INACTIVE"]}}

## Validation Rules

- `private_zone_requires_vpc`: at least one vpc_associations entry is required when is_private is true (AWS creates a private zone attached to its first VPC)
- `vpc_associations_require_private`: vpc_associations can only be set when is_private is true (public zones resolve globally, not per-VPC)
- `delegation_set_public_only`: delegation_set_id cannot be set for private zones (reusable delegation sets apply to public zones only)
- `accelerated_recovery_public_only`: enable_accelerated_recovery cannot be set for private zones
- `query_logging_public_only`: query_logging cannot be set for private zones (Route 53 query logging supports public hosted zones only)
- `dnssec_public_only`: dnssec cannot be set for private zones (DNSSEC signing supports public hosted zones only)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsRoute53Zone, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.zone_id` | `string` | The hosted zone ID (e.g. "Z1D633PJN98FT9"). The identifier used by every resource that composes onto this zone. |
| `status.outputs.zone_name` | `string` | The zone's domain name as normalized by Route 53 (trailing dot removed), e.g. "example.com". |
| `status.outputs.nameservers` | `[]string` | The four authoritative name servers assigned to the zone. For a public zone, these are the values to set as the domain's NS delegation at the registrar. |
| `status.outputs.primary_name_server` | `string` | The first (primary) name server of the zone's delegation set — the one used as the SOA MNAME. |
| `status.outputs.zone_arn` | `string` | The Amazon Resource Name of the hosted zone (arn:aws:route53:::hostedzone/<zone_id>). Used in IAM policies scoping route53:ChangeResourceRecordSets to specific zones. |
| `status.outputs.ds_record` | `string` | The DS (Delegation Signer) record for the zone's key-signing key, only populated when DNSSEC signing is enabled. This is the value to register with the PARENT zone (the domain registrar) to complete the DNSSEC chain of trust — signing without it protects nobody, because resolvers have no trust anchor to validate against. |
| `status.outputs.dnskey_record` | `string` | The DNSKEY record for the zone's key-signing key (the public key in DNS record form), only populated when DNSSEC signing is enabled. Useful for out-of-band verification and for parents that take a DNSKEY instead of a DS. |
| `status.outputs.key_signing_key_tag` | `string` | The key tag of the key-signing key, only populated when DNSSEC signing is enabled — the short integer identifier registrars display next to a DS record (matches the first field of ds_record). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.vpcAssociations[].vpcId` | AwsVpc | `status.outputs.vpc_id` |
| `spec.queryLogging.cloudwatchLogGroupArn` | AwsCloudwatchLogGroup | `status.outputs.log_group_arn` |
| `spec.dnssec.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsAlb | `spec.dns.route53ZoneId` | `status.outputs.zone_id` |
| AwsCertManagerCert | `spec.route53HostedZoneId` | `status.outputs.zone_id` |
| AwsNlb | `spec.dns.route53ZoneId` | `status.outputs.zone_id` |
| AwsRoute53DnsRecord | `spec.zoneId` | `status.outputs.zone_id` |
| KubernetesExternalDns | `spec.awsRoute53.zoneIdFilters` | `status.outputs.zone_id` |

## See Also

- [Overview](../README.md)
