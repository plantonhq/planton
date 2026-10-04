// Package catalogbundle defines the catalog bundle: the catalog shipped as
// DATA. A bundle carries everything a runtime needs to serve every kind --
// schemas with their validation rules, kind metadata, conversion specs, and
// presets -- so a new kind or kind version reaches a deployment by loading a
// new bundle, never by rebuilding a binary.
//
// Layout inside the zip:
//
//	manifest.yaml         format version, build info, per-entry sha256s
//	descriptors.binpb     FileDescriptorSet of the whole proto module --
//	                      schemas, buf.validate rules, and the kind registry
//	                      (enum options ARE the kind metadata) in one artifact
//	conversions/**        every authored ConversionSpec, provider/kind layout
//	presets/**            every kind's preset manifests, provider/kind layout
//	entries/**            every user-facing kind's catalog entry (title,
//	                      description, slug, logo, contract links, official
//	                      IaC module directories, and -- for covered
//	                      kinds -- fact-sheet summaries), provider/kind
//	                      layout
//	costs/**              covered kinds' cost profiles, provider/kind
//	                      layout, byte-identical to the tree's cost.yaml
//	controls/**           covered kinds' control profiles (controls.yaml)
//	permissions/**        covered kinds' permission manifests
//	                      (iac/permissions.yaml)
//	estimates/**          covered kinds' generated per-preset cost
//	                      estimates, provider/kind layout
//	derivations/**        derived kinds' machine-executable cost
//	                      derivations (value-to-quantity rules a server-side
//	                      estimator evaluates against live manifests),
//	                      provider/kind layout
//	compliance/**         the central control catalog and the framework
//	                      crosswalks the control profiles are read against
//	pricebooks/**         the pinned per-provider price books the estimates
//	                      were generated from and the derivations price by
//
// Fact-sheet coverage is presence-based: a kind without the cost/
// controls/permissions sidecars ships no cargo and its entry carries no
// summaries -- absence means "not yet covered", never "free".
//
// The manifest's checksums make the bundle self-verifying; release signing
// wraps the whole zip (the signature travels beside the artifact, not in
// it).
//
// Consumers load through Load in this package (Go) or the platform's
// equivalent loader (Java); the conformance check in this package is the
// gate that a bundle serves EXACTLY what the compiled-in registry serves.
package catalogbundle

// FormatVersion identifies the bundle layout. Consumers refuse versions they
// do not understand -- loudly, with the version named.
const FormatVersion = "2"

// Manifest is the bundle's self-description (manifest.yaml).
type Manifest struct {
	// FormatVersion of the bundle layout.
	FormatVersion string `json:"formatVersion"`
	// ReleaseTag that built the bundle (empty for local builds).
	ReleaseTag string `json:"releaseTag,omitempty"`
	// Checksums maps every entry path in the zip (except manifest.yaml
	// itself) to its sha256, hex-encoded.
	Checksums map[string]string `json:"checksums"`
}
