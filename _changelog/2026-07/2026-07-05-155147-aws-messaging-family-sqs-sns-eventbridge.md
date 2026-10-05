# AWS Messaging Family — SQS, SNS, and EventBridge to 90/10

**Date:** 2026-07-05
**Scope:** Kinds #26–#30 — `AwsSqsQueue`, `AwsSnsTopic`, `AwsSnsSubscription` (new kind), `AwsEventBridgeBus`, `AwsEventBridgeRule`

## Summary

The five messaging kinds brought to the 90/10 bar in one session: SQS gains the
redrive-allow permission surface, SNS topics are rebuilt with subscriptions
split out into a first-class `AwsSnsSubscription` kind (enum 353), and the
EventBridge pair closes a live cross-engine parity defect (the Terraform bus
module silently dropped KMS/DLQ/logging) while the rule gains the full typed
target-block surface. All ten live dual-engine E2E lanes green with a
zero-orphan sweep — the family's first-ever live coverage.

## Product impact

- **Subscriptions become graph nodes** — a topic owner and a queue owner can
  now be different teams: the subscription is its own resource with its own
  ARN, filter policy, raw delivery, redrive DLQ, archived-message replay, and
  HTTP/S confirmation handshake. Subscribing an owned endpoint to a
  cross-account topic works by literal ARN. The topic keeps identity, policy,
  and delivery posture; nine protocols supported with the couplings
  CEL-enforced (firehose requires a role; confirmation knobs are HTTP/S-only).
- **SNS topics model the full provider surface** — FIFO message archiving
  (subscriptions replay via their own `replay_policy`), per-protocol
  delivery-status logging as five structured blocks (application/firehose/
  http/lambda/sqs, each with IAM role refs), and data-protection policies
  (PII/PHI audit/mask/deny) folded as the single-per-topic satellite.
- **SQS queues own both sides of dead-lettering** — `dead_letter_config`
  points a queue's failures AT a DLQ; the new `redrive_allow_policy` on the
  DLQ controls WHO may point at it (`allowAll`/`denyAll`/`byQueue` with 1–10
  source-queue refs, the coupling CEL-enforced).
- **EventBridge buses stop lying across engines** — the Terraform module now
  implements `kms_key_identifier`, `dead_letter_config`, and `log_config`
  (previously Pulumi-only), plus the new `resource_policy` fold for
  cross-account PutEvents grants. Creating a bus named "default" fails fast
  at plan time in BOTH engines with the same message.
- **EventBridge rules invoke real services** — rule-level `role_arn`,
  `force_destroy`, the `ENABLED_WITH_ALL_CLOUDTRAIL_MANAGEMENT_EVENTS` state,
  and five typed target blocks: `sqs_target` (renamed from `sqs_config` for
  provider-authentic naming), `kinesis_target`, `http_target`, `batch_target`,
  and the full `ecs_target` RunTask surface (capacity strategies, awsvpc
  networking, placement strategies/constraints) — at most one per target,
  CEL-enforced.

## Design decisions

- **Split, not fold (SNS subscriptions):** many-per-topic, independent
  lifecycle, and FK edges to SQS/Lambda/Firehose/IAM that a folded block
  buried — the event-source-mapping precedent. The topic's folded
  `subscriptions` block and `subscription_arns` output are removed (breaking).
- **Registry prerequisites stay honest:** `AwsSnsSubscription` declares
  `[AwsSnsTopic]` only; the SQS endpoint is optional composition declared
  per-scenario via the `planton.dev/e2e-prerequisites` annotation.
- **Two constraints documented, not CEL'd, with recorded reasons:**
  schedule-expressions-only-on-the-default-bus (the bus name is a
  `StringValueOrRef` resolving at deploy time) and bus-name ≠ "default"
  (derives from `metadata.name`, invisible to spec CEL) — the latter is an
  identical fail-fast in both IaC modules instead.

## Breaking changes

| Kind | Change |
|------|--------|
| AwsSnsTopic | Folded `subscriptions` block + `subscription_arns` output removed; new fields renumbered contiguously; new outputs `owner`, `beginning_archive_time` |
| AwsEventBridgeRule | Target `sqs_config` renamed `sqs_target`; new top-level `role_arn`, `force_destroy` |
| All five kinds | Legacy `type = any` TF contracts replaced by generator-owned `variables.tf` on the v6 provider floor (invisible to manifests) |

Chart debt recorded for the charts wave (not fixed): `serverless-api`'s
messaging template sets a non-existent `spec.queueType` on AwsSqsQueue and
omits required `region`; any chart carrying inline SNS subscriptions must move
to `AwsSnsSubscription` nodes.

## Live defects found and fixed

- **Struct-typed spec fields treated as JSON strings in five TF modules** —
  the proto→tfvars layer passes `google.protobuf.Struct` values as nested
  objects; modules comparing them to `""` (and jsonencoding nothing) failed
  at apply with "string required". `tofu validate` passes on the `any`-typed
  contract; only the live SNS-topic Terraform lane exposed it. All Struct
  fields now `jsonencode()` behind `!= null` guards; folded into the
  Terraform authoring rule.
- **SNS topic TF crashed on absent `delivery_feedback`** — HCL's
  non-short-circuiting `&&` dereferenced null protocol blocks; the blocks are
  now normalized to zero-value objects once in `locals.tf`
  (`try(coalesce(...), zero)`), and the idiom is in the authoring rule.
- **SQS E2E scenario declared its own kind as a prerequisite** — the harness
  rejects self-references (and a FIFO queue's DLQ must itself be FIFO, so the
  standard fixture was doubly wrong); the scenario is now the full FIFO
  surface with the DLQ arm offline-proven, while SQS-as-DLQ wiring runs live
  in the EventBridge bus lane.
- **All five E2E profiles used `status: pending`** — not a valid
  `CatalogKindE2EProfile` enum value; the profile loader aborts on the first
  invalid profile, breaking CI matrix discovery for the entire AWS provider.

## Validation

- Offline gate green: Ginkgo spec tests for all five kinds (156 cases; the
  rule suite grew from 36 to 58 covering every typed-block CEL), variables.tf
  drift guard, outputs conformance, `tofu validate` × 5, Go + Bazel builds,
  `validate-refs --check`, `secret-coverage --check`, `validate-outputs` × 5,
  `planton e2e discover --provider aws`.
- **Live dual-engine E2E — all ten lanes green:** SQS 2m58s/1m51s, SNS topic
  26s/42s, SNS subscription 3m42s/4m06s (topic prerequisite + composed queue,
  proving the queue-policy coupling live), bus 3m23s/3m42s (DLQ + log_config —
  the two arms of the closed TF parity gap), rule 3m43s/3m58s (event-pattern
  rule on a custom bus with an SQS target + input transformer). All five
  profiles `status: green`.
- Zero-orphan sweep verified clean across topics, subscriptions, queues,
  buses, and rules.
- New harness verifiers: `sns_topic` (GetTopicAttributes), `sns_subscription`
  (GetSubscriptionAttributes), `eventbridge_bus` (DescribeEventBus),
  `eventbridge_rule` (DescribeRule, an OutputsVerifier parsing the bus from
  the rule ARN); sns + eventbridge SDKs added to go.mod + MODULE.bazel.
