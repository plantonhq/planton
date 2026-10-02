# GcpDialogflowCxAgent — Terraform Implementation

This directory contains the Terraform implementation for a Dialogflow CX agent and its folded children from the Planton spec: one `google_project_service` (API enablement), one `google_dialogflow_cx_agent`, and, each through `for_each`, one `google_dialogflow_cx_webhook` per `spec.webhooks[]` entry (keyed by display name), one `google_dialogflow_cx_tool` per `spec.tools[]` entry (keyed by display name), one `google_dialogflow_cx_tool_version` per `spec.tools[].versions[]` entry (keyed `"{tool}/{version}"`), one `google_dialogflow_cx_version` per `spec.versions[]` entry (keyed `"{flow_id}/{display_name}"`), one `google_dialogflow_cx_environment` per `spec.environments[]` entry (keyed by display name), and one `google_dialogflow_cx_generative_settings` per `spec.generative_settings[]` entry (keyed by language).

## Provider

- **Provider**: `hashicorp/google` `~> 8.3`
- Credentials via the ambient environment (Application Default Credentials or the runner's provider configuration)

## File Organization

| File | Purpose |
|------|---------|
| `provider.tf` | Terraform block and Google provider configuration |
| `variables.tf` | `metadata` and `spec` variable definitions (generated from the spec; the tfvars converter flattens refs to plain strings) |
| `locals.tf` | Ambient project fallback, the display name defaulted from `metadata.name`, null-for-empty optionals, the wildcard default-playbook path, the start flow id, the child maps |
| `main.tf` | `google_project_service` and the seven Dialogflow CX resources |
| `outputs.tf` | `name`, `agent_id`, `location`, `start_flow`, and the six child name lists |

## Send Posture

- **`display_name`** -- `spec.display_name`, defaulting to `metadata.name` (Google requires one) -- PARITY with the Pulumi module.
- **Optional booleans** -- sent only when true; turning one off plans the change back to false.
- **`start_playbook`** -- `projects/-/locations/-/agents/-/playbooks/00000000-0000-0000-0000-000000000000` when `start_with_default_playbook` is set (the wildcard form Google accepts before the agent's id exists).
- **`advanced_settings`** -- emitted only when declared (Optional+Computed); each nested block only when declared.
- **Lifted blocks** -- `answer_feedback_settings`, `personalization_settings`, `speech_to_text_settings`, `text_to_speech_settings`, a webhook's `service_account_auth_config`, and the security settings' `insights_export_settings` are emitted only when their single leaf is set.
- **Webhooks** -- the endpoint block is written out twice (direct and under `service_directory`), field for field.
- **Tools and tool versions** -- the tool definition is written out twice (the tool and the version's `tool` snapshot); a data store spec always carries Google's empty `fallback_prompt {}`.
- **Versions and environments** -- a version's `parent` is `{agent}/flows/{flow_id}` (empty `flow_id` is the start flow `00000000-0000-0000-0000-000000000000`); an environment's `version_configs[].version` is the declared version's `id` or `{agent}/flows/{flow_id}/versions/{version_id}`.
- **`deletion_policy`** -- fanned from the spec to every child that has one (generative settings has none).
- **Name lists** -- in manifest order (comprehensions over `spec.*`, not over the `for_each` maps), so both engines export the same lists.

## Usage

```shell
planton tofu init --manifest <manifest.yaml>
planton tofu plan --manifest <manifest.yaml>
planton tofu apply --manifest <manifest.yaml>
```
