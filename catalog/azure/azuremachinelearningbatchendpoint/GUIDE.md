# Azure Machine Learning Batch Endpoint -- Operational Guide

Judgment that saves real time when running batch endpoints. The field reference lives in the API Explorer; this is the operational layer above it.

## The endpoint is a routing object -- treat it as free and permanent

A batch endpoint provisions no compute and bills nothing at rest: it is a stable address plus a pointer. Create it early, keep it forever, and let deployments (the recipes) churn behind it. Schedulers and pipelines should know ONLY the endpoint address -- never a deployment name -- so recipe rollouts never touch the caller side.

## Authentication is Entra-only; stop looking for keys

Every job submission presents a Microsoft Entra token for a principal (user or service principal). There are no endpoint keys: ARM's shared enum advertises `Key` and `AMLToken`, but the batch service rejects both with "AuthMode must be 'AADToken'" -- this kind's validation stops those values before they reach Azure. Plan the caller side accordingly: the submitting principal needs rights to create jobs in the workspace, and CI schedulers authenticate as service principals, not with copied secrets.

## The submitter's identity runs the job -- grant THAT, not the endpoint

The identity story inverts the online endpoint's: jobs run under the SUBMITTING principal's token, and data access happens through the compute cluster's managed identity or the workspace datastores' stored credentials. The endpoint's own identity is optional and sits outside the data path. When a job cannot read its input, look at the compute cluster's identity grants and the datastore's credential mode -- not at the endpoint.

## Set the default-deployment pointer as the LAST step of a rollout

`defaultDeploymentName` names the deployment that answers submissions which do not name one -- it is batch's whole routing dial (there is no traffic percentage map). It can reference a deployment that does not exist yet, and ARM will accept it, so the safe rollout order is: create the new deployment, invoke it BY NAME until it earns trust, then move the pointer (an in-place update). Automation that creates endpoint and pointer together, before any deployment exists, produces an endpoint whose default submissions fail -- legal, and useless.

## Reachability is the workspace's business

Unlike the online endpoint, the batch surface has no public-network toggle -- job submission reachability follows the WORKSPACE's network settings (public access, private endpoints). If batch submissions must be private, fix it at the workspace layer; nothing on this component can do it.

## Deleting an endpoint deletes its deployments

The endpoint is the ARM parent: deleting it takes every deployment recipe with it. Jobs already running on the compute cluster are their own objects -- check for in-flight jobs before deleting an endpoint a scheduler still submits to, or the next scheduled run fails with a missing endpoint rather than a graceful drain.
