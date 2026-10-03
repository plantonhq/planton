// Package setdeploy deploys a SET of catalog object manifests as one
// operation: the preflight wall, the dependency-ordered execution loop, and
// the output-fed reference resolution between nodes.
//
// The package composes two layers that already exist and adds no second
// opinion on either: pkg/manifestgraph answers what depends on what, what
// resolves inside the set, and what points out of it (the one rule set every
// lane shares); the engine's single-manifest run paths (tofumodule,
// pulumistack) answer how one resource deploys. setdeploy's own job is the
// part neither owns — verifying EVERYTHING verifiable before the first IaC
// handoff and walking the order with each node's captured outputs feeding the
// next nodes' references.
//
// The preflight wall is the product surface here: every check the deploy
// environment can fail is probed up front — schema, identity, references,
// backend-resolved value refusal, cycles, binaries, modules, state backend,
// provider credentials — and ALL failures land in one report, never
// fail-at-first. A failure the wall could have caught surfacing as an IaC
// stack trace twenty minutes into an apply is the defect class this package
// exists to end. Probes are injected (see Probes) so the wall's logic is
// exhaustively testable without a cloud.
//
// The library never prints and never exits. Preflight returns a structured
// Report whose severity vocabulary is the caller's to render; Execute streams
// progress through the Events seam. Each consumer (the engine CLI today, the
// platform CLI's offline arm as a library consumer) renders its own surface —
// a library that hardcoded printing would force one product's voice onto the
// other.
//
// Execution isolation: every tofu/terraform node runs in its own stable,
// identity-keyed workspace (a per-node copy of the module under the CLI's
// workspace root). Module cache directories are shared per kind — running
// nodes in them directly would collide same-kind state files and leak one
// node's backend config into the next. The identity-keyed copy makes local
// backend state safe (it persists per node across runs, so re-running the
// same command is the recovery story) and gives remote-backend nodes a clean
// backend.tf of their own. Pulumi nodes need no copy: their state lives
// behind the backend URL, never in the working directory.
package setdeploy
