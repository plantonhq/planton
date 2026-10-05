# AwsBedrockGuardrail

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

AwsBedrockGuardrailSpec defines the desired configuration for an Amazon
Bedrock guardrail - a set of content-safety policies evaluated on model
inputs and outputs (content filters, denied topics, word filters,
sensitive-information handling, and contextual grounding checks) that
applies uniformly across foundation models, agents, and knowledge bases.

The guardrail's name is taken from `metadata.name` (1-50 characters,
letters/digits/hyphen/underscore only - AWS rejects spaces and dots).

A guardrail always has a mutable working draft (version "DRAFT"). Editing
the spec updates the draft in place; production consumers should pin a
numbered version published through `versions` so draft edits never change
live behavior. Guardrails are free to create - AWS charges per text unit
evaluated at invocation time.

At least one policy family (content_policy, topic_policy, word_policy,
sensitive_information_policy, contextual_grounding_policy) should be
configured - a guardrail that evaluates nothing still intercepts nothing;
AWS validates policy shapes server-side at CreateGuardrail.

Credentials, region, and deployment workflow live outside this spec in
IaC inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockGuardrail
metadata:
  name: test-assistant-guardrail
  id: test-assistant-guardrail
  org: test-org
  env: dev
spec:
  region: us-west-2
  description: Full-surface Bedrock guardrail hack manifest
  blockedInputMessaging: Sorry, I can't help with that request.
  blockedOutputsMessaging: Sorry, I can't provide that response.
  kmsKeyArn:
    value: arn:aws:kms:us-west-2:123456789012:key/abc-123
  # The STANDARD tier requires cross-region inference (AWS rejects it
  # otherwise); the geography-qualified profile id is the portable shape.
  crossRegionProfile: us.guardrail.v1:0
  contentPolicy:
    tier: STANDARD
    filters:
      - type: HATE
        inputStrength: HIGH
        outputStrength: HIGH
      - type: VIOLENCE
        inputStrength: MEDIUM
        outputStrength: MEDIUM
        inputModalities: [TEXT, IMAGE]
        outputModalities: [TEXT]
      - type: PROMPT_ATTACK
        inputStrength: HIGH
        outputStrength: NONE
  topicPolicy:
    topics:
      - name: investment-advice
        definition: Providing investment advice or recommending specific financial products.
        examples:
          - Should I buy this stock?
  wordPolicy:
    profanityFilter: {}
    customWords:
      - text: codename-atlas
  sensitiveInformationPolicy:
    piiEntities:
      - type: EMAIL
        action: ANONYMIZE
      - type: US_SOCIAL_SECURITY_NUMBER
        action: BLOCK
    regexes:
      - name: employee-id
        pattern: EMP-[0-9]{6}
        description: Internal employee identifiers
        action: ANONYMIZE
  contextualGroundingPolicy:
    filters:
      - type: GROUNDING
        threshold: 0.75
      - type: RELEVANCE
        threshold: 0.5
  versions:
    - name: prod
      description: Initial production pin
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.blockedInputMessaging` | `string` | yes |  |  |
| `spec.blockedOutputsMessaging` | `string` | yes |  |  |
| `spec.kmsKeyArn` | `string \| valueFrom` |  |  | AwsKmsKey (`status.outputs.key_arn`) |
| `spec.contentPolicy` | `AwsBedrockGuardrailContentPolicy` |  |  |  |
| `spec.contentPolicy.tier` | `string` |  |  |  |
| `spec.contentPolicy.filters` | `[]AwsBedrockGuardrailContentFilter` | yes |  |  |
| `spec.contentPolicy.filters[].type` | `string` |  |  |  |
| `spec.contentPolicy.filters[].inputStrength` | `string` |  |  |  |
| `spec.contentPolicy.filters[].outputStrength` | `string` |  |  |  |
| `spec.contentPolicy.filters[].inputAction` | `string` |  |  |  |
| `spec.contentPolicy.filters[].outputAction` | `string` |  |  |  |
| `spec.contentPolicy.filters[].inputEnabled` | `bool` |  |  |  |
| `spec.contentPolicy.filters[].outputEnabled` | `bool` |  |  |  |
| `spec.contentPolicy.filters[].inputModalities` | `[]string` |  |  |  |
| `spec.contentPolicy.filters[].outputModalities` | `[]string` |  |  |  |
| `spec.topicPolicy` | `AwsBedrockGuardrailTopicPolicy` |  |  |  |
| `spec.topicPolicy.tier` | `string` |  |  |  |
| `spec.topicPolicy.topics` | `[]AwsBedrockGuardrailTopic` | yes |  |  |
| `spec.topicPolicy.topics[].name` | `string` | yes |  |  |
| `spec.topicPolicy.topics[].definition` | `string` | yes |  |  |
| `spec.topicPolicy.topics[].examples` | `[]string` |  |  |  |
| `spec.wordPolicy` | `AwsBedrockGuardrailWordPolicy` |  |  |  |
| `spec.wordPolicy.profanityFilter` | `AwsBedrockGuardrailManagedWordList` |  |  |  |
| `spec.wordPolicy.profanityFilter.inputAction` | `string` |  |  |  |
| `spec.wordPolicy.profanityFilter.outputAction` | `string` |  |  |  |
| `spec.wordPolicy.profanityFilter.inputEnabled` | `bool` |  |  |  |
| `spec.wordPolicy.profanityFilter.outputEnabled` | `bool` |  |  |  |
| `spec.wordPolicy.customWords` | `[]AwsBedrockGuardrailCustomWord` |  |  |  |
| `spec.wordPolicy.customWords[].text` | `string` | yes |  |  |
| `spec.wordPolicy.customWords[].inputAction` | `string` |  |  |  |
| `spec.wordPolicy.customWords[].outputAction` | `string` |  |  |  |
| `spec.wordPolicy.customWords[].inputEnabled` | `bool` |  |  |  |
| `spec.wordPolicy.customWords[].outputEnabled` | `bool` |  |  |  |
| `spec.sensitiveInformationPolicy` | `AwsBedrockGuardrailSensitiveInformationPolicy` |  |  |  |
| `spec.sensitiveInformationPolicy.piiEntities` | `[]AwsBedrockGuardrailPiiEntity` |  |  |  |
| `spec.sensitiveInformationPolicy.piiEntities[].type` | `string` |  |  |  |
| `spec.sensitiveInformationPolicy.piiEntities[].action` | `string` |  |  |  |
| `spec.sensitiveInformationPolicy.piiEntities[].inputAction` | `string` |  |  |  |
| `spec.sensitiveInformationPolicy.piiEntities[].outputAction` | `string` |  |  |  |
| `spec.sensitiveInformationPolicy.piiEntities[].inputEnabled` | `bool` |  |  |  |
| `spec.sensitiveInformationPolicy.piiEntities[].outputEnabled` | `bool` |  |  |  |
| `spec.sensitiveInformationPolicy.regexes` | `[]AwsBedrockGuardrailRegex` |  |  |  |
| `spec.sensitiveInformationPolicy.regexes[].name` | `string` | yes |  |  |
| `spec.sensitiveInformationPolicy.regexes[].pattern` | `string` | yes |  |  |
| `spec.sensitiveInformationPolicy.regexes[].description` | `string` |  |  |  |
| `spec.sensitiveInformationPolicy.regexes[].action` | `string` |  |  |  |
| `spec.sensitiveInformationPolicy.regexes[].inputAction` | `string` |  |  |  |
| `spec.sensitiveInformationPolicy.regexes[].outputAction` | `string` |  |  |  |
| `spec.sensitiveInformationPolicy.regexes[].inputEnabled` | `bool` |  |  |  |
| `spec.sensitiveInformationPolicy.regexes[].outputEnabled` | `bool` |  |  |  |
| `spec.contextualGroundingPolicy` | `AwsBedrockGuardrailContextualGroundingPolicy` |  |  |  |
| `spec.contextualGroundingPolicy.filters` | `[]AwsBedrockGuardrailContextualGroundingFilter` | yes |  |  |
| `spec.contextualGroundingPolicy.filters[].type` | `string` |  |  |  |
| `spec.contextualGroundingPolicy.filters[].threshold` | `double` |  |  |  |
| `spec.crossRegionProfile` | `string` |  |  |  |
| `spec.versions` | `[]AwsBedrockGuardrailVersion` |  |  |  |
| `spec.versions[].name` | `string` | yes |  |  |
| `spec.versions[].description` | `string` |  |  |  |
| `spec.versions[].keepOnDelete` | `bool` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the guardrail will be created.
Example: "us-west-2", "us-east-1"

- rule: {"string":{"minLen":"1"}}

### spec.description

`string`

Human-readable description shown in the Bedrock console and
GetGuardrail responses (up to 200 characters). Updates in place.

- rule: {"string":{"maxLen":"200"}}

### spec.blockedInputMessaging

`string` · required

Message returned to the application when the guardrail BLOCKS a model
INPUT (1-500 characters). Required by AWS for every guardrail.
Example: "Sorry, I can't help with that request."

- rule: {"string":{"minLen":"1","maxLen":"500"}}

### spec.blockedOutputsMessaging

`string` · required

Message returned to the application when the guardrail BLOCKS a model
OUTPUT (1-500 characters). Required by AWS for every guardrail.

- rule: {"string":{"minLen":"1","maxLen":"500"}}

### spec.kmsKeyArn

`string | valueFrom`

Customer-managed KMS key ARN for encrypting the guardrail at rest.
Without it, AWS uses a Bedrock-managed key. Updates in place.

- references: AwsKmsKey (`status.outputs.key_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_arn}} -- a bare string does not parse

### spec.contentPolicy

`AwsBedrockGuardrailContentPolicy`

Content filters - strength-tiered detection of harmful content
categories (sexual, violence, hate, insults, misconduct, prompt
attacks) on model inputs and outputs.

- rule: each content filter type may appear at most once

### spec.contentPolicy.tier

`string`

Safeguard tier for content filters. STANDARD enables the 2025
cross-lingual tier (more languages, robustness levels); CLASSIC is the
original English-centric tier. Omitted = AWS default (CLASSIC).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["CLASSIC","STANDARD"]}}

### spec.contentPolicy.filters

`[]AwsBedrockGuardrailContentFilter` · required

One filter per harmful-content category. AWS evaluates each configured
category at the configured strengths.

- rule: {"repeated":{"minItems":"1"}}

### spec.contentPolicy.filters[].type

`string`

The harmful-content category this filter detects. PROMPT_ATTACK
detects jailbreak/prompt-injection attempts (input side only in
practice - AWS requires its output strength to be NONE).

- rule: {"string":{"in":["SEXUAL","VIOLENCE","HATE","INSULTS","MISCONDUCT","PROMPT_ATTACK"]}}

### spec.contentPolicy.filters[].inputStrength

`string`

Detection strength for model INPUTS: NONE, LOW, MEDIUM, or HIGH.
Higher strengths block more aggressively.

- rule: {"string":{"in":["NONE","LOW","MEDIUM","HIGH"]}}

### spec.contentPolicy.filters[].outputStrength

`string`

Detection strength for model OUTPUTS: NONE, LOW, MEDIUM, or HIGH.

- rule: {"string":{"in":["NONE","LOW","MEDIUM","HIGH"]}}

### spec.contentPolicy.filters[].inputAction

`string`

What to do when the filter matches an INPUT: BLOCK (default) rejects
the request with blocked_input_messaging; NONE detects and reports
(visible in trace/observability) without intervening. Sent only when
set - omitted lets AWS default to BLOCK.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["BLOCK","NONE"]}}

### spec.contentPolicy.filters[].outputAction

`string`

What to do when the filter matches an OUTPUT: BLOCK (default) or NONE
(detect-only). Sent only when set.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["BLOCK","NONE"]}}

### spec.contentPolicy.filters[].inputEnabled

`bool` · optional (explicit presence)

Evaluate this filter on inputs. AWS defaults to true; set false
explicitly to skip input evaluation while keeping output evaluation
(both engines send the explicit value whenever this field is set).

### spec.contentPolicy.filters[].outputEnabled

`bool` · optional (explicit presence)

Evaluate this filter on outputs. AWS defaults to true; set false
explicitly to skip output evaluation.

### spec.contentPolicy.filters[].inputModalities

`[]string`

Input modalities the filter evaluates: TEXT and/or IMAGE. Omitted =
AWS default (TEXT). IMAGE requires a model with image understanding.

- rule: {"repeated":{"items":{"string":{"in":["TEXT","IMAGE"]}}}}

### spec.contentPolicy.filters[].outputModalities

`[]string`

Output modalities the filter evaluates: TEXT and/or IMAGE. Omitted =
AWS default (TEXT).

- rule: {"repeated":{"items":{"string":{"in":["TEXT","IMAGE"]}}}}

### spec.topicPolicy

`AwsBedrockGuardrailTopicPolicy`

Denied topics - natural-language topic definitions the model must not
engage with, each with optional example phrases.

- rule: topic names must be unique

### spec.topicPolicy.tier

`string`

Safeguard tier for topic evaluation. STANDARD enables the 2025
cross-lingual tier; CLASSIC is the original tier. Omitted = AWS
default (CLASSIC).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["CLASSIC","STANDARD"]}}

### spec.topicPolicy.topics

`[]AwsBedrockGuardrailTopic` · required

Topics the model must not engage with (at least one). All topics are
DENY topics - AWS defines no other topic type; the modules send the
type constant.

- rule: {"repeated":{"minItems":"1"}}

### spec.topicPolicy.topics[].name

`string` · required

Topic name shown in the console and in guardrail traces (1-100
characters; letters, digits, space, hyphen, underscore, !?.).

- rule: {"string":{"minLen":"1","maxLen":"100","pattern":"^[0-9a-zA-Z-_ !?.]+$"}}

### spec.topicPolicy.topics[].definition

`string` · required

Natural-language definition of the topic (1-1000 characters). Write it
the way you would brief a human reviewer - AWS classifies against this
definition. Example: "Providing investment advice or recommending
specific financial products."

- rule: {"string":{"minLen":"1","maxLen":"1000"}}

### spec.topicPolicy.topics[].examples

`[]string`

Optional example phrases that belong to the topic (each 1-100
characters) - they sharpen classification near the boundary.

- rule: {"repeated":{"items":{"string":{"minLen":"1","maxLen":"100"}}}}

### spec.wordPolicy

`AwsBedrockGuardrailWordPolicy`

Word filters - the AWS-managed profanity list and/or exact custom
words/phrases to intercept.

- rule: custom_words entries must have unique text
- rule: word_policy must enable profanity_filter and/or define custom_words

### spec.wordPolicy.profanityFilter

`AwsBedrockGuardrailManagedWordList`

Enable the AWS-managed profanity list (the only managed list AWS
defines; the modules send its PROFANITY type constant). Presence of
this message enables the list; the action/enabled arms tune it.

### spec.wordPolicy.profanityFilter.inputAction

`string`

What to do on an INPUT match: BLOCK (default) or NONE (detect-only).
Sent only when set.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["BLOCK","NONE"]}}

### spec.wordPolicy.profanityFilter.outputAction

`string`

What to do on an OUTPUT match: BLOCK (default) or NONE (detect-only).
Sent only when set.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["BLOCK","NONE"]}}

### spec.wordPolicy.profanityFilter.inputEnabled

`bool` · optional (explicit presence)

Evaluate the list on inputs. AWS defaults to true; explicit false
skips input evaluation (sent whenever set, both engines).

### spec.wordPolicy.profanityFilter.outputEnabled

`bool` · optional (explicit presence)

Evaluate the list on outputs. AWS defaults to true.

### spec.wordPolicy.customWords

`[]AwsBedrockGuardrailCustomWord`

Custom words and phrases to intercept (each up to 3 words long per
AWS's matching model; at most 10,000 entries).

- rule: {"repeated":{"maxItems":"10000"}}

### spec.wordPolicy.customWords[].text

`string` · required

The exact word or phrase to match (case-insensitive; up to three
words).

- rule: {"string":{"minLen":"1"}}

### spec.wordPolicy.customWords[].inputAction

`string`

What to do on an INPUT match: BLOCK (default) or NONE (detect-only).
Sent only when set.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["BLOCK","NONE"]}}

### spec.wordPolicy.customWords[].outputAction

`string`

What to do on an OUTPUT match: BLOCK (default) or NONE (detect-only).
Sent only when set.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["BLOCK","NONE"]}}

### spec.wordPolicy.customWords[].inputEnabled

`bool` · optional (explicit presence)

Evaluate this word on inputs. AWS defaults to true.

### spec.wordPolicy.customWords[].outputEnabled

`bool` · optional (explicit presence)

Evaluate this word on outputs. AWS defaults to true.

### spec.sensitiveInformationPolicy

`AwsBedrockGuardrailSensitiveInformationPolicy`

Sensitive-information handling - PII entity types and custom regex
patterns, each blocked or anonymized (masked) in model I/O.

- rule: each PII entity type may appear at most once
- rule: regex names must be unique
- rule: sensitive_information_policy must define pii_entities and/or regexes

### spec.sensitiveInformationPolicy.piiEntities

`[]AwsBedrockGuardrailPiiEntity`

PII entity types to detect, each with its own action.

### spec.sensitiveInformationPolicy.piiEntities[].type

`string`

The PII entity type to detect (AWS's closed catalog of general,
finance, IT, US-, CA-, and UK-specific identifiers).

- rule: {"string":{"in":["ADDRESS","AGE","AWS_ACCESS_KEY","AWS_SECRET_KEY","CA_HEALTH_NUMBER","CA_SOCIAL_INSURANCE_NUMBER","CREDIT_DEBIT_CARD_CVV","CREDIT_DEBIT_CARD_EXPIRY","CREDIT_DEBIT_CARD_NUMBER","DRIVER_ID","EMAIL","INTERNATIONAL_BANK_ACCOUNT_NUMBER","IP_ADDRESS","LICENSE_PLATE","MAC_ADDRESS","NAME","PASSWORD","PHONE","PIN","SWIFT_CODE","UK_NATIONAL_HEALTH_SERVICE_NUMBER","UK_NATIONAL_INSURANCE_NUMBER","UK_UNIQUE_TAXPAYER_REFERENCE_NUMBER","URL","USERNAME","US_BANK_ACCOUNT_NUMBER","US_BANK_ROUTING_NUMBER","US_INDIVIDUAL_TAX_IDENTIFICATION_NUMBER","US_PASSPORT_NUMBER","US_SOCIAL_SECURITY_NUMBER","VEHICLE_IDENTIFICATION_NUMBER"]}}

### spec.sensitiveInformationPolicy.piiEntities[].action

`string`

Base action when this entity is detected: BLOCK rejects the request/
response; ANONYMIZE masks the entity (e.g. {NAME}) and lets the
request through; NONE detects and reports without intervening.

- rule: {"string":{"in":["BLOCK","ANONYMIZE","NONE"]}}

### spec.sensitiveInformationPolicy.piiEntities[].inputAction

`string`

Override the action for INPUTS only. AWS materializes a default from
`action` when omitted; once set on a created guardrail this cannot be
reverted to "AWS-derived" - only to another explicit value.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["BLOCK","ANONYMIZE","NONE"]}}

### spec.sensitiveInformationPolicy.piiEntities[].outputAction

`string`

Override the action for OUTPUTS only. Same materialize-once semantics
as input_action.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["BLOCK","ANONYMIZE","NONE"]}}

### spec.sensitiveInformationPolicy.piiEntities[].inputEnabled

`bool` · optional (explicit presence)

Evaluate this entity on inputs. AWS defaults to true.

### spec.sensitiveInformationPolicy.piiEntities[].outputEnabled

`bool` · optional (explicit presence)

Evaluate this entity on outputs. AWS defaults to true.

### spec.sensitiveInformationPolicy.regexes

`[]AwsBedrockGuardrailRegex`

Custom regex patterns for organization-specific identifiers (employee
IDs, ticket numbers, ...), each with its own action.

### spec.sensitiveInformationPolicy.regexes[].name

`string` · required

Pattern name shown in traces and anonymization masks (1-100
characters).

- rule: {"string":{"minLen":"1","maxLen":"100"}}

### spec.sensitiveInformationPolicy.regexes[].pattern

`string` · required

The regular expression (RE2-compatible; AWS validates server-side).

- rule: {"string":{"minLen":"1"}}

### spec.sensitiveInformationPolicy.regexes[].description

`string`

Optional description of what the pattern captures (1-1000 characters
when set).

- rule: {"string":{"maxLen":"1000"}}

### spec.sensitiveInformationPolicy.regexes[].action

`string`

Base action on a match: BLOCK, ANONYMIZE (mask), or NONE
(detect-only).

- rule: {"string":{"in":["BLOCK","ANONYMIZE","NONE"]}}

### spec.sensitiveInformationPolicy.regexes[].inputAction

`string`

Override the action for INPUTS only (materialize-once, like PII
entities).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["BLOCK","ANONYMIZE","NONE"]}}

### spec.sensitiveInformationPolicy.regexes[].outputAction

`string`

Override the action for OUTPUTS only.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["BLOCK","ANONYMIZE","NONE"]}}

### spec.sensitiveInformationPolicy.regexes[].inputEnabled

`bool` · optional (explicit presence)

Evaluate this pattern on inputs. AWS defaults to true.

### spec.sensitiveInformationPolicy.regexes[].outputEnabled

`bool` · optional (explicit presence)

Evaluate this pattern on outputs. AWS defaults to true.

### spec.contextualGroundingPolicy

`AwsBedrockGuardrailContextualGroundingPolicy`

Contextual grounding - thresholds that reject model responses that are
not grounded in the provided source material (GROUNDING) or not
relevant to the user's query (RELEVANCE). Applies to RAG and
summarization flows that supply grounding sources at invocation.

- rule: each contextual grounding filter type may appear at most once

### spec.contextualGroundingPolicy.filters

`[]AwsBedrockGuardrailContextualGroundingFilter` · required

One filter per check type (at least one).

- rule: {"repeated":{"minItems":"1"}}

### spec.contextualGroundingPolicy.filters[].type

`string`

The check type: GROUNDING scores whether the response is supported by
the source material supplied at invocation; RELEVANCE scores whether
the response addresses the user's query.

- rule: {"string":{"in":["GROUNDING","RELEVANCE"]}}

### spec.contextualGroundingPolicy.filters[].threshold

`double`

Confidence threshold in [0, 1]. Responses scoring BELOW the threshold
are blocked. 0 blocks nothing; values near 1 block aggressively
(0.99 is the practical maximum AWS documents).

- rule: {"double":{"lte":1,"gte":0}}

### spec.crossRegionProfile

`string`

The AWS system-defined guardrail profile that lets this guardrail
evaluate traffic routed through cross-region inference. Accepts the
geography-qualified profile identifier (e.g. "us.guardrail.v1:0" -
the portable shape; both modules compose it into the deploying
account's profile ARN at deploy time, live-verified 2026-08-13) or
the full profile ARN
(arn:aws:bedrock:<region>:<account>:guardrail-profile/<id>). AWS
defines the available profiles per geography; this cannot reference
a customer resource. REQUIRED whenever any policy family sets the
STANDARD tier - AWS rejects STANDARD without cross-region inference
("Can't configure guardrail policy tier. Enable cross-Region
inference for your guardrail to use Standard tier.").

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"pattern":"^([a-z]{2,4}\\.guardrail\\.v[0-9]+:[0-9]+|arn:aws[a-z-]*:bedrock:[a-z0-9-]+:[0-9]{12}:guardrail-profile/.+)$"}}

### spec.versions

`[]AwsBedrockGuardrailVersion`

Immutable numbered versions published from the guardrail's current
draft definition. Each entry publishes ONE version; AWS assigns the
version number sequentially (1, 2, ...) - the entry's `name` is a
stable local key only, never sent to AWS. The published number for
each entry is exported in the `version_numbers` output map keyed by
that name. Entries publish AFTER the draft update in the same deploy,
so a spec edit plus a new entry captures the edited definition.
Removing an entry deletes that published version (unless
`keep_on_delete` is set); versions in use by agents fail deletion
server-side.

### spec.versions[].name

`string` · required

Stable local key for this entry (1-100 characters; letters, digits,
hyphen, underscore). Used as the for_each key on both engines and as
the key in the `version_numbers` output map - NEVER sent to AWS.
Renaming an entry destroys and republishes its version under the new
key.

- rule: {"string":{"minLen":"1","maxLen":"100","pattern":"^[0-9a-zA-Z-_]+$"}}

### spec.versions[].description

`string`

Optional description recorded on the published version (up to 200
characters). Changing it destroys and republishes the version - AWS
version descriptions are create-time-immutable.

- rule: {"string":{"maxLen":"200"}}

### spec.versions[].keepOnDelete

`bool`

Keep the published version in AWS when this entry is removed (or the
guardrail destroyed) instead of deleting it. The version then lives
outside this manifest's management.

## Validation Rules

- `version_names_unique`: versions entries must have unique names
- `content_standard_tier_requires_cross_region`: content_policy.tier STANDARD requires cross_region_profile (AWS rejects the Standard tier without cross-region inference)
- `topic_standard_tier_requires_cross_region`: topic_policy.tier STANDARD requires cross_region_profile (AWS rejects the Standard tier without cross-region inference)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsBedrockGuardrail, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.guardrail_id` | `string` | The unique guardrail identifier (e.g. "gr-abc123..."). The join key model invocations and agents use together with a version. |
| `status.outputs.guardrail_arn` | `string` | The Amazon Resource Name of the guardrail - the canonical key for IAM policies and cross-account references. |
| `status.outputs.draft_version` | `string` | The guardrail's mutable working version - always the literal "DRAFT". Production consumers should pin an entry from `version_numbers` instead. |
| `status.outputs.version_numbers` | `map<string, string>` | AWS-assigned numbers of the versions published through spec.versions, keyed by each entry's `name`. Example: {"prod": "1", "canary": "2"}. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.kmsKeyArn` | AwsKmsKey | `status.outputs.key_arn` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsBedrockAgent | `spec.guardrail.guardrailId` | `status.outputs.guardrail_id` |
| AwsBedrockFlow | `spec.definition.nodes[].prompt.guardrail.guardrailId` | `status.outputs.guardrail_id` |
| AwsBedrockFlow | `spec.definition.nodes[].knowledgeBase.guardrail.guardrailId` | `status.outputs.guardrail_id` |

## See Also

- [Overview](../README.md)
