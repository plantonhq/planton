# A DigitalOcean database cluster is the room its users, databases, pools, and topics live in

## What changed

- **`DigitalOceanDatabaseCluster` is a container kind.** Six kinds are created on a managed cluster and API-addressed under its id -- `DigitalOceanDatabaseUser`, `DigitalOceanDatabaseDb`, `DigitalOceanDatabaseConnectionPool`, `DigitalOceanDatabaseFirewall`, `DigitalOceanDatabaseKafkaTopic`, and `DigitalOceanDatabaseKafkaSchema` -- and none can exist before it. The kind metadata now says what their own specs already said ("an additional user ON a managed database cluster"; "a topic ON a DigitalOcean managed Kafka cluster"): on a diagram the cluster is the room they stand in.
- **Three references into a cluster are containment-exempt, travelling with the mark.** `DigitalOceanDatabaseReplicaSpec.cluster` names the primary a read replica FOLLOWS -- a replica is a single-node cluster of its own, in the primary's region or another, not something created inside it. `DigitalOceanMonitorAlertSpec.database_cluster_ids` names the clusters an alert policy WATCHES. `DigitalOceanAppDatabase.cluster_name` names an existing cluster an App Platform app ATTACHES as a dependency it connects to. Without these three lines the mark alone would have drawn a replica inside its primary, an alert inside the database it watches, and an app inside the database it connects to.
- The containment-decision registry (`shared/cloudresourcekind/testdata/containment_decisions.txt`) gains exactly nine lines: the six children `contained` in the cluster, the three references `exempt`; nothing else in the registry moved.

## Why

`container_kind` says a kind is a box other resources nest inside, and `containment_exempt` says a reference into such a box is access, not placement. The two are decided together because marking a box changes the meaning of every reference into it: the moment the cluster became a room, every `StringValueOrRef` typed to it became placement by omission, and the registry showed the three that are not placement at all. A DigitalOcean operator who runs a PostgreSQL cluster with a read replica, an app that attaches it, and an alert on its CPU should see the users, the pool, the databases, and the firewall inside the cluster -- and the replica, the app, and the alert outside it, each with a line in.

## How to check

```bash
go test ./shared/cloudresourcekind/...   # green; the golden carries the six contained and three exempt lines
grep -n "container_kind: true" -B12 shared/cloudresourcekind/cloud_resource_kind.proto | grep -A12 "DigitalOceanDatabaseCluster ="
grep -n containment_exempt catalog/digitalocean/digitaloceandatabasereplica/v1alpha1/spec.proto catalog/digitalocean/digitaloceanmonitoralert/v1alpha1/spec.proto catalog/digitalocean/app_spec.proto
```
