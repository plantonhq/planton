# Issue Catalog

Fix patterns for errors observed in the compile loop. Always prefer
`planton explain <Kind>` over guessing when a field or enum is
rejected.

## How to read an issue

Issues arrive in `issues[]` from `planton chart build -o json`.

- When `file` is set → open that template file first.
- When `file` is empty → the message often echoes the rendered manifest;
  match `resourceKind` / `resourceName` or search templates for the offending
  field named in the message.
- After each fix, rebuild — errors are layered; fixing one reveals the next.

## Unknown or misspelled spec field

**Symptom:**

```
Cannot find field: nonExistentField in message …AwsKmsKeySpec
```

**Fix:**

1. `planton explain AwsKmsKey` — find the correct field name (or drill:
   `planton explain aws-kms-key.spec`).
2. Rename or remove the field in the template. Protojson camelCase only
   (`cidrBlock`, not `cidr_block`).

## Invalid valueFrom output

**Symptom:**

```
Invalid valueFrom references: Field 'no_such_output' not found in …StackOutputs for kind: AwsIamRole
```

**Fix:**

1. `planton explain AwsIamRole` — read the OUTPUTS section.
2. Update `fieldPath: status.outputs.<correctLeaf>`.
3. Ensure producer `metadata.name` matches the `valueFrom.name`.

## Sensitive field rejected (plaintext, or a secret reference not found)

**Symptom** (at apply/deploy — sensitive-field validation runs on the
control plane, not in `chart build`):

```
one or more sensitive fields must reference an existing org secret (use '$secret/<slug>'); the following are invalid:
  - spec.password: secret not found for environment 'staging' slug 'db-password'
```

**Fix** (per offending field path — the error lists them all):

1. Plaintext or a placeholder in the field → replace it with a `$secret/...`
   reference; a sensitive field never accepts a literal
   (`config-references.md`).
2. "secret not found for organization slug …" → the reference is org-scoped
   but no org secret has that slug. `planton secret list -o json` — if the
   secret is environment-scoped (its record carries an `env`), add the sigil:
   `$secret/@<env>/<slug>`.
3. "secret not found for environment '<env>' …" → wrong env in the sigil, or
   the secret genuinely does not exist yet — create it
   (`planton secret set <slug> value=<value> --env <env>`) or hand the user
   that exact command with the reference already in place.

Scope is strict: org-scoped and env-scoped references never fall back to
each other. The reference must encode the scope the secret actually has.

## Secret reference in a field every viewer reads

**Symptom** (at apply or deploy, before anything is created. On a service,
saving the configuration is accepted and the deploy is refused before any
of the environment's manifests applies; the run's environment row carries
the sentence after the manifest's number, kind and name):

```
spec.containers[0].env[1].value (STRIPE_KEY) holds a secret reference, but its value is stored where anyone who can view the resource reads it -- move the reference to secretValue, the sibling field that keeps it in a secret store the workload reads by reference
```

**Cause:** the field is marked `(no secrets: use <field>)` on the
component's page. The runner resolves a `$secret/...` to its plain value
before the module runs, so in this field the secret itself would be written
into the revision, task definition, or pod spec.

**Fix:** move the reference, unchanged, to the field the sentence names --
on Cloud Run the same env entry's `secretValue` (and drop its `value`), on
ECS the container's `secretEnvironment` map (and remove the key from
`environment`), on a Kubernetes workload an entry under `env.secrets`
(`config-references.md`, "Where a secret reference may go"). Never replace
the reference with a literal to make the refusal go away: a literal in a
field every viewer reads is the same leak, typed by hand. The secret homes
are themselves `sensitive`, so once moved, the reference must name a secret
that exists at its scope -- "Sensitive field rejected" above.

## YAML parse error in values.yaml

**Symptom:**

```
failed to parse params from values.yaml: … did not find expected ',' or ']'
```

**Fix:** Fix the YAML syntax at the indicated line. Common causes: unquoted
`:` in a description string, bad indentation in the `params` list.

## Jinjava / template render failure

**Symptom:** Message references a template expression or "Error rendering
template".

**Fix:**

- Check `{% if %}` / `{% endif %}` pairing — wrap whole documents including
  leading `---`.
- Quote substitutions: `"{{ values.aws_region }}"`.
- Use `| bool` for bool params in conditionals.

## Enum or constraint violation

**Symptom:** Message mentions an invalid enum value or a buf.validate
constraint.

**Fix:**

1. Schema report — read `enum` on the field and `constraints[]`.
2. Apply `specRules[]` for cross-field fixes.

## Missing required field

**Symptom:** Required field missing or empty after render.

**Fix:** Schema report shows `required: true`. Add the field with a sensible
default or a values.yaml param.

## Conditional swallowed a resource

**Symptom:** Phase 1 plan included a resource but `resources[]` omits it.

**Fix:** The `{% if %}` evaluated false. Rebuild with the toggle param flipped
(`true` / `false`) and compare `resources[]` — both branches should produce
valid charts when the param is meant to gate optional infrastructure.

## Wrong apiVersion or kind

**Symptom:** Unknown kind or apiVersion mismatch.

**Fix:** Schema report shows `kind` and `apiVersion` at the top — copy exactly
into the manifest header.

## Quoted-number corruption

**Symptom:** A string that looks numeric was mangled by the YAML parser after
render — a Kubernetes version becomes a truncated float (`"1.30"` → `1.3`), or
an account ID loses its leading zero (`"066380525333"` → `66380525333`) — and
the build rejects the field or the value silently changes.

**Fix:** Quote in the template: `"{{ values.cluster_version }}"` and quote
defaults in values.yaml when they look numeric.

## No provider connection at deploy time (Kubernetes workloads)

**Symptom** (at deploy, not at build — the resource fails the instant its
pipeline node starts):

```
No provider connection available for CloudResource creation.
```

**Fix:** The resource carries no `planton.dev/connection` annotation and the
org/env has no default Kubernetes connection. If the cluster is in the SAME
chart, the annotation is missing — add it to every Kubernetes-kind resource
(the same-chart pattern in `kubernetes-on-cluster.md`). If the cluster lives
elsewhere, the organization has no cluster connection yet: a cluster deployed
through Planton (or connected via the desktop) establishes the default
binding automatically, so this error means no such cluster exists — deploy or
connect one; do not paper over it with annotations or params. The same error
on NON-Kubernetes kinds means the organization has no connection for that
cloud at all — the machine's own login may be the fastest way forward
(`machine-deploy.md`: probe, offer, consent), and an organization cloud
account connected through the console is the standing alternative.

## Connection slug not found at deploy time

**Symptom:**

```
kubernetes-provider-connection with slug 'dev-my-cluster' not found in org …
```

**Fix:** The annotation is present but its value does not match any
connection. In a one-run composition this means the producer and consumer ends
disagree: the workload's `planton.dev/connection` must render to the identical
string as the cluster's `planton.dev/connection-name` (or, when no override is
set, to `<env>-<cluster resource metadata.name>`). Use one values expression on
both ends. For an existing cluster, verify the real connection slug via the
CLI instead of guessing.

## When to stop fixing

- **Exit 2** — environment problem. Do not patch templates.
- Same error after two schema-grounded attempts — re-read the preset or fleet
  example for that kind; you may be modeling the wrong shape entirely.
