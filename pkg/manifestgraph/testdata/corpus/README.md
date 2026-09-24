# The manifest-graph behavior corpus

One scenario per ordering/classification semantics point. Each directory holds the manifest set (sorted filename order is the authored order) and a committed `golden.yaml`: the dependency order, derived placement targets, and classification verdicts (`<class> <node> <fieldPath>` — deliberately class-and-location, not message text, so wording can improve without a semantics ceremony).

These goldens are a CONTRACT, not a snapshot: every lane that orders a set of manifests — offline deployment, chart validation's edge pass, and the platform's server-side orchestrator — must produce these orders and verdicts for these inputs. A change that shifts a golden is a semantics change for every lane at once and is made deliberately, in the open, never by regenerating and moving on.

Regenerate (then REVIEW the diff as a semantics review):

```
PLANTON_REGEN_MANIFESTGRAPH_GOLDENS=1 go test ./pkg/manifestgraph/
```

The scenarios, by the semantics point each pins:

| Scenario | Pins |
|---|---|
| `two-node-real-kinds` | The canonical two-node composition on real catalog kinds: an annotation-riding reference orders the producer first |
| `annotation-riding-ref` | Defaults materialize BEFORE ordering: a reference with no explicit `kind:` still forms the edge its annotation implies |
| `explicit-kind-ref` | The fully-spelled reference forms the same edge |
| `relationships-edge` | `metadata.relationships` entries are ordering facts on their own, with no `valueFrom` anywhere |
| `derived-namespace` | A literal namespace on a placement-annotated field is a derived target when the set does not deploy it — recorded, never refused, never deployed |
| `namespace-edge` | The same literal becomes a real edge when the set DOES deploy the namespace |
| `connection-placement` | A Kubernetes workload whose `planton.dev/connection` names the connection a sibling cluster publishes (`planton.dev/connection-name`) runs on that cluster — the cluster orders first with no relationship authored |
| `connection-placement-default-name` | The same edge against the cluster's DEFAULT published name `<env>-<name>` when no `connection-name` annotation is set |
| `literal-sibling` | A LITERAL in a `default_kind` field that names a sibling of that kind by slug (a route naming its gateway) is an edge — the shape the wizard writes and kubectl users think in |
| `literal-sibling-no-match` | A literal that names nothing in the set (a cloud id in a VPC field) forms no edge and mints no node — silence, never a phantom |
| `operator-prerequisite` | The kind metadata's operator prerequisite orders first when exactly one instance is in the set, with no field referencing it |
| `operator-prerequisite-ambiguous` | Two instances of the prerequisite kind: the set does not say which reconciles the operand, so no edge forms and the authored order stands |
| `derived-edge-cycle` | An inferred edge that closes a cycle with an authored edge is dropped and reported (`derived-edge-dropped`); the authored order stands and nothing refuses |
| `map-ref` | References inside map-typed fields form edges (the traversal has no map blind spot) |
| `cycle` | A dependency cycle yields no order and names the chain |
| `external-valuefrom` | A `valueFrom` target outside the set is the external classification (consumer policy decides warning vs refusal) |
| `env-external` | A reference naming another env explicitly is its own classification — cross-environment by design |
| `external-relationship` | A relationship target outside the set is a stated assumption, not a broken reference — no value is needed |
| `duplicate-identity` | Two manifests with one identity: the first stays the node, the duplicate is a finding |
| `phantom-node-slug` | Node identity and edge-target identity derive through the ONE slug function: a name with spaces/case joins the node it means |
| `ref-rule-violation` | A reference that overrides the annotated kind without spelling a field path violates the strict rules (and its target is honestly classified too) |

## What the corpus does not pin

Infra-chart validation (`pkg/infrachart`) builds its own edge pass from `valueFrom` references and checks it for cycles; the inferred sources (namespace placement, connection placement, literal siblings, operator prerequisites) are not part of a chart's validity and are not run there. A chart's order at deploy time comes from this package through the set lane, and that order is what the corpus pins.
