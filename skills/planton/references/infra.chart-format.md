# Chart Format

The exact shape of the three chart components. A chart is the
organization's own architecture, written for its clouds and its
environments: Planton ships none. Ground every kind in its reference page
(`catalog.kind-grounding.md`), and let `planton chart build` judge the
chart before it is published.

## Chart.yaml

Chart.yaml IS a Planton manifest of kind `InfraChart` -- the same
apiVersion/kind/metadata/spec structure every Planton resource uses:

```yaml
apiVersion: infra-hub.planton.ai/v1alpha1
kind: InfraChart
metadata:
  name: EKS Environment            # human-readable display name
spec:
  description: Configurable EKS environment with custom VPC, Route 53,
    autoscaled managed node group and toggleable Kubernetes add-ons.
  iconUrl: https://downloads.planton.ai/catalog/logos/{provider}/{kind}/logo.svg   # optional
  webLinks:                                        # optional
    chartWebUrl: https://github.com/org/repo/tree/main/charts/aws/eks-environment
    readmeRawUrl: https://raw.githubusercontent.com/org/repo/main/charts/aws/eks-environment/README.md
```

Required for composition: `apiVersion`, `kind`, `metadata.name`,
`spec.description`. Everything else is catalog polish.

Chart.yaml names no owner. A chart belongs to the organization that
publishes it: `planton chart publish` records the organization in your CLI
context (or the one `--org <org>` names) in the chart's `metadata.org`.
Every chart has an organization; a chart without one is refused.

`planton explain infra-chart` renders the InfraChart schema itself —
including which spec fields the tooling assembles from the chart directory
(marked `(assembled)`: templates, values, params) and must therefore never
appear in Chart.yaml. This file covers the practical shape; the explain
report is the field-name authority.

## Versions are counted by the platform

A chart has no `version:` line — never write one in Chart.yaml. The
platform counts versions itself: a chart is **v1** when first published,
and becomes the next version (v2, v3, …) each time its templates or
`values.yaml` change. Publishing unchanged files keeps the version, so a
version always means the same files; each version carries a digest the
platform computes from them. `planton chart publish` prints the version it
made, and `planton chart list` shows each
chart's version.

A stack installed from a catalog chart records the chart and the version
it took ("app-infra v3"). When the chart publishes a newer version, the
stack reads **Update Available**, and **Upgrade** moves it there while
keeping its values and any file it customised
(`infra.stack-promotion.md`). Publishing never changes a deployed stack.

## values.yaml

A single `params` list. Each param:

```yaml
params:
  - name: aws_region                  # snake_case; referenced as values.aws_region
    description: AWS region for every resource   # REQUIRED in practice -- it is the user's docs
    value: us-east-1                  # the default; strings unless typed

  - name: dns_enabled                 # bool params drive conditionals
    description: Create the Route53 zone and DNS records
    type: bool
    value: true

  - name: node_count
    description: Number of worker nodes
    type: number
    value: 2
```

- `name`: snake_case, unique.
- `description`: always write one -- it renders as the user's form help.
- `type`: `string` (default), `bool`, `number`, `list`, `string_enum`.
- `value`: the default. **The YAML scalar's own type must match the declared
  `type`** -- the platform stores params as typed values and rejects a
  mismatch at build time:

  ```yaml
  # WRONG — a quoted number is a STRING; the build rejects it for type: number
  - name: node_count
    type: number
    value: "2"

  # RIGHT — number and bool values are written bare
  - name: node_count
    type: number
    value: 2
  ```

  The inverse rule protects strings: keep values that could parse as numbers
  quoted when they are semantically strings (versions like `"1.31"`, account
  IDs, ports-as-strings) -- unquoted, YAML would corrupt them.

Design guidance:

- A param earns its place when the person DEPLOYING genuinely decides it
  (image, hostname, region, CIDRs, sizes, toggles). Internal wiring never
  becomes a param -- an id, ARN, or endpoint another resource produces is a
  `valueFrom` reference, even when the producer lives in a different chart
  (`infra.dependencies.md`, the cross-boundary check).
- Fewer params is a feature: every exposed knob is a question the user must
  answer before they can deploy. Default everything that has a sane default
  and name the default in the description; lead the list with the
  developer-facing values (image, port, hostname) so what matters most
  reads first.
- Every bool param should gate something real -- a whole resource or a
  meaningful behavior change -- and the chart must build validly with the
  toggle in BOTH positions.
- Group related params with comment headers (`# ─── Network ───`) in larger
  charts.

## templates/

- Any number of `.yaml` files, nested directories allowed. The file path
  relative to `templates/` is the file name that build issues are attributed
  to.
- **Prefer subfolders by concern with ONE resource per file** when
  authoring (`network/vpc.yaml`, `network/igw.yaml`, `cluster/eks.yaml`,
  `kubernetes/addons/cert-manager.yaml`): the tree reads as the
  architecture, build issues and canvas nodes point at exactly one file,
  and diffs stay reviewable. Multiple manifests per file separated by `---`
  remain legal -- many existing charts use them; respect a chart's existing
  layout when editing, adopt the per-file layout when composing fresh.
- Every manifest is a full catalog object:

```yaml
apiVersion: aws.planton.dev/v1alpha1        # <provider>.planton.dev/v1
kind: AwsVpc                          # PascalCase kind
metadata:
  name: "{{ values.env }}-vpc"        # ALWAYS env-prefixed
spec:
  …                                   # exactly the kind's schema
```

- `metadata.org` and `metadata.env` are injected by the platform at
  deployment -- never write them in templates.
- A comment header at the top of each template file explaining what it
  provisions and any non-obvious topology decision pays for itself.

## Naming conventions

- Resource names: `"{{ values.env }}-<role>"` (`-vpc`, `-igw`, `-public-1`,
  `-nat-1`). The env prefix is what lets one chart deploy into many
  environments without collisions.
- When a chart has a user-visible name param (like a cluster name), compose:
  `"{{ values.env }}-{{ values.cluster_name }}"`.
- **Exception — semantic names.** Some kinds carry their real-world identity
  in `metadata.name` itself (an `AwsRoute53Zone`'s name IS the domain; the
  schema has no separate domain field). Prefixing would corrupt the value
  (`dev-example.com` is not a valid zone). For these, name the resource with
  the user's param directly (`"{{ values.dns_zone_name }}"`) and skip the env
  prefix. Tell-tale sign: the spec schema has no field for the value you
  expected to set, because `metadata.name` carries it.
- Template file names: lowercase, by concern, `.yaml`.
- **Never write an environment's name literally.** The same templates
  travel from qa to prod when a stack is promoted; a literal `qa` in a
  name or a hostname would read the same in prod. A promotion's plan warns
  ("templates/db.yaml names qa (line 12). It reads the same in prod. Use
  {{ values.env }}."); write `{{ values.env }}` instead.

## What never goes in a chart

- `planton.dev/provisioner` labels (the deploying org's setting, not yours).
- `metadata.org` / `metadata.env` (platform-injected).
- Literal resource IDs copied between resources (use valueFrom).
- Secrets or credentials of any sort — a sensitive field holds a
  `$secret/...` reference to a managed secret instead
  (`infra.config-references.md` has the grammar and the lookup workflow).
