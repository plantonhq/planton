package provider

import "context"

// IdentityProjectAnnotation names, on a scenario, the project (or the
// provider's equivalent scope) its declared identity is granted in, when the
// kind under test targets one other than the harness's own: a policy that
// lives in a second test project is deployed by an identity granted there,
// exactly as a customer's connection is granted on the project the
// deployment targets. A kind that acts on two projects at once (a Shared VPC
// attachment joins a host and a service project) names both, comma-separated,
// and is granted in each. The value is a manifest value, so tokens such as
// ${E2E_ENV:NAME} expand in it; without it the identity is granted in the
// harness's own project. Harness vocabulary like the identity spec itself:
// the runner never reads it.
const IdentityProjectAnnotation = "planton.dev/e2e-identity-project"

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
// console deploy uses. The engine subprocesses that deploy, re-plan and
// destroy the kind also get the identity's credential as environment
// variables, after everything they inherit: a Terraform run reads them, and a
// Pulumi program's default provider would otherwise fall back to the
// harness's own credential. The test process's environment never changes,
// and the fixture chain keeps the harness's own posture.
type IdentityProvisioner interface {
	// ProvisionIdentity creates the identity the spec describes for this lane
	// and returns the path of a provider configuration file that authenticates
	// as it, plus the cleanup that removes the identity after the lane.
	ProvisionIdentity(ctx context.Context, tc *KindTestContext, spec string) (providerConfigPath string, cleanup func(), err error)
}

// GrantPropagationJudge is the optional capability that tells a grant still
// on its way from a missing one. Some services check permissions through
// their own caches, which can lag the provider's IAM by minutes after an
// identity is created, so a lane can be refused a permission its identity
// already holds. The runner retries a declared lane's deploy, after a wait,
// while the harness says the refusal names a held permission; any other
// failure fails the phase at once, so a permission the identity lacks is
// still a manifest finding.
type GrantPropagationJudge interface {
	// HeldPermissionRefused returns the permission a deploy failure names
	// and whether the lane's identity was granted it.
	HeldPermissionRefused(tc *KindTestContext, deployErr error) (permission string, held bool)
}
