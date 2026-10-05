# Weekly Patch Window

This preset creates the classic Sunday-night patch window: four hours,
new work stops in the last hour, and a rate-controlled
`AWS-RunPatchBaseline` install runs across the tagged fleet.

## When to Use

- Production fleets whose patching must happen inside a declared,
  auditable window rather than whenever State Manager gets around to
  it
- Anywhere the change calendar says "Sunday night only"

## What You Get

- A window opening Sundays 02:00 (your timezone) for 4 hours, cutoff 1
  hour before close, with running installs CANCELLED at the cutoff
- A tag-registered target set and a patch-install task moving 10% of
  the fleet at a time, stopping at 5% failures

## Customize

- Set `Operation: Scan` first to observe before installing (pair with
  [AWS SSM Association](/infra-catalog/aws-ssm-association) for
  continuous scanning)
- Add `outputLocation`/`cloudwatchConfig` on the invocation to keep
  command output
- Register more targets and reference them from tasks with the
  `WindowTargetIds` key using the `target_ids` output
