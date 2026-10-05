<p align="center">
  <img src="logo.svg" alt="AWS CloudTrail" width="80"/>
</p>

# AWS CloudTrail

Manage a [CloudTrail trail](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-user-guide.html)
— the account's API audit log, delivered as log files to S3 with
optional CloudWatch Logs mirroring, SNS delivery notices, Insights
anomaly detection, and organization-wide capture.

## What Gets Managed

- **The trail** (`metadata.name` is the trail name): home region,
  multi-region capture, global-service events, log-file validation
  digests, and the logging on/off switch.
- **Delivery**: the S3 bucket (which must carry the
  `cloudtrail.amazonaws.com` bucket policy — see
  [AwsS3Bucket](../awss3bucket) `spec.policy`), optional SSE-KMS
  encryption, and optional SNS notification per delivered file.
- **Event scope**: classic event selectors (management scope + coarse
  S3/Lambda/DynamoDB data events) OR advanced event selectors
  (fine-grained field matching) — AWS keeps exactly one style per
  trail.
- **CloudTrail Insights**: call-rate and error-rate anomaly engines.
- **Organization capture**: `is_organization_trail` plus the optional
  account-global delegated-administrator registration.

Destroying this component **deletes the trail** — already-delivered
log files stay in the bucket. Management-event delivery is free for
the account's first trail copy; data events and Insights bill per
event.

CloudTrail Lake (event data stores) is deliberately NOT part of this
kind — a data store deploys with zero trails and owns its own
billing, retention, and termination-protection lifecycle, so it ships
as its own kind.

See [v1alpha1/reference.md](v1alpha1/reference.md) for the full field
reference generated from the spec proto.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
