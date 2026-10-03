# DigitalOcean Database Db -- Operational Guide

What experience with this kind teaches that the field reference cannot.

## The name is a contract, not a label

`database_name` is how every client, pool, and user grant addresses this database. Both spec fields are create-only: editing either one REPLACES the logical database, which means the old one is DROPPED -- data and all -- and an empty successor appears. Choose names like API endpoints, and treat any rename as a data migration you plan (dump, create new, restore, cut over), never as a manifest edit.

## One database per workload

Logical databases are free and instant. Give each application, service, or experiment its own instead of piling tables into `defaultdb` -- deletion blast radius then matches workload boundaries exactly.

## Deletion is immediate and total

Deleting this resource drops the database and its data with no recycle bin of its own. The cluster's automatic backups (a DigitalOceanDatabaseCluster property) are the only recovery path -- restoring means provisioning a new cluster from a backup, not un-dropping the database.

## Composition order

Pools reference this database by name (`db_name`), and grants ride users. In a chart: cluster, then database, then user, then pool -- reference wiring enforces the cluster edge automatically; the database-before-pool edge is yours to keep (the pool's `db_name` is a plain string by design, composable with either this kind or `defaultdb`).

## What is deliberately NOT here

Per-database sizing or placement (a logical database is a namespace, not hardware), credentials (see DigitalOceanDatabaseUser), and in-database schema (data-plane migrations, not IaC).
