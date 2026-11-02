---
title: "Flow Control"
description: "What a resource's annotations and its environment decide about how its Infra Jobs run: unmanaged resources, jobs that wait for approval, a pause between preview and apply, and what a stack's plan reads."
icon: guide
order: 60
tags:
  - Infrastructure
  - Flow Control
  - Governance
---

# Flow Control

Flow control decides how cautiously an [Infra Component](/docs/infrastructure/infra-components)'s [Infra Jobs](/docs/infrastructure/infra-jobs) run: whether a change to the resource starts a job at all, whether that job waits for a person before it begins, and whether it pauses between its preview and its apply. It belongs to the resource, written where the resource is declared. Who may change an environment, and whether a change there needs an approver, belongs to the environment: a protected environment adds its review on top of whatever the resource says.

## Why Flow Control Exists

Most resources should simply deploy when their declaration changes. A few should not: a resource another team created by hand and you only want recorded, a shared database whose every change should wait for someone, a resource whose plan you want to read before it applies. Flow control says so once, on the resource, and every job for it follows.

## A Resource's Own Settings

Each setting is off unless the resource turns it on.

### Unmanaged: Lifecycle Events Disabled

```yaml
metadata:
  annotations:
    planton.dev/infra-job.disable-on-lifecycle-events: "true"
```

Creating or updating the resource starts no Infra Job: Planton records it without deploying anything. Inside an Infra Stack or a Service deployment it reads **Unmanaged**: the run records it and never applies or destroys it. A stack's plan reads it Unmanaged; removing it from the templates plans **Keep** (named by its provider: Keep in AWS), with the choice locked; a stack's undeploy or purge releases it and leaves it running; a service's deletion leaves it too.

Deliberate acts on that one resource are never held back by it: deleting or purging the Infra Component, creating an Infra Job for it, or a targeted destroy run as asked — unless the resource is protected from destroy (below).

### Manual: Jobs Wait for Approval

```yaml
metadata:
  annotations:
    planton.dev/infra-job.is-manual: "true"
```

A job created by a change to the resource is born waiting for approval, and nothing runs until someone approves it. This holds for a change made to the resource directly. Inside a stack's or a service's pipeline a resource never waits this way: the pipeline's own approval carries it — a request in a protected environment, or a manual gate on the resource's node (`planton.dev/gate.type: manual`, see [Infra Pipelines](/docs/infrastructure/infra-pipelines#node-level-gates)).

### Pause Between Preview and Apply

The Infra Component's typed `spec.flow_control` holds both settings above and a third: `pause_between_preview_and_update_or_destroy`. With it on, a job that updates or destroys runs its preview, then holds the apply or destroy until someone approves the plan it showed. A value set on the typed field wins over the annotations.

A single job may also ask for this pause on its own — a destroy you want to read before it runs — without changing the resource's standing settings.

## What the Environment Adds

- **A protected environment** forces the pause between preview and apply on every job that changes a resource there directly, and the approval must come from someone other than the person who started it while another approver exists. A stack's run into a protected environment is a **request** instead: the approver reads the exact plan once, before anything applies, and the jobs it runs do not ask again (see [Infra Stacks](/docs/infrastructure/infra-stacks#in-a-protected-environment-a-change-is-a-request)).
- **An environment in a Promotion Only stage** changes by promotion from an earlier stage of its path. Promotion Only belongs to the stage on its [Promotion Path](/docs/ci-cd/deployment-environments#promotion-paths-the-order-a-push-walks), not to the environment: it covers that stage and every later one, so moving an environment to another stage changes whether it is Promotion Only. For a service, pushes stop there. For an Infra Stack, a template edit that did not arrive by promotion is cautioned on its plan ("prod changes by promotion from staging. This edit makes prod differ from staging.") and never refused; protection still decides who may make it.

## Annotations a Stack's Plan Reads

Two annotations on a resource decide what an [Infra Stack](/docs/infrastructure/infra-stacks) may do when it lets go of the resource. A plan reads both before anything runs, and adding or removing either changes no deployment — the plan reads **Update Record**: only Planton's record changes.

| Annotation | Whenever the stack lets go of the resource | In a plan that removes it |
|------------|---------------------------------------------|---------------------------|
| `planton.dev/keep-on-provider: "true"` | Released, never destroyed: removed from the templates, undeployed or purged, it keeps running on its provider, and another stack can adopt it or a service can take it over. | **Keep** (named by its provider: Keep in AWS), the choice locked: "Marked planton.dev/keep-on-provider: its stack never destroys it". |
| `planton.dev/prevent-destroy: "true"` | Refused: an undeploy or purge is refused before anything starts, naming the resource. | **Blocked** from destroying — choose Keep to apply the plan, or remove the annotation to destroy it. |

The annotation is not the only thing that blocks a destroy: the resource's own deletion protection, or anything else still reading it (another stack, a service, a resource deployed on its own), blocks it the same way, and the row says which.

A [Service](/docs/ci-cd/what-is-a-service) reads both marks when a run removes a resource it no longer declares: a resource marked keep-on-provider reads **Kept** (Kept in AWS) and keeps running, and one marked prevent-destroy reads **Left in Place**, naming the mark, until it is lifted and the next run destroys it ([Removing a Resource](/docs/ci-cd/what-is-a-service#removing-a-resource)). It honours both the same way when it is deleted: a resource marked keep-on-provider is released and keeps running, and one marked prevent-destroy refuses the delete before anything starts, naming the resource ("checkout-api can't be deleted: orders-db in prod is marked prevent-destroy. Remove the mark, or delete with Keep Resources.").

Keep answers "this resource outlives its owner"; prevent-destroy answers "stop anyone from tearing this down by mistake". A removed resource without either is destroyed when its plan says Destroy and you agreed to that plan.

## Common Patterns

| You want | Write |
|----------|-------|
| Changes apply as soon as they are declared | Nothing: every setting is off by default |
| A resource Planton records and never deploys | `planton.dev/infra-job.disable-on-lifecycle-events: "true"` |
| Every direct change to a resource waits for someone | `planton.dev/infra-job.is-manual: "true"` |
| Read the plan before each update or destroy applies | `spec.flow_control.pause_between_preview_and_update_or_destroy: true` |
| Every change in an environment reviewed by a second person | Protect the environment |
| A resource that outlives its stack or service | `planton.dev/keep-on-provider: "true"` |
| A resource nobody tears down by mistake | `planton.dev/prevent-destroy: "true"` |

## Where to See It

The Infra Job's page shows the flow control it ran with — the resource's settings, with the environment's pause added — so what ran and what is shown never differ.

## Related Documentation

- [Infra Jobs](/docs/infrastructure/infra-jobs) — The execution units flow control governs
- [Infra Components](/docs/infrastructure/infra-components) — The resources that carry these settings
- [Infra Pipelines](/docs/infrastructure/infra-pipelines) — Requests and per-resource gates for multi-resource deployments
- [Infra Stacks](/docs/infrastructure/infra-stacks) — Plans, revisions, and what a stack does with what it no longer declares
