<p align="center">
  <img src="logo.svg" alt="AWS GuardDuty" width="80"/>
</p>

# AWS GuardDuty

Manage this account+region's [GuardDuty](https://docs.aws.amazon.com/guardduty/latest/ug/what-is-guardduty.html)
threat-detection posture: the detector, its protection-plan features,
finding filters, trusted/threat IP lists, findings export, and — for
organization administrators — member-account management.

This is a **region singleton**: AWS allows exactly ONE detector per
account per region, and the detector has no name (its AWS-assigned ID
is the identity), so `metadata.name` never reaches AWS. Deploy at most
one instance per region.

## What Gets Managed

- **The detector**: monitoring on/off and the finding re-publish
  frequency.
- **Protection plans** (`features`): S3 data events, EKS audit logs,
  runtime monitoring, RDS login events, Lambda network logs, AI
  protection — each a patch onto the detector, with agent-management
  sub-toggles.
- **Finding filters**: match criteria that auto-archive (or organize)
  findings — the noise-control surface.
- **Trusted and threat IP lists**: S3-hosted list files GuardDuty
  reads as `guardduty.amazonaws.com`.
- **Findings export**: an S3 destination (bucket policy + KMS key
  policy must grant GuardDuty — see [AwsS3Bucket](../awss3bucket) and
  [AwsKmsKey](../awskmskey)).
- **Organization administration**: delegation, auto-enrollment
  posture (NEW/ALL/NONE), org-wide feature enablement, member
  accounts with per-member feature overrides — or, member-side,
  accepting an administrator's invitation.

Destroying this component **deletes the detector** — findings and
every satellite go with it. Feature and organization-configuration
arms are patches (AWS has no delete for them); removing an arm from
the spec reverts nothing on its own. Malware Protection for S3 is
deliberately NOT part of this kind — a protection plan guards an
S3 bucket, not a detector, and ships as its own kind.

See [v1alpha1/reference.md](v1alpha1/reference.md) for the full field
reference generated from the spec proto.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
