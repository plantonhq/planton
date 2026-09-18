# DigitalOcean Droplet Autoscale Pool -- Operational Guide

What experience with this component teaches that the field reference cannot.

## Destroy destroys the droplets -- there is no other delete

DigitalOcean's only delete for an autoscale pool is the "dangerous" variant: the pool AND every member droplet it owns are terminated together. There is no adopt-the-members teardown, no orphan mode. Before destroying, drain traffic (load balancers, DNS) as if you were terminating that many droplets by hand -- because you are. The inverse hazard also holds: a FORGOTTEN pool keeps real droplets billing indefinitely, so pool hygiene is bill hygiene.

## The first destroy reports an error DigitalOcean has already accepted

At the current provider (v2.100.1, Pulumi bridge v4.79.1) every destroy of a pool with members fails 5-6 seconds in with `Error waiting for Droplet autoscale pool (...) to become be deleted: unexpected state 'deleting', wanted target 'Not Found'`. The provider's delete waiter only expects the pool to answer "OK" and then vanish, but DigitalOcean reports the pool `deleting` while it terminates the members, and the provider treats that word as a failure. Nothing is actually wrong: DigitalOcean completes the deletion (the pool and every member 404 within about ten seconds), and the engine is simply left believing a resource exists that does not. Recovery differs by engine. Terraform: run destroy again -- the refresh reads the 404 and drops the pool from state. Pulumi: run `pulumi refresh` first, then destroy -- a second destroy without the refresh calls delete on a pool that is gone, DigitalOcean answers 404, and the provider fails on that too. Until the upstream waiter accepts `deleting` as a pending state (a one-line fix), treat this error as "wait ten seconds, refresh, destroy again", never as a stuck pool. Verify with the control panel or `GET /v2/droplets/autoscale/{pool_id}` if in doubt; a lingering pool would keep real droplets billing.

## Members always join a VPC -- unset means the region's default, sent explicitly

Leaving `dropletTemplate.vpc` unset places members in the region's default VPC, exactly as DigitalOcean would. Both modules look that VPC up and send its UUID rather than omitting the field: DigitalOcean reports the UUID back on every read, and the provider's field is optional but not computed, so an omitted value beside a populated read-back would re-plan on every apply. The manifest stays simple; the cloud and the plan agree.

Members get a public IPv4 by default. `dropletTemplate.publicNetworking: false` creates them with no public interface at all -- reachable only inside the VPC, which is the right shape for a fleet behind a DigitalOceanLoadBalancer. It is a template setting, so changing it rolls the whole fleet. Either way, restrict reachability with a DigitalOceanFirewall targeting the pool's tags.

## Members are cattle -- keep state out of them

The pool creates and destroys members on its own schedule (scale events, health replacement, template rollouts). Anything written to a member's local disk is one scale-in from gone. Point members at managed databases, Spaces, or volumes owned elsewhere; use `userData` to bootstrap them identically on every boot.

## Dynamic scaling decides on agent metrics -- keep the agent on

CPU and memory targets are evaluated from the droplet monitoring agent's telemetry. Ship dynamic pools with `withDropletAgent: true` (the quick-start default here); without it, memory-based scaling has no data source at all. Static pools can skip the agent, though replacement health still benefits from it.

## Target the pool with tags, never with droplet IDs

Member droplet IDs churn with every scale event -- any firewall rule or load-balancer target list naming them goes stale immediately. The template's `tags` (plus the Planton labels both engines always apply) follow the membership automatically; tag-targeted firewall rules and load-balancer tag targets are the ONLY reliable way to address the fleet.

## Template changes roll the fleet

Editing the template (size, image, user data) applies in place on the pool, and DigitalOcean replaces members to converge on the new shape. Plan template edits like deployments: capacity dips while members roll. The image also has a read-back quirk -- DigitalOcean reports it as a numeric image ID even when you configured a slug; the modules keep your configured value, but a freshly IMPORTED pool will show an image update (numeric id back to your slug) on its first plan. Applying it is safe: re-sending the identical template does not roll the members (measured live -- no member replaced, no history event, `updated_at` unchanged).

## Size the bounds for the bill, not just the load

A static pool bills `target_instances` droplets around the clock; a dynamic pool bills between `minInstances` and `maxInstances`. The max bound is your cost ceiling under sustained load (or under a runaway feedback loop) -- set it from budget, not optimism. Scale-ups respect `cooldownMinutes`, so short traffic spikes may ride on the existing members.

## What is deliberately NOT here

per-member identity knobs (names are generated from the pool name -- members are interchangeable by design); and `current_utilization` / timestamps as outputs (volatile metrics, not identity -- read them from DigitalOcean monitoring).
