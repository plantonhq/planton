# GcpDeployPolicy Guide

The judgment this guide protects: a deploy policy is a release-wide rule, not a pipeline setting. One policy governs many pipelines and targets, a pipeline can fall under many policies, and the override is the emergency door -- not the normal path.

## Choosing what it governs

Selectors pick pipelines and targets in the policy's own project and location. A selector's `deliveryPipeline` and `target` each match by `id` (a reference to the pipeline's `delivery_pipeline_id` or the target's `target_id`, a literal ID, or `*` for all of them in the location) and by `labels`; every attribute you give in one selector must match. Across selectors it is the other way round: the policy applies when any selector matches. Label selectors age best -- a new production target labeled `env: prod` is covered the moment it exists, while an ID list has to be edited. Selector labels are match criteria, so the platform's attribution labels never go into them; they go only on the policy's own `labels`.

## Writing the rules

A rule is a rollout restriction. `actions` narrows what is blocked (`CREATE` stops new rollouts and promotions; add `APPROVE`, `ADVANCE`, `ROLLBACK`, or the job actions as needed), and `invokers` narrows who (`USER`, `DEPLOY_AUTOMATION`); an empty list blocks everything. Think twice before blocking `ROLLBACK` -- a freeze that stops you from undoing a bad release is worse than no freeze. Every restriction needs time windows in one IANA time zone: weekly windows recur on `daysOfWeek` (empty means every day) between `startTime` and `endTime` (both or neither; neither means the whole day, and `24:00` is the end of a day), and one-time windows run from a start date and time to an end date and time.

## Living with it

A blocked action fails with a policy violation, so pick policy and restriction IDs people will recognize. The caller can override with the policy's `deploy_policy_id` when they hold `clouddeploy.deployPolicies.override`; grant that to a small on-call group, not to the deploy automation. To lift a freeze early, set `suspended: true` and keep the definition for next time. Everything but `location` and `deployPolicyId` updates in place; destroy deletes the policy and does nothing to the pipelines and targets it governed.
