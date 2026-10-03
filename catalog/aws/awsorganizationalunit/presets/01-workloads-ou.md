# Workloads OU

This preset creates the standard first-level container: a Workloads OU
directly under the organization root, wired by reference.

## When to Use

- The first OU in any fresh organization — production and staging
  accounts live here, guardrails attach here
- Any first-level grouping (swap the name for Security, Sandbox,
  Infrastructure)

## What You Get

- One OU under the root, with the parent resolved from your
  AwsOrganization resource's `root_id` output — no hand-copied IDs

## Customize

- `ouName` takes spaces and arbitrary characters ("Core Services") —
  renames apply in place
- Place accounts in it with
  [AWS Organization Account](/infra-catalog/aws-organization-account)
  (`parentId` → this OU's `ou_id` output)
- Attach guardrails with
  [AWS Organization Policy](/infra-catalog/aws-organization-policy)
