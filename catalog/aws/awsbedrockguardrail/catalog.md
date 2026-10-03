# AWS Bedrock Guardrail

Deploys an Amazon Bedrock guardrail — content-safety policies evaluated on every model input and output, applied uniformly across foundation models, agents, and flows. The policy families cover content filters (six harm categories at four strengths), denied topics defined in natural language, word filters, sensitive-information handling (31 PII entity types plus custom regexes, each blocked or anonymized), and contextual grounding thresholds. Editing the spec updates the mutable DRAFT in place; the `versions` entries publish immutable numbered versions for production pinning, and the cost driver is text units evaluated at invocation time.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Bedrock Guardrail** — the guardrail with every policy family you declare: content filters, denied topics, the managed profanity list and custom words, PII entities and regex patterns, and grounding/relevance thresholds, plus the required blocked-input and blocked-output messages
- **Published Versions** — created only when `versions` entries exist: one immutable numbered version per entry, published AFTER the draft update in the same deploy so each capture includes the current edit; AWS assigns the numbers, exported in `version_numbers` keyed by your entry names

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with Bedrock guardrail permissions (`bedrock:CreateGuardrail`, `bedrock:CreateGuardrailVersion`, and their read/update/delete siblings). Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- **Bedrock available in the target region** — guardrails are supported in all Bedrock commercial regions.
- **A customer-managed KMS key with Bedrock key-use permissions** — only when the guardrail definition must be encrypted under your own key (referenced by `kmsKeyArn`).
- **Cross-region inference profile availability** — only when any policy family uses the STANDARD tier; `crossRegionProfile` names an AWS system-defined profile (e.g. `us.guardrail.v1:0`), not a customer resource.

## Deploy

### Console

Open the deployment store, find **AWS Bedrock Guardrail**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: region, blocked messagings, and the policy families. Start from the **Content Safety Baseline** preset in the [Presets](#presets) tab for the all-categories starting point.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsBedrockGuardrail
metadata:
  name: assistant-guardrail
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  blockedInputMessaging: Sorry, I can't help with that request.
  blockedOutputsMessaging: Sorry, I can't provide that response.
  contentPolicy:
    filters:
      - type: HATE
        inputStrength: HIGH
        outputStrength: HIGH
  sensitiveInformationPolicy:
    piiEntities:
      - type: EMAIL
        action: ANONYMIZE
  versions:
    - name: prod
```

```shell
planton apply -f guardrail.yaml
```

This creates a guardrail that blocks hate content at high strength, masks email addresses, and publishes version 1 under the `prod` entry for consumers to pin. An Infra Job tracks the provisioning in real time.

### InfraChart

When the guardrail deploys alongside its KMS key in one chart, wire the reference via ValueFromRef:

```yaml
spec:
  region: us-west-2
  blockedInputMessaging: Sorry, I can't help with that request.
  blockedOutputsMessaging: Sorry, I can't provide that response.
  kmsKeyArn:
    valueFrom:
      kind: AwsKmsKey
      name: guardrail-key
      fieldPath: status.outputs.key_arn
  contentPolicy:
    filters:
      - type: HATE
        inputStrength: HIGH
        outputStrength: HIGH
```

The InfraPipeline resolves the dependency graph, deploys the key first, then creates the guardrail encrypted under it.

## Key Configuration

These are the most important decisions when configuring a guardrail. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Pin versions in production** — the spec edits the mutable DRAFT; agents and applications should reference a number from `version_numbers`, so a draft edit never changes live behavior unannounced. The deployment story is: publish a new `versions` entry, move consumers to the new number, then retire the old entry.

**Trial new policies detect-only** — set `inputAction`/`outputAction` to NONE to observe matches in traces against real traffic without intervening, then flip to BLOCK when the false-positive rate is acceptable. This works per content filter, per word, per PII entity, and per regex.

**Blocked messagings are user-facing copy** — `blockedInputMessaging` and `blockedOutputsMessaging` are returned verbatim to the calling application whenever the guardrail intervenes. Write them as product copy, not debug strings.

**BLOCK or ANONYMIZE for sensitive information** — BLOCK rejects the whole request or response; ANONYMIZE masks the entity (e.g. `{NAME}`) and lets the text through, which is usually right for assistants that must keep working while staying compliant. One caution: the per-direction overrides (`inputAction`/`outputAction` on PII entities and regexes) are materialize-once — set on a created guardrail, they can move to another explicit value but never revert to AWS-derived.

**The STANDARD tier requires cross-region inference** — AWS rejects a Standard-tier content or topic policy without `crossRegionProfile`; the manifest validation front-loads that contract. Prefer the geography-qualified identifier shape (`us.guardrail.v1:0`) — it never embeds an account ID, so the manifest stays portable; the modules compose it into the deploying account's profile ARN.

**PROMPT_ATTACK is input-side** — AWS requires its `outputStrength` to be NONE: the filter classifies user input for jailbreak attempts, and model output is not a prompt.

**Contextual grounding only fires with sources** — the GROUNDING and RELEVANCE thresholds evaluate invocations that supply source content (RAG flows, Converse with grounding). Responses scoring below the threshold are blocked: 0 blocks nothing, values near 1 block aggressively, and around 0.75 is a common GROUNDING starting point.

**Version entries are keyed by name, numbered by AWS** — the entry `name` is local identity (the key in `version_numbers`), never sent to AWS. Removing an entry deletes that published version unless `keepOnDelete` is set; deleting a version still in use by an agent fails server-side — repoint the consumer first.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsKmsKey** | `kmsKeyArn` | `status.outputs.key_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `guardrail_id` | The unique guardrail identifier | An AwsBedrockAgent's `guardrail.guardrailId`; a flow node's guardrail reference — always paired with a version |
| `guardrail_arn` | The guardrail's ARN | IAM policies and cross-account references |
| `version_numbers` | AWS-assigned version numbers keyed by each `versions` entry's name | The version a production consumer pins alongside `guardrail_id` |

`draft_version` is the constant "DRAFT" — a record, not what production consumers should pin.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Content safety baseline** — all six harm categories filtered on inputs and outputs, with a published version for consumers to pin. The right first guardrail for any user-facing assistant; tune per-category strengths as trace data accumulates. Start from the **Content Safety Baseline** preset.

**PII redaction** — ANONYMIZE on the identity-bearing entity types so conversations keep flowing with masked values, BLOCK reserved for credentials-class entities (AWS keys, passwords) where masking is not enough. Add custom regexes for organization-specific identifiers like employee IDs. Start from the **PII Redaction** preset.

**Topic denylist for RAG** — denied topics written as natural-language definitions plus GROUNDING/RELEVANCE thresholds, for retrieval-augmented assistants that must stay on-domain and answer only from their sources. Definition quality is the lever: write each topic the way you would brief a human reviewer, and add boundary examples. Start from the **Topic Denylist for RAG** preset.

## Works With

- [**AWS Bedrock Agent**](/infra-catalog/aws-bedrock-agent) — attaches this guardrail (at a pinned version) to every model input and output the agent produces
- [**AWS Bedrock Flow**](/infra-catalog/aws-bedrock-flow) — pins this guardrail onto prompt and knowledge-base nodes
- [**AWS KMS Key**](/infra-catalog/aws-kms-key) — customer-managed encryption of the guardrail definition via `kmsKeyArn`
