---
title: "Presets"
description: "Ready-to-deploy configuration presets for Cosmos DB Account"
type: "preset-list"
componentSlug: "cosmos-db-account"
componentTitle: "Cosmos DB Account"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-sql-api"
    rank: "01"
    title: "Cosmos DB with SQL API"
    excerpt: "This preset creates an Azure Cosmos DB account with the SQL (NoSQL) API, Session consistency, a single geo-location, and one SQL database containing a container with a partition key. The SQL API is..."
  - slug: "02-mongodb-api"
    rank: "02"
    title: "Cosmos DB with MongoDB API"
    excerpt: "This preset creates an Azure Cosmos DB account with the MongoDB wire-protocol API, MongoDB server version 6.0, Session consistency, and one MongoDB database containing a sharded collection...."
  - slug: "03-serverless"
    rank: "03"
    title: "Cosmos DB Serverless (SQL API)"
    excerpt: "This preset creates an Azure Cosmos DB account in serverless mode with the SQL (NoSQL) API. Serverless mode uses pay-per-request pricing with no provisioned throughput -- you only pay for the Request..."
---

# Cosmos DB Account Presets

Ready-to-deploy configuration presets for Cosmos DB Account. Each preset is a complete manifest you can copy, customize, and deploy.
