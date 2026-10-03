# Custom Document Schedule

This preset binds YOUR document — wired by reference to an
[AWS SSM Document](/infra-catalog/aws-ssm-document) component — to
tagged instances on a weekly rate schedule, pinned to the document's
default version.

## When to Use

- Keeping fleet software converged with your own runbook (agent
  installs, config enforcement) instead of an AWS-managed document
- Chart wiring: the association follows the document component through
  the reference, so releasing a new default version updates what runs
  with no association edit

## What You Get

- A weekly run of the referenced document's `$DEFAULT` version with a
  per-run parameter
- MEDIUM-severity compliance findings when a target fails

## Customize

- Pin `documentVersion` to a number for strict rollouts, or `$LATEST`
  to track unreleased edits
- Add `applyOnlyAtCronInterval: true` (with a cron schedule) to skip
  the immediate first run
- For Automation documents, set `automationTargetParameterName` to
  fan the targets into a runbook parameter
