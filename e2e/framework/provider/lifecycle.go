package provider

// FirstActManifestPathKey is the context key that carries, during a second
// act (the upgrade a scenario declares with planton.dev/e2e-upgrade-manifest),
// the path of the FIRST act's manifest. ManifestPathKey then names the second
// manifest, so a verifier that stored what the first deploy created (keyed by
// the first manifest) can compare it with what the upgrade left: the same
// object changed in place, or a new one that replaced it.
type FirstActManifestPathKey struct{}

// SecondActManifestPathKey is the context key that carries, while the lane's
// identity is provisioned (IdentityProvisioner), the path of the scenario's
// second act, prepared the way the runner will deploy it. One identity
// deploys both acts, so a harness that grants what a manifest uses reads
// both: a feature only the upgrade turns on is the identity's to hold, never
// a refusal that would read as a gap in the kind's permissions. Absent when
// the scenario declares no second act.
type SecondActManifestPathKey struct{}

// ExpectUpgradeAnnotation, on a second-act manifest, declares what the upgrade
// does to the deployed object's identity: UpgradeInPlace (the same object,
// changed) or UpgradeReplaced (a new object, the old one gone the way the
// kind's destroy leaves it). The runner never reads it -- whether an object
// kept its identity is only knowable through the provider's own API -- so it
// is vocabulary for harnesses whose verifiers can tell, read from the second
// manifest on the context. A harness that reads it should refuse a second act
// that omits it rather than pass an upgrade it never judged.
const ExpectUpgradeAnnotation = "planton.dev/e2e-expect-upgrade"

// The values ExpectUpgradeAnnotation takes.
const (
	UpgradeInPlace  = "in-place"
	UpgradeReplaced = "replaced"
)
