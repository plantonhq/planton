// Releasing what the provider made on the lane's behalf -- an optional Harness
// extension the runner discovers by type assertion.

package provider

import "context"

// DependentsReleaser removes the objects a provider creates by itself when the
// kind under test is deployed, which outlive the kind's own destroy: a
// Dialogflow agent's data-store tool makes Google create chat engines that
// keep holding the store, and a KMS key handle makes Autokey create a key
// that keeps billing. The engines never knew those objects, so neither
// engine's destroy removes them, and left alone they either block the
// fixtures' teardown (a store an engine still uses refuses deletion) or stay
// behind as orphans.
//
// Activated by the scenario annotation `planton.dev/e2e-release-dependents:
// <class>`, whose class the harness names and interprets (an unknown class
// is the harness's error, so a typo never passes silently). The runner calls
// it once, after the kind's last destroy and before the fixtures are torn
// down, whatever the lane's verdict, with ManifestPathKey on ctx naming the
// first act's manifest. Implementations find the objects by what this lane
// deployed (its run-scoped names), never by listing everything of a type, so
// lanes running side by side never release each other's; they confirm what
// they release reads gone (or, for objects the provider never deletes,
// reads in its final state) before returning. Teaching the customer what the
// provider leaves behind is the kind's GUIDE's job; this only keeps the
// lane's account clean.
type DependentsReleaser interface {
	ReleaseDependents(ctx context.Context, tc *KindTestContext, class string) error
}
