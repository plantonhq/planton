---
title: "Presets"
description: "Ready-to-deploy configuration presets for DynamoDB"
type: "preset-list"
componentSlug: "dynamodb"
componentTitle: "DynamoDB"
provider: "aws"
icon: "package"
order: 200
presets:
  - slug: "01-on-demand-simple"
    rank: "01"
    title: "On-Demand Simple Table"
    excerpt: "This preset creates a DynamoDB table with on-demand (pay-per-request) billing and a simple partition key. On-demand pricing automatically scales to handle any traffic level without capacity planning...."
  - slug: "02-provisioned-production"
    rank: "02"
    title: "Provisioned Production Table"
    excerpt: "This preset creates a DynamoDB table with provisioned capacity, a composite primary key (partition + sort key), and server-side encryption with a customer-managed key. Provisioned mode is more..."
---

# DynamoDB Presets

Ready-to-deploy configuration presets for DynamoDB. Each preset is a complete manifest you can copy, customize, and deploy.
