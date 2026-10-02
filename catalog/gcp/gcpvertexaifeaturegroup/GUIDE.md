# GcpVertexAiFeatureGroup Guide

The judgment this guide protects: a feature group is a contract over a BigQuery table, not a copy of it. Get the table's shape right, register the columns models actually use, and let online stores decide what to serve.

## Feature Store reads BigQuery in place

Vertex AI Feature Store (the current generation, built on feature groups and online stores) keeps feature data in BigQuery. A feature group records which table holds which entities' features; registering a feature records which column is that feature. Nothing moves until an online store's feature view syncs the features it selects. The older Feature Store (featurestores and entity types) is superseded and not part of the catalog.

## The table's shape is Google's contract

The source table or view needs at least one entity ID column and a `TIMESTAMP` column named `feature_timestamp`: each row is one entity's feature values at one moment, and the latest row per entity is what an online store serves. Name the entity ID columns in `entityIdColumns` (one column, or several that together form the key); leave it empty when the column is called `entity_id`. Jobs under the group read the table as the Vertex AI Service Agent, which needs `roles/bigquery.dataViewer` on it.

The source is a `GcpBigQueryTable` reference (its `project.dataset.table` name) or a literal; Google stores it as `bq://project.dataset.table`, and both modules add the prefix when it is missing. The source is fixed at creation.

## Features name columns

A feature's id is the column it reads unless `versionColumnName` points elsewhere -- useful when the column carries a unit or a version suffix (`ltv_usd`) and the feature should not. Register only the columns a model uses; an online store can serve only registered features, by feature id.

## IDs use underscores

Group and feature ids allow lowercase letters, digits, and underscores, up to 128 characters, not starting with a digit. That rules out the platform's kebab-case resource names, so the group id is always set explicitly. The provider does not mark the group id force-new; treat it as permanent and replace the block to change it.

## Destroy

`deletionPolicy` fans to every feature. `DELETE` removes the features and the group -- the BigQuery table is never touched -- `PREVENT` makes destroy fail, and `ABANDON` leaves them in place. An online store whose views reference the group should be changed first; Google refuses to delete a group a view still serves.
