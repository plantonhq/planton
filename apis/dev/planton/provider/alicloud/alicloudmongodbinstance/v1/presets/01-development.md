# Preset: Development MongoDB Instance

A minimal MongoDB 7.0 replica-set instance for development and testing.

## Use Case

- Local development and integration testing
- Non-production workloads with minimal cost
- Quick prototyping with MongoDB

## Configuration

- **Engine**: MongoDB 7.0 with WiredTiger storage engine
- **Instance Class**: `dds.mongo.mid` (entry-level)
- **Storage**: 20 GB
- **Replication**: 3-node replica set (primary + secondary + hidden)
- **Billing**: PostPaid (pay-as-you-go, default)

## What's Not Included

- Multi-zone HA (single AZ deployment)
- Encryption (TDE/SSL)
- Backup configuration
- Read replicas
- Release protection
