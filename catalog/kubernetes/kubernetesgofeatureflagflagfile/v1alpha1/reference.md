# KubernetesGoFeatureFlagFlagFile

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesGoFeatureFlagFlagFileSpec** declares a GO Feature Flag flag
file - typed, validated flags with their variations, targeting rules,
percentage, progressive and scheduled rollouts - and renders it as a
ConfigMap (named after the resource) that a KubernetesGoFeatureFlag relay
reads through its `config_map` retriever.

WHY A SEPARATE RESOURCE: flags change far more often than the relay, are
many per relay, and are often owned by a different team than the one that
runs it. Keeping them here means a flag flip touches this resource only -
the relay is never re-applied, its permissions can be granted separately,
and the relay serves the change on its next poll (its polling interval,
60 seconds by default) with no restart. Several flag files can feed one
relay; when two define the same flag, the retriever listed later wins.

Every rule GO Feature Flag enforces on a flag (at least one variation, one
value type across variations, a query on every enabled targeting rule, a
default rule that resolves to a value, rules and rollouts that name
existing variations, a rollout that ramps forward in time and share,
unique rule names, real dates) is enforced here at plan time, so such a
flag never reaches the relay - where it would be dropped (with an error
logged) and every evaluation of it would return the caller's default
value, and where an undecodable date would stop the whole file from
loading. Query syntax is the one check left to the relay: it parses each
targeting query when it loads the file and drops a flag whose query does
not parse.

Not modeled: a scheduled step's `bucketingKey` and `metadata` - the
format accepts them on a step, and the relay's step merge ignores them.

## Example

```yaml
# Full-surface development manifest - every flag shape GO Feature Flag
# evaluates: boolean, string, number, object and list variations; targeting
# by query; a percentage split; a progressive rollout; a disabled rule; a
# scheduled rollout with every overlay; an experimentation window; metadata,
# a bucketing key and event export turned off.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesGoFeatureFlagFlagFile
metadata:
  name: goff-dev-flags
spec:
  namespace:
    value: goff-dev
  key: flags.goff.yaml
  flags:
    assistant:
      variations:
        enabled:
          boolValue: true
        disabled:
          boolValue: false
      targeting:
        - name: first-organizations
          query: org in ["planton", "acme"]
          variation: enabled
        - name: beta-ramp
          query: beta eq true
          progressiveRollout:
            initial:
              variation: disabled
              percentage: 0
              date: "2026-11-01T09:00:00Z"
            end:
              variation: enabled
              percentage: 100
              date: "2026-11-15T09:00:00Z"
        - name: parked
          query: org eq "globex"
          variation: enabled
          disable: true
      defaultRule:
        variation: disabled
      metadata:
        description: The Planton Assistant
        owner: platform
    checkout-theme:
      variations:
        light:
          stringValue: light
        dark:
          stringValue: dark
      targeting:
        - name: split
          query: country eq "FR"
          percentage:
            dark: 25
            light: 75
      defaultRule:
        variation: light
      bucketingKey: companyId
      trackEvents: false
      version: "3"
      experimentation:
        start: "2026-11-01T00:00:00Z"
        end: "2026-12-01T00:00:00Z"
      scheduledRollout:
        - date: "2026-11-08T09:00:00Z"
          targeting:
            - name: split
              percentage:
                dark: 50
                light: 50
        - date: "2026-12-01T09:00:00Z"
          variations:
            contrast:
              stringValue: contrast
          defaultRule:
            variation: dark
          disable: false
          trackEvents: true
          version: "4"
    request-limit:
      variations:
        small:
          numberValue: 10
        large:
          numberValue: 100
      defaultRule:
        percentage:
          small: 90
          large: 10
    limits:
      variations:
        standard:
          objectValue:
            rps: 10
            burst: 20
        premium:
          objectValue:
            rps: 100
            burst: 200
      defaultRule:
        variation: standard
    regions:
      variations:
        default:
          listValue:
            - us-east-1
            - eu-west-1
      defaultRule:
        variation: default
      disable: true
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.key` | `string` |  | `flags.goff.yaml` |  |
| `spec.flags` | `map<string, KubernetesGoFeatureFlagFlag>` |  |  |  |
| `spec.flags.*.variations` | `map<string, KubernetesGoFeatureFlagValue>` | yes |  |  |
| `spec.flags.*.variations.*.boolValue` | `bool` |  |  |  |
| `spec.flags.*.variations.*.stringValue` | `string` |  |  |  |
| `spec.flags.*.variations.*.numberValue` | `double` |  |  |  |
| `spec.flags.*.variations.*.objectValue` | `object` |  |  |  |
| `spec.flags.*.variations.*.listValue` | `[]any` |  |  |  |
| `spec.flags.*.targeting` | `[]KubernetesGoFeatureFlagRule` |  |  |  |
| `spec.flags.*.targeting[].name` | `string` |  |  |  |
| `spec.flags.*.targeting[].query` | `string` |  |  |  |
| `spec.flags.*.targeting[].variation` | `string` |  |  |  |
| `spec.flags.*.targeting[].percentage` | `map<string, double>` |  |  |  |
| `spec.flags.*.targeting[].progressiveRollout` | `KubernetesGoFeatureFlagProgressiveRollout` |  |  |  |
| `spec.flags.*.targeting[].progressiveRollout.initial` | `KubernetesGoFeatureFlagProgressiveRolloutStep` | yes |  |  |
| `spec.flags.*.targeting[].progressiveRollout.initial.variation` | `string` | yes |  |  |
| `spec.flags.*.targeting[].progressiveRollout.initial.percentage` | `double` |  |  |  |
| `spec.flags.*.targeting[].progressiveRollout.initial.date` | `string` | yes |  |  |
| `spec.flags.*.targeting[].progressiveRollout.end` | `KubernetesGoFeatureFlagProgressiveRolloutStep` | yes |  |  |
| `spec.flags.*.targeting[].progressiveRollout.end.variation` | `string` | yes |  |  |
| `spec.flags.*.targeting[].progressiveRollout.end.percentage` | `double` |  |  |  |
| `spec.flags.*.targeting[].progressiveRollout.end.date` | `string` | yes |  |  |
| `spec.flags.*.targeting[].disable` | `bool` |  |  |  |
| `spec.flags.*.defaultRule` | `KubernetesGoFeatureFlagRule` | yes |  |  |
| `spec.flags.*.defaultRule.name` | `string` |  |  |  |
| `spec.flags.*.defaultRule.query` | `string` |  |  |  |
| `spec.flags.*.defaultRule.variation` | `string` |  |  |  |
| `spec.flags.*.defaultRule.percentage` | `map<string, double>` |  |  |  |
| `spec.flags.*.defaultRule.progressiveRollout` | `KubernetesGoFeatureFlagProgressiveRollout` |  |  |  |
| `spec.flags.*.defaultRule.progressiveRollout.initial` | `KubernetesGoFeatureFlagProgressiveRolloutStep` | yes |  |  |
| `spec.flags.*.defaultRule.progressiveRollout.initial.variation` | `string` | yes |  |  |
| `spec.flags.*.defaultRule.progressiveRollout.initial.percentage` | `double` |  |  |  |
| `spec.flags.*.defaultRule.progressiveRollout.initial.date` | `string` | yes |  |  |
| `spec.flags.*.defaultRule.progressiveRollout.end` | `KubernetesGoFeatureFlagProgressiveRolloutStep` | yes |  |  |
| `spec.flags.*.defaultRule.progressiveRollout.end.variation` | `string` | yes |  |  |
| `spec.flags.*.defaultRule.progressiveRollout.end.percentage` | `double` |  |  |  |
| `spec.flags.*.defaultRule.progressiveRollout.end.date` | `string` | yes |  |  |
| `spec.flags.*.defaultRule.disable` | `bool` |  |  |  |
| `spec.flags.*.bucketingKey` | `string` |  |  |  |
| `spec.flags.*.trackEvents` | `bool` |  |  |  |
| `spec.flags.*.disable` | `bool` |  |  |  |
| `spec.flags.*.version` | `string` |  |  |  |
| `spec.flags.*.metadata` | `object` |  |  |  |
| `spec.flags.*.experimentation` | `KubernetesGoFeatureFlagExperimentation` |  |  |  |
| `spec.flags.*.experimentation.start` | `string` | yes |  |  |
| `spec.flags.*.experimentation.end` | `string` | yes |  |  |
| `spec.flags.*.scheduledRollout` | `[]KubernetesGoFeatureFlagScheduledStep` |  |  |  |
| `spec.flags.*.scheduledRollout[].date` | `string` | yes |  |  |
| `spec.flags.*.scheduledRollout[].variations` | `map<string, KubernetesGoFeatureFlagValue>` |  |  |  |
| `spec.flags.*.scheduledRollout[].variations.*.boolValue` | `bool` |  |  |  |
| `spec.flags.*.scheduledRollout[].variations.*.stringValue` | `string` |  |  |  |
| `spec.flags.*.scheduledRollout[].variations.*.numberValue` | `double` |  |  |  |
| `spec.flags.*.scheduledRollout[].variations.*.objectValue` | `object` |  |  |  |
| `spec.flags.*.scheduledRollout[].variations.*.listValue` | `[]any` |  |  |  |
| `spec.flags.*.scheduledRollout[].targeting` | `[]KubernetesGoFeatureFlagRule` |  |  |  |
| `spec.flags.*.scheduledRollout[].targeting[].name` | `string` |  |  |  |
| `spec.flags.*.scheduledRollout[].targeting[].query` | `string` |  |  |  |
| `spec.flags.*.scheduledRollout[].targeting[].variation` | `string` |  |  |  |
| `spec.flags.*.scheduledRollout[].targeting[].percentage` | `map<string, double>` |  |  |  |
| `spec.flags.*.scheduledRollout[].targeting[].progressiveRollout` | `KubernetesGoFeatureFlagProgressiveRollout` |  |  |  |
| `spec.flags.*.scheduledRollout[].targeting[].progressiveRollout.initial` | `KubernetesGoFeatureFlagProgressiveRolloutStep` | yes |  |  |
| `spec.flags.*.scheduledRollout[].targeting[].progressiveRollout.initial.variation` | `string` | yes |  |  |
| `spec.flags.*.scheduledRollout[].targeting[].progressiveRollout.initial.percentage` | `double` |  |  |  |
| `spec.flags.*.scheduledRollout[].targeting[].progressiveRollout.initial.date` | `string` | yes |  |  |
| `spec.flags.*.scheduledRollout[].targeting[].progressiveRollout.end` | `KubernetesGoFeatureFlagProgressiveRolloutStep` | yes |  |  |
| `spec.flags.*.scheduledRollout[].targeting[].progressiveRollout.end.variation` | `string` | yes |  |  |
| `spec.flags.*.scheduledRollout[].targeting[].progressiveRollout.end.percentage` | `double` |  |  |  |
| `spec.flags.*.scheduledRollout[].targeting[].progressiveRollout.end.date` | `string` | yes |  |  |
| `spec.flags.*.scheduledRollout[].targeting[].disable` | `bool` |  |  |  |
| `spec.flags.*.scheduledRollout[].defaultRule` | `KubernetesGoFeatureFlagRule` |  |  |  |
| `spec.flags.*.scheduledRollout[].defaultRule.name` | `string` |  |  |  |
| `spec.flags.*.scheduledRollout[].defaultRule.query` | `string` |  |  |  |
| `spec.flags.*.scheduledRollout[].defaultRule.variation` | `string` |  |  |  |
| `spec.flags.*.scheduledRollout[].defaultRule.percentage` | `map<string, double>` |  |  |  |
| `spec.flags.*.scheduledRollout[].defaultRule.progressiveRollout` | `KubernetesGoFeatureFlagProgressiveRollout` |  |  |  |
| `spec.flags.*.scheduledRollout[].defaultRule.progressiveRollout.initial` | `KubernetesGoFeatureFlagProgressiveRolloutStep` | yes |  |  |
| `spec.flags.*.scheduledRollout[].defaultRule.progressiveRollout.initial.variation` | `string` | yes |  |  |
| `spec.flags.*.scheduledRollout[].defaultRule.progressiveRollout.initial.percentage` | `double` |  |  |  |
| `spec.flags.*.scheduledRollout[].defaultRule.progressiveRollout.initial.date` | `string` | yes |  |  |
| `spec.flags.*.scheduledRollout[].defaultRule.progressiveRollout.end` | `KubernetesGoFeatureFlagProgressiveRolloutStep` | yes |  |  |
| `spec.flags.*.scheduledRollout[].defaultRule.progressiveRollout.end.variation` | `string` | yes |  |  |
| `spec.flags.*.scheduledRollout[].defaultRule.progressiveRollout.end.percentage` | `double` |  |  |  |
| `spec.flags.*.scheduledRollout[].defaultRule.progressiveRollout.end.date` | `string` | yes |  |  |
| `spec.flags.*.scheduledRollout[].defaultRule.disable` | `bool` |  |  |  |
| `spec.flags.*.scheduledRollout[].trackEvents` | `bool` |  |  |  |
| `spec.flags.*.scheduledRollout[].disable` | `bool` |  |  |  |
| `spec.flags.*.scheduledRollout[].version` | `string` |  |  |  |
| `spec.flags.*.scheduledRollout[].experimentation` | `KubernetesGoFeatureFlagExperimentation` |  |  |  |
| `spec.flags.*.scheduledRollout[].experimentation.start` | `string` | yes |  |  |
| `spec.flags.*.scheduledRollout[].experimentation.end` | `string` | yes |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace of the rendered ConfigMap. Accepts a literal namespace name
or a reference to a KubernetesNamespace resource. Put it where the
relay runs, or name this namespace in the relay's retriever.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.key

`string` · optional (explicit presence)

ConfigMap data key the flag file is written under. The relay's
retriever names the same key.

- default: `flags.goff.yaml`
- rule: The key may contain only letters, digits, '-', '_' and '.'.
- rule: {"string":{"maxLen":"253"}}

### spec.flags

`map<string, KubernetesGoFeatureFlagFlag>`

The flags, keyed by flag name (the key OpenFeature SDKs evaluate). An
empty file is valid: every evaluation then returns the caller's
default value.

- rule: Flag names may not be empty or contain whitespace.
- rule: Every variation of a flag must hold the same value type (all booleans, all strings, all numbers, all objects or all lists).
- rule: The default rule applies when no targeting rule matches, so it takes no query.
- rule: The default rule must resolve to a value: set a variation, a percentage split, or a progressive rollout.
- rule: The default rule cannot be disabled - it is what every unmatched evaluation returns.
- rule: Every enabled targeting rule must resolve to a value: set a variation, a percentage split, or a progressive rollout.
- rule: Every enabled targeting rule needs a query (the default rule is the one without).
- rule: A rule resolves one way: a variation, a percentage split, or a progressive rollout - when several are set, GO Feature Flag silently uses one.
- rule: Percentage shares cannot be negative, and a split must give some variation a share.
- rule: Rules and progressive rollouts may only name variations the flag defines.
- rule: Targeting rule names must be unique within a flag (scheduled steps update rules by name).

### spec.flags.*.variations

`map<string, KubernetesGoFeatureFlagValue>` · required

The values this flag can return, keyed by variation name (e.g.
enabled: true, disabled: false). Every variation must hold the same
value type.

- rule: {"map":{"minPairs":"1"}}

### spec.flags.*.variations.*.boolValue

`bool`

A boolean value (the common on/off flag).

### spec.flags.*.variations.*.stringValue

`string`

A string value.

### spec.flags.*.variations.*.numberValue

`double`

A number value (integer or decimal).

### spec.flags.*.variations.*.objectValue

`object`

A JSON object value.

### spec.flags.*.variations.*.listValue

`[]any`

A JSON array value.

### spec.flags.*.targeting

`[]KubernetesGoFeatureFlagRule`

Targeting rules, checked in order; the first whose query matches the
evaluation context decides the variation. A context matching no rule
falls to default_rule.

### spec.flags.*.targeting[].name

`string`

Rule name. Required for a rule a scheduled step updates (steps match
rules by name).

### spec.flags.*.targeting[].query

`string`

Which evaluation contexts the rule matches: GO Feature Flag's query
language - e.g. `org in ["acme", "globex"]`, `targetingKey sw "beta-"`,
`email ew "@example.com" and country eq "FR"` - or a JSONLogic
expression. Required on every enabled targeting rule; not allowed on
the default rule. The relay checks the syntax when it loads the file:
a query that does not parse drops the whole flag, and every evaluation
of it returns the caller's default value.

### spec.flags.*.targeting[].variation

`string`

The variation returned when the rule matches.

### spec.flags.*.targeting[].percentage

`map<string, double>`

A split across variations (variation name -> share, e.g. enabled: 10,
disabled: 90), bucketed on the targeting key (or bucketing_key) so a
caller keeps its variation across evaluations. Shares are relative
weights (they need not add up to 100). In a scheduled step, a negative
share removes that variation from the split it updates.

### spec.flags.*.targeting[].progressiveRollout

`KubernetesGoFeatureFlagProgressiveRollout`

A gradual ramp from one variation share to another between two
dates.

- rule: A progressive rollout ramps forward: the end share must be at least the initial share.
- rule: A progressive rollout moves callers between two variations - name different ones at its two ends (a single variation for everyone is a plain variation).
- rule: A progressive rollout's end date must come after its initial date.

### spec.flags.*.targeting[].progressiveRollout.initial

`KubernetesGoFeatureFlagProgressiveRolloutStep` · required

Where the ramp starts (served before its date).

- rule: {"required":true}

### spec.flags.*.targeting[].progressiveRollout.initial.variation

`string` · required

The variation this end of the ramp serves. Callers outside the moving
share get the initial variation: everyone gets it before the initial
date. Callers inside the share get the end variation: everyone gets it
after the end date when the end share is 100.

- rule: {"required":true}

### spec.flags.*.targeting[].progressiveRollout.initial.percentage

`double` · optional (explicit presence)

Share of callers on the END variation at this step's date. Between the
two dates the share moves linearly from the initial step's value to the
end step's. Initial: empty = 0. End: empty or 0 = 100. The end share
must be at least the initial share.

- rule: {"double":{"lte":100,"gte":0}}

### spec.flags.*.targeting[].progressiveRollout.initial.date

`string` · required

When this step applies, as an RFC 3339 timestamp (e.g.
2026-11-01T09:00:00Z).

- rule: Dates are real RFC 3339 timestamps such as 2026-11-01T09:00:00Z.
- rule: {"required":true}

### spec.flags.*.targeting[].progressiveRollout.end

`KubernetesGoFeatureFlagProgressiveRolloutStep` · required

Where the ramp ends (served after its date).

- rule: {"required":true}

### spec.flags.*.targeting[].progressiveRollout.end.variation

`string` · required

The variation this end of the ramp serves. Callers outside the moving
share get the initial variation: everyone gets it before the initial
date. Callers inside the share get the end variation: everyone gets it
after the end date when the end share is 100.

- rule: {"required":true}

### spec.flags.*.targeting[].progressiveRollout.end.percentage

`double` · optional (explicit presence)

Share of callers on the END variation at this step's date. Between the
two dates the share moves linearly from the initial step's value to the
end step's. Initial: empty = 0. End: empty or 0 = 100. The end share
must be at least the initial share.

- rule: {"double":{"lte":100,"gte":0}}

### spec.flags.*.targeting[].progressiveRollout.end.date

`string` · required

When this step applies, as an RFC 3339 timestamp (e.g.
2026-11-01T09:00:00Z).

- rule: Dates are real RFC 3339 timestamps such as 2026-11-01T09:00:00Z.
- rule: {"required":true}

### spec.flags.*.targeting[].disable

`bool`

Skip this rule (it is kept in the file but never matches). Not
allowed on the default rule.

### spec.flags.*.defaultRule

`KubernetesGoFeatureFlagRule` · required

The rule applied when no targeting rule matches. It takes no query.

- rule: {"required":true}

### spec.flags.*.defaultRule.name

`string`

Rule name. Required for a rule a scheduled step updates (steps match
rules by name).

### spec.flags.*.defaultRule.query

`string`

Which evaluation contexts the rule matches: GO Feature Flag's query
language - e.g. `org in ["acme", "globex"]`, `targetingKey sw "beta-"`,
`email ew "@example.com" and country eq "FR"` - or a JSONLogic
expression. Required on every enabled targeting rule; not allowed on
the default rule. The relay checks the syntax when it loads the file:
a query that does not parse drops the whole flag, and every evaluation
of it returns the caller's default value.

### spec.flags.*.defaultRule.variation

`string`

The variation returned when the rule matches.

### spec.flags.*.defaultRule.percentage

`map<string, double>`

A split across variations (variation name -> share, e.g. enabled: 10,
disabled: 90), bucketed on the targeting key (or bucketing_key) so a
caller keeps its variation across evaluations. Shares are relative
weights (they need not add up to 100). In a scheduled step, a negative
share removes that variation from the split it updates.

### spec.flags.*.defaultRule.progressiveRollout

`KubernetesGoFeatureFlagProgressiveRollout`

A gradual ramp from one variation share to another between two
dates.

- rule: A progressive rollout ramps forward: the end share must be at least the initial share.
- rule: A progressive rollout moves callers between two variations - name different ones at its two ends (a single variation for everyone is a plain variation).
- rule: A progressive rollout's end date must come after its initial date.

### spec.flags.*.defaultRule.progressiveRollout.initial

`KubernetesGoFeatureFlagProgressiveRolloutStep` · required

Where the ramp starts (served before its date).

- rule: {"required":true}

### spec.flags.*.defaultRule.progressiveRollout.initial.variation

`string` · required

The variation this end of the ramp serves. Callers outside the moving
share get the initial variation: everyone gets it before the initial
date. Callers inside the share get the end variation: everyone gets it
after the end date when the end share is 100.

- rule: {"required":true}

### spec.flags.*.defaultRule.progressiveRollout.initial.percentage

`double` · optional (explicit presence)

Share of callers on the END variation at this step's date. Between the
two dates the share moves linearly from the initial step's value to the
end step's. Initial: empty = 0. End: empty or 0 = 100. The end share
must be at least the initial share.

- rule: {"double":{"lte":100,"gte":0}}

### spec.flags.*.defaultRule.progressiveRollout.initial.date

`string` · required

When this step applies, as an RFC 3339 timestamp (e.g.
2026-11-01T09:00:00Z).

- rule: Dates are real RFC 3339 timestamps such as 2026-11-01T09:00:00Z.
- rule: {"required":true}

### spec.flags.*.defaultRule.progressiveRollout.end

`KubernetesGoFeatureFlagProgressiveRolloutStep` · required

Where the ramp ends (served after its date).

- rule: {"required":true}

### spec.flags.*.defaultRule.progressiveRollout.end.variation

`string` · required

The variation this end of the ramp serves. Callers outside the moving
share get the initial variation: everyone gets it before the initial
date. Callers inside the share get the end variation: everyone gets it
after the end date when the end share is 100.

- rule: {"required":true}

### spec.flags.*.defaultRule.progressiveRollout.end.percentage

`double` · optional (explicit presence)

Share of callers on the END variation at this step's date. Between the
two dates the share moves linearly from the initial step's value to the
end step's. Initial: empty = 0. End: empty or 0 = 100. The end share
must be at least the initial share.

- rule: {"double":{"lte":100,"gte":0}}

### spec.flags.*.defaultRule.progressiveRollout.end.date

`string` · required

When this step applies, as an RFC 3339 timestamp (e.g.
2026-11-01T09:00:00Z).

- rule: Dates are real RFC 3339 timestamps such as 2026-11-01T09:00:00Z.
- rule: {"required":true}

### spec.flags.*.defaultRule.disable

`bool`

Skip this rule (it is kept in the file but never matches). Not
allowed on the default rule.

### spec.flags.*.bucketingKey

`string`

Evaluation-context field to bucket percentages on instead of the
targeting key (e.g. companyId, so a whole company gets one variation).

### spec.flags.*.trackEvents

`bool` · optional (explicit presence)

Send this flag's evaluations to the relay's exporters. Empty = true.

### spec.flags.*.disable

`bool`

Turn the flag off: every evaluation returns the caller's default
value.

### spec.flags.*.version

`string`

Free-form version shown in notifications and exported events.

### spec.flags.*.metadata

`object`

Free-form information about the flag (a description, an issue link,
an owner), returned with every evaluation.

### spec.flags.*.experimentation

`KubernetesGoFeatureFlagExperimentation`

Run the flag only inside a time window; outside it every evaluation
returns the caller's default value.

### spec.flags.*.experimentation.start

`string` · required

Window start, as an RFC 3339 timestamp.

- rule: Dates are real RFC 3339 timestamps such as 2026-11-01T09:00:00Z.
- rule: {"required":true}

### spec.flags.*.experimentation.end

`string` · required

Window end, as an RFC 3339 timestamp.

- rule: Dates are real RFC 3339 timestamps such as 2026-11-01T09:00:00Z.
- rule: {"required":true}

### spec.flags.*.scheduledRollout

`[]KubernetesGoFeatureFlagScheduledStep`

Changes applied automatically at given times (e.g. widen a percentage
every day), in date order.

### spec.flags.*.scheduledRollout[].date

`string` · required

When the change applies, as an RFC 3339 timestamp.

- rule: Dates are real RFC 3339 timestamps such as 2026-11-01T09:00:00Z.
- rule: {"required":true}

### spec.flags.*.scheduledRollout[].variations

`map<string, KubernetesGoFeatureFlagValue>`

Variations added or replaced.

### spec.flags.*.scheduledRollout[].variations.*.boolValue

`bool`

A boolean value (the common on/off flag).

### spec.flags.*.scheduledRollout[].variations.*.stringValue

`string`

A string value.

### spec.flags.*.scheduledRollout[].variations.*.numberValue

`double`

A number value (integer or decimal).

### spec.flags.*.scheduledRollout[].variations.*.objectValue

`object`

A JSON object value.

### spec.flags.*.scheduledRollout[].variations.*.listValue

`[]any`

A JSON array value.

### spec.flags.*.scheduledRollout[].targeting

`[]KubernetesGoFeatureFlagRule`

Targeting rules merged by name, field by field (a rule with a new name
is added). A negative percentage share removes that variation from the
rule's split.

### spec.flags.*.scheduledRollout[].targeting[].name

`string`

Rule name. Required for a rule a scheduled step updates (steps match
rules by name).

### spec.flags.*.scheduledRollout[].targeting[].query

`string`

Which evaluation contexts the rule matches: GO Feature Flag's query
language - e.g. `org in ["acme", "globex"]`, `targetingKey sw "beta-"`,
`email ew "@example.com" and country eq "FR"` - or a JSONLogic
expression. Required on every enabled targeting rule; not allowed on
the default rule. The relay checks the syntax when it loads the file:
a query that does not parse drops the whole flag, and every evaluation
of it returns the caller's default value.

### spec.flags.*.scheduledRollout[].targeting[].variation

`string`

The variation returned when the rule matches.

### spec.flags.*.scheduledRollout[].targeting[].percentage

`map<string, double>`

A split across variations (variation name -> share, e.g. enabled: 10,
disabled: 90), bucketed on the targeting key (or bucketing_key) so a
caller keeps its variation across evaluations. Shares are relative
weights (they need not add up to 100). In a scheduled step, a negative
share removes that variation from the split it updates.

### spec.flags.*.scheduledRollout[].targeting[].progressiveRollout

`KubernetesGoFeatureFlagProgressiveRollout`

A gradual ramp from one variation share to another between two
dates.

- rule: A progressive rollout ramps forward: the end share must be at least the initial share.
- rule: A progressive rollout moves callers between two variations - name different ones at its two ends (a single variation for everyone is a plain variation).
- rule: A progressive rollout's end date must come after its initial date.

### spec.flags.*.scheduledRollout[].targeting[].progressiveRollout.initial

`KubernetesGoFeatureFlagProgressiveRolloutStep` · required

Where the ramp starts (served before its date).

- rule: {"required":true}

### spec.flags.*.scheduledRollout[].targeting[].progressiveRollout.initial.variation

`string` · required

The variation this end of the ramp serves. Callers outside the moving
share get the initial variation: everyone gets it before the initial
date. Callers inside the share get the end variation: everyone gets it
after the end date when the end share is 100.

- rule: {"required":true}

### spec.flags.*.scheduledRollout[].targeting[].progressiveRollout.initial.percentage

`double` · optional (explicit presence)

Share of callers on the END variation at this step's date. Between the
two dates the share moves linearly from the initial step's value to the
end step's. Initial: empty = 0. End: empty or 0 = 100. The end share
must be at least the initial share.

- rule: {"double":{"lte":100,"gte":0}}

### spec.flags.*.scheduledRollout[].targeting[].progressiveRollout.initial.date

`string` · required

When this step applies, as an RFC 3339 timestamp (e.g.
2026-11-01T09:00:00Z).

- rule: Dates are real RFC 3339 timestamps such as 2026-11-01T09:00:00Z.
- rule: {"required":true}

### spec.flags.*.scheduledRollout[].targeting[].progressiveRollout.end

`KubernetesGoFeatureFlagProgressiveRolloutStep` · required

Where the ramp ends (served after its date).

- rule: {"required":true}

### spec.flags.*.scheduledRollout[].targeting[].progressiveRollout.end.variation

`string` · required

The variation this end of the ramp serves. Callers outside the moving
share get the initial variation: everyone gets it before the initial
date. Callers inside the share get the end variation: everyone gets it
after the end date when the end share is 100.

- rule: {"required":true}

### spec.flags.*.scheduledRollout[].targeting[].progressiveRollout.end.percentage

`double` · optional (explicit presence)

Share of callers on the END variation at this step's date. Between the
two dates the share moves linearly from the initial step's value to the
end step's. Initial: empty = 0. End: empty or 0 = 100. The end share
must be at least the initial share.

- rule: {"double":{"lte":100,"gte":0}}

### spec.flags.*.scheduledRollout[].targeting[].progressiveRollout.end.date

`string` · required

When this step applies, as an RFC 3339 timestamp (e.g.
2026-11-01T09:00:00Z).

- rule: Dates are real RFC 3339 timestamps such as 2026-11-01T09:00:00Z.
- rule: {"required":true}

### spec.flags.*.scheduledRollout[].targeting[].disable

`bool`

Skip this rule (it is kept in the file but never matches). Not
allowed on the default rule.

### spec.flags.*.scheduledRollout[].defaultRule

`KubernetesGoFeatureFlagRule`

Fields merged into the default rule.

### spec.flags.*.scheduledRollout[].defaultRule.name

`string`

Rule name. Required for a rule a scheduled step updates (steps match
rules by name).

### spec.flags.*.scheduledRollout[].defaultRule.query

`string`

Which evaluation contexts the rule matches: GO Feature Flag's query
language - e.g. `org in ["acme", "globex"]`, `targetingKey sw "beta-"`,
`email ew "@example.com" and country eq "FR"` - or a JSONLogic
expression. Required on every enabled targeting rule; not allowed on
the default rule. The relay checks the syntax when it loads the file:
a query that does not parse drops the whole flag, and every evaluation
of it returns the caller's default value.

### spec.flags.*.scheduledRollout[].defaultRule.variation

`string`

The variation returned when the rule matches.

### spec.flags.*.scheduledRollout[].defaultRule.percentage

`map<string, double>`

A split across variations (variation name -> share, e.g. enabled: 10,
disabled: 90), bucketed on the targeting key (or bucketing_key) so a
caller keeps its variation across evaluations. Shares are relative
weights (they need not add up to 100). In a scheduled step, a negative
share removes that variation from the split it updates.

### spec.flags.*.scheduledRollout[].defaultRule.progressiveRollout

`KubernetesGoFeatureFlagProgressiveRollout`

A gradual ramp from one variation share to another between two
dates.

- rule: A progressive rollout ramps forward: the end share must be at least the initial share.
- rule: A progressive rollout moves callers between two variations - name different ones at its two ends (a single variation for everyone is a plain variation).
- rule: A progressive rollout's end date must come after its initial date.

### spec.flags.*.scheduledRollout[].defaultRule.progressiveRollout.initial

`KubernetesGoFeatureFlagProgressiveRolloutStep` · required

Where the ramp starts (served before its date).

- rule: {"required":true}

### spec.flags.*.scheduledRollout[].defaultRule.progressiveRollout.initial.variation

`string` · required

The variation this end of the ramp serves. Callers outside the moving
share get the initial variation: everyone gets it before the initial
date. Callers inside the share get the end variation: everyone gets it
after the end date when the end share is 100.

- rule: {"required":true}

### spec.flags.*.scheduledRollout[].defaultRule.progressiveRollout.initial.percentage

`double` · optional (explicit presence)

Share of callers on the END variation at this step's date. Between the
two dates the share moves linearly from the initial step's value to the
end step's. Initial: empty = 0. End: empty or 0 = 100. The end share
must be at least the initial share.

- rule: {"double":{"lte":100,"gte":0}}

### spec.flags.*.scheduledRollout[].defaultRule.progressiveRollout.initial.date

`string` · required

When this step applies, as an RFC 3339 timestamp (e.g.
2026-11-01T09:00:00Z).

- rule: Dates are real RFC 3339 timestamps such as 2026-11-01T09:00:00Z.
- rule: {"required":true}

### spec.flags.*.scheduledRollout[].defaultRule.progressiveRollout.end

`KubernetesGoFeatureFlagProgressiveRolloutStep` · required

Where the ramp ends (served after its date).

- rule: {"required":true}

### spec.flags.*.scheduledRollout[].defaultRule.progressiveRollout.end.variation

`string` · required

The variation this end of the ramp serves. Callers outside the moving
share get the initial variation: everyone gets it before the initial
date. Callers inside the share get the end variation: everyone gets it
after the end date when the end share is 100.

- rule: {"required":true}

### spec.flags.*.scheduledRollout[].defaultRule.progressiveRollout.end.percentage

`double` · optional (explicit presence)

Share of callers on the END variation at this step's date. Between the
two dates the share moves linearly from the initial step's value to the
end step's. Initial: empty = 0. End: empty or 0 = 100. The end share
must be at least the initial share.

- rule: {"double":{"lte":100,"gte":0}}

### spec.flags.*.scheduledRollout[].defaultRule.progressiveRollout.end.date

`string` · required

When this step applies, as an RFC 3339 timestamp (e.g.
2026-11-01T09:00:00Z).

- rule: Dates are real RFC 3339 timestamps such as 2026-11-01T09:00:00Z.
- rule: {"required":true}

### spec.flags.*.scheduledRollout[].defaultRule.disable

`bool`

Skip this rule (it is kept in the file but never matches). Not
allowed on the default rule.

### spec.flags.*.scheduledRollout[].trackEvents

`bool` · optional (explicit presence)

Turn event export on or off from this date.

### spec.flags.*.scheduledRollout[].disable

`bool` · optional (explicit presence)

Turn the flag off (true) or back on (false) from this date.

### spec.flags.*.scheduledRollout[].version

`string`

A replacement version.

### spec.flags.*.scheduledRollout[].experimentation

`KubernetesGoFeatureFlagExperimentation`

A replacement experimentation window.

### spec.flags.*.scheduledRollout[].experimentation.start

`string` · required

Window start, as an RFC 3339 timestamp.

- rule: Dates are real RFC 3339 timestamps such as 2026-11-01T09:00:00Z.
- rule: {"required":true}

### spec.flags.*.scheduledRollout[].experimentation.end

`string` · required

Window end, as an RFC 3339 timestamp.

- rule: Dates are real RFC 3339 timestamps such as 2026-11-01T09:00:00Z.
- rule: {"required":true}

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesGoFeatureFlagFlagFile, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.config_map_name` | `string` | Name of the rendered ConfigMap. |
| `status.outputs.key` | `string` | ConfigMap data key holding the flag file. |
| `status.outputs.namespace` | `string` | Namespace of the rendered ConfigMap. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.namespace` | KubernetesNamespace | `spec.name` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| KubernetesGoFeatureFlag | `spec.flagSource.retrievers[].configMap.configMapName` | `status.outputs.config_map_name` |
| KubernetesGoFeatureFlag | `spec.flagSource.retrievers[].configMap.key` | `status.outputs.key` |
| KubernetesGoFeatureFlag | `spec.flagSets.items[].source.retrievers[].configMap.configMapName` | `status.outputs.config_map_name` |
| KubernetesGoFeatureFlag | `spec.flagSets.items[].source.retrievers[].configMap.key` | `status.outputs.key` |

## See Also

- [Overview](../README.md)
