---
title: "Preset: Basic Data Catalog"
description: "A minimal Glue Data Catalog database with a description. Tables are created without a default storage location — each table specifies its own S3 path."
type: "preset"
rank: "01"
presetSlug: "01-basic-data-catalog"
componentSlug: "glue-catalog-database"
componentTitle: "Glue Catalog Database"
provider: "aws"
icon: "package"
order: 1
---

# Preset: Basic Data Catalog

A minimal Glue Data Catalog database with a description. Tables are created
without a default storage location — each table specifies its own S3 path.

## What This Configures

- A named database in the Glue Data Catalog.
- A description documenting the database's purpose.
- No default storage location — tables define their own paths.

## When to Use

- Quick-start setup for Athena analytics or Glue ETL exploration.
- Environments where tables point to diverse S3 locations (no shared prefix).
- Development or experimentation databases.

## Customization Points

- Add `locationUri` to set a default S3 storage location for tables.
- Change `description` to document the specific use case.
