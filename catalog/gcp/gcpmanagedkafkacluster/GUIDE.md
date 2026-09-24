# GcpManagedKafkaCluster Guide

The judgment this guide protects: a Kafka cluster is shared infrastructure that application teams build on, so its shape should change rarely and its contents -- topics, access rules, connectors -- should change often, in other people's manifests. Size it, attach it to the networks its clients live in, and leave the rest to the blocks that reference it.

## Sizing

Capacity is two numbers for the whole cluster: `vcpuCount` (at least 3) and `memoryBytes` (1-8 GiB per vCPU, written in bytes -- 3 GiB is 3221225472). Google spreads them across brokers in three zones and bills both around the clock whether or not messages flow. Both scale in place; after a scale-up, `rebalanceMode: AUTO_REBALANCE_ON_SCALE_UP` moves partitions onto the new brokers, where the default leaves existing partitions where they are and only new ones land on the new brokers. `brokerDiskSizeGib` (at least 100) sets each broker's local disk.

## Networks and addresses

Each `networkConfigs` entry attaches the cluster to one VPC network through one subnet in the cluster's region -- up to ten networks, and the subnet's project may differ, so a Shared VPC host's subnet works. Google creates the bootstrap and broker addresses and their private DNS names in every attached network, so a client anywhere in those networks connects by name. The bootstrap address is fixed for the cluster's life, but its format can differ between clusters and the pinned Pulumi SDK cannot read it yet, so it is not an output: read it once with `gcloud managed-kafka clusters describe` and hand it to clients as configuration. Internet access (Google's public cluster option) waits for pulumi-gcp v10 for the same reason.

## Authentication and authorization

Every client authenticates with Google IAM over SASL (port 9092): the principal needs `roles/managedkafka.client`, and Kafka ACLs (`GcpManagedKafkaAcl`) decide what it may do on each topic and consumer group. `tlsConfig` adds mutual TLS (port 9192) with client certificates issued by the Certificate Authority Service pools in `caPools`; `sslPrincipalMappingRules` turns a certificate's distinguished name into the short principal your ACLs name, and changing it restarts the brokers one by one. An empty `tlsConfig` is how you turn mTLS back off -- leaving the field out keeps whatever Google has.

## Encryption and destroy

`kmsKey` encrypts the cluster with your own key; it must be in the cluster's region, the Managed Kafka service agent needs encrypt/decrypt on it, and it can never change. `deletionPolicy` defaults to `DELETE`, which removes every topic and message -- `PREVENT` is the right choice for a production cluster.
