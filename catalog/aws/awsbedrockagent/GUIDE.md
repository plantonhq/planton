# AwsBedrockAgent — Kind Guide

Authored operational judgment for the Bedrock agent component: the design
decisions behind the spec's shape, and what to know before running agents
in production.

## Design decisions

- **The satellites are folded, name-keyed collections.** The provider
  ships the agent, action groups, aliases, collaborators, and
  knowledge-base associations as five resources; the platform folds them
  into one component because they share one lifecycle pivot (the DRAFT
  version) and no satellite is meaningful without its agent. Every entry
  carries a stable `name` — the for_each key on both engines and the key
  in the per-satellite output maps (`alias_arns`, `action_group_ids`,
  ...). Renaming an entry destroys and recreates that satellite.
- **The DRAFT pivot orders everything.** Action groups, collaborators,
  and associations attach to DRAFT; an alias snapshots DRAFT into an
  immutable numbered version at creation. Both modules therefore create
  aliases LAST — an alias created mid-assembly would snapshot a
  half-built agent.
- **`prepare_agent` and `skip_resource_in_use_check` are not spec
  surface.** Both are apply/delete-behavior knobs, not desired state: the
  platform always prepares (a declarative system never deliberately ships
  an unserved draft) and keeps AWS's in-use check on delete. Recorded as
  exclusions in the parity manifest.
- **Prompt overrides are always OVERRIDDEN.** The provider strips
  non-overridden template configurations from state on read, so a
  DEFAULT-mode entry would vanish and drift forever. Authoring an entry
  IS the override; the modules send the creation-mode constant.
- **The executor union is a bool plus a reference.** AWS defines exactly
  one custom-control method (RETURN_CONTROL), so the spec models it as
  `return_control: true` rather than a one-value vocabulary field, XOR a
  Lambda reference.
- **`map_block_key` is called `name`.** The provider's function-schema
  parameter key keeps a legacy name for its own compatibility; the spec
  calls it what it is, and the parity manifest records the rename.

## Running agents in production

- **CreateAgent is allowlisted after 2026-07-30.** AWS put Bedrock
  Agents Classic into maintenance mode: accounts that did not call
  `CreateAgent` or `InvokeInlineAgent` in the 12 months before that
  date get `AccessDeniedException` (HTTP 403, "Bedrock Agents is in
  Maintenance Mode") on new creates. There is no exception process.
  Existing agents and every other API (Get/Update/Delete/Prepare,
  satellites, InvokeAgent) stay available to all accounts. New agent
  work on a non-allowlisted account belongs on AgentCore (Q35), not
  this kind. Live-caught 2026-08-14 on the Planton E2E account
  (`859666865785`) — both engines, both scenarios, same 403.
- **Version through aliases.** Treat `aliases` entries like release
  channels: `live` for production, a second entry for canary. Each new
  alias snapshots the current draft; repointing consumers between aliases
  is the rollback path.
- **Model choice gates account setup.** Auto-enabled AWS models (Nova,
  Titan, Mistral, Meta) need nothing; Anthropic models need the
  account's use-case form on file and marketplace models need an
  agreement — compose `AwsBedrockModelAccess` first and let the chart
  order the dependency.
- **The role is load-bearing at prepare time.** AWS validates
  assumability and model permissions when the agent prepares — a role
  that misses `bedrock:InvokeModel` fails the deploy, not the first
  invocation.
- **Supervisor teardown has a service-side dance.** Removing the last
  collaborator from a SUPERVISOR agent makes AWS refuse the prepare; the
  provider detects this and flips collaboration off and back on around
  the release. Expect delete of collaborator-carrying agents to take a
  few extra prepare cycles.
- **Memory is per-session-summary only.** AWS defines one memory type;
  `memory.storage_days` bounds retention (their documented cap moved
  from 30 to 365 — the platform validates the documented range and the
  live gate watches the service's actual acceptance).

## Cost model

Creating, preparing, and deleting agents is free. Costs start when
aliases serve traffic: per-invocation model tokens (plus knowledge-base
retrieval and guardrail evaluation when attached). Idle agents cost
nothing. Verified per-preset figures live in the generated estimate at
`catalog/_pricing/estimates/awsbedrockagent.yaml`, computed from the
pinned, source-dated price book.
