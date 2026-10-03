# AWS Architecture Judgment

Generic infrastructure vocabulary is not enough on AWS: the same requirement
has several viable service combinations, and picking well — secure and
cost-efficient for THIS user's motive — is the composer's job. This reference
is judgment, not a service catalog. Ground every kind's exact fields with
`planton explain` as usual.

## Choosing the compute shape

Match the requirement to the cheapest shape that honestly serves it — do not
default to the heaviest thing the user mentioned:

| Requirement smells like | Reach for | Not | Why |
|---|---|---|---|
| A containerized web service or API, no k8s requirement | ECS on Fargate | EKS | No control-plane fee, no nodes to manage; EKS bills an always-on control-plane charge before one pod runs (the figure is in its cost fact-sheet) |
| Team already lives in Kubernetes, or needs k8s-native tooling (Istio, operators, Helm ecosystems) | EKS | ECS | The ecosystem is the requirement |
| Event-driven, bursty, or request-scale-to-zero | Lambda (+ API Gateway/ALB) | always-on containers | Pay per request; zero idle cost |
| Static site / SPA | S3 + CloudFront | anything with servers | Cents per month |
| A database for relational data | RDS Postgres | self-managed on EC2/k8s | Managed backups/patching beat the instance premium for small teams |
| Key-value at unpredictable scale | DynamoDB on-demand | provisioned RDS | Scales to zero cost when idle |

When the user names a service explicitly, honor it — but if a materially
cheaper or simpler shape fits, say so once, with the delta, and let them
choose.

## Network shape

- One VPC, two AZs is the sane default; more AZs only when the motive is
  production resilience.
- Public subnets for load balancers and NAT; private subnets for compute and
  data. Databases never get public IPs.
- NAT: single gateway for dev (per-AZ only for production), and question
  whether NAT is needed at all — a workload that only serves inbound traffic
  through an ALB may not need a standing egress path at all, and every NAT
  gateway bills around the clock (its cost fact-sheet has the figure).
- Security groups: least privilege between tiers (ALB → service port only,
  service → database port only). Never 0.0.0.0/0 ingress except 80/443 on the
  public-facing load balancer.

## Security defaults (non-negotiable, cost-free)

These cost nothing and are always on in charts you compose: encryption at
rest where the kind offers it (S3, RDS, EBS), block-public-access on S3
buckets unless the bucket IS a public website, IAM roles over long-lived
keys, private DB subnet placement, HTTPS termination at the edge (ACM certs
are free). When a user asks to loosen one, comply after saying the risk in
one sentence.

## Cost posture

`cost-transparency.md` governs the duty; the AWS specifics that matter most:
the always-on trio (EKS control plane, NAT, load balancers) usually IS the
bill at small scale — architecture choices that remove one of them outweigh
any instance-type tuning. Prefer Graviton (ARM) instance types where the
workload allows; prefer on-demand-priced serverless (Lambda, DynamoDB
on-demand, Fargate) for spiky workloads.

## Composing beyond AWS primitives

The catalog also carries Kubernetes workload kinds — when the user's EKS
cluster is the platform, in-cluster kinds (ingress-nginx, cert-manager,
Istio) compose per `kubernetes-on-cluster.md`. The AWS judgment still
applies underneath: the cluster's endpoint exposure, node sizing, and NAT
shape follow the motive.
