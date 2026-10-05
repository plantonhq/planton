# Percentage Rollout

A flag rolled out to a percentage of callers with flagd's `fractional` operation, plus a shared evaluator that always gives staff the new variant. Raise the percentage by editing the two weights; callers keep their bucket across evaluations because `fractional` hashes the targeting key.

## When to Use

- Gradually exposing a new implementation behind a string or number variant
- Testing with internal users first, then a slice of everyone else

## Key Configuration Choices

- `evaluators.isStaff` is rendered as `$evaluators` and referenced with `{"$ref": "isStaff"}`; an evaluator cannot reference another
- `fractional` takes `[variant, weight]` pairs; the weights here are 10 and 90
- `defaultVariant: current` is the answer when targeting returns `null`

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<namespace>` | The KubernetesFlagd's namespace | The flagd resource's `spec.namespace` |
| `@example.com` | Your staff email domain | Your identity provider |

## Related Presets

- **Release Flags** -- an on/off flag targeted by organization
