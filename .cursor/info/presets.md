# Preset Authoring Guide

Purpose: create production-quality, deployable YAML presets with companion markdown for a deployment component.

## Inputs to read
- `api.proto` — `api_version` and `kind` constant values
- `spec.proto` — all fields, types, validations, default annotations (`recommended_default`, `default`)
- `docs/README.md` — design rationale and common deployment patterns
- `iac/hack/manifest.yaml` — structural reference for KRM envelope
- Existing `presets/` directory — avoid duplication, determine next available rank

## Path
- `apis/org/openmcf/provider/<provider>/<kindfolder>/v1/presets/`

## File naming
- `{rank}-{description}.yaml` + `{rank}-{description}.md`
- Rank: zero-padded two-digit (`01` = most common, `02` = next)
- Description: lowercase, hyphenated, 3-5 words (e.g., `internet-facing-https`)
- Every `.yaml` MUST have a companion `.md` with the same base name

## YAML skeleton
```yaml
apiVersion: <provider>.openmcf.org/v1
kind: <Kind>
metadata:
  name: my-<descriptive-name>
spec:
  # Required fields with placeholders or real values
  # Optional fields with sensible defaults and comments
```

- `apiVersion` and `kind` MUST match the constants in `api.proto`
- `metadata` contains only `name` (no `org`, `env`, or `version`)
- `name` should be prefixed with `my-` and describe the preset purpose
- Do not include `status` — it is system-managed
- Use bare (unquoted) values wherever YAML allows — only quote when syntax requires it

## StringValueOrRef fields — CRITICAL

For fields typed as `StringValueOrRef` (subnets, security groups, ARNs, zone IDs), always use the `value:` wrapper:

CORRECT:
```yaml
subnets:
  - value: "<public-subnet-id-az1>"
  - value: "<public-subnet-id-az2>"
certificateArn:
  value: "<acm-certificate-arn>"
```

WRONG (will not deserialize):
```yaml
subnets:
  - "<public-subnet-id-az1>"
certificateArn: "<acm-certificate-arn>"
```

Do NOT use `valueFrom:` in presets. Use `value:` with descriptive angle-bracket placeholders.

## Placeholder convention
- Angle brackets: `<lowercase-hyphenated-description>`
- Describe what the value represents, not the field name
- Add context for repeated fields: `<public-subnet-id-az1>`, `<public-subnet-id-az2>`

## Default values
- If a field has `(org.openmcf.shared.options.recommended_default)` or `(org.openmcf.shared.options.default)` in `spec.proto`, use that value with a comment:
  ```yaml
  idleTimeoutSeconds: 60  # recommended_default from spec
  ```
- For opinionated production choices, use real values with explanatory comments:
  ```yaml
  deleteProtectionEnabled: true  # Recommended for production
  ```

## MD skeleton
```markdown
# <Preset Title>

<2-4 sentence description of what this preset configures and why.>

## When to Use

- <Scenario 1>
- <Scenario 2>

## Key Configuration Choices

- **<Choice>** (`field: value`) -- <rationale>

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<placeholder>` | What it is | Where to get it |
```

Optional section: `## Related Presets` — reference other presets in the same component.

## Ranking
- `01` = the config you'd deploy with 30 seconds to decide (usually standard production)
- `02` = second most common (often the opposite: internal vs external, dev vs prod)
- `03+` = specialized patterns (HA, cost-optimized, use-case variants)
- Rank by real-world frequency, NOT by complexity

## How many presets
- Minimum: 1 per component
- Recommended: 2-4
- Maximum: 5-6 for high-variety components
- Do NOT create presets for edge cases serving <10% of deployments

## Notes
- Presets are deployable artifacts, not documentation. README.md serves the documentation role.
- Read `architecture/presets.md` for the full convention reference.
- Quality over quantity: 1 excellent preset > 5 mediocre ones.
