---
title: "Presets"
description: "Ready-to-deploy configuration presets for Glue Catalog Database"
type: "preset-list"
componentSlug: "glue-catalog-database"
componentTitle: "Glue Catalog Database"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-basic-data-catalog"
    rank: "01"
    title: "Preset: Basic Data Catalog"
    excerpt: "A minimal Glue Data Catalog database with a description. Tables are created without a default storage location — each table specifies its own S3 path."
  - slug: "02-s3-data-lake"
    rank: "02"
    title: "Preset: S3 Data Lake"
    excerpt: "A Glue Data Catalog database with a default S3 storage location for an organized data lake. Tables created in this database inherit the base S3 path, keeping data organized under a consistent prefix."
---

# Glue Catalog Database Presets

Ready-to-deploy configuration presets for Glue Catalog Database. Each preset is a complete manifest you can copy, customize, and deploy.
