# Autoscale Only

## Use Case

Capacity-based pricing without a standing bill: a Standard reservation with no baseline that autoscales to 100 slots while queries run, routing one project's queries onto it.

## When to Use

- Moving a project off on-demand bytes pricing
- Spiky or unpredictable query load

## What This Creates

- A zero-baseline Standard reservation in `US` autoscaling to 100 slots, with `analytics-project`'s QUERY jobs assigned

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `autoscaleMaxSlots` | `100` | The ceiling on what a burst can cost. |
| `edition` | `STANDARD` | `ENTERPRISE` for commitments and more features. Fixed at creation. |
