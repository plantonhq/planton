# Custom Slice Monitor

This preset watches ONE team's spend slice (a cost-allocation tag)
with real-time SNS alerts, filtered to anomalies that are both $50+
and 10%+ above normal.

## When to Use

- A team, product, or environment whose spend deserves its own
  anomaly stream
- Feeding anomalies into chat or incident tooling in real time

## What You Get

- A CUSTOM monitor over the slice a Cost Explorer expression selects
  (here: the `team: platform` tag, spelled `user:team` — Cost
  Explorer's canonical form for user-defined tag keys; the expression
  is the AWS Expression document verbatim)
- IMMEDIATE individual alerts to an SNS topic, noise-filtered by a
  composed absolute-AND-percentage threshold

## Customize

- Watch a member account instead:
  `monitorSpecification: {Dimensions: {Key: LINKED_ACCOUNT, Values: ["123456789012"]}}`
- The tag key must be an ACTIVATED cost-allocation tag or the slice
  matches nothing — and keep the document in its canonical form: the
  `user:` prefix on the key (`aws:` for AWS-generated tags) and every
  root member present with `null` when unused. A sparser or unprefixed
  document deploys fine, then proposes a replacement on every re-plan
- The SNS topic's policy must allow costalerts.amazonaws.com to
  publish ([AWS SNS Topic](/infra-catalog/aws-sns-topic))
