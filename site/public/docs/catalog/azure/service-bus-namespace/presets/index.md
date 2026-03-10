---
title: "Presets"
description: "Ready-to-deploy configuration presets for Service Bus Namespace"
type: "preset-list"
componentSlug: "service-bus-namespace"
componentTitle: "Service Bus Namespace"
provider: "azure"
icon: "package"
order: 200
presets:
  - slug: "01-standard-messaging"
    rank: "01"
    title: "Standard Messaging Service Bus"
    excerpt: "This preset creates an Azure Service Bus namespace on the Standard tier with a single queue and topic — the fastest path to reliable messaging for most applications. The Standard tier provides shared..."
  - slug: "02-premium-enterprise"
    rank: "02"
    title: "Premium Enterprise Service Bus"
    excerpt: "This preset creates an Azure Service Bus namespace on the Premium tier with zone redundancy, private networking, and advanced messaging features. Premium tier provides dedicated resources (1..."
  - slug: "03-event-driven-microservices"
    rank: "03"
    title: "Event-Driven Microservices Service Bus"
    excerpt: "This preset creates an Azure Service Bus namespace configured for an event-driven microservices architecture with queue chaining and dead-letter forwarding. The Standard tier keeps costs low..."
---

# Service Bus Namespace Presets

Ready-to-deploy configuration presets for Service Bus Namespace. Each preset is a complete manifest you can copy, customize, and deploy.
