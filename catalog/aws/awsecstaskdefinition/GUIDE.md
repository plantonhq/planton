# AwsEcsTaskDefinition Guide

The judgment this guide protects: a task definition is an immutable,
versioned document. Every apply registers a NEW revision of the family
(named from `metadata.name`), old revisions stay describable, and a
referencing `AwsEcsService` rolls because the revision-carrying ARN changed.
Everything below follows from that -- including why a secret in the wrong
field is a permanent leak rather than a mistake you can edit away.

## Three places a variable can live -- one keeps a Planton secret secret

A container's variables come from three maps, and each name may appear in
only one of them:

- `environment` is written into the task definition itself. Anyone with
  `ecs:DescribeTaskDefinition` in the account reads it, and because
  revisions are immutable and never deleted (only deregistered), every
  value ever written there stays readable. Configuration only. On Planton a
  `$secret/...` reference here is refused before anything deploys: the
  platform resolves references to plain values before the module runs, so
  the reference would be registered as the secret itself, forever.
- `secretEnvironment` is for a secret Planton holds -- on Planton it takes
  only a `$secret/...` reference (outside Planton, the value itself): the
  component stores each entry in its own Secrets Manager secret named
  `<family>/<container>/<name>` (unique, because the family already is
  within an account and region), attaches a resource policy that lets only
  the `executionRole` read it, and puts that secret's ARN -- pinned to the
  stored version -- into the container's secrets. A changed value registers
  a new revision and the service rolls, so a rotation is a deploy you can
  see; changing the Planton secret redeploys nothing by itself, and every
  deploy resolves the reference again. Destroying the task definition
  deletes those secrets immediately (no recovery window: the copy is
  derived from whoever supplied the value).
- `secrets` takes the ARN of a Secrets Manager secret or SSM parameter YOU
  own. Its lifecycle, rotation, and read grant stay yours; the execution
  role must be allowed to read it or every task start fails at secret
  resolution.

Use `secretEnvironment` for anything referenced as `$secret/...`; use
`secrets` when another owner rotates the secret or several services share
it.

## Two roles, never one

`executionRole` is the ECS agent's identity: it pulls private images,
fetches every secret above, and writes log streams. `taskRole` is your
code's identity for its own AWS calls. Merging them hands the application
the agent's secret-reading power and the agent the application's data
access. The execution role is required as soon as the task uses the
default awslogs logging on Fargate or any `secretEnvironment` -- the spec
refuses both without one, at validate time rather than at task start.

## On the diagram

The task definition is referenced by `AwsEcsService` through
`status.outputs.task_definition_arn` and references `AwsIamRole` twice
(execution and task identity). The secrets `secretEnvironment` creates are
owned by this node and do not render as their own; a secret that other
resources also read belongs in a dedicated `AwsSecretsManagerSecret` node,
referenced here through `secrets`, so the diagram shows who shares it.

## Pairs well with

- `AwsEcsService` -- runs and rolls the revisions this kind registers.
- `AwsIamRole` -- one execution role, one task role (judgment above).
- `AwsSecretsManagerSecret` -- a secret with its own owner or several
  readers, referenced through `secrets`.
- `AwsEcsCluster` -- where the service schedules the tasks.
