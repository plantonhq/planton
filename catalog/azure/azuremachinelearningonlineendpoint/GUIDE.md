# Azure Machine Learning Online Endpoint -- Operational Guide

Judgment that saves real time when running online endpoints. The field reference lives in the API Explorer; this is the operational layer above it.

## The endpoint is the contract; deployments are the churn

Treat the endpoint as the long-lived, boring object: its name is in your applications' configuration and its auth mode is a promise to callers. Everything volatile -- model versions, instance sizes, probes -- lives in deployments that come and go behind it. If you find yourself replacing endpoints routinely, the volatile thing is modeled at the wrong layer.

## Names are region-wide real estate

The endpoint name is reserved per subscription across the WHOLE region -- not per workspace -- because it becomes part of the scoring DNS. A deleted endpoint releases its name only after deletion completes. Pick application-scoped names (`fraud-scoring`, not `endpoint1`), and never build automation that recreates endpoints under the same name in a tight loop; the reservation lag will bite.

## Choose the auth mode like a security posture, not a default

`Key` is the pragmatic start: static keys that never expire -- which is exactly their weakness; treat them like any other long-lived credential and rotate deliberately. `AMLToken` expires and refreshes but callers must handle refresh. `AADToken` is the keyless posture worth defending for internal services: no secret to store at all, at the cost of Entra plumbing in every caller. Changing the mode updates in place, but every caller breaks until reconfigured -- coordinate it like an API version change.

## Give the identity its grants BEFORE the first deployment

The endpoint's identity does the pulling: container images from the registry, model artifacts from storage. A system-assigned identity exists only after the endpoint does, so its grants (AcrPull on the ACR, Storage Blob Data Reader on the model store) can land only between endpoint creation and the first deployment -- exactly the window automation forgets. In charts, prefer a user-assigned identity created and granted FIRST, then referenced by the endpoint; deployments then work on the first try.

## Move traffic in steps, and mirror before you shift

The traffic map updates in place and is the whole point of the endpoint layer. The rollout that works: deploy green at `0`, mirror 10-20% to it (`mirrorTraffic`) and watch its logs and latency, then walk traffic `90/10 → 50/50 → 0/100`, then delete blue. Mirrored traffic doubles the load for its share -- account for it in instance counts. A traffic map summing below 100 silently drops the remainder -- ARM accepts it; your callers see errors.

## Keys: bring your own or fetch, never persist

ARM never returns key values on any read -- by design, and this kind honors it: keys are not outputs. Either bring your own keys from a secret store (`initialAuthKeys`, Key mode) so rotation is your secret manager's job, or let the service mint them and read them at deploy time with `az ml online-endpoint get-credentials`. Anything that copies keys into files or outputs is building the leak.

## Deleting an endpoint deletes its deployments

The endpoint is the ARM parent: deleting it takes every deployment (and their running instances) with it, and in-flight requests fail immediately. Drain first -- move traffic away, watch the request count fall, then delete. The same applies in reverse: you cannot delete the last deployment while the traffic map still routes to it; zero the map first.
