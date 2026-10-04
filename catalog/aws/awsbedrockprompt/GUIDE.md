# AwsBedrockPrompt — Kind Guide

Authored operational judgment for the Bedrock prompt kind: the
design decisions behind the spec's shape, and what to know before running
managed prompts in production.

## Design decisions

- **`template_type` is derived, never asked for.** A variant sets exactly
  one of `text`/`chat` (CEL-guarded) and the modules derive AWS's
  TEXT/CHAT discriminator — a leaf that must agree with structure is
  drift surface, not configuration. The same rule derives the
  model-vs-agent target from `model_id` XOR `agent_alias_arn`.
- **The gen-AI resource wrapper flattens to one ARN.** The provider
  spells the agent target `gen_ai_resource.agent.agent_identifier`; the
  spec carries `agent_alias_arn` — two single-member wrappers dropped,
  recorded in the parity manifest.
- **Content blocks flatten onto their turns.** A chat message's content
  is a one-element block holding text XOR a cache point; the spec puts
  `text` and `cache_point: true` directly on the message (and system
  block, and tool entry). The cache-point TYPE has exactly one legal
  value ("default") — the modules own the constant.
- **`metadata` is a map.** The provider's repeated {key, value} entries
  are a key-value map in the spec; both modules render entries
  name-sorted for deterministic plans.
- **Tool choice is a three-arm union.** AUTO and ANY are EMPTY provider
  blocks (presence IS the value) modeled as bools; forcing a named tool
  is `tool_name` — exactly one arm, CEL-guarded.

## Running managed prompts in production

- **Treat the prompt as the interface, variants as implementations.**
  Applications invoke the prompt ID; `default_variant` decides what runs.
  Promote a candidate by flipping the default — no application deploy.
- **The DRAFT moves; published versions do not.** This component manages
  the draft only. When a formulation ships, publish a numbered version
  (console/API) and pin critical consumers to it.
- **Declare every `{{variable}}`.** AWS matches template placeholders
  against `input_variables` at invocation; an undeclared variable
  surfaces as a runtime invocation error, not a deploy failure.
- **Tools describe, models decide.** The tool catalog's descriptions are
  the model's only signal for tool selection — write them like API docs.
  Use `tool_choice.any` to force SOME tool call (structured-output
  extraction), `tool_name` to force one.
- **additional_model_request_fields is the escape hatch** for
  model-specific parameters outside the standard set (e.g. Anthropic
  `top_k`) — it passes through as JSON, unvalidated until invocation.
- **Importing an existing prompt shows a one-time inference-float
  reconcile.** Bedrock stores `temperature`/`top_p` as 32-bit floats, so
  a value that is not float32-exact (0.9, 0.7) reads back slightly
  widened (0.8999999761581421) on import, and the first plan proposes an
  in-place update back to your manifest's value. Applying it is a
  server-side no-op and the plan is clean thereafter. Normal deploys
  never see this — state keeps the manifest's value.

## Cost model

Creating, updating, and deleting prompts is free. The target model bills
per invocation when the prompt executes; prompt caching (where the model
supports it) reduces repeated-prefix token costs.
