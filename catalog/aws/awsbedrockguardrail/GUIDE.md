# AwsBedrockGuardrail — Kind Guide

Authored operational judgment for the Bedrock guardrail kind: the
design decisions behind the spec's shape, and what to know before running
guardrails in production.

## Design decisions

- **The policy families drop the provider's `_config` suffix.** The
  provider spells every family `*_policy_config` with nested
  `filters_config`/`topics_config`/`words_config` lists; the spec carries
  `content_policy.filters`, `topic_policy.topics`, `word_policy.custom_words`
  — the suffix is provider plumbing, not meaning. The parity manifest
  records every rename.
- **One-value vocabularies are module constants.** AWS defines exactly one
  topic type (`DENY`) and one managed word list (`PROFANITY`). The spec
  never asks for a field with one legal value: topics are simply topics,
  and the profanity list is a presence-typed `profanity_filter` message.
  If AWS ever adds a second value, the spec grows the field THEN — with
  the change visible in the parity gate the day the pin advances.
- **The safeguard tier is a leaf, not a block.** The provider nests
  `tier_config { tier_name }` single-entry lists inside the content and
  topic families; the spec carries `tier: CLASSIC|STANDARD` directly.
  Omitted means AWS's default (CLASSIC) — the modules then send nothing
  and pin whatever AWS materializes.
- **Versions are name-keyed entries, numbered by AWS.** A published
  guardrail version has no user-assignable identity — AWS numbers them
  sequentially. Each `versions` entry's `name` is a stable local key (the
  for_each key on both engines, never sent to AWS) and the AWS-assigned
  number lands in the `version_numbers` output map under that key. This is
  the same fold idiom as EventBridge archives and Redshift endpoint
  accesses: local identity for graph stability, cloud identity in outputs.
- **Versions publish AFTER the draft update in the same deploy.** Both
  engines order every version resource behind the guardrail resource, so a
  spec edit plus a new entry captures the edited definition — publish-then-
  edit races are structurally impossible within one deploy.
- **Per-direction action/enabled arms are send-when-set.** AWS defaults
  actions to BLOCK and enabled to true. The spec's optional arms transmit
  whenever set — including explicit false — so disabling one direction is
  expressible (the un-disable class). The PII/regex per-direction arms are
  Optional+Computed at the provider: once set on a created guardrail they
  never revert to "AWS-derived", only to another explicit value (taught on
  the spec fields).

## Operational judgment

- **Draft vs. published versions is the guardrail's deployment story.**
  Treat the DRAFT as your staging surface and pin production consumers
  (agents, application inference configs) to a published number. Publish a
  new entry, move consumers, then retire the old entry.
- **Trial new policies detect-only.** `input_action: NONE` /
  `output_action: NONE` reports matches in traces without intervening —
  run a new topic or filter that way against real traffic before flipping
  to BLOCK.
- **Blocked messagings are user-facing copy.** They are returned verbatim
  to the calling application whenever the guardrail intervenes — write
  them as product copy, not debug strings.
- **PROMPT_ATTACK output strength must be NONE** (AWS contract): the
  filter classifies user input for jailbreak attempts; model output is not
  a prompt.
- **Contextual grounding needs grounding sources at invocation.** The
  GROUNDING/RELEVANCE thresholds only fire on invocations that supply
  source content (RAG flows, Converse with grounding). A threshold of 0
  blocks nothing; ~0.75 is a common starting point for GROUNDING.
- **Deleting a version in use by an agent fails server-side.** Detach or
  repoint the consumer first, or mark the entry `keep_on_delete` and
  manage the version outside the manifest.

## Coverage decisions

- Every configurable argument of `aws_bedrock_guardrail` and
  `aws_bedrock_guardrail_version` at the pinned provider is modeled,
  mapped, or excluded with a reason in `iac/provider-parity.yaml` (zero
  findings at forge time).
- `cross_region_profile` accepts the geography-qualified profile
  identifier ("us.guardrail.v1:0") or the full profile ARN. The AWS API
  resolves either (live-verified 2026-08-13; GetGuardrail echoes both the
  id and the account-resolved ARN), but the upstream provider's schema
  demands an ARN — so both modules compose the id shape into the
  deploying account's profile ARN at deploy time (a caller-identity
  lookup). The identifier is the portable shape for committed manifests:
  it never embeds an account id. The STANDARD policy tier REQUIRES the
  field: AWS rejects CreateGuardrail with a Standard tier and no
  cross-region config, and the spec CEL front-loads that contract at
  manifest time.
- The version entry's `name` never reaches AWS — it is for_each identity
  and the `version_numbers` output key (recorded as a specExclusion).
