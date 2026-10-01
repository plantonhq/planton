# Load Balancer 5xx Error-Rate Alert

This preset pages when a load balancer's 5xx error rate climbs -- the user-facing "the site is broken" signal, wired to the balancer by reference and delivered to both email and a Slack incidents channel.

## When to Use

- Any production load balancer: 5xx rate is the closest built-in metric to user pain
- Pairing with a droplet CPU policy so cause (hot backends) and effect (errors) page together

## Key Configuration Choices

- **Balancer by reference** (`valueFrom`) -- wires to a DigitalOceanLoadBalancer in the same chart or environment; swap in a literal UUID for an existing balancer.
- **`value: 5` percent over `5m`** -- catches real degradation without paging on a single bad request; tune to traffic volume.
- **Slack webhook URL is a secret** -- the field is sensitive, so the platform accepts only a managed-secret reference: store your real webhook as a managed secret and point `$secret/slack-incidents-webhook` at it (rename to taste). A literal URL is rejected at create.
- **Email recipients must be verified team members** -- DigitalOcean rejects any other address at create time; invite the on-call inbox to the team first.

## What You Get

One policy on the referenced balancer, its `alert_id` exported for reference.
