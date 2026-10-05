# TestCatalogKindGeneric

Permanent, hermetic test infrastructure. This kind exists so the machinery
that carries every real kind — manifest loading, validation, defaults,
value-or-reference resolution, iac-input assembly, dual-engine IaC
execution, and output capture — can be exercised end to end with zero
cloud credentials, zero network, and zero cost.

It is **never shipped to users**: release content packaging, the catalog
site, and module auto-tagging all exclude the `_test` provider explicitly,
and the exclusion is guarded in CI. What ships to certification and test
surfaces is exactly what those surfaces need — the registry entry, generated
stubs, and these modules.

## Shape

The spec deliberately carries one of every generic field class (scalars with
and without defaults, a nested message, value-or-reference strings including
a fully annotated foreign-key fixture, maps, repeated fields, and sensitive
fields), so a machinery change that mishandles any class fails a test here
before it can touch a real kind.

Both IaC engines are implemented against a no-op target: Terraform manages a
single builtin `terraform_data` resource (a real lifecycle with no cloud
behind it), and the Pulumi module exports outputs derived deterministically
from inputs. Output names mirror `outputs.proto` exactly.

## Kubernetes-specific testing

For Kubernetes-specific field shapes (container resources, env, ports), use
the sibling `TestCatalogKindKubernetes` instead.

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
