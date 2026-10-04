# AwsConfigAggregator — Kind Guide

Authored operational judgment for the Config aggregation kind:
the design decisions behind the spec's shape, and what to know before
operating aggregators in production.

## Design decisions

- **Its own kind, not a recorder arm.** Aggregation references no
  Config recorder — it works in an account with zero recorders (the
  data comes from SOURCE accounts' recorders), so it never belonged on
  AwsConfigRecorder.
- **Both sides of the relationship in one kind.** The aggregator (the
  collector) and the reciprocal authorization grants (the
  source-account side) are two arms of one spec, at-least-one
  required. The two arms deploy in OPPOSITE accounts of a
  cross-account topology — the same precedent as AwsGuardDuty's
  member/invite-accepter arms. An account may honestly declare both
  (run its own rollup AND grant a sibling's).
- **Exactly one source shape.** AWS accepts an account list or an
  organization source, never both; the spec CEL mirrors the
  provider's exclusivity, and each source requires regions or the
  all-regions posture (the provider documents this contract).
- **Grants are keyed by identity.** Each grant's
  `{account_id}:{authorized_aws_region}` pair is the provider's own
  import ID; the modules key instances by it, so reordering the spec
  list never churns resources, and the outputs echo a same-keyed ARN
  map for blind imports.
- **The deprecated `region` alias is not modeled.**
  `authorized_aws_region` is the surviving provider argument; the
  alias is a recorded exclusion.

## Operating aggregators in production

- **Pending-authorization is a state, not an error.** An account-list
  aggregator naming a source that has not granted it shows the source
  as pending; data flows the moment the source account deploys its
  grant arm. Organization sources skip grants entirely.
- **The aggregator sees only what recorders record.** No recorder in
  a source account/region means no data from it — pair the rollup
  with AwsConfigRecorder deployments where coverage matters.
- **Prefer the organization source at org scale.** It self-discovers
  accounts (new accounts join the view automatically) and needs the
  `AWSConfigRoleForOrganizations` role from the management or
  delegated-admin account.
- **Replacement semantics are asymmetric.** Adding a source block to
  an existing aggregator replaces it; editing or removing one updates
  in place (the provider's own diff rule — the modules encode
  nothing).

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
