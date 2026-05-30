# Proto hygiene + E2E fixture retirement

Three cleanups across the API surface and the E2E harness: fixed the Gateway API
CRDs stack-outputs contract, eliminated the protobuf `reserved` keyword from the
repo, and retired the now-unused E2E fixture-override mechanism.

## 1. KubernetesGatewayApiCrds stack-outputs contract

The `installed_crds` output was a `(version x channel)` derivation hardcoded
version-blind -- it was frozen at the component's `v1.2.1` default (where GRPCRoute
and TLSRoute were still experimental-only) and so reported the wrong CRD set for
the v1.5.1 the Gateway API family targets. It was also mis-rendered in the console
(a `repeated string` bound into a single string field) and carried no DAG value
(downstream resources depend on the CRDs *existing*, not on parsing a name list).

- Removed `installed_crds` from `stack_outputs.proto`, the Pulumi `outputs.go`,
  and the Terraform `outputs.tf`/`locals.tf`.
- Added `installed_manifest_url` in its place (field 3) -- the exact CRD bundle URL
  applied, which encodes version + channel in a single always-truthful value. This
  also resolves a three-way drift: Terraform already emitted `manifest_url` while
  the proto contract and Pulumi did not. The TF output is renamed to
  `installed_manifest_url` to match.
- Fixed stale channel docs in `spec.proto` and the IaC READMEs (GRPCRoute/TLSRoute
  are standard as of v1.5.1; experimental adds TCPRoute/UDPRoute).

Cross-repo stub/catalog/console propagation happens on the next openmcf release +
`make upgrade-openmcf`; the console row was updated in lockstep (planton repo).

## 2. Eliminated the `reserved` keyword

OpenMCF is pre-1.0 with no external wire-compatibility obligations, so `reserved`
is unnecessary ceremony. Both real usages were removed:

- `cloudflarer2bucket/v1/spec.proto`: dropped `reserved 5` / `"versioning_enabled"`
  and renumbered `custom_domain` `6 -> 5` so fields are sequential again.
- `cloud_resource_kind.proto`: dropped `reserved 2523, 2524` (the removed OpenStack
  Magnum kinds). The freed numbers are left as free *primary* slots within the
  OpenStack block (the block documents `2507-2524` for primaries, `2525+` for
  companion kinds), marked with a one-line breadcrumb. **No enum value was
  renumbered** -- kind numbers are permanent identities and the provider-block
  allocation scheme is preserved.

The convention is now codified in
`_rules/coding-guidelines/protobuf-validations.mdc` ("Removing Fields and Enum
Values"): delete + renumber message fields, never `reserved`; never renumber an
existing kind; preserve the per-provider block grouping and its sub-conventions.

## 3. Retired the E2E fixture-override mechanism

The E2E harness (T08) supported two dependency sources: registry `prerequisites`
(source of truth) and hand-authored `v1/e2e/fixtures/` overrides. The only six
fixtures in the repo (Postgres/Kafka/MongoDB/ClickHouse/Elasticsearch/Solr
operators) were exact -- or name/namespace-only -- duplicates of the operators'
own `scenarios/minimal.yaml`, and each consumer already declares its operator as a
registry `prerequisite`. So the fixtures were dead duplication and the override
path had zero users.

- Deleted the six `fixtures/` directories.
- Removed the dead machinery from `dependencies.go`: `resolveFixtureDependencies`,
  `kindSlugFromFixtureFilename`, the fixture/registry merge + dedupe, and the whole
  `DependencySource`/`Source` abstraction (it existed only to distinguish the two
  sources). `ResolveDependencies` is now a straight registry resolution.
- Updated `dependencies_test.go`, the `runner.go` phase comments, and
  `e2e/README.md` accordingly.

Operator-dependent components now install their operator from the operator's
`scenarios/minimal.yaml` via the registry path -- the same path the Gateway API
family already uses. Re-add an explicit per-consumer override only if a real need
arises (YAGNI).

## 4. .gitignore: terraform local artifacts

`terraform init`/`validate` (run to validate the CRDs module) creates a
`.terraform/` provider directory and lock/state files that must never be
committed. The ignore list only covered `.terraform.lock.hcl` and
`terraform.tfplan`, so added a complete block: `**/.terraform/`, `*.tfstate`,
`*.tfstate.*`, and `*.tfplan`.

## Incidental fixes surfaced by `make protos`

- `pkg/crkreflect/BUILD.bazel`: gazelle added the Gateway API family deps
  (gateway, gatewayclass, http/grpc/tls/tcp routes, referencegrant) that
  `kind_map_gen.go` already imports -- a latent BUILD/source sync gap from when
  the family was forged. Healed automatically.
- `kubernetesgrpcroute/v1/spec_test.go`: one gofmt alignment fix.

## Validation

- `make protos` -- clean (buf lint/format/generate, Java compile gate, gazelle).
- `make generate-cloud-resource-kind-map` -- regenerated with **no diff** to the
  kind map, confirming no enum value changed (only the two `reserved` lines and
  the `cloudflarer2bucket` field number moved).
- `go build` of all touched packages (kubernetesgatewayapicrds, cloudflarer2bucket,
  cloudresourcekind, e2e, crkreflect) -- clean.
- `go vet` + `go test ./e2e/framework/runner/...` (the rewritten dependency engine
  + updated unit tests) -- green.
- `make e2e-build` + `make e2e-vet` -- clean.
- `terraform validate` on the CRDs module (with the renamed output, no dangling
  `installed_crds`/`*_crds` locals) -- valid; init artifacts removed afterward.
- Stub spot-check: `InstalledManifestUrl` present / `InstalledCrds` gone in
  `stack_outputs.pb.go`; `custom_domain` is field 5 in the cloudflare stub;
  OpenStack companion kinds still 2525/2526 in `cloud_resource_kind.pb.go`.
- Planton web checks intentionally deferred (per owner): the console row edit is
  forward-compatible and propagates with the next openmcf release + `make
  upgrade-openmcf`.

## Pending (handoff)

Part B's six operator-dependent e2e tests (Postgres/Kafka/MongoDB/ClickHouse/
Elasticsearch/Solr, both Pulumi + Terraform) need a live `kind` run to confirm the
registry path installs each operator from its `scenarios/minimal.yaml`. Four are
byte-identical to the deleted fixture (no-op); MongoDB and Solr differ only in the
operator install name/namespace (benign, since the operator is a cluster-scoped
controller). Run on a kind cluster:

```
go test -tags=e2e ./e2e/... -run 'TestKubernetes(Postgres|Kafka|Mongodb|Clickhouse|Elasticsearch|Solr)_(Pulumi|Terraform)'
```

Expected: all pass; each installs its operator as a registry prerequisite, applies
the data CR, verifies, and tears down.
