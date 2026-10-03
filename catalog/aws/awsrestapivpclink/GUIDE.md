# AwsRestApiVpcLink — Kind Guide

Authored operational judgment for the REST API VPC-link component: the
design decisions behind the spec's shape, and what to know before
fronting private NLBs with REST APIs.

## Design decisions

- **A link is not an API field.** One NLB attachment is shared by many
  APIs and outlives any of them — the same reason AwsHttpApiVpcLink
  is its own kind.
- **Exactly one NLB, immutable.** AWS accepts one balancer per link
  and has no update for it. A different NLB is a new link; integrations
  re-home onto the new ID.
- **Not interchangeable with the HTTP API link.** v1 links front an
  NLB; v2 links attach to subnets (ALB, NLB, or Cloud Map). Pointing a
  REST integration at an AwsHttpApiVpcLink fails at apply.

## Fronting private backends in production

- **Internal NLB is the usual shape.** A public NLB works but defeats
  the point of the link.
- **Create is free; wait is not.** Provisioning takes several minutes
  while AWS builds the attachment — size E2E timeouts accordingly
  (this kind's profile is 25 minutes).
- **One link per NLB, many APIs.** Do not create a link per API; share
  it through `vpc_link_id`.
- **Destroy last.** Integrations still referencing the link fail when
  it disappears; drain the APIs first.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
