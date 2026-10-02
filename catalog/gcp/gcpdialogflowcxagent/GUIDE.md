# GcpDialogflowCxAgent Guide

The judgment this guide protects: an agent is two things with two owners. Conversation design -- flows, pages, intents, playbook instructions -- belongs to the people who write it in the Dialogflow console. The infrastructure that design calls -- webhooks, tools, releases, environments, generative settings, the agent's own settings -- belongs here, versioned and deployed like the rest of the stack. Keep each on its side and neither steps on the other.

## Flows or playbooks

Every agent is created with a default start flow and a default playbook. A flow agent routes each turn through state machines you design, matched by intents -- predictable and cheap per turn. `startWithDefaultPlaybook` begins conversations in the playbook instead: a generative agent that follows natural-language instructions and calls the tools declared here. Google bills playbook turns at a higher rate, and when a playbook calls a flow, every turn of that conversation bills as a playbook turn. Google allows only the default playbook as a start playbook, which is why the setting is a yes or no.

## Location

`global` works in any project immediately. A region keeps the agent's data there, but a project's first regional agent needs Google's one-time location settings, which only the Dialogflow console can set. The location, like the default language, is fixed at creation; security settings, data stores, and Service Directory services the agent uses must be in a compatible location.

## Webhooks and tools

A webhook is your code that a flow or page calls -- look up an order, book a slot. A tool is an API, a data store, or a client-run function a playbook calls, described so the model knows when to use it; write the tool's `description` for the model. Authenticate without stored secrets where you can: `serviceAgentAuth: ID_TOKEN` has the Dialogflow service agent mint a token Cloud Run or Cloud Functions accept, and a webhook's `serviceAccount` sends that account's token (the service agent needs `roles/iam.serviceAccountTokenCreator` on it). When a third party needs a secret, name a Secret Manager secret version (`secretVersionFor...`) rather than writing the value; the spec marks every raw credential sensitive, but Google stores whatever it is sent.

Children are keyed by display name, the only name you write -- Google assigns every webhook, tool, version, and environment a random id. Google requires webhook, tool, and environment names unique within the agent; this block also requires flow versions unique per flow and tool versions unique per tool. Google renames in place, but here a rename replaces the child and gives it a new resource name, so console content pointing at the old name must be re-pointed.

## Tool versions are frozen

A tool version captures one definition of a tool so a playbook can pin it while the tool keeps changing. Every field is immutable, so the snapshot is written out in full -- a version that copied its tool would be re-created whenever the tool changed and would stop being a snapshot.

## Releases and environments

A version freezes one flow -- its pages, routes, and trained model -- and an environment serves a chosen version of each flow. Declare versions of the start flow by leaving `flowId` empty; name console-authored flows by their id, because a full flow path contains the agent's id, which does not exist before the first apply. An environment names a declared version by display name, or a version made in the console by `versionId`. Google requires a version for every flow reachable from the start flow; a missing one fails the apply. Creating a version waits for the flow's model to train.

## Generative settings

Generative settings hold the fallback prompts, banned phrases, knowledge-connector persona, and model for one language, and Google keeps exactly one set per agent and language. Declaring an entry overwrites Google's defaults for that language; removing it only stops managing it, so the last settings stay in Google. Declare an entry for the default language and each supported language that needs its own.

## Linking a search engine

Link from one side. A `GcpVertexAiSearchEngine` chat engine that names this agent in `dialogflowAgentToLink` needs nothing here. `genAppBuilderSettings.engine` is for pinning the engine Google created when a data store was connected in the console, and `deleteChatEngineOnDestroy` removes that engine with the agent -- refused when the engine is another block's, because that block owns its lifecycle.

## Destroy

`deletionPolicy` fans to every webhook, tool, tool version, version, and environment. `DELETE` removes the agent and all of its console-authored content with it; `PREVENT` makes destroy fail; `ABANDON` leaves everything in place. `locked` additionally makes Google reject every change except a restore -- unlock before changing anything else. Export content to GitHub (`gitIntegrationSettings`) if it must survive the agent.
