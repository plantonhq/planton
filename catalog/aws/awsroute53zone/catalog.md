# AWS Route 53 Zone

Deploys a Route53 hosted zone that can serve as either a public internet-facing DNS zone or a private VPC-scoped zone. Public zones support DNSSEC signing backed by a KMS key, query logging to CloudWatch, reusable delegation sets for vanity name servers, and accelerated recovery. The zone's domain name comes from `metadata.name` and is create-time immutable.

Individual DNS records are deliberately not part of the zone: each record is its own AwsRoute53DnsRecord resource referencing this zone's `zone_id` output, so records can be created, repointed, and deleted without touching the zone.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Route53 Hosted Zone** -- a public hosted zone resolving globally on the internet, or a private hosted zone resolving only within associated VPCs
- **VPC Associations** -- created only when `isPrivate` is `true`; each association allows the private zone to resolve DNS queries from the specified VPC (cross-region associations supported via `vpcRegion`)
- **DNSSEC Signing** -- created only when the `dnssec` block is configured; a key-signing key is created from the referenced KMS key (status configurable: `ACTIVE` by default, `INACTIVE` as the diagnostics lever) and the zone's signing status is switched on. The signed zone exports `ds_record`, `dnskey_record`, and `key_signing_key_tag` outputs -- the values the registrar needs to complete the chain of trust
- **Query Logging** -- created only when the `queryLogging` block is configured; sends DNS query logs to the referenced CloudWatch Logs group for debugging and security monitoring
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **A VPC** (required for private zones) -- one or more VPCs with `enableDnsHostnames` and `enableDnsSupport` enabled. Provide VPC IDs directly or reference an AwsVpc Infra Component via ValueFromRef.
- **A KMS key** (required for DNSSEC) -- must live in us-east-1, be an asymmetric key with key spec ECC_NIST_P256 and SIGN_VERIFY usage, and its key policy must allow the Route53 DNSSEC service principal (dnssec-route53.amazonaws.com). Reference an AwsKmsKey Infra Component or pass the key ARN.
- **A CloudWatch Log Group** (required for query logging) -- must live in us-east-1 (Route53 delivers query logs there regardless of the zone's region), and an account-level CloudWatch Logs resource policy must allow the route53.amazonaws.com service principal to write. Reference an AwsCloudwatchLogGroup Infra Component or pass the log group ARN.
- **Domain registrar DS record** (DNSSEC only) -- signing is half the chain of trust; after deployment, register the zone's DS record with the domain registrar to complete it. The value is exported as the `ds_record` output (with `key_signing_key_tag` for registrar forms that ask for it) -- no console fishing required.

## Deploy

### Console

Open the deployment store, find **AWS Route 53 Zone**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Public DNS Zone** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRoute53Zone
metadata:
  name: example.com
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  comment: production apex zone — owned by platform team
  isPrivate: false
```

```shell
planton apply -f route53-zone.yaml
```

This creates a public hosted zone for `example.com` with fresh Route53 name servers, no DNSSEC, and no query logging. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying a private zone as part of a multi-resource environment, use ValueFromRef to wire the zone to a VPC deployed in the same InfraPipeline:

```yaml
spec:
  isPrivate: true
  vpcAssociations:
    - vpcId:
        valueFrom:
          kind: AwsVpc
          name: production-vpc
          fieldPath: status.outputs.vpc_id
      vpcRegion: us-east-1
```

The InfraPipeline resolves the dependency graph, deploys the VPC first, then provisions the private hosted zone with the resolved VPC ID.

## Key Configuration

These are the most important decisions when configuring a Route53 zone. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Public vs. private zone** -- Set `isPrivate: false` (default) for a zone that resolves globally on the internet. Set `isPrivate: true` for split-horizon DNS that resolves only within associated VPCs. Private zones require at least one VPC association; every other feature below is public-zone only.

**DNSSEC** -- Off by default. Configure the `dnssec` block with a `kmsKeyArn` (literal ARN or AwsKmsKey reference) to cryptographically sign DNS responses and prevent spoofing. Optionally name the key-signing key via `keySigningKeyName`. Requires the DS record at your domain registrar to complete the chain of trust.

**Query logging** -- Off by default. Configure the `queryLogging` block with a `cloudwatchLogGroupArn` (literal ARN or AwsCloudwatchLogGroup reference) to send DNS query logs to CloudWatch. Useful for security monitoring and debugging resolution issues, but generates significant log volume on high-traffic domains.

**Delegation set** -- Leave `delegationSetId` blank (default) and Route53 assigns four fresh name servers. Set it to assign the zone's name servers from an existing reusable delegation set (vanity name servers, bulk migrations). Create-time immutable.

**Teardown contract** -- Leave `forceDestroy: false` (default) so deletion fails while the zone still carries live records — protection against accidental teardown. Set it to `true` for development zones so deletion purges every record (except NS/SOA) and disables DNSSEC in one motion.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsVpc** (private zones) | `vpcAssociations[].vpcId` | `status.outputs.vpc_id` |
| **AwsKmsKey** (DNSSEC) | `dnssec.kmsKeyArn` | `status.outputs.key_arn` |
| **AwsCloudwatchLogGroup** (query logging) | `queryLogging.cloudwatchLogGroupArn` | `status.outputs.log_group_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `zone_id` | Route53 hosted zone ID | AwsRoute53DnsRecord zone reference, ACM DNS validation |
| `zone_name` | Hosted zone domain name (trailing dot removed) | Application configuration, subdomain delegation |
| `nameservers` | The four authoritative name servers | Domain registrar NS delegation |
| `primary_name_server` | The delegation set's first name server (SOA MNAME) | Zone-transfer and SOA tooling |
| `zone_arn` | Hosted zone ARN | IAM policies scoping route53:ChangeResourceRecordSets to specific zones |
| `ds_record` | The DS record of the key-signing key (DNSSEC only) | Registering with the domain registrar to complete the chain of trust |
| `dnskey_record` | The DNSKEY record of the key-signing key (DNSSEC only) | Out-of-band verification; parents that take a DNSKEY instead of a DS |
| `key_signing_key_tag` | The KSK's key tag (DNSSEC only) | Registrar forms that ask for the key tag beside the DS record |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Public zone** -- A globally resolving hosted zone for internet-facing domains. Suitable for websites, APIs, and services that need public DNS resolution. Start from the **Public DNS Zone** preset.

**Private VPC zone** -- A VPC-scoped zone for internal service discovery and split-horizon DNS. Only resolves within associated VPCs, keeping internal hostnames private. Start from the **Private VPC DNS Zone** preset.

**DNSSEC-signed zone** -- A public zone signed with a KSK backed by a us-east-1 KMS key, exporting the `ds_record` the registrar needs to complete the chain of trust. Start from the **DNSSEC-Signed Public Zone** preset.

## Works With

- [**AWS VPC**](/infra-catalog/aws-vpc) -- provides the VPC ID for private hosted zone associations
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) -- provides the asymmetric signing key for DNSSEC
- [**AWS CloudWatch Log Group**](/infra-catalog/aws-cloudwatch-log-group) -- receives DNS query logs
- [**AWS Route 53 DNS Record**](/infra-catalog/aws-route53-dns-record) -- the first-class home for this zone's records, referencing `zone_id`
