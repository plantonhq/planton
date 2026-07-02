# Overview

The AwsIamPolicy API resource provisions a customer-managed IAM policy: a
standalone, versioned permission document with its own ARN that can be attached
to many roles and users at once, or used as a permissions boundary.

## Why We Created This API Resource

A managed policy is the reusable unit of AWS permissions. Modeling it as a
first-class component -- instead of burying permission documents inside every
role -- lets one definition serve an entire architecture. It lets you:

- **Define permissions once, attach everywhere**: roles and users reference the
  policy through their `managedPolicyArns` fields, so "read-only access to the
  analytics bucket" exists in exactly one place.
- **Set permissions boundaries**: the same policy ARN plugs into a role's or
  user's `permissionsBoundary` field to cap the maximum permissions a principal
  can ever have.
- **Update centrally**: changing the document updates every attached principal
  at once -- AWS versions the document, and the modules keep the version history
  within AWS's 5-version limit automatically.

## Key Features

### Permission Document

- **Free-form JSON document**: the full IAM policy language (statements,
  conditions, NotAction, resource patterns) with no schema loss.
- **Versioned updates**: document changes create a new default policy version;
  older versions are pruned automatically so updates never hit AWS's 5-version
  cap.

### Organization

- **IAM path**: group policies into path hierarchies (e.g.
  `/service-boundaries/`) that other IAM policies can match with wildcards.
- **Description**: a human-readable statement of intent, shown in the IAM
  console.

## Benefits

- **Composability**: roles, users, and boundaries reference the policy by ARN
  through `valueFrom`, so the architecture graph shows exactly which principals
  carry which permissions.
- **Least-privilege at scale**: shared definitions make narrow, well-reviewed
  permission sets practical across many principals.
- **Consistency**: identical behavior across Terraform and Pulumi.

## Stack outputs

- `policy_arn`: ARN of the managed policy (what attachments and boundaries reference)
- `policy_id`: stable unique ID AWS assigns to the policy
- `policy_name`: friendly name of the policy
