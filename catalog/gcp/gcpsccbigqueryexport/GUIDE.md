# GcpSccBigQueryExport Guide

The judgment this guide protects: an export without its writer grant writes nothing, and an export without a filter fills the dataset with every finding ever updated. Grant the principal, filter deliberately, and plan the dataset's retention.

## Prerequisites

1. **Activation.** Security Command Center must be active on the scope.
2. **The dataset.** Create it first (`GcpBigQueryDataset`) in the location your analysts query.
3. **The writer.** The export's `principal` output is the identity Security Command Center writes as. Grant it `roles/bigquery.dataEditor` on the dataset through a dataset access entry. Without it no rows arrive.

## What lands in the dataset

Security Command Center creates and manages a findings table and appends a row for every new or updated finding that matches the filter, within minutes. `state = "ACTIVE" AND NOT mute = "MUTED"` keeps resolved and muted findings out; an empty filter exports every create and update. Use the dataset's default table and partition expiration to bound storage.

## Lifecycle

Destroy deletes the export; the dataset and the rows already written stay. Google's API may report the dataset "still in use" for a minute after an export is deleted, so a dataset destroyed in the same run can need a retry. The organization export's resource name is composed by both modules from the organization, location, and ID, because Google's organization resource takes the name as an argument.
