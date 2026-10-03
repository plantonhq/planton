# AwsCloudTrail — Kind Guide

Authored operational judgment for the CloudTrail component: the design
decisions behind the spec's shape, and what to know before operating
audit trails in production.

## Design decisions

- **The bucket policy is the consumer's contract.** AWS validates the
  delivery bucket's POLICY at trail creation ("Incorrect S3 bucket
  policy is detected") — the policy lives on AwsS3Bucket
  (`spec.policy`, granting `cloudtrail.amazonaws.com` GetBucketAcl +
  PutObject), never inside this component. Create the bucket WITH its
  policy before the trail.
- **One selector style per trail, enforced early.** Classic
  `event_selectors` and `advanced_event_selectors` are mutually
  exclusive in AWS; the spec fails the manifest before it reaches the
  cloud. Prefer the advanced style for new trails — it is the one AWS
  extends (network-activity events, eventCategory matching).
- **CloudWatch mirroring travels as a pair.** AWS requires the log
  group and the delivery role together, so the spec models them as one
  presence-typed message — half-wired mirroring cannot be expressed.
  Both engines normalize the group ARN to AWS's `:*` suffix form.
- **CloudTrail Lake split out.** The event data store carries no trail
  edge (it deploys with zero trails, owns its own billing/retention/
  termination protection), so it is its own component rather than an
  arm here.
- **The delegated-admin registration is account-global.** One
  delegation per organization, performed from the management account —
  it rides this kind because it exists to run organization trails, but
  most deployments leave it unset.

## Operating audit trails in production

- **Multi-region + validation is the audit posture.** A single-region
  trail misses activity everywhere else; without digest files,
  tampering with delivered logs is undetectable. The compliance preset
  ships both on.
- **Management events are free once.** The first copy of management
  events per account is free; a second trail delivering the same
  events bills. Data events and Insights always bill per event —
  scope `data_resources` deliberately.
- **KMS delivery needs the key's consent.** With `kms_key_id` set, the
  key POLICY must carry the "Allow CloudTrail to encrypt logs" grant
  for `cloudtrail.amazonaws.com`; without it, creation fails.
- **Organization trails run from the management account** (or the
  delegated administrator, with all-features Organizations enabled).
  Member accounts see the trail read-only.
- **Destroy deletes the trail, not the evidence.** Delivered log
  files stay in the bucket under their lifecycle rules; plan bucket
  retention as the real audit retention.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
