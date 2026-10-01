// Out-of-band deletion -- an optional Harness extension the runner discovers by
// type assertion. It exists for the promise every GUIDE makes about drift:
// what happens to the next plan when someone deletes the object in the
// provider's own console, behind the engine's back, and how the declared file
// gets it back.

package provider

import "context"

// OutOfBandDeleter deletes the object a lane deployed, through the provider's
// own API rather than the engine, the way a person deleting it in a console
// would. Activated by the scenario annotation
// `planton.dev/e2e-out-of-band-delete`. It runs after VERIFY-RES, so stack
// outputs are on tc and the harness has already stored the deployed identity
// (via ManifestPathKey on ctx). Implementations must confirm the delete took
// (the object reads back absent) before returning: the act's later phases
// prove what the ENGINE does with a missing object, so a delete that silently
// failed would let a broken recovery pass.
type OutOfBandDeleter interface {
	DeleteOutOfBand(ctx context.Context, tc *ComponentTestContext) error
}
