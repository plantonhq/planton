# AwsBedrockAgentCoreGateway — Kind Guide

Authored operational judgment for the AgentCore gateway component: the
design decisions behind the spec's shape, and what to know before
fronting production tools.

## Design decisions

- **The backend is a six-arm union with the protocol wrappers
  collapsed.** The provider nests targets under `http`/`mcp` protocol
  blocks; the spec's `backend` message carries the six arms directly
  (exactly one, CEL-guarded) — the wrapper choice is derivable from the
  arm and a leaf that must agree with structure is drift surface.
- **The tool-schema tree is typed to exactly AWS's depth.** The Lambda
  target's JSON schemas nest three typed levels, then bottom out in
  raw-JSON leaves (`items_json`/`properties_json` as Structs) — exactly
  where AWS's own configuration surface stops unrolling its recursion.
  Never deepen or flatten one side alone; the parity manifest records
  the `property` → `properties` renames at every level.
- **OpenAPI targets need a `servers` array with an HTTPS URL.** AWS validates the schema document's content when the target creates — a document without a non-empty `servers` entry, or with a non-HTTPS URL, lands the target FAILED with named validation errors (live-verified 2026-08-14). Nothing calls the URL at create; it just has to be present and HTTPS.
- **One-value vocabularies are module constants**: protocol type (MCP),
  search type (SEMANTIC → `enable_semantic_search`), exception level
  (DEBUG → `expose_debug_exceptions`).
- **`jwt_passthrough` is presence-as-value.** The provider's arm is an
  EMPTY block; the spec models it as a bool and the modules render the
  block on true (recorded as a spec exclusion in the parity manifest).
- **At most one credential arm** — the provider expresses it as pairwise
  conflicts; the spec's CEL is the symmetric closure.

## Fronting production tools

- **Descriptions are model-facing prose.** Tool and schema descriptions
  are what the model reads to decide when and how to call — write them
  like documentation for a sharp intern, not like code comments.
- **Prefer semantic search past a dozen tools.** Without it agents list
  every tool into context; `enable_semantic_search` lets them query for
  the relevant ones.
- **Outbound credentials belong in AgentCore Identity.** API keys and
  OAuth clients live as Identity credential providers; targets reference
  the provider ARN — rotating a credential never touches the gateway.
- **A remote MCP server must be reachable at create** in DEFAULT listing
  mode (AWS snapshots its tools); use DYNAMIC when the server's tool set
  changes or its availability is not guaranteed at deploy time.
- **Roll out Cedar in LOG_ONLY first.** ENFORCE blocks denied calls
  immediately; LOG_ONLY shows what WOULD be denied without breaking
  agents mid-conversation.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
