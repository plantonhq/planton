// Package providerparity measures the catalog against the Terraform
// provider it deploys through: for each cloud provider, what the pinned
// Terraform provider can configure versus what the catalog's kinds expose.
//
// Naming note -- this is PROVIDER parity, a different axis from the
// cross-engine parity the audit machinery already enforces:
//
//   - Cross-engine parity (pkg/iac/MODULE_PARITY.md, PARITY-EXCEPTION
//     comments, the audit rule's --parity focus): the Terraform and Pulumi
//     modules of ONE kind implement the same contract identically.
//   - Provider parity (this package): the catalog as a whole, kind by kind,
//     covers the full configurable surface of the pinned Terraform provider,
//     with every omission recorded as a decision.
//
// The measurement has three independent censuses, each mechanical:
//
//   - The provider side is self-describing: committed, distilled
//     `providers schema -json` artifacts per pinned provider version
//     (schemas/, produced by the distiller sub-package). No source parsing.
//   - The catalog's contract side is a proto field census from descriptors
//     via pkg/crkreflect (spec_census.go) -- the same registry walk as
//     pkg/secretcoverage; a reader who knows one walk knows both.
//   - The catalog's module side is a consumed-resource, provider-pin, and
//     provider-attachment census over every `*.tf` file of every kind's
//     Terraform module (module_census.go). Every file, never main.tf
//     alone: modules may split resources across sibling files, and a
//     main.tf-only scan undercounts silently. The attachment census (which
//     resource blocks set `provider = <name>`) is what makes a secondary
//     channel visible rather than inferred.
//
// Parity is always measured against a NAMED provider version -- the pin --
// never against "latest"; that is what makes a freshness promise
// well-defined.
//
// On top of the measurement sits the total-accounting layer (accounting.go):
// every configurable, non-deprecated argument of every consumed resource is
// exact-matched to a spec field, mapped by the kind's recorded judgment
// (manifest.go -- catalog/<provider>/<kind>/iac/provider-parity.yaml), or
// excluded there with a reason; in reverse, every spec leaf reaches provider
// surface or carries an exclusion (the reverse-drift check); and every GA
// resource carries exactly one breadth disposition (dispositions.go). The
// matcher carries zero name heuristics -- divergence is recorded, never
// guessed. Gaps gate through the burn-down baseline (baseline.go) in the
// pkg/anatomy / pkg/secretcoverage grain, surfaced by the `planton
// provider-parity` developer command and the lint.provider-parity CI lane.
//
// Capability a provider publishes only in a SECONDARY CHANNEL (Google's
// google-beta) enters through the admission list (admissions.go --
// admissions/<channel>.yaml: one entry per resource per consuming kind,
// with the reason and the promotion-tracking path). The baseline schema is
// always the yardstick for a resource it serves; a channel-only resource is
// accounted against the channel's schema exactly when it is admitted for
// the kind and the module attaches it through the channel's provider, and
// every other combination -- unadmitted consumption, an admission the
// module does not attach, an attachment on a GA-served resource, an
// admission the GA schema has since made unnecessary -- is a finding. The
// public page renders the list. hack/guards/ensure_beta_admissions.sh is
// the dependency-free half of the same gate, in the terraform-modules lane.
//
// The same total-accounting rule covers the PROVIDER BLOCK itself
// (provider_config_accounting.go): the provider's own configuration
// arguments -- credentials, role-assumption chains, default tags, endpoint
// overrides, retry tuning -- are accounted against the provider-config proto
// (catalog/<provider>/provider.proto, censused in
// provider_config_census.go) under a provider-level manifest
// (provider_config_manifest.go -- catalog/<provider>/provider-config-parity.yaml),
// and arguments set inside catalog modules' own provider blocks (found by
// the module census's HCL walk) must carry recorded judgment too. Enrollment
// is manifest presence; findings ride the same baseline under the
// "provider:<cloud>" key class.
//
// The accounting also renders the PUBLIC parity report (publicreport.go +
// e2eproof.go): a committed, drift-tested markdown page per provider at
// catalog/<provider>/terraform-parity.md, carried onto the docs site by the
// site build. Generated, never hand-authored -- the page and the CI gate
// render from the same accounting, so the published numbers can never
// disagree with the check.
package providerparity
