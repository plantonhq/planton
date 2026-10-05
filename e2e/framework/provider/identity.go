package provider

import "context"

// IdentityProvisioner is the optional harness capability behind the
// planton.dev/e2e-identity scenario annotation: the lane deploys as an
// identity the harness creates for it, instead of the harness's own
// (administrative) credentials. The spec is provider-interpreted, exactly as
// the failure-mode annotations' values are; the Kubernetes harness reads
// "declared" (a ServiceAccount bound to the rules the kind's
// iac/permissions.yaml declares) and "declared-minus:<group>/<resource>:<verbs>"
// (the same, with named verbs withheld), which is how a lane proves that a
// module's least-privilege claim holds, or that a missing right is refused
// with the right named.
//
// The identity reaches the engines the way a real deploy's connection does:
// as a provider configuration file the runner binds instead of the kind's
// fixture, so both engines receive it through the same iac-input path a
// console deploy uses. Nothing about the process environment changes; the
// fixture chain keeps the harness's own posture.
type IdentityProvisioner interface {
	// ProvisionIdentity creates the identity the spec describes for this lane
	// and returns the path of a provider configuration file that authenticates
	// as it, plus the cleanup that removes the identity after the lane.
	ProvisionIdentity(ctx context.Context, tc *KindTestContext, spec string) (providerConfigPath string, cleanup func(), err error)
}
