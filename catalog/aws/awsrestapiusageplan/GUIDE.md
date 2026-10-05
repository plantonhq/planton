# AwsRestApiUsagePlan — Kind Guide

Authored operational judgment for the REST API usage-plan kind:
the design decisions behind the spec's shape, and what to know before
metering API consumers.

## Design decisions

- **A plan is not an API field.** One plan covers many APIs and
  stages; putting it on AwsRestApiGateway would force every consumer
  cohort to share an API lifecycle.
- **Keys are created and attached here.** A key with no plan is an
  orphan; a plan with no keys meters nobody. The component owns both
  so the attachment cannot be forgotten.
- **Key values are not outputs.** They are secrets. IDs and ARNs are
  exported; the value is read from AWS when you distribute it.
- **Quota and throttle are independent.** Omit quota for throttle
  only; omit throttle to inherit the account ceiling.

## Metering APIs in production

- **Keys are not authentication.** Pair `apiKeyRequired` with IAM,
  Cognito, or a Lambda authorizer — otherwise anyone who sees a key
  is a caller.
- **Start with a daily quota.** Weekly / monthly periods hide burn
  until the window rolls; day is the honest unit for a new plan.
- **Per-method throttles are for the expensive paths.** Put the
  plan-wide ceiling on `throttle` and a tighter cap on `/search/GET`
  (or whichever method is the costly one).
- **Rotate by adding a key, then removing the old one.** There is no
  in-place value rotate; treat keys as replaceable identities.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
