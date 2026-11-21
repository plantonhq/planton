# GcpVertexAiAgentEngine Guide

The judgment this guide protects: the agent is code you own running in a
service Google runs. Decide where the code comes from, who the agent is
when it calls other APIs, how big it may grow, and whether it remembers --
then let Agent Engine build, host, and scale it.

## Where the code comes from

Exactly one path. `sourceCodeSpec` has Vertex AI build the agent from
source: an `inlineSource` archive (a base64 .tar.gz of the source root),
a `developerConnectSource` repository at a Git ref, or an
`agentConfigSource` -- an ADK agent described entirely by JSON config.
Pair the source with one build recipe: `pythonSpec` (Vertex AI installs
`requirements.txt` and imports `entrypointObject` from `entrypointModule`,
`root_agent` from `agent` by ADK convention) or `imageSpec` (the
Dockerfile at the source root). The build runs as a Cloud Build in your
project. `containerSpec` skips the build and runs an image you built that
implements the Agent Engine serving contract; the project's Reasoning
Engine service agent needs `roles/artifactregistry.reader` on the
repository to pull it. `packageSpec` is the older
pickled-object shape the Vertex AI SDK produced; it still works.

An inline archive is input-only -- Google never reads it back -- so a new
archive is detected by the manifest diff and redeploys the code in place.
The readable source of the sample agent the scenarios deploy lives under
`e2e/fixtures/agent-source/`; its docstring has the archive command.

## Who the agent is

By default the agent runs as the project's Vertex AI Reasoning Engine
service agent. Name a `GcpServiceAccount` in `agent.serviceAccount` to run
as a custom identity you grant roles to (the deploying principal needs
`iam.serviceAccounts.actAs` on it). `identityType: AGENT_IDENTITY` gives
the agent its own Agent Identity instead; `serviceAccount` must then be
unset. Secrets never go in `env`; `secretEnv` injects Secret Manager versions at instance start. A `secretEnv` entry either references a secret you manage with `secretRef`, and then granting the agent's identity `roles/secretmanager.secretAccessor` is yours, or carries a `value`. For a `value`, the module creates one secret per variable in `location`, stores the value as a pinned version, and grants the runtime identity access to that secret alone before the engine is created. That identity is `serviceAccount`, or the project's Reasoning Engine service agent when it is empty. A `value` cannot be combined with `identityType: AGENT_IDENTITY`: that identity exists only once the engine does, and the engine's instances read their secrets while it is being created, so there is nothing to grant to in time. Use `secretRef` to a secret the agents' identities can already read when the agent is created, since the create reads it; once the agent exists, its own identity is the `effectiveIdentity` output.

## How big it may grow

`deploymentSpec` sizes the service: `minInstances` (1-75, Google's
default 1; Google accepts 0, but its provider never sends an explicit 0, so
the agent would keep one instance billing around the clock and re-plan on
every deploy -- the spec refuses 0),
`maxInstances`, `containerConcurrency` (2 x cpu + 1 is Google's guidance),
and `resourceLimits` with `cpu` (1, 2, 4, 6, 8) and `memory` (up to
32Gi). The minimum is the committed spend. `pscInterfaceConfig` gives
the agent a Private Service Connect interface into a VPC through a
network attachment, with DNS peering into other projects' private zones;
`agentGatewayConfig` routes inbound and outbound traffic through Agent
Gateway.

## Whether it remembers

`contextSpec.memoryBankConfig` turns on the Memory Bank. Name the Gemini
model that generates memories and when it runs (`generationRule`: after
N events, at a fixed interval, or after idleness; none means immediately),
the embedding model that finds them, and the TTLs that expire them.
`customizationConfigs` tune generation per scope (`user_id`,
`session_id`): Google's managed topics, your custom topics, worked
examples of conversation-to-memory, third-person phrasing.
`structuredMemoryConfigs` generate into fixed JSON schemas. Both models
are named as full publisher-model paths in the agent's location.

Turning the Memory Bank off is not a manifest change: removing
`memoryBankConfig` sends nothing, and Google keeps the bank and its
memories. An empty `memoryBankConfig: {}` is never sent either, so a bank is
on only once it names at least one of its models.

## Two knobs held out, on purpose

`buildSpec.serviceAccount` (the Cloud Build builder's identity) and the
`audioTranscription` payload of an example-conversation part exist in
Google's provider but not in the pinned Pulumi SDK. Both engines hold
them out so a manifest means the same thing everywhere; they join when
the SDK catches up. `buildSpec.workerPool` is modeled: reference a
`GcpCloudBuildWorkerPool` when the source build must reach a private
network, such as an internal package index.

## Destroy

`DELETE` deletes the agent without force, so Google refuses it while the
agent still holds sessions or memories -- an agent that has served a
conversation needs `FORCE`, which deletes it together with them. `PREVENT`
guards a production agent; `ABANDON` leaves it running when the block
leaves management.

`ABANDON` keeps the agent only. The secrets this component stores for
`secretEnv[].value`, their versions, and the identity's grant on them are
deleted on both engines, so the abandoned agent's next instance start can no
longer read them. An agent meant to outlive its manifest takes its secrets
through `secretRef`, from a Secret Manager secret you own.
