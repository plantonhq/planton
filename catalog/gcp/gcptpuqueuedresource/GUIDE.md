# GcpTpuQueuedResource Guide

The judgment this guide protects: a queued request turns into billing TPUs whenever Google finds capacity -- possibly hours later, possibly at night -- and its nodes belong to it. Ask only for what the job needs, and destroy the request, not just the nodes, when the work is done.

## When to queue instead of create

A `GcpTpuVm` create fails when the zone has no chips to give. A queued resource asks for the same capacity and waits: its state moves through `ACCEPTED` and `WAITING_FOR_RESOURCES` to `PROVISIONING` and `ACTIVE` when the nodes exist. A request can also end `FAILED` (a late-found problem, or capacity that never came within its constraints), and moves through `SUSPENDING` to `SUSPENDED` when its nodes are deleted. The deploy reports none of these: Google's provider exports no state for a request, so read it with `gcloud compute tpus queued-resources describe`. Use it for scarce generations and large slices, and for jobs that can start whenever capacity arrives.

## Beta-only by Google

Google publishes `google_tpu_v2_queued_resource` only in its beta Terraform provider. The Terraform module attaches `provider = google-beta` to it under the catalog's recorded admission; Pulumi serves it from its single provider.

## Describing the nodes

`nodeSpecs` lists the nodes, provisioned together. Each has a required `nodeId` (unique; the provider neither reads back an id Google generates nor changes one in place, so an unnamed node would replace the request on every deploy), and a `node` with `runtimeVersion` (required, matching the generation), `acceleratorType` (default `v2-8`), a description, and a network interface. Unset, the interface is the project's network and subnetwork named `default`, which many organizations never create -- a project without them must name a network. The modules derive every node's parent -- the request's project and zone -- because a node is always created where its request lives.

## What the pinned provider does not carry

At the pinned provider a queued request carries no spot or guaranteed options, no reservation, no labels, service account, or data disks. When you need those, create a `GcpTpuVm` directly once capacity is available.

## Lifecycle

Everything is immutable: any change replaces the request and its nodes. Google deletes a request without force only once its nodes are already gone or while it is `ACCEPTED`, `FAILED`, or `SUSPENDED`, and the provider never sends force. So destroying -- or changing -- a request whose TPUs are up (`ACTIVE`, or still `PROVISIONING` or `WAITING_FOR_RESOURCES`) fails while the TPUs keep billing. Delete it with `gcloud compute tpus queued-resources delete --force`, then run the destroy or the change again. `ABANDON` leaves the request and its nodes running -- and billing -- outside management. The nodes bill from the moment they are provisioned, not from when the request is created.
