# Template Language

Chart templates are rendered with Jinjava -- a Java implementation of Jinja2.
The practical subset below is everything the official chart fleet uses; stay
inside it.

## Context: what variables exist

Every param from values.yaml is available two ways -- as
`values.<param_name>` (preferred, always use this) and as a bare top-level
variable. Two extra variables always exist without being declared:

| Variable | Meaning |
|----------|---------|
| `values.env` | The environment the chart is deployed into (e.g. `dev`, `prod`) |
| `values.org` | The organization deploying the chart |

`values.env` is the backbone of resource naming: `"{{ values.env }}-vpc"`.

## Substitution

```yaml
metadata:
  name: "{{ values.env }}-{{ values.cluster_name }}"
spec:
  region: "{{ values.aws_region }}"
```

Quote template expressions inside YAML values -- `name: "{{ … }}"` -- so the
file stays parseable YAML regardless of what the expression renders to.

## Conditionals

All params render as strings; the `| bool` filter converts a bool param for
truth testing. Wrap the ENTIRE document (including its `---` separator)
inside the conditional:

```yaml
{% if values.cert_manager_enabled | bool %}
---
apiVersion: kubernetes.planton.dev/v1alpha1
kind: KubernetesCertManager
metadata:
  name: "{{ values.cluster_name }}-cert-manager"
spec:
  …
{% endif %}
```

Compound conditions and string comparison:

```yaml
{% if values.https_enabled | bool and values.dns_enabled | bool %} … {% endif %}
{% if values.nat_mode != "none" %} … {% endif %}
{% if values.service_port|trim not in ["80", "443"] %} … {% endif %}
```

(Older fleet charts use camelCase param names; the platform accepts both, but
new charts follow the snake_case convention from `infra.chart-format.md`.)

A conditional can also gate a single field or list item, not just whole
documents -- but whole-document toggles are the common, legible case.

## Loops

```yaml
{% for az in values.availability_zones %}
---
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSubnet
metadata:
  name: "{{ values.env }}-subnet-{{ loop.index }}"
spec:
  availabilityZone: "{{ az }}"
{% endfor %}
```

Loops are rare in the fleet -- most charts write each subnet/AZ explicitly
because they need distinct CIDR params anyway. Prefer explicit resources
unless the count itself is a parameter.

## Config references compose with templating

`$var`/`$secret` reference strings (`infra.config-references.md`) are plain field
values, and slugs can never contain `@` — so the env-scope sigil composes
cleanly with substitution. The canonical per-environment secret:

```yaml
spec:
  password: $secret/@{{ values.env }}/db-password
```

The template renders to `$secret/@staging/db-password` in staging; the
reference resolves inside the deployment runner, never at render time.

## Filters

The ones the fleet actually uses: `| bool`, `| trim`, `| upper`, `| lower`.
Do not reach for exotic filters; if logic gets complex, reshape the params
instead.

## Rendering model (what to expect from the compiler)

- Each template file renders independently with the full context, then every
  rendered document is parsed as YAML and validated against its kind's
  schema.
- Render or YAML-syntax failures are attributed to the template file; a
  file+message issue means look at that file first.
- A document whose conditional evaluates false simply does not exist in the
  output -- it appears nowhere in the build report's resources array. That
  is how you verify a toggle actually works: build once per position with
  `--set the_toggle=true` / `--set the_toggle=false` (no file edits) and
  compare the resources arrays.
- Rendering merges everything; there are no line numbers in build issues.
  Issues carry file + message + (when attributable) resource kind/name.
