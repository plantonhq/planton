# Automation Runbook

This preset creates an Automation document — a multi-step runbook
driving AWS APIs (here: stop, resize, restart an EC2 instance) with
typed, per-run parameters.

## When to Use

- Operational procedures that touch AWS APIs rather than run scripts
  on nodes — resizing, snapshotting, remediation
- Runbooks that maintenance windows or incident automation should
  execute unattended

## What You Get

- A schema-0.3 Automation document in YAML with declared parameters
- A `versionName` release label pinning this content

## Customize

- Add `aws:approve` steps for human gates on destructive operations
- Run it on a schedule inside
  [AWS SSM Maintenance Window](/infra-catalog/aws-ssm-maintenance-window)
  (task type `AUTOMATION`), or rate-controlled across many instances
  via [AWS SSM Association](/infra-catalog/aws-ssm-association) with
  `automationTargetParameterName: InstanceId`
