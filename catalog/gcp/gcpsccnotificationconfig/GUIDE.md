# GcpSccNotificationConfig Guide

The judgment this guide protects: a notification config that cannot publish fails silently. Activate Security Command Center, grant the publisher, and filter to what responders act on, or the stream is either empty or noise.

## The three prerequisites

1. **Activation.** Security Command Center must be active on the scope: the organization, or the project on its own (Standard tier is free at project level when the organization has not activated it).
2. **The topic.** Folder and organization configs require one; Google lets a project config omit it, which records the config but sends nothing.
3. **The publisher.** The config's `service_account` output is the identity Security Command Center publishes as. Grant it `roles/pubsub.publisher` on the topic. Creating the config does not check this, and without it every notification is dropped.

## Writing the filter

Restrictions `<field> <operator> <value>` combined with `AND` and `OR` (`OR` binds tighter), negated with a leading `-`. Common shapes: `state = "ACTIVE" AND severity = "HIGH"`, `category = "OPEN_FIREWALL"`, `state = "ACTIVE" AND NOT mute = "MUTED"` (respect mute rules). Operators: `=` for every type, `>`, `<`, `>=`, `<=` for integers, `:` for substring match. An empty filter streams every create and update.

## Scope

A folder config covers every project beneath it; an organization config covers everything. One organization config to the SOC's topic plus project configs for the teams that own them is a common split. `configId` must be unique within the parent.
