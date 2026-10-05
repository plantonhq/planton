# GcpDeployPolicy

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpDeployPolicySpec declares a Cloud Deploy deploy policy
(`google_clouddeploy_deploy_policy`): rollout restrictions -- freeze
windows -- that block chosen actions on chosen delivery pipelines and
targets during chosen times.

A policy is how a team says "no production deploys on Friday
afternoons" or "nothing ships over the year-end freeze". Its selectors
pick the pipelines and targets it governs (by ID, by "*" for all in
the location, or by labels); its rules say which actions are blocked,
for whom, and when. One policy can govern many pipelines, and one
pipeline can fall under many policies, so the policy is its own block
rather than part of a pipeline.

A blocked action fails with a policy violation unless whoever runs it
overrides the policy by name (which needs the
clouddeploy.deployPolicies.override permission) -- the override is the
emergency door, the policy is the default.

Important behavioral notes:

  - The policy lives in the same project and location as the pipelines
    and targets it governs; selectors never reach across locations.
  - location and deploy_policy_id are create-time decisions; everything
    else, including rules and selectors, updates in place.
  - suspended keeps the policy without enforcing it -- the way to lift a
    freeze early without deleting its definition.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDeployPolicy
metadata:
  name: prod-freeze
spec:
  projectId:
    value: my-gcp-project
  location: us-central1
  deployPolicyId: prod-freeze
  description: No production rollouts on weekends or over the year-end freeze
  labels:
    team: release-eng
  annotations:
    owner: release-eng
  rules:
    - rolloutRestriction:
        id: weekend-freeze
        actions:
          - CREATE
          - APPROVE
        invokers:
          - USER
          - DEPLOY_AUTOMATION
        timeWindows:
          timeZone: America/New_York
          weeklyWindows:
            - daysOfWeek:
                - SATURDAY
                - SUNDAY
            - daysOfWeek:
                - FRIDAY
              startTime:
                hours: 15
                minutes: 30
              endTime:
                hours: 24
    - rolloutRestriction:
        id: year-end
        timeWindows:
          timeZone: America/New_York
          oneTimeWindows:
            - startDate:
                year: 2026
                month: 12
                day: 20
              startTime:
                hours: 17
              endDate:
                year: 2027
                month: 1
                day: 3
              endTime:
                hours: 9
  selectors:
    - deliveryPipeline:
        id:
          value: web
      target:
        id:
          value: web-prod
    - target:
        labels:
          env: prod
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.deployPolicyId` | `string` |  |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.suspended` | `bool` |  |  |  |
| `spec.rules` | `[]GcpDeployPolicyRule` | yes |  |  |
| `spec.rules[].rolloutRestriction` | `GcpDeployPolicyRolloutRestriction` |  |  |  |
| `spec.rules[].rolloutRestriction.id` | `string` | yes |  |  |
| `spec.rules[].rolloutRestriction.actions` | `[]string` |  |  |  |
| `spec.rules[].rolloutRestriction.invokers` | `[]string` |  |  |  |
| `spec.rules[].rolloutRestriction.timeWindows` | `GcpDeployPolicyTimeWindows` | yes |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.timeZone` | `string` | yes |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows` | `[]GcpDeployPolicyOneTimeWindow` |  |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].startDate` | `GcpDeployPolicyDate` | yes |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].startDate.year` | `int32` |  |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].startDate.month` | `int32` |  |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].startDate.day` | `int32` |  |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].startTime` | `GcpDeployPolicyTimeOfDay` | yes |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].startTime.hours` | `int32` |  |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].startTime.minutes` | `int32` |  |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].startTime.seconds` | `int32` |  |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].startTime.nanos` | `int32` |  |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].endDate` | `GcpDeployPolicyDate` | yes |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].endDate.year` | `int32` |  |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].endDate.month` | `int32` |  |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].endDate.day` | `int32` |  |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].endTime` | `GcpDeployPolicyTimeOfDay` | yes |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].endTime.hours` | `int32` |  |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].endTime.minutes` | `int32` |  |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].endTime.seconds` | `int32` |  |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].endTime.nanos` | `int32` |  |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.weeklyWindows` | `[]GcpDeployPolicyWeeklyWindow` |  |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.weeklyWindows[].daysOfWeek` | `[]string` |  |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.weeklyWindows[].startTime` | `GcpDeployPolicyTimeOfDay` |  |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.weeklyWindows[].startTime.hours` | `int32` |  |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.weeklyWindows[].startTime.minutes` | `int32` |  |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.weeklyWindows[].startTime.seconds` | `int32` |  |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.weeklyWindows[].startTime.nanos` | `int32` |  |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.weeklyWindows[].endTime` | `GcpDeployPolicyTimeOfDay` |  |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.weeklyWindows[].endTime.hours` | `int32` |  |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.weeklyWindows[].endTime.minutes` | `int32` |  |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.weeklyWindows[].endTime.seconds` | `int32` |  |  |  |
| `spec.rules[].rolloutRestriction.timeWindows.weeklyWindows[].endTime.nanos` | `int32` |  |  |  |
| `spec.selectors` | `[]GcpDeployPolicySelector` | yes |  |  |
| `spec.selectors[].deliveryPipeline` | `GcpDeployPolicyDeliveryPipelineSelector` |  |  |  |
| `spec.selectors[].deliveryPipeline.id` | `string \| valueFrom` |  |  | GcpDeliveryPipeline (`status.outputs.delivery_pipeline_id`) |
| `spec.selectors[].deliveryPipeline.labels` | `map<string, string>` |  |  |  |
| `spec.selectors[].target` | `GcpDeployPolicyTargetSelector` |  |  |  |
| `spec.selectors[].target.id` | `string \| valueFrom` |  |  | GcpDeployTarget (`status.outputs.target_id`) |
| `spec.selectors[].target.labels` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The project the policy lives in -- the project of the pipelines and
targets it governs: a literal project ID or a GcpProject reference.
Empty means the provider's default project. Immutable.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The region the policy lives in, e.g. "us-central1" -- the region of
the pipelines and targets it governs. Required. Immutable.

- rule: {"required":true}

### spec.deployPolicyId

`string`

The policy's ID, unique in the project and location: 1-63 lowercase
letters, digits, and hyphens, starting with a letter and not ending
with a hyphen. Defaults to metadata.name. Immutable.

- rule: deploy_policy_id must be 1-63 lowercase letters, digits, or hyphens, starting with a letter and not ending with a hyphen

### spec.description

`string`

A description of the policy, up to 255 characters -- say what the
freeze is for and who to ask for an override.

- rule: {"string":{"maxLen":"255"}}

### spec.labels

`map<string, string>`

Labels on the policy resource. The platform attribution labels are
added on top and win on key conflicts. These label the policy itself;
to select pipelines or targets by label, use selectors[].

### spec.annotations

`map<string, string>`

Annotations on the policy (user metadata Cloud Deploy never reads).
Only the keys declared here are managed.

### spec.suspended

`bool`

True keeps the policy but stops enforcing it: actions it would block
go through. Use it to lift a freeze early without deleting it.

### spec.rules

`[]GcpDeployPolicyRule` · required

What the policy restricts. At least one rule; a blocked action is one
that any rule matches.

- rule: {"repeated":{"minItems":"1"}}

### spec.rules[].rolloutRestriction

`GcpDeployPolicyRolloutRestriction`

A rollout restriction: which actions are blocked, for whom, and when.
It is the only kind of rule Cloud Deploy defines, so set it on every
rule -- a rule without one restricts nothing.

### spec.rules[].rolloutRestriction.id

`string` · required

The restriction's ID, unique in the policy and shown in the violation
message: 1-63 lowercase letters, digits, and hyphens, starting with a
letter and not ending with a hyphen. Required.

- rule: rollout_restriction.id must be 1-63 lowercase letters, digits, or hyphens, starting with a letter and not ending with a hyphen
- rule: {"required":true}

### spec.rules[].rolloutRestriction.actions

`[]string`

The rollout actions blocked. Empty blocks every action:
  "ADVANCE"          -- advancing a rollout to its next phase
  "APPROVE"          -- approving a rollout
  "CANCEL"           -- cancelling a rollout
  "CREATE"           -- creating a rollout (a deploy or a promotion)
  "IGNORE_JOB"       -- ignoring a failed job
  "RETRY_JOB"        -- retrying a failed job
  "ROLLBACK"         -- rolling a target back
  "TERMINATE_JOBRUN" -- terminating a running job

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["ADVANCE","APPROVE","CANCEL","CREATE","IGNORE_JOB","RETRY_JOB","ROLLBACK","TERMINATE_JOBRUN"]}}}}

### spec.rules[].rolloutRestriction.invokers

`[]string`

Who is blocked. Empty blocks both:
  "USER"              -- a person or a script calling Cloud Deploy
  "DEPLOY_AUTOMATION" -- the pipeline's own automations (promotions,
                         advances, repairs)

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["USER","DEPLOY_AUTOMATION"]}}}}

### spec.rules[].rolloutRestriction.timeWindows

`GcpDeployPolicyTimeWindows` · required

When the actions are blocked. Required.

- rule: {"required":true}

### spec.rules[].rolloutRestriction.timeWindows.timeZone

`string` · required

The IANA time zone every window is read in, e.g. "America/New_York"
or "Europe/Berlin". Required.

- rule: {"required":true}

### spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows

`[]GcpDeployPolicyOneTimeWindow`

Dated windows that happen once, e.g. a year-end freeze from December
20 at 17:00 to January 3 at 09:00.

### spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].startDate

`GcpDeployPolicyDate` · required

The first day of the window. Required.

- rule: {"required":true}

### spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].startDate.year

`int32`

The year, 1-9999.

- rule: {"int32":{"lte":9999,"gte":0}}

### spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].startDate.month

`int32`

The month, 1-12.

- rule: {"int32":{"lte":12,"gte":0}}

### spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].startDate.day

`int32`

The day of the month, 1-31 and valid for the month.

- rule: {"int32":{"lte":31,"gte":0}}

### spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].startTime

`GcpDeployPolicyTimeOfDay` · required

The time on start_date the window opens (inclusive); 00:00 is the
beginning of the day. Required.

- rule: {"required":true}

### spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].startTime.hours

`int32`

Hours, 0-23; 24 is accepted for the end of the day.

- rule: {"int32":{"lte":24,"gte":0}}

### spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].startTime.minutes

`int32`

Minutes, 0-59.

- rule: {"int32":{"lte":59,"gte":0}}

### spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].startTime.seconds

`int32`

Seconds, 0-59 (60 where a leap second is allowed).

- rule: {"int32":{"lte":60,"gte":0}}

### spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].startTime.nanos

`int32`

Fractions of a second in nanoseconds, 0-999999999.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].endDate

`GcpDeployPolicyDate` · required

The last day of the window. Required.

- rule: {"required":true}

### spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].endDate.year

`int32`

The year, 1-9999.

- rule: {"int32":{"lte":9999,"gte":0}}

### spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].endDate.month

`int32`

The month, 1-12.

- rule: {"int32":{"lte":12,"gte":0}}

### spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].endDate.day

`int32`

The day of the month, 1-31 and valid for the month.

- rule: {"int32":{"lte":31,"gte":0}}

### spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].endTime

`GcpDeployPolicyTimeOfDay` · required

The time on end_date the window closes (exclusive); 24:00 is the end
of the day. Required.

- rule: {"required":true}

### spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].endTime.hours

`int32`

Hours, 0-23; 24 is accepted for the end of the day.

- rule: {"int32":{"lte":24,"gte":0}}

### spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].endTime.minutes

`int32`

Minutes, 0-59.

- rule: {"int32":{"lte":59,"gte":0}}

### spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].endTime.seconds

`int32`

Seconds, 0-59 (60 where a leap second is allowed).

- rule: {"int32":{"lte":60,"gte":0}}

### spec.rules[].rolloutRestriction.timeWindows.oneTimeWindows[].endTime.nanos

`int32`

Fractions of a second in nanoseconds, 0-999999999.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.rules[].rolloutRestriction.timeWindows.weeklyWindows

`[]GcpDeployPolicyWeeklyWindow`

Windows that recur every week, e.g. every Friday from 15:00 to 24:00,
or all of Saturday and Sunday.

- rule: a weekly window sets both start_time and end_time, or neither (neither blocks the whole day)

### spec.rules[].rolloutRestriction.timeWindows.weeklyWindows[].daysOfWeek

`[]string`

The days the window recurs on: MONDAY, TUESDAY, WEDNESDAY, THURSDAY,
FRIDAY, SATURDAY, SUNDAY. Empty means every day.

- rule: {"repeated":{"unique":true,"items":{"string":{"in":["MONDAY","TUESDAY","WEDNESDAY","THURSDAY","FRIDAY","SATURDAY","SUNDAY"]}}}}

### spec.rules[].rolloutRestriction.timeWindows.weeklyWindows[].startTime

`GcpDeployPolicyTimeOfDay`

The time each day the window opens (inclusive); 00:00 is the
beginning of the day. Set together with end_time; leave both unset to
block the whole of each day.

### spec.rules[].rolloutRestriction.timeWindows.weeklyWindows[].startTime.hours

`int32`

Hours, 0-23; 24 is accepted for the end of the day.

- rule: {"int32":{"lte":24,"gte":0}}

### spec.rules[].rolloutRestriction.timeWindows.weeklyWindows[].startTime.minutes

`int32`

Minutes, 0-59.

- rule: {"int32":{"lte":59,"gte":0}}

### spec.rules[].rolloutRestriction.timeWindows.weeklyWindows[].startTime.seconds

`int32`

Seconds, 0-59 (60 where a leap second is allowed).

- rule: {"int32":{"lte":60,"gte":0}}

### spec.rules[].rolloutRestriction.timeWindows.weeklyWindows[].startTime.nanos

`int32`

Fractions of a second in nanoseconds, 0-999999999.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.rules[].rolloutRestriction.timeWindows.weeklyWindows[].endTime

`GcpDeployPolicyTimeOfDay`

The time each day the window closes (exclusive); 24:00 is midnight at
the end of the day. Set together with start_time.

### spec.rules[].rolloutRestriction.timeWindows.weeklyWindows[].endTime.hours

`int32`

Hours, 0-23; 24 is accepted for the end of the day.

- rule: {"int32":{"lte":24,"gte":0}}

### spec.rules[].rolloutRestriction.timeWindows.weeklyWindows[].endTime.minutes

`int32`

Minutes, 0-59.

- rule: {"int32":{"lte":59,"gte":0}}

### spec.rules[].rolloutRestriction.timeWindows.weeklyWindows[].endTime.seconds

`int32`

Seconds, 0-59 (60 where a leap second is allowed).

- rule: {"int32":{"lte":60,"gte":0}}

### spec.rules[].rolloutRestriction.timeWindows.weeklyWindows[].endTime.nanos

`int32`

Fractions of a second in nanoseconds, 0-999999999.

- rule: {"int32":{"lte":999999999,"gte":0}}

### spec.selectors

`[]GcpDeployPolicySelector` · required

Which pipelines and targets the policy applies to. At least one
selector; the policy applies when ANY selector matches, and within a
selector EVERY attribute given must match (its delivery_pipeline and
its target, by ID and by labels).

- rule: {"repeated":{"minItems":"1"}}

### spec.selectors[].deliveryPipeline

`GcpDeployPolicyDeliveryPipelineSelector`

The delivery pipelines this selector matches.

### spec.selectors[].deliveryPipeline.id

`string | valueFrom`

The pipeline's ID (the last segment of its name, in the policy's
project and location): a GcpDeliveryPipeline reference, a literal ID,
or "*" for every pipeline in the location. Empty matches by labels
alone.

- references: GcpDeliveryPipeline (`status.outputs.delivery_pipeline_id`)
- rule: delivery_pipeline.id must be a pipeline ID (lowercase letters, digits, hyphens) or "*"
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpDeliveryPipeline, name: <that resource's name>, fieldPath: status.outputs.delivery_pipeline_id}} -- a bare string does not parse

### spec.selectors[].deliveryPipeline.labels

`map<string, string>`

Labels a pipeline must carry, all of them, to match.

### spec.selectors[].target

`GcpDeployPolicyTargetSelector`

The targets this selector matches.

### spec.selectors[].target.id

`string | valueFrom`

The target's ID (the last segment of its name, in the policy's
project and location): a GcpDeployTarget reference, a literal ID, or
"*" for every target in the location. Empty matches by labels alone.

- references: GcpDeployTarget (`status.outputs.target_id`)
- rule: target.id must be a target ID (lowercase letters, digits, hyphens) or "*"
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpDeployTarget, name: <that resource's name>, fieldPath: status.outputs.target_id}} -- a bare string does not parse

### spec.selectors[].target.labels

`map<string, string>`

Labels a target must carry, all of them, to match.

### spec.deletionPolicy

`string`

What destroy does:
  "" / "DELETE" -- the policy is deleted
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the policy leaves management and stays in Google

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `spec.unique_rule_ids`: rollout_restriction.id must be unique within the policy

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpDeployPolicy, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name: projects/{project}/locations/{location}/deployPolicies/{deploy_policy_id}. |
| `status.outputs.deploy_policy_id` | `string` | The policy's ID, the name a deploy policy override (for example `gcloud deploy releases promote --override-deploy-policies`) takes. |
| `status.outputs.uid` | `string` | Google's unique identifier for the policy. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.selectors[].deliveryPipeline.id` | GcpDeliveryPipeline | `status.outputs.delivery_pipeline_id` |
| `spec.selectors[].target.id` | GcpDeployTarget | `status.outputs.target_id` |

## See Also

- [Overview](../README.md)
