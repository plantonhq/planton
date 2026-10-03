# AwsBedrockAgentCoreMemory — Kind Guide

Authored operational judgment for the AgentCore memory component: the
design decisions behind the spec's shape, and what to know before
running agent memory in production.

## Design decisions

- **The memory name is an explicit spec field.** AWS's charset (letter
  first, then letters/digits/underscore) rejects hyphens; deriving from
  `metadata.name` would fail most manifests at apply.
- **`namespace_templates` is required on every strategy.** The provider
  pairs it exactly-one with the deprecated `namespaces` twin, so
  omitting both fails at plan; the spec requires the living surface and
  the deprecated twin is excluded by design.
- **CUSTOM pairs exactly with `custom`.** The override configuration is
  present when and only when the type is CUSTOM (CEL both directions);
  within it, EPISODIC_OVERRIDE requires `reflection` and SUMMARY_OVERRIDE
  forbids `extraction` — AWS's own pairing rules, enforced at manifest
  time.
- **Kinesis delivery flattens three wrappers.** The provider's
  `stream_delivery_resources.resource.kinesis` chain is one
  `kinesis_delivery` message; MEMORY_RECORDS is the only content type
  AWS defines — a module constant.
- **`event_expiry_days` names the unit** the provider's
  `event_expiry_duration` leaves implicit.

## Running agent memory in production

- **Size the window to the conversation, not the archive.** Short-term
  events exist to feed extraction; long-term records outlive them. A
  30-day window suits most assistants — pay for retention only where
  raw replay matters.
- **Namespaces are your query API.** Design templates around retrieval
  (`/facts/{actorId}`, `/preferences/{actorId}`) before data lands —
  changing them later strands existing records under old paths.
- **Batch strategy edits.** Each change serializes through the parent
  memory (provider timeout 45m per strategy operation); five separate
  applies take five serialized waits. Live timing datapoint (2026-08-14): a four-strategy attach ran ~13 minutes per direction per engine, individual attaches 2–7.5 minutes.
- **Reflection namespaces must stay inside the episodic namespace.** AWS requires each reflection namespace to equal, or be a hierarchical (whole-segment) prefix of, an episodic namespace — the spec rejects the contradiction at manifest time, and the server enforces the rest.
- **After importing an EPISODIC strategy, expect a one-time reconcile.** The provider's import does not read back `reflection_configuration`, so the first plan after import re-adds it — applying is a server-side no-op that only heals the state (upstream gap recorded 2026-08-14; the platform's import round-trips tolerate exactly this path).
- **Indexed keys replace the memory.** They are create-time structure
  (ForceNew at the provider) — choose filter keys up front.
- **CUSTOM strategies need the execution role** with invoke access to
  their models; built-in strategies run entirely AWS-managed.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
