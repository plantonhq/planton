---
title: "Standard VPC"
description: "This preset creates a Scaleway VPC in the Paris region with routing disabled. A VPC is a regional container that groups Private Networks. This is the simplest and most common starting configuration,..."
type: "preset"
rank: "01"
presetSlug: "01-standard"
componentSlug: "vpc"
componentTitle: "VPC"
provider: "scaleway"
icon: "package"
order: 1
---

# Standard VPC

This preset creates a Scaleway VPC in the Paris region with routing disabled. A VPC is a regional container that groups Private Networks. This is the simplest and most common starting configuration, suitable for single-tier architectures or environments where cross-network communication is not needed.

## When to Use

- Single-application environments with one Private Network
- Development and testing environments
- Getting started with Scaleway networking before planning multi-tier architectures

## Key Configuration Choices

- **Paris region** (`region: fr-par`) -- Scaleway's primary region with the broadest service availability
- **Routing disabled** (`enableRouting: false`) -- Private Networks in this VPC are isolated from each other; enable routing only when cross-network communication is needed (note: once enabled, routing cannot be disabled)
- **Custom routes propagation disabled** (`enableCustomRoutesPropagation: false`) -- no VPN or network appliance routes are advertised between Private Networks

## Placeholders to Replace

No placeholders -- this preset is ready to deploy as-is. Change `region` to `nl-ams` or `pl-waw` for Amsterdam or Warsaw.

## Related Presets

- **02-routing-enabled** -- Use instead when Private Networks in the VPC need to communicate with each other (e.g., multi-tier architectures, Kapsule cluster talking to RDB in a separate Private Network)
