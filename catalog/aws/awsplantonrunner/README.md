# AwsPlantonRunner

Runs a standing Planton runner appliance inside your AWS network: an
always-on worker that receives deploy operations from the Planton control
plane and executes them from within the VPC -- with an outbound-only
network posture (the runner dials out; nothing dials in).

## Purpose

Some infrastructure is reachable only from inside the network. The
canonical case is a Kubernetes cluster with a **private API endpoint**: no
hosted runner fleet can reach it, so nothing outside the VPC can deploy
into it. Placing a runner inside the VPC makes that cluster deployable and
operable -- initial installs, day-2 updates, destroys, and live resource
browsing -- without opening a single inbound port.

The appliance is standing infrastructure, not a bootstrap step. It
survives rebuilds of the clusters it deploys to, which is what makes
teardown orderly: in-cluster workloads are destroyed through the runner,
the cluster is destroyed over the AWS path, and the runner itself is
destroyed last.

The compute substrate is **ECS Fargate**: serverless containers, no hosts
to patch, restarted automatically if the runner ever exits. The spec
deliberately does not model the substrate -- it models intent (placement,
sizing, version, identity, and the token the runner joins with), so the
API stays stable however the implementation evolves.

## Key Features

- **Outbound-only networking** -- the runner initiates every connection
  (control plane, its work queue, image pulls). The security group created
  for it allows no inbound traffic at all.
- **Token-first enrollment** -- the runner is born with a runner TOKEN,
  never an identity. On first boot it presents the token, registers
  ITSELF, and receives its own individually revocable identity; revoking
  the token never touches runners it already admitted, and task
  replacement re-joins with the same token (the token's lineage re-admits
  the runner it originally admitted).
- **Pull-based execution** -- the runner polls its own queue for deploy
  operations; work waits in the queue while the runner boots, so ordering
  never depends on timing. There is no mode knob: the runner
  self-configures its execution mode from the join response.
- **First-class runtime identity** -- the runner holds an IAM role at
  runtime (reference your own `AwsIamRole`, or a permissionless one is
  created), the seam that lets keyless cloud access run through the
  runner without long-lived keys.
- **The token handled as a secret end to end** -- the runner token is
  stored in AWS Secrets Manager and injected into the container at start
  by AWS itself; it never appears in any launch configuration or task
  definition.
- **Auditable by design** -- every operation the runner executes lands in
  its CloudWatch log group, with retention you control.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsPlantonRunner
metadata:
  name: vpc-runner
spec:
  region: us-west-2
  subnets:
    - valueFrom: { kind: AwsSubnet, name: private-a, fieldPath: status.outputs.subnet_id }
    - valueFrom: { kind: AwsSubnet, name: private-b, fieldPath: status.outputs.subnet_id }
  token: $secret/vpc-runner-token
```

There is no manual credential step: before the infrastructure applies, the
platform mints a runner token and writes it at exactly the managed-secret
reference the manifest declares. The token only authorizes the runner to
join -- the runner registers itself on first boot and receives its own
individually revocable identity. Pick any secret slug and deploy with:

```shell
planton apply -f runner.yaml
```

Both a Pulumi module and a Terraform/OpenTofu module implement this
kind at full behavioral parity; the provisioner is an execution
detail.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
