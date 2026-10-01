# GcpWorkflow

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpWorkflowSpec defines a Cloud Workflows workflow — a serverless
orchestrator that executes a sequence of steps (HTTP calls, connector
calls to GCP services, conditionals, retries) defined in YAML or JSON.
Workflows are the glue between services: an Eventarc trigger fires an
execution, the workflow calls Cloud Run / BigQuery / Pub/Sub in order,
handles errors, and records the result.

Every deployment of new source mints a new REVISION; executions started
before a deploy finish on the revision they started with. The
revision_id stack output tracks the deployed revision.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpWorkflow
metadata:
  name: my-sample-workflow
spec:
  # The workflow definition executed on each run (≤128KB). Every change
  # deploys a NEW revision; running executions finish on the old one.
  sourceContents: |
    main:
      steps:
        - hello:
            return: "hello world"

  # Co-locate with the services the steps call.
  region: us-central1

  # What this workflow orchestrates (shown in the console).
  description: Sample two-step workflow — returns a constant.

  # The guard is ON by default; the sample allows teardown so the
  # manifest destroys cleanly.
  deletionProtection: false

  # What a destroy does: DELETE (cancels running executions and erases
  # history), PREVENT (the posture for production orchestrations), or
  # ABANDON (keep running unmanaged).
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.region` | `string` |  |  |  |
| `spec.workflowName` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.sourceContents` | `string` | yes |  |  |
| `spec.serviceAccount` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.cryptoKey` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`), GcpKmsKeyHandle (`status.outputs.kms_key`) |
| `spec.callLogLevel` | `string` |  |  |  |
| `spec.executionHistoryLevel` | `string` |  |  |  |
| `spec.userEnvVars` | `map<string, string>` |  |  |  |
| `spec.resourceManagerTags` | `map<string, string>` |  |  |  |
| `spec.deletionProtection` | `bool` |  | `true` |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project to create the workflow in. Can be a literal project ID
or a reference to a GcpProject resource. If omitted, the provider's
default project is used. Immutable: changing it replaces the workflow.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.region

`string`

The region the workflow runs in (e.g. us-central1). If omitted, the
provider's default region is used. Immutable: changing it replaces the
workflow.

### spec.workflowName

`string`

The workflow name in GCP. Defaults to metadata.name when left empty.
Immutable: changing it replaces the workflow (a new workflow with a
fresh execution history — running executions on the old name are
unaffected until it is deleted).

### spec.description

`string`

What this workflow orchestrates — shown in the console list. At most
1000 unicode characters (the provider's own documented cap).

- rule: {"string":{"maxLen":"1000"}}

### spec.labels

`map<string, string>`

User labels attached to the workflow (merged with the platform's
standard labels by the module).

### spec.sourceContents

`string` · required

The workflow definition — the YAML (or JSON) source executed on each
run (https://cloud.google.com/workflows/docs/reference/syntax). The
API caps the size at 128KB. REQUIRED: a workflow without source has
never been deployable through the API — the provider still marks the
argument optional but its own 8.0.0 upgrade note says it "will become
REQUIRED ... to align with API constraints"; this spec models the API
truth today. Every source change deploys a NEW revision.

- rule: {"required":true,"string":{"maxBytes":"131072"}}

### spec.serviceAccount

`string | valueFrom`

The service account the workflow's executions run AS — the identity
its HTTP calls and connector calls carry (format: bare email; the
module renders the projects/{project}/serviceAccounts/{email} form the
API stores). A literal email or a reference to a GcpServiceAccount
resource. If omitted, the project's default compute service account is
used — fine for experiments, wrong for production (grant a dedicated
account only the roles the steps need). Changing the service account
deploys a NEW revision.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.cryptoKey

`string | valueFrom`

Customer-managed encryption key (CMEK) used to encrypt workflow
definitions and execution data at rest. The full crypto key resource
name (projects/{p}/locations/{l}/keyRings/{r}/cryptoKeys/{k}) — a
literal or a reference to a GcpKmsKey resource. Grant the Workflows
service agent roles/cloudkms.cryptoKeyEncrypterDecrypter on the key
BEFORE deploying, or the deploy fails. Omit for Google-managed
encryption.

- references: GcpKmsKey (`status.outputs.key_id`), GcpKmsKeyHandle (`status.outputs.kms_key`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.callLogLevel

`string`

How much call detail execution history records:
  ""                          -- the API default (errors only)
  "CALL_LOG_LEVEL_UNSPECIFIED" -- explicit API default
  "LOG_ALL_CALLS"             -- every call and its result (debugging;
                                 highest log volume and cost)
  "LOG_ERRORS_ONLY"           -- failed calls only
  "LOG_NONE"                  -- nothing

- rule: call_log_level must be one of: CALL_LOG_LEVEL_UNSPECIFIED, LOG_ALL_CALLS, LOG_ERRORS_ONLY, LOG_NONE

### spec.executionHistoryLevel

`string`

How much step-level detail execution history keeps:
  ""                                    -- the API default (basic)
  "EXECUTION_HISTORY_LEVEL_UNSPECIFIED" -- explicit API default
  "EXECUTION_HISTORY_BASIC"             -- step names and status
  "EXECUTION_HISTORY_DETAILED"          -- step inputs/outputs too
                                           (needed for step-level
                                           debugging in the console)

- rule: execution_history_level must be one of: EXECUTION_HISTORY_LEVEL_UNSPECIFIED, EXECUTION_HISTORY_BASIC, EXECUTION_HISTORY_DETAILED

### spec.userEnvVars

`map<string, string>`

Environment variables visible to the workflow source via sys.get_env().
At most 20 entries (the API's own cap, enforced here); each value up to
4KiB. Keys must be non-empty and must NOT start with "GOOGLE" or
"WORKFLOWS" (reserved prefixes the API rejects — key-shape rules live
here in the comment because map KEYS are not CEL-addressable). Changing
env vars deploys a NEW revision.

- rule: user_env_vars accepts at most 20 entries (the API cap)

### spec.resourceManagerTags

`map<string, string>`

Resource manager tags bound at workflow creation, keyed
tagKeys/{tag_key_id} with values tagValues/{tag_value_id} — the
org-policy / cost-attribution tag system (distinct from labels).
Immutable: changing tags REPLACES the workflow (provider ForceNew) —
plan tag changes deliberately.

### spec.deletionProtection

`bool` · optional (explicit presence)

Guard rail: while true (the DEFAULT), destroying this resource FAILS
before touching the workflow. Set false explicitly to allow destroy.
Both IaC engines send the value explicitly on every apply so a
true -> false transition always reaches the engine (the
send-true-or-omit class would silently keep the guard up).

- default: `true`

### spec.deletionPolicy

`string`

Deletion policy — what happens when this resource is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the workflow and its execution history are deleted;
               running executions are cancelled
  "PREVENT" -- destroy FAILS (defense in depth alongside
               deletion_protection)
  "ABANDON" -- the workflow is removed from management but keeps
               running in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpWorkflow, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.workflow_id` | `string` | The full workflow resource name (projects/{project}/locations/{region}/workflows/{name}) — the value Eventarc consumes: a GcpEventarcTrigger's destination.workflow and a GcpEventarcMessageBus pipeline's destination.workflow both take exactly this form. |
| `status.outputs.workflow_name` | `string` | The short workflow name (the last segment of workflow_id). |
| `status.outputs.revision_id` | `string` | The currently deployed revision (e.g. "000001-a4d"). A new revision is minted on every source / env-var / service-account change — compare across applies to confirm a deploy actually rolled. |
| `status.outputs.state` | `string` | The workflow state reported by GCP ("ACTIVE" when deployable and executable; "UNAVAILABLE" while the underlying KMS key is unusable). |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.serviceAccount` | GcpServiceAccount | `status.outputs.email` |
| `spec.cryptoKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.cryptoKey` | GcpKmsKeyHandle | `status.outputs.kms_key` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpEventarcMessageBus | `spec.pipelines[].destination.workflow` | `status.outputs.workflow_id` |
| GcpEventarcTrigger | `spec.destination.workflow` | `status.outputs.workflow_id` |

## See Also

- [Overview](../README.md)
