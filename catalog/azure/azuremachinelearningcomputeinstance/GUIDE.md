# Azure Machine Learning Compute Instance -- Operational Guide

Judgment that saves real time when running compute instances. The field reference lives in the API Explorer; this is the operational layer above it.

## An instance is a workstation, not a training rig

Size it for notebooks and debugging (general-purpose DS-series) and push anything heavy to a compute cluster -- the cluster scales to zero between jobs, the instance never does. The strongest cost lever an ML platform team has is keeping this distinction: personal instances small, shared clusters doing the real work.

## It bills around the clock until someone stops it

A forgotten instance is the classic ML cost leak: a VM billing 24/7 for eight hours of weekday use. The resource itself has no schedule surface here -- build the habit (or the automation) of stopping instances outside working hours (`az ml compute stop`), and delete instances whose owners moved on. A stopped instance keeps its disk and configuration; only compute billing pauses.

## Nothing updates in place -- treat the instance as disposable

The provider has NO update path: resizing, re-tagging, changing SSH -- everything replaces the instance, and its OS disk (local files, installed packages, cloned repos) goes with it. The discipline that makes this painless: keep data in datastores, code in git, environments in the registry. An instance that can be recreated in ten minutes without loss is configured correctly; one that cannot is holding state it should not.

## Names are reserved region-wide -- pick personal, specific names

Instance names are unique per Azure REGION per subscription, not per workspace -- `dev-instance` will collide with a colleague's in another workspace. Name instances after their owner (`alice-dev`). A just-deleted name can also stay reserved briefly; if a recreate fails on the name, wait a few minutes before suspecting anything else. Live-proven: a `${E2E_RUN_ID}` token in the cloud-side name expands on both engines (shape `ci-<8-hex>-<p|t>`, 13 characters) and the verifier reads the expanded ARM id -- use a unique suffix whenever two engines or two reruns share a subscription.

## Provision FOR people, not AS people

The platform-team pattern is `authorizationType: personal` + `assignToUser` with the colleague's tenant and object IDs -- the instance is theirs from first boot, without sharing the deploying credentials. Without `assignToUser`, whoever's credentials ran the deploy owns the instance -- rarely what a platform team wants.

## The subnet rule depends on the workspace -- know your isolation mode

With `nodePublicIpEnabled: false`, the provider demands a `subnetId` UNLESS the workspace runs a managed network (isolation mode AllowInternetOutbound / AllowOnlyApprovedOutbound) -- the check happens at apply time against the live workspace, not at manifest time. Corollary: on a managed-network workspace, do NOT set a subnet (Azure networks the instance itself and the provider will not read the subnet back); on a plain workspace, private instances need the subnet explicitly.

## SSH stays off unless someone actually needs a terminal

Absent `ssh` block = port closed (the default worth keeping; the studio and VS Code remote cover most needs). When enabled, the SERVICE assigns the username and port -- read `ssh_username` / `ssh_port` from the outputs rather than assuming 22/azureuser. Live-proven without an ssh block: both engines emit empty username and port 0 (Terraform via `try(..., "")` / `try(..., 0)`; Pulumi must `Elem()` the optional SSH fields -- a raw nil export fails the output int32 parse). The identity principal output populates on both engines when `identity.type` is system-assigned.
