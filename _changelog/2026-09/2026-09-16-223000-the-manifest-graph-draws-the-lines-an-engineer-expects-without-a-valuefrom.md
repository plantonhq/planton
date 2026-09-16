# The manifest graph draws the lines an engineer expects without a `valueFrom`

## What changed

- **Two new edge sources in `pkg/manifestgraph`, the fifth and sixth.** A route that names its gateway by a literal name, a certificate that names its issuer, a workload that names its service: the shape the console's wizard writes and `kubectl` users think in, and until now a shape the set lane ordered nothing by. **Literal siblings**: a literal in any field annotated with a `default_kind` becomes an edge when a resource of that kind with that slug is in the set, in the same env. A Kubernetes name is an identity and matches; a cloud id (`vpc-0abc123`) slugs to nothing the set holds and never does; no node is minted for an unmatched literal and no finding is raised, because a literal that names nothing in the set is the common, legitimate case. The namespace stays the placement source's own (it alone mints a derived target when absent). **Operator prerequisites**: the kind metadata's `prerequisites` list, which the proto comment has long said the platform orders by and nothing read, becomes an edge for every prerequisite in the Kubernetes operators-and-controllers service group (operators, controllers, and the CRD bundles that admit a kind's objects) when exactly one instance of that kind is in the set. Two candidates mean the set does not say which one reconciles the operand, and the graph stays silent rather than choose.

- **Every edge carries its source.** `Graph.DependsOn` is `[][]Dependency{Producer, Source, FieldPath}` rather than bare indexes: the graph can now say why a node orders where it does, and the one consumer that iterated the old shape (`pkg/setdeploy`) reads the same facts. `Graph.Producers()` hands the plain adjacency list to the ordering and cycle routines, whose signatures are unchanged.

- **An inference yields to the author.** Two of the six sources are the author's own words (`valueFrom`, `metadata.relationships`); four are inferences from metadata and matching names. When an inferred edge closes a cycle with authored edges, `BuildGraph` drops it and reports `derived-edge-dropped`, so a set the author never wrote a cycle into stays orderable; a cycle among authored edges stands and `TopoOrder` reports it as before. The offline deploy lane treats the drop as a stated assumption, never a refusal.

- **One traversal, two collectors.** `CollectLiteralUses` is the literal-arm twin of `CollectRefUses`; both ride the same walk (`walkPopulated` takes a visitor), so a reference and a literal can never be found by two walkers that disagree about a container. `Node.LiteralUses()` is cached from set construction beside `RefUses()`.

- **Five corpus scenarios pin the semantics** for every lane that orders a set: `literal-sibling`, `literal-sibling-no-match`, `operator-prerequisite`, `operator-prerequisite-ambiguous`, `derived-edge-cycle`. No existing golden moved. Unit tests pin the two matching rules and the drop at the rule, so a drift is named there rather than discovered as a reordered golden.

- **The offline policy pin is green again.** `pkg/setdeploy`'s corpus policy test names every scenario or fails; the two connection-placement scenarios landed earlier today without an entry. Both are pinned as schema refusals (their cluster fixtures are deliberately schema-minimal), the five new scenarios as deploys.

- **Three Cloudflare guides teach cost drivers instead of quoting prices.** The custom-hostname allotment, the load balancer's rule cap by tier, and Advanced Certificate Manager's per-zone add-on are said as what bills and why, never as a dollar figure; the price-provenance gate is green again after four pushes red.

## Why

The platform's own front door renders as seven islands: every route names its gateway, every certificate its issuer, by the literal name an engineer would type, and the platform drew no line for any of them. Every hosted platform names the operator it needs in its kind metadata and no lane read it, so an operator stood in its own namespace with no line to the clusters it reconciles. Both facts were already in the manifests and the catalog; the graph now reads them, in the one package every ordering lane shares, so `planton apply` orders by them today and the platform's picture draws them the day it pins this release.

## Validation

`go test ./pkg/manifestgraph/ ./pkg/setdeploy/ ./pkg/infrachart/ ./pkg/priceprovenance/ ./pkg/cataloglogo/` green; `go vet` on the package and every importer; `planton validate-refs --check` and `planton secret-coverage --check` green from a fresh build. No `.proto` changed, so the stubs and the containment ledger are as released.
