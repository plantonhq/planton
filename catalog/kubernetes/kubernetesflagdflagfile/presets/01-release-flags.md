# Release Flags

A release flag that hides a new capability from everyone except the organizations listed in its targeting. The flag defaults to `off`, and the JSONLogic rule turns it `on` when the evaluation context's `org` is in the list. Widening the release is a one-line edit to this flag file -- the KubernetesFlagd serving it is never re-applied.

## When to Use

- Shipping a feature dark and enabling it organization by organization
- Any boolean switch keyed on a context attribute your services send

## Key Configuration Choices

- `defaultVariant: "off"` makes the safe answer the default; an organization not in the list never sees the feature
- `targeting` is JSONLogic: `{"if": [{"in": [{"var": "org"}, [...]]}, "on", "off"]}`
- `metadata.flagSetId` lets SDKs and gRPC sync selectors filter on this file's flags
- `metadata.owner` on the flag records who retires it

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<namespace>` | The KubernetesFlagd's namespace | The flagd resource's `spec.namespace` |
| `<flag-set-id>` | Identifier for this file's flags | Your naming plan (for example the owning service) |
| `<first-organization>` | Organization that sees the feature first | The `org` value your services put in the evaluation context |
| `<owning-team>` | Team that owns and retires the flag | Your team directory |

## Related Presets

- **Percentage Rollout** -- the same file shape with fractional bucketing and a shared evaluator
