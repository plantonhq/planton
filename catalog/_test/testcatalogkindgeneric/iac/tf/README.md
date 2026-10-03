# TestCatalogKindGeneric Terraform Module

Hermetic module for exercising the proto→tfvars→plan→outputs pipeline with
zero cloud dependencies. Manages a single builtin `terraform_data` resource:
its `input` carries the full metadata + spec, so a change to any field class
produces a visible plan diff, and replacement semantics give the kind a real
create/update/destroy lifecycle without provisioning anything.

Variable shapes mirror `spec.proto` exactly (see `variables.tf`); outputs
mirror `outputs.proto` exactly (see `outputs.tf`). There is no
provider block on purpose — nothing outside the engine's own state is ever
touched, so `init`/`plan`/`apply` run offline.
