---
title: "Presets"
description: "Ready-to-deploy configuration presets for Firestore Database"
type: "preset-list"
componentSlug: "firestore-database"
componentTitle: "Firestore Database"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-default-native"
    rank: "01"
    title: "Default Firestore Native Database"
    excerpt: "This preset creates the project's default Firestore Native database in the US multi-region (nam5) with delete protection enabled. This is the primary database that client libraries connect to when no..."
  - slug: "02-named-native-pitr"
    rank: "02"
    title: "Named Firestore Native Database with PITR"
    excerpt: "This preset creates a named Firestore Native database with point-in-time recovery enabled (7-day version retention) and delete protection. Suitable for production workloads that need disaster..."
  - slug: "03-enterprise-cmek"
    rank: "03"
    title: "Enterprise Firestore Database with CMEK"
    excerpt: "This preset creates an Enterprise-edition Firestore Native database with customer-managed encryption (CMEK), point-in-time recovery, and delete protection. Designed for regulated and enterprise..."
---

# Firestore Database Presets

Ready-to-deploy configuration presets for Firestore Database. Each preset is a complete manifest you can copy, customize, and deploy.
