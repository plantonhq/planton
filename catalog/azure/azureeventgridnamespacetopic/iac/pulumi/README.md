# AzureEventgridNamespaceTopic Pulumi Module

## Overview

Creates one named CloudEvents stream inside an Azure Event Grid namespace. Many topics share one namespace with independent lifecycles -- like consumer groups on an Event Hub.

## Resources Created

- `eventgrid.NamespaceTopic` -- the stream (namespace binding, name, delivery retention)

## Outputs

- `namespace_topic_id` -- the topic's ARM resource ID (`{namespace_id}/topics/{name}`)
- `namespace_topic_name` -- the topic's name

## Behavior Notes

- **Azure pins the schema and publisher type**: CloudEvents v1.0 and "Custom" -- the provider sends both; neither is configurable.
- **Retention is the only updatable property** (1-7 days, platform default 7 always sent); name and namespace changes replace the topic -- and drop its buffered events.
- **No tags**: the provider exposes none on this resource.
- **Cost**: the topic itself adds nothing; throughput bills on the namespace's capacity.

## Usage

The module is executed by the Planton platform with an IaC input containing the target `AzureEventgridNamespaceTopic` resource and an Azure provider configuration. For a manifest example, see `../../e2e/manifest.yaml`.
