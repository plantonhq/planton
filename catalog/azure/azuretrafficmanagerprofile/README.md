# Overview

The **AzureTrafficManagerProfile** component deploys a Traffic Manager profile -- Azure's DNS-based traffic director. The profile owns a public DNS name and answers each lookup with the address of one of its endpoints, chosen by routing method (nearest, failover order, weighted spread, caller's geography or subnet, or all-at-once) and endpoint health.

## Purpose

- **Global traffic steering without a data path**: the decision happens in DNS, so Traffic Manager fronts anything with a resolvable address -- across regions, clouds, and on-premises -- and adds no hop, no throughput limit, and no TLS termination.
- **Health-driven answers**: the profile continuously probes every endpoint and only hands out healthy ones; failover is automatic at DNS speed (the TTL).
- **Routing method as a dial**: switching from Performance to Priority to Weighted re-steers traffic in place, without touching endpoints.

## Key Features

- Full azurerm v5 surface: all six routing methods, the complete monitor configuration (protocol/port/path, custom status ranges, probe headers, fast-interval probing), MultiValue's answer cap, Traffic View, profile enable/disable.
- The provider's cross-field contracts front-loaded as validation: MultiValue requires max_return; the fast probe interval narrows the timeout window.
- Chart-ready: publishes the profile id endpoints reference (AzureTrafficManagerEndpoint) and the trafficmanager.net FQDN your domain CNAMEs to.

## Use Cases

- **Multi-region failover**: a Priority profile over per-region deployments -- traffic holds on the primary and fails to standbys by health.
- **Latency routing**: a Performance profile sending each user to the nearest healthy region.
- **Gradual cutovers and A/B**: a Weighted profile shifting traffic percentages between old and new stacks by editing weights.

## Future Enhancements

- Endpoint objects are a separate kind (AzureTrafficManagerEndpoint) by design -- one profile serves many endpoints with independent lifecycles.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
