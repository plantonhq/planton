# AwsSagemakerImage — Kind Guide

Authored operational judgment for the SageMaker image component: the
design decisions behind the spec's shape, and what to know before
running custom images in production.

## Design decisions

- **Versions fold into the image spec.** An image without versions is
  an empty shelf; the version pointing at your ECR image is why the
  entry exists. One manifest carries the registry entry and its
  versions together instead of scattering satellites.
- **Entries are keyed by position.** AWS assigns version numbers
  sequentially and never reuses them, so no spec field can be the
  stable identity — the entry's POSITION in `versions` is. The modules
  key satellites by index, which is why the spec teaches append-only:
  add at the end, never reorder.
- **`base_image` is the version's identity.** Changing it replaces the
  version under a NEW AWS-assigned number (the old number stays
  retired). The compatibility metadata — `job_type`, `ml_framework`,
  `processor`, `programming_lang`, `vendor_guidance` — updates in
  place.
- **The image's AWS name derives from `metadata.name`.** SageMaker
  accepts the same charset, so there is no rename field to drift.
- **`role_arn` defaults to an `AwsIamRole` reference.** The ECR-pull
  role is the image's one hard dependency; changing it replaces the
  image.

## Running SageMaker images in production

- **Budget a minute per create.** The provider sleeps ~1 minute before
  `CreateImage` for IAM propagation — every create is at least a
  minute, even for an empty registry entry.
- **Versions attach one at a time.** AWS serializes version creation
  per image (the provider holds a mutex); ten new versions in one
  apply land sequentially.
- **Let aliases carry meaning, not numbers.** Aliases (`latest`,
  `stable`) move freely between versions and update in place — point
  consumers at aliases and promote by moving them, instead of chasing
  AWS-assigned numbers.
- **Versions carry no tags upstream** (by provider design) — tag the
  image; don't expect cost allocation per version.
- **ECR images must live in the same account and region** as the
  image. Cross-account registries need a replication step before the
  version can register.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
