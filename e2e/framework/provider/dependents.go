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

// DestroyPreparer takes, before the kind under test is destroyed, a step the
// cloud requires of the resource's owners that neither engine's destroy
// takes: Google deletes a VPC peering in CONSENSUS mode only after the owners
// of both networks have requested its removal, and neither provider makes
// that request. The kind's GUIDE teaches the customer the step; the harness
// takes it in their place, so the lane proves the GUIDE's teardown.
//
// Activated by the scenario annotation `planton.dev/e2e-prepare-destroy:
// <class>`, whose class the harness names and interprets (an unknown class
// is the harness's error, so a typo never passes silently). The runner calls
// it before every destroy attempt of the kind under test -- DESTROY, the
// cleanup destroy after a failed phase, DESTROY-AGAIN, and each scenario-
// declared retry -- with ManifestPathKey on ctx naming the resolved first
// act, so it works when the deploy failed before any output existed.
// Implementations are idempotent: a request already made, or a resource
// already gone, is success.
type DestroyPreparer interface {
	PrepareDestroy(ctx context.Context, tc *KindTestContext, class string) error
}

// DependencyTeardownPreparer removes what a cloud leaves holding a fixture
// after the objects that made it are gone, which the fixture's own destroy
// cannot remove: deleting a Managed Kafka Connect cluster leaves the network
// attachment Google made on its subnetwork, and the subnetwork then refuses
// deletion while the attachment exists. The runner calls it before every
// destroy attempt of each dependency, with the dependency's kind and its
// captured outputs. Implementations remove only objects nothing serves any
// more (an attachment with no connection endpoints), never one in use, and
// return an error only when such a removal fails; the destroy is attempted
// either way.
type DependencyTeardownPreparer interface {
	PrepareDependencyDestroy(ctx context.Context, kindSlug string, outputs map[string]interface{}) error
}
