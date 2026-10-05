package mappingeval

import "strings"

// The platform's seeding fingerprints: the tags Planton's own IaC modules
// write onto every resource they deploy. On a SEEDED exam account these
// tags ARE the answer key (planton.ai/resource-kind literally names the
// kind; planton.ai/environment names the environment), so a
// proposer that reads them is not being graded on mapping judgment at all.
const (
	seedTagPrefix = "planton.ai/"
	// The e2e fixture-marker labels ride manifests as metadata.labels and
	// become cloud tags. managed-by is stripped ONLY when it carries the
	// e2e marker value -- managed-by itself is a realistic tag key real
	// accounts use (managed-by: terraform), and scrubbing it wholesale
	// would over-sanitize the exam.
	seedManagedByValue = "planton-e2e"
	seedKindLabelKey   = "e2e-catalog-kind"
)

// RedactSeedFingerprints removes the platform's own seeding fingerprints
// from every scanned resource's Tags property, so a seeded exam account
// presents to the proposer exactly as a stranger's account would: Name
// tags and realistic user tags stay (real accounts have those, and they
// are legitimate mapping signals); only the deploy machinery's identity
// tags leave. The tags remain ON the provider resources (fixture sweeps key
// off them by convention) -- this redacts the proposer's VIEW, nothing
// else. It is deliberately NOT a general tag scrubber: the strip list is
// exactly the fingerprints Planton's seeding writes, and must never grow
// signals a genuinely foreign account could carry.
//
// Both eval lanes (offline fixture tests and the live e2e lane) apply this
// between scan and proposer, so the exam's honesty is a property of the
// pipeline rather than of any recorded fixture.
func RedactSeedFingerprints(scan *Scan) {
	for _, resource := range scan.Resources {
		tags, ok := resource.Properties["Tags"].([]any)
		if !ok {
			continue
		}
		kept := tags[:0]
		for _, raw := range tags {
			if isSeedFingerprintTag(raw) {
				continue
			}
			kept = append(kept, raw)
		}
		resource.Properties["Tags"] = kept
	}
}

// isSeedFingerprintTag reports whether one {Key, Value} tag entry is a
// seeding fingerprint. Entries that do not parse as tag objects are kept
// -- redaction must never eat data it does not understand.
func isSeedFingerprintTag(raw any) bool {
	tag, ok := raw.(map[string]any)
	if !ok {
		return false
	}
	key, _ := tag["Key"].(string)
	value, _ := tag["Value"].(string)
	switch {
	case strings.HasPrefix(key, seedTagPrefix):
		return true
	case key == seedKindLabelKey:
		return true
	case key == "managed-by" && value == seedManagedByValue:
		return true
	}
	return false
}
