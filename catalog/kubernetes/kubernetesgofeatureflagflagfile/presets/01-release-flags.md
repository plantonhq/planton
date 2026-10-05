# Release Flags

A release flag file: each flag hides something new from everyone except the organizations it is ready for. Turning a flag on for one more organization is a one-line change to the targeting query; the relay serves it within its polling interval, with no restart.

## When to Use

- Shipping work dark and turning it on organization by organization.
- Any boolean switch whose audience is a list you name.
- The starting point for every flag file a KubernetesGoFeatureFlag relay reads.

## Key Configuration Choices

- `variations` holds the two values; every variation of a flag must hold the same type.
- `targeting[0].query` matches the evaluation context; `org in [...]` targets the organizations that callers send as the `org` attribute.
- `defaultRule.variation: disabled` keeps the flag off for everyone else - the default rule takes no query.
- `metadata` travels with every evaluation; an owner and a description make the flag's purpose obvious to whoever retires it.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<namespace>` | Namespace of the rendered ConfigMap - the relay's namespace, or one its retriever names | Your relay's namespace |
| `<organization-slug>` | Organizations the flag is on for | Your organizations' identifiers as callers send them |
| `<team>` | Team that owns the flag | Your team directory |
| `<what the flag hides>` | One line on the feature behind the flag | The feature's description |

## Related Presets

- **Progressive Rollout** - widen a flag by percentage between two dates.
