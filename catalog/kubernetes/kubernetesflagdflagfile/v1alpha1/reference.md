# KubernetesFlagdFlagFile

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `kubernetes.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

**KubernetesFlagdFlagFileSpec** declares a flagd flag definition file -
typed flags with their variants, default variant and JSONLogic targeting,
plus shared evaluators - and renders it as JSON (with the flagd `$schema`)
into a ConfigMap named after the resource, which a KubernetesFlagd
`config_map` source mounts as a directory.

WHY A SEPARATE RESOURCE: flags change far more often than the daemon, are
many per daemon, and are often owned by a different team. A flag flip
touches this resource only - flagd is never re-applied, and it serves the
change once the kubelet syncs the mounted volume (typically one to two
minutes), with no restart.

## Example

```yaml
# Full-surface development manifest - boolean, string, number and object
# variants; JSONLogic targeting with flagd's operations (fractional
# bucketing, sem_ver, ends_with) and a shared evaluator; a disabled flag; a
# flag that defers to the caller's code default; flag-set metadata.
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesFlagdFlagFile
metadata:
  name: flagd-dev-flags
spec:
  namespace:
    value: flagd-dev
  key: flags.flagd.json
  metadata:
    flagSetId: web
  evaluators:
    isStaff:
      ends_with:
        - var: email
        - "@example.com"
  flags:
    assistant:
      state: ENABLED
      variants:
        "on":
          boolValue: true
        "off":
          boolValue: false
      defaultVariant: "off"
      targeting:
        if:
          - in:
              - var: org
              - - planton
                - acme
          - "on"
          - "off"
      metadata:
        owner: platform
    checkout-color:
      state: ENABLED
      variants:
        red:
          stringValue: "#c00"
        blue:
          stringValue: "#00c"
      defaultVariant: red
      targeting:
        fractional:
          - - red
            - 50
          - - blue
            - 50
    staff-banner:
      state: ENABLED
      variants:
        shown:
          boolValue: true
        hidden:
          boolValue: false
      defaultVariant: hidden
      targeting:
        if:
          - $ref: isStaff
          - shown
          - hidden
    request-limit:
      state: ENABLED
      variants:
        standard:
          numberValue: 10
        premium:
          numberValue: 100
      defaultVariant: standard
      targeting:
        if:
          - sem_ver:
              - var: appVersion
              - ">="
              - 2.0.0
          - premium
          - standard
    limits:
      state: DISABLED
      variants:
        small:
          objectValue:
            rps: 10
        large:
          objectValue:
            rps: 100
      defaultVariant: small
    code-default:
      state: ENABLED
      variants:
        "on":
          boolValue: true
        "off":
          boolValue: false
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.namespace` | `string \| valueFrom` | yes |  | KubernetesNamespace (`spec.name`) |
| `spec.key` | `string` |  | `flags.flagd.json` |  |
| `spec.flags` | `map<string, KubernetesFlagdFlag>` |  |  |  |
| `spec.flags.*.state` | `string` | yes |  |  |
| `spec.flags.*.variants` | `map<string, KubernetesFlagdValue>` | yes |  |  |
| `spec.flags.*.variants.*.boolValue` | `bool` |  |  |  |
| `spec.flags.*.variants.*.stringValue` | `string` |  |  |  |
| `spec.flags.*.variants.*.numberValue` | `double` |  |  |  |
| `spec.flags.*.variants.*.objectValue` | `object` |  |  |  |
| `spec.flags.*.defaultVariant` | `string` |  |  |  |
| `spec.flags.*.targeting` | `object` |  |  |  |
| `spec.flags.*.metadata` | `object` |  |  |  |
| `spec.evaluators` | `map<string, object>` |  |  |  |
| `spec.metadata` | `object` |  |  |  |

## Field Details

### spec.namespace

`string | valueFrom` · required

Namespace of the rendered ConfigMap - the KubernetesFlagd's namespace
(pod volumes cannot cross namespaces). Accepts a literal namespace name
or a reference to a KubernetesNamespace resource.

- references: KubernetesNamespace (`spec.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: KubernetesNamespace, name: <that resource's name>, fieldPath: spec.name}} -- a bare string does not parse

### spec.key

`string` · optional (explicit presence)

ConfigMap data key the definitions are written under. It must end in
.json (flagd picks the parser from the extension).

- default: `flags.flagd.json`
- rule: The key must end in .json and contain only letters, digits, '-', '_' and '.'.

### spec.flags

`map<string, KubernetesFlagdFlag>`

The flags, keyed by flag key (the key OpenFeature SDKs evaluate).

- rule: Flag keys may not be empty or contain whitespace.
- rule: Every variant of a flag must hold the same value type (all booleans, all strings, all numbers or all objects).
- rule: default_variant must name one of the flag's variants.

### spec.flags.*.state

`string` · required

ENABLED serves the flag; DISABLED makes every evaluation return the
caller's default value with reason DISABLED.

- rule: state must be ENABLED or DISABLED.
- rule: {"required":true}

### spec.flags.*.variants

`map<string, KubernetesFlagdValue>` · required

The values this flag can return, keyed by variant name. Every variant
must hold the same value type.

- rule: Variant names cannot be empty.
- rule: {"map":{"minPairs":"1"}}

### spec.flags.*.variants.*.boolValue

`bool`

A boolean value.

### spec.flags.*.variants.*.stringValue

`string`

A string value.

### spec.flags.*.variants.*.numberValue

`double`

A number value (integer or decimal).

### spec.flags.*.variants.*.objectValue

`object`

A JSON object value.

### spec.flags.*.defaultVariant

`string`

The variant returned when targeting does not choose one. Empty = the
caller's code default (reason DEFAULT, no value).

### spec.flags.*.targeting

`object`

JSONLogic deciding the variant from the evaluation context, with
flagd's operations: fractional (percentage bucketing), sem_ver,
starts_with, ends_with - e.g. {"if": [{"ends_with": [{"var": "email"},
"@example.com"]}, "on", "off"]}. It returns a variant name: null falls
back to default_variant, and a name the flag does not define is an
evaluation error.

### spec.flags.*.metadata

`object`

Information about the flag (a description, an owner), returned with
evaluations: scalar values only (strings, numbers, booleans).
flagSetId and version must be strings.

- rule: Metadata values must be strings, numbers or booleans, and flagSetId and version must be strings.

### spec.evaluators

`map<string, object>`

Shared JSONLogic fragments (`$evaluators`), reused from any flag's
targeting with {"$ref": "<name>"} (one level; an evaluator cannot
reference another).

- rule: An evaluator cannot be empty (flagd rejects the whole file).

### spec.metadata

`object`

Flag-set metadata merged into every flag's metadata: scalar values only
(strings, numbers, booleans) - e.g. flagSetId, which gRPC source
selectors and SDKs filter on. flagSetId and version must be strings.

- rule: Metadata values must be strings, numbers or booleans, and flagSetId and version must be strings.

## Outputs

Reference an output from another manifest as `valueFrom: {kind: KubernetesFlagdFlagFile, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.config_map_name` | `string` | Name of the rendered ConfigMap. |
| `status.outputs.key` | `string` | ConfigMap data key holding the definitions. |
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
| KubernetesFlagd | `spec.sources[].configMap.configMapName` | `status.outputs.config_map_name` |
| KubernetesFlagd | `spec.sources[].configMap.key` | `status.outputs.key` |

## See Also

- [Overview](../README.md)
