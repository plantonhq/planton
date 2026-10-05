# AwsBedrockAgentCoreEvaluation — Kind Guide

Authored operational judgment for the AgentCore Evaluations kind:
the design decisions behind the spec's shape, and what to know before
scoring agents in production.

## Design decisions

- **One bundle, three arms.** Evaluators, harnesses, and online
  configs are AWS-standalone resources, but a team provisions them
  together — name-keyed collections, each arm optional (at least one,
  CEL-guarded).
- **Harness model vendors are exactly-one.** Bedrock, Gemini, and
  OpenAI are mutually exclusive; the spec rejects a silent first-arm
  pick the provider would otherwise make.
- **Rating scales are exactly-one.** Categorical and numerical are
  alternatives; mixing them is a manifest error, not an AWS surprise.
- **Runtime environment is optional and computed.** A harness does
  not need an agent runtime; when omitted, AWS assigns the default.
  The explicit arm is there for the day you pin one.
- **Float32 leaves are import-normalized.** The harness's model
  temperature / top-p values, the summarization truncation ratio, and
  the memory retrieval relevance score are Float32 upstream — the
  import catalog writes normalized attributes for exactly those leaves
  so a no-op apply stays a no-op. Integer parameters (max-tokens) and
  the evaluator's Float64 inference values need no tolerance and carry
  none.

## Scoring agents in production

- **Harnesses finish at READY; evaluators and online configs at
  ACTIVE.** The asymmetry is AWS's, not a bug: a harness that sits in
  CREATING and lands READY is done, and a stuck deploy is diagnosed
  against the right terminal word per arm.
- **Applies can legitimately take up to 30 minutes per resource.**
  All three resource types carry 30-minute create/update/delete
  waiters upstream. A slow apply is AWS provisioning, not a hang.
- **An evaluator in use by an active online config is LOCKED.**
  Editing or destroying an evaluator that an online config references
  waits on AWS's conflict retries and then fails. Disable or delete
  the referencing online config first; the day-two symptom is an
  apply that hangs on the evaluator and errors with a conflict.
- **Start with a code evaluator.** Create does not invoke the Lambda,
  so the first evaluator is fixture-cheap and independent of Bedrock
  model access.
- **Give the judge an inference-profile model id.** CreateEvaluator validates the judge model against the region's INFERENCE set — a bare on-demand foundation-model id fails with "not available in region" even when that model is regionally listed, while the cross-region profile form (`us.amazon.nova-2-lite-v1:0`) creates cleanly. The harness's model field has no such create-time gate (it validated nothing at create in live runs); its model is exercised only when a run executes.
- **Judge instructions must embed the level's placeholders.** Each evaluator level requires at least one of its allowed single-brace placeholders in the instructions — for SESSION the set is `{available_tools}`, `{context}`, `{actual_tool_trajectory}`, `{expected_tool_trajectory}`, `{assertions}` — and CreateEvaluator rejects a plain-prose prompt naming the set. The spec front-loads the SESSION contract as a CEL.
- **A memory-less harness still has a memory.** AWS auto-provisions a managed memory when `memory` is omitted (as it does the runtime environment). Import reads that auto-memory back, so the first plan after adopting an existing harness shows a memory diff that reconciles as a server-side no-op — declared import-normalized in the catalog, not drift.
- **Online configs sample production, they do not create it.** Point
  them at log groups AgentCore observability actually writes to;
  sampling an empty group scores nothing and still bills the config's
  idle path only when AWS starts runs.
- **Builtin evaluator IDs** (`Builtin.Helpfulness` and siblings) are
  AWS-owned; custom IDs come from this bundle's `evaluator_ids` map.
  Do not invent a builtin name.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
