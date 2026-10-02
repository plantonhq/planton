# GcpDialogflowCxAgent — Pulumi Implementation

This directory contains the Pulumi implementation for a Dialogflow CX agent and its folded children from the Planton spec: one `gcp.projects.Service` (API enablement), one `gcp.diagflow.CxAgent`, and one `gcp.diagflow.CxWebhook`, `CxTool`, `CxToolVersion`, `CxVersion`, `CxEnvironment`, or `CxGenerativeSettings` per declared entry, each parented to the agent (tool versions to their tool).

## File Organization

| File | Purpose |
|------|---------|
| `main.go` | Module entry point; invokes `Resources()` which wires locals, provider, the agent, and each child family |
| `module/locals.go` | The display name defaulted from `metadata.name`; the start flow id and the wildcard default-playbook path; the send-only-when-set helpers |
| `module/agent.go` | Enables the API; maps the agent and its settings; exports `name`, `agent_id`, `location`, `start_flow` |
| `module/webhooks.go` | Maps every webhook; the direct and Service Directory endpoint builders mirror each other |
| `module/tools.go` | Maps every tool and tool version; the tool and snapshot builders mirror each other |
| `module/versions.go` | Maps flow versions, then environments resolving their version configs |
| `module/generative_settings.go` | Maps generative settings per language |
| `module/outputs.go` | Output key constants |

## Send Posture (parity with Terraform)

- **`DisplayName`** -- `spec.display_name`, defaulting to `metadata.name`.
- **Optional booleans** -- sent only when true.
- **`StartPlaybook`** -- the wildcard default-playbook path when `start_with_default_playbook` is set.
- **`AdvancedSettings`** and the lifted one-leaf blocks -- emitted only when declared or set.
- **Credentials** -- `pulumi.ToSecret` on the GitHub access token, webhook and tool OAuth client secrets, API keys, and bearer tokens.
- **Keys and names** -- children keyed by display name (versions by `{flow_id}/{display_name}`, generative settings by language); versions' parents and outside version paths composed from the agent's ID; a data store spec always carries the empty `FallbackPrompt`.
- **`DeletionPolicy`** -- fanned from the spec to every child that has one.
- **Name lists** -- the created children's IDs in manifest order.

## Usage

```shell
planton pulumi up --manifest <manifest.yaml>
```
