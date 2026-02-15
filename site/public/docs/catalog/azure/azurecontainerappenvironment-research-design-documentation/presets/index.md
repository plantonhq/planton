---
title: "Presets"
description: "Ready-to-deploy configuration presets for AzureContainerAppEnvironment: Research & Design Documentation"
type: "preset-list"
componentSlug: "azurecontainerappenvironment-research-design-documentation"
componentTitle: "AzureContainerAppEnvironment: Research & Design Documentation"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-consumption"
    rank: "01"
    title: "Consumption Plan Environment"
    excerpt: "This preset creates a minimal Azure Container App Environment using the default Consumption (serverless) plan with no VNet injection. Apps deployed to this environment share Azure-managed networking..."
  - slug: "02-workload-profiles-vnet"
    rank: "02"
    title: "Workload Profiles with VNet Integration"
    excerpt: "This preset creates a production-grade Azure Container App Environment with VNet injection, internal load balancer (no public internet exposure), zone redundancy, and a D4 dedicated workload profile...."
---

# AzureContainerAppEnvironment: Research & Design Documentation Presets

Ready-to-deploy configuration presets for AzureContainerAppEnvironment: Research & Design Documentation. Each preset is a complete manifest you can copy, customize, and deploy.
