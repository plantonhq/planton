# State Import — Adopting Provider Resources That Exist but Aren't in State

Read this when a deploy fails saying something **already exists** — the
signature of an orphaned resource: an earlier run created it in the cloud,
then died before the IaC state recorded it (an auth failure or crash mid-way
through a long create). The cloud has the resource; the state file does not;
every rerun tries to create it again and collides.

Recognize the signature in the failed node's engine logs (step 3 of the
diagnosis workflow in `craft.planton-cli.md`):

```
googleapi: Error 409: Already exists: projects/.../clusters/prod-cluster   # GCP
...AlreadyExistsException / BucketAlreadyOwnedByYou / EntityAlreadyExists  # AWS
...Conflict: ... already exists                                            # Azure
Error: ... a resource with the ID "..." already exists                     # generic tofu
```

**The repair is import, not delete-and-retry.** The platform has first-class
state-import commands: they run an Infra Job that adopts the existing provider
resource into the InfraComponent's IaC state — **the cloud is never touched,
only the state file is updated** — and then a fresh apply reconciles the
desired configuration against what was adopted.

## The commands

The provisioner family matters (check `.planton/stack.yaml` or the Infra
Job record; OpenTofu and Terraform are interchangeable here):

```
# OpenTofu / Terraform — the state entry is a resource ADDRESS (type.name):
planton tofu state import <IC_ID | Kind name> \
  --address "<tf_type.tf_name>" --id "<cloud-provider-id>" \
  -m "adopt orphaned resource created by failed run" [--dry-run]

# Pulumi — the state entry is a TYPE plus a logical NAME:
planton pulumi state import <IC_ID | Kind name> \
  --type "<pulumi:type:Token>" --name "<logical-name>" --id "<cloud-provider-id>" \
  -m "adopt orphaned resource created by failed run" [--dry-run]
```

- The target is the **InfraComponent whose stack owns the orphan**: pass its
  id (`ic_...`) or `<Kind> <name>` under the current org/env context (e.g.
  `GcpGkeCluster prod-cluster`). Find it with
  `planton search infra-components --org <org> -e <env>`.
- `--dry-run` prints the equivalent native `tofu import`/`pulumi import`
  command without creating anything — always show the user a dry run before
  the real one.
- The import job is **idempotent** (importing an already-tracked resource
  succeeds) and runs `init → import → refresh → preview → capture` as one
  Infra Job. It reports drift but does not apply it — reconciling is the
  follow-up deploy.
- One import command adopts ONE resource; run it once per orphan.

Related state commands when the repair needs them: `planton <prov> state
list <CR…>` (what the state tracks now — run it after the import to
confirm), `state show --address/--urn`, `state rm` (the inverse: forget a
state entry without touching the cloud).

## Finding the two identifiers

**The address (or type/name)** — what the IaC module calls the resource:

- The failed run's engine logs name the address being created at the moment
  of failure (`google_container_cluster.this: Creating...`).
- `planton <prov> state list <CR…>` shows the addresses of everything
  already tracked — the orphan's siblings reveal the module's naming shape.

**The cloud ID** — the provider's identifier, in the format the provider's
import documentation defines for that resource type. Look it up read-only
with the provider's own CLI (the read-only regime in
`cloud-exploration.md`; describe/list calls run freely):

```
# GKE cluster → projects/<project>/locations/<location>/clusters/<name>
gcloud container clusters list --format="value(name,location)"

# AWS: most resources import by their id or name
aws ec2 describe-vpcs --query 'Vpcs[].VpcId'      # vpc-0abc...
aws s3api list-buckets --query 'Buckets[].Name'   # bucket name IS the id

# Azure: usually the full ARM resource ID
az resource list --name <name> --query '[].id'
```

When unsure of the exact format, check the tofu/terraform provider's import
documentation for that resource type — the `--dry-run` output is a safe way
to show the user exactly what would run before committing.

## Consent and boundaries

- An import is a **platform mutation** (it changes the stack's recorded
  state): explain what it adopts and get a yes — one confirmation per
  import. `--dry-run` and every lookup above run freely.
- Never repair by deleting the provider resource so the rerun "works" unless
  the user explicitly chooses that instead — deletion destroys whatever the
  resource already holds and is a cloud mutation with its own confirmation.
- After a successful import, the deploy is still pending: save/rerun (its
  own consent, per `infra.deployed-stacks.md`) and confirm the run goes green.
