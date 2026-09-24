# GcpManagedKafkaConnectCluster Guide

The judgment this guide protects: a Connect cluster is shared compute for many teams' pipelines. Size it for the connectors it will run, give it network paths to everything they reach, and let each team own its connectors as separate blocks.

## Attachment

`kafkaCluster` names the Managed Kafka cluster the workers belong to -- where Kafka Connect keeps its internal config, offset, and status topics, and the default cluster connectors read and write. A reference to a `GcpManagedKafkaCluster` yields its full path; a literal must be that path. Attaching needs `managedkafka.clusters.attachConnectCluster` on the Kafka cluster.

## Capacity

`vcpuCount` (at least 3) and `memoryBytes` (at least 3 GiB, 1-8 GiB per vCPU) size the workers every connector on the cluster shares, billed around the clock. `tasks.max` in each connector's configuration spreads its work across them; scale the cluster when connectors' tasks queue up.

## Networks and DNS

Each `networkConfigs` entry gives the workers a Private Service Connect interface in a private subnet of the Connect cluster's region, and through it they reach the Kafka cluster and any endpoint in that network. `dnsDomainNames` makes other private DNS domains resolvable -- MirrorMaker 2 replicating from a second Managed Kafka cluster needs that cluster's domain (its bootstrap address without the leading `bootstrap.` label and the port).

## Identity

Connectors act as the Managed Kafka service agent. A Pub/Sub sink needs it to hold `roles/pubsub.publisher` on the topic, a BigQuery sink `roles/bigquery.dataEditor` on the dataset -- grants made outside this block, before the connector runs.
