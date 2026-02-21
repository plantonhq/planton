---
title: "Standard DevOps Project"
description: "This preset creates an OCI DevOps project wired to an Oracle Notification Service (ONS) topic for pipeline event delivery. The project serves as the organizational container for build pipelines,..."
type: "preset"
rank: "01"
presetSlug: "01-standard-project"
componentSlug: "devops-project"
componentTitle: "DevOps Project"
provider: "oci"
icon: "package"
order: 1
---

# Standard DevOps Project

This preset creates an OCI DevOps project wired to an Oracle Notification Service (ONS) topic for pipeline event delivery. The project serves as the organizational container for build pipelines, deploy pipelines, code repositories, artifact references, and triggers. All downstream DevOps resources reference the project by its OCID, making this the required first step for any CI/CD workflow on OCI.

## When to Use

- Setting up a new CI/CD pipeline for an application or microservice
- Creating a shared DevOps namespace for a team that manages multiple pipelines and repositories
- Bootstrapping OCI DevOps before adding build pipelines, deploy pipelines, or code repositories
- Any project that needs automated notifications for build completions, deployment successes, and failures

## Key Configuration Choices

- **Notification topic** (`notificationTopicId`) -- every DevOps project requires an ONS topic. Pipeline events (build started, deployment succeeded, approval required) are published to this topic. Subscribers can be email, Slack webhooks, PagerDuty, or custom HTTPS endpoints configured on the ONS topic itself.
- **Description** (`description`) -- a human-readable summary that appears in the OCI Console and API responses. Helps teams identify the project's purpose when multiple projects exist in the same compartment.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the project will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<ons-topic-ocid>` | OCID of the ONS notification topic for pipeline events | OCI Console > Developer Services > Notifications > Topics |
