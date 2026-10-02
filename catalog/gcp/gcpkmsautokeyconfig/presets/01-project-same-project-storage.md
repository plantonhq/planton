# Project Same-Project Storage

## Use Case

Let every team in one project get customer-managed encryption keys on demand, with each key living beside the resource it protects.

## When to Use

- A project that must use CMEK for buckets, disks, datasets, or topics, without a central key team
- Trying Autokey on one project before rolling it out to a folder
- A project with no organization folder above it

## What This Creates

- The Cloud KMS API on the project
- The project's Autokey configuration with same-project key storage

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `scope.projectId` | `my-gcp-project` | Reference a `GcpProject` instead of a literal. |
| `keyProjectResolutionMode` | `RESOURCE_PROJECT` | `DISABLED` to opt this project out of its folder's Autokey. |
| `deletionPolicy` | `DELETE` | `PREVENT` when the project depends on Autokey staying on. |
