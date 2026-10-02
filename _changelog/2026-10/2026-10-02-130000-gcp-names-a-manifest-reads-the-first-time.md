# Three GCP Fields Named the Way a Manifest Reads Them

**Date**: October 2, 2026
**Type**: Fix (breaking changes on `v1alpha1` fields)
**Components**: `GcpVertexAiAgentEngine`, `GcpLogBucket`, `GcpIamDenyPolicy`

## Summary

**The Agent Engine's agent block is `agent`, not `spec`.** Google's API calls the block that holds an agent's code, identity, and deployment shape `spec`, and the kind copied the name, so every manifest read `spec: spec:`. Readers took the inner `spec` for the outer one; the kind's own docs did too. The field is now `agent`, which is what its comment always called it. Parity accounting records the name as a mapping onto Google's `spec` block.

**A log bucket's folder and a deny policy's folder are links to `GcpFolder`.** Every other folder field in the catalog already links to the folder kind. These two were plain text, so a chart that creates a folder could not hand its ID to the folder's `_Default` log bucket or to a deny policy guarding it, and the log bucket's scope no longer matched the logging sink's, which a console editor shares between them. Both take a typed reference now (the folder's `folder_id` output) or a literal.

## Breaking changes

All three are `v1alpha1` fields, changed in place. The kinds shipped one release ago.

### `GcpVertexAiAgentEngine.spec.spec` is `spec.agent`

```yaml
# Before
spec:
  location: us-central1
  spec:
    agentFramework: google-adk
# After
spec:
  location: us-central1
  agent:
    agentFramework: google-adk
```

The message type is renamed to match: `GcpVertexAiAgentEngineSpecConfig` is `GcpVertexAiAgentEngineAgent`. The rule ids on it read `agent.container_xor_source`, `agent.agent_identity_forbids_service_account`, and `agent.agent_identity_forbids_stored_secret_values`. The Terraform module's variable is `spec.agent`; the resource's `spec` block and its arguments are unchanged.

### `GcpLogBucket.spec.scope.folderId` and `GcpIamDenyPolicy.spec.parent.folderId` take a value or a link

```yaml
# Before
spec:
  scope:
    folderId: "987654321098"
# After: a literal
spec:
  scope:
    folderId:
      value: "987654321098"
# After: a link
spec:
  scope:
    folderId:
      valueFrom:
        kind: GcpFolder
        name: platform
        fieldPath: status.outputs.folder_id
```

The deny policy's `parent.folderId` changes the same way. Both keep accepting the `folders/` prefix on a literal. The rules that test whether a folder arm is set (`at_most_one_scope`, `locked_is_project_scope_only`, `analytics_is_project_scope_only`, `scope_settings_need_folder_or_org_scope`, `at_most_one_parent`) read the reference's content.

## Verification

- Spec tests for the three kinds; their Pulumi modules build; the Agent Engine module's tests.
- `variables.tf` regenerated for the three modules; the drift gate is green.
- `provider-parity` at total accounting for all three (Agent Engine 104 arguments, log bucket 59, deny policy 13).
- Offline plans: the Agent Engine manifest, both scenarios, and both presets; the log bucket manifest and two scenarios; the deny policy manifest and scenario; plus a folder-scoped bucket and a folder-attached deny policy. The rendered plans show the agent block reaching Google's `spec`, the folder ID reaching `google_logging_folder_bucket_config.folder`, and the deny policy's parent rendered as `cloudresourcemanager.googleapis.com%2Ffolders%2F987654321098`.
- `make generate-reference` regenerated the reference pages and the documentation index.
