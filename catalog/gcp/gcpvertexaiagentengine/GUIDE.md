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
implements the Agent Engine serving contract. `packageSpec` is the older
pickled-object shape the Vertex AI SDK produced; it still works.

An inline archive is input-only -- Google never reads it back -- so a new
archive is detected by the manifest diff and redeploys the code in place.
The readable source of the sample agent the scenarios deploy lives under
`e2e/fixtures/agent-source/`; its docstring has the archive command.

## Who the agent is

By default the agent runs as the project's Vertex AI Reasoning Engine
service agent. Name a `GcpServiceAccount` in `spec.serviceAccount` to run
as a custom identity you grant roles to (the deploying principal needs
`iam.serviceAccounts.actAs` on it). `identityType: AGENT_IDENTITY` gives
the agent its own Agent Identity instead; `serviceAccount` must then be
unset. Secrets never go in `env`; `secretEnv` injects Secret Manager
versions at instance start, and the agent's identity needs
`roles/secretmanager.secretAccessor` on each.

## How big it may grow

`deploymentSpec` sizes the service: `minInstances` (Google's default 1;
0 scales to nothing between requests at the cost of cold starts),
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

## Two knobs held out, on purpose

`buildSpec.serviceAccount` (the Cloud Build builder's identity) and the
`audioTranscription` payload of an example-conversation part exist in
Google's provider but not in the pinned Pulumi SDK. Both engines hold
them out so a manifest means the same thing everywhere; they join when
the SDK catches up. `buildSpec.workerPool` is modeled.

## Destroy

`DELETE` removes the agent with its sessions and memories; `PREVENT`
guards a production agent; `ABANDON` leaves it running when the block
leaves management.
