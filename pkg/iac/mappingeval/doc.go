// Package mappingeval is the machine-scoring harness for import mapping:
// given a known ground truth (kind manifests deployed into a real
// account) and a proposer's ImportMappingProposal (produced blind, from a
// read-only scan of that account), it grades the proposal on the three axes
// that define mapping quality:
//
//   - grouping: did each discovered provider resource land in the right
//     proposed component instance?
//   - spec: did the proposed manifests reconstruct the settings the
//     ground-truth manifests declare?
//   - refs: did the proposal wire value_from references where the ground
//     truth has them, instead of freezing literals?
//
// The harness exists so mapping quality is MEASURED, never hoped for: any
// proposer -- the deterministic baseline in this package's baseline
// subpackage today, an AI mapping agent later -- takes the same exam and
// gets the same impartial grade. The grader itself is deliberately free of
// judgment: every comparison is structural, driven by the kinds' own proto
// schemas and the shared value_from encoding, so it works for any kind
// on any provider with zero per-kind grading code.
//
// Honesty rules the scorer enforces (they keep scores trustworthy on a
// shared, messy account):
//
//   - Scoring is defined over the ground-truth universe only. A scan of a
//     real region also sees AWS-implicit resources, other tests' leftovers,
//     and unrelated infrastructure; proposals about those are reported for
//     information but never rewarded or penalized -- no answer key exists
//     for them.
//   - Resources whose type is not scan-discoverable (no
//     cloud_control_type_name in the provider import catalog) are
//     structurally invisible: they are excluded from the grouping
//     denominator, because no proposer can claim what the scan cannot show
//     it.
//   - Instances are matched by kind plus claim overlap, never by name: a
//     blind proposer cannot know the seeded fixtures' internal names.
//   - Spec fields declared config-only in the provider import catalog
//     (values that exist only in IaC configuration, never cloud-side) are
//     excluded from the spec axis -- expecting them would penalize every
//     proposer for physics.
package mappingeval
