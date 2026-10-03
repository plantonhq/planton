# Apply captures outputs (masked display), pulumi state backend becomes configurable, and multi-document input refuses loudly

## What changed

- **`planton apply` now shows the stack's outputs after a successful
  apply** — the behavior every IaC user expects from `terraform apply`,
  previously absent entirely. The engine reads outputs back while the
  module workspace is still alive (`tofu output -json` / a pulumi
  output read pair), transforms them through the kind's typed
  Outputs contract (module-shipped transform overrides honored),
  and the CLI renders a compact summary. **Sensitive values never
  print**: tofu's per-output `sensitive` flag and a masked-vs-shown
  comparison on the pulumi side drive `(sensitive)` rendering, because
  apply runs in CI as much as on laptops and CI logs are persistent.
  A capture failure after a successful apply reports honestly and never
  fails the apply — the infrastructure is already deployed. The seam is
  an opt-in functional option (`WithOutputCapture`) on both engines' run
  functions, so every existing call site compiles untouched and callers
  that skip the option pay no capture cost.

- **The pulumi state backend joins the configuration model tofu has had
  all along.** Previously the pulumi lane had NO backend configuration —
  state landed wherever the machine's ambient `pulumi login` pointed,
  which is ambiguous in CI. Now the backend URL resolves through the
  same three-layer precedence the tofu backend follows — `--backend-url`
  flag > `pulumi.planton.dev/backend.url` manifest annotation >
  `PLANTON_BACKEND_URL` environment variable — and pins
  `PULUMI_BACKEND_URL` for the run and its output reads. Nothing
  configured keeps today's ambient-login behavior.

- **Stack selection precedence is normalized** (behavior change): the
  `--stack` flag now wins over the `pulumi.planton.dev/stack.fqdn`
  annotation. Previously the annotation silently overrode an explicit
  flag — the inverse of every other flag-vs-annotation precedence in the
  CLI. The annotation still applies whenever the flag is absent.

- **Multi-document YAML input refuses loudly instead of silently
  deploying one document.** The manifest loader reads exactly one YAML
  document; previously a multi-document input — most dangerously a
  kustomize overlay bundling several resources — was silently truncated
  to its FIRST document: one resource deployed, success reported, the
  rest dropped without a word. The loader now refuses multi-document
  input naming every document found (kind/name), at the one funnel every
  load passes through, so every command inherits the honesty. Callers
  that legitimately handle multi-document streams split first via the
  new `SplitDocuments` helper.

- **Preset validation now checks EVERY document of a composition
  preset.** Multi-document presets (e.g. the AzurePublicIpPrefix
  two-step composition) previously had only their first document
  validated — later steps could have shipped broken without a gate
  noticing. The preset-validity gate now splits and validates each
  document.

- **The catalog-page baseline burns to zero.** The one remaining
  baseline entry (gcpsslcertificate's recorded invalid-manifest) was
  stale — the manifest validates today — so the entry is removed and no
  catalog page in the tree sits below the ONE catalog-page standard.

## Why

Captured, typed outputs are the composition primitive everything
downstream consumes — a deploy whose outputs vanish into state can only
be composed by hand. The masked-by-default display and the loud
multi-document refusal both close silent-failure classes: secrets in CI
logs, and the overlay that deploys a fraction of itself while claiming
success.

## Verification

- `go build` + `go vet` clean across every touched package and the e2e
  tree; gazelle BUILD regeneration included.
- Unit suites added: tofu envelope unwrapping with sensitivity flags,
  pulumi masked-vs-shown secret detection, backend-URL precedence (all
  layer combinations), multi-document refusal (document naming,
  separator noise, malformed-YAML deference), and the masking law
  itself (a sensitive value never appears in rendered output).
- Gates re-run green: `pkg/outputs`, `pkg/iac/...`, `internal/manifest`,
  `internal/cli/ui`, `pkg/presetvalidity`, `pkg/certification`,
  `pkg/catalogpage`.
