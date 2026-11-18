package runner

import (
	"strings"
	"testing"
)

// Short compositions pass through untouched -- the cap must not perturb the
// names every existing green lane already uses.
func TestGenerateStackNameShortPassthrough(t *testing.T) {
	got := GenerateStackName("dep-azureresourcegroup-rg", "0a1b2c3d")
	want := "e2e-dep-azureresourcegroup-rg-0a1b2c3d"
	if got != want {
		t.Fatalf("short name changed: got %q want %q", got, want)
	}
}

// The run id is kept whole: an engine-scoped id's engine suffix is what keeps
// a kind's Pulumi and Terraform lanes on separate fixture stacks.
func TestGenerateStackNameKeepsTheEngineSuffix(t *testing.T) {
	pulumi := GenerateStackName("dep-x", EngineScopedRunID("0a1b2c3d", "pulumi"))
	terraform := GenerateStackName("dep-x", EngineScopedRunID("0a1b2c3d", "terraform"))
	if pulumi != "e2e-dep-x-0a1b2c3d-p" || terraform != "e2e-dep-x-0a1b2c3d-t" {
		t.Fatalf("engine suffix lost: pulumi %q, terraform %q", pulumi, terraform)
	}
}

// Long compositions respect the cap.
func TestGenerateStackNameCapped(t *testing.T) {
	got := GenerateStackName(strings.Repeat("a", 80), "0a1b2c3d")
	if len(got) != maxStackNameLen {
		t.Fatalf("capped name length = %d, want %d (%q)", len(got), maxStackNameLen, got)
	}
}

// The regression that motivated the hash tail: an install profile's
// documents share a long common prefix and differ only in suffixes past the
// cap (manifest-name tails, per-document indexes). A blind prefix cut
// collapsed all four subnet fixture stacks into ONE shared stack, so each
// deploy silently replaced the previous document's resources. Distinct full
// compositions must yield distinct capped names.
func TestGenerateStackNameLongLabelsStayDistinct(t *testing.T) {
	runID := "0a1b2c3d"
	labels := []string{
		"dep-azuresubnet-planton-oss-e2e-azure-fixture-subnet",
		"dep-azuresubnet-planton-oss-e2e-azure-fixture-firewall-subnet-1",
		"dep-azuresubnet-planton-oss-e2e-azure-fixture-gateway-subnet-2",
		"dep-azuresubnet-planton-oss-e2e-azure-fixture-pls-subnet-3",
	}
	seen := map[string]string{}
	for _, label := range labels {
		name := GenerateStackName(label, runID)
		if len(name) > maxStackNameLen {
			t.Fatalf("name for %q exceeds cap: %q (%d chars)", label, name, len(name))
		}
		if prev, dup := seen[name]; dup {
			t.Fatalf("labels %q and %q collide on stack name %q", prev, label, name)
		}
		seen[name] = label
	}
}

// The derivation is deterministic: dependency stacks are keyed by run id so
// every scenario in a run reuses the same stack -- a capped name must come
// out identical on every call.
func TestGenerateStackNameDeterministic(t *testing.T) {
	label := strings.Repeat("dep-long-label-", 6)
	first := GenerateStackName(label, "0a1b2c3d")
	second := GenerateStackName(label, "0a1b2c3d")
	if first != second {
		t.Fatalf("derivation not deterministic: %q vs %q", first, second)
	}
}

// The two dependency labels that motivated boundStackName: a two-database
// Firestore chain whose labels agree on the first 50 characters. A plain
// head-truncate collapsed them onto one stack (the second up REPLACED the
// first fixture; teardown then failed "no stack named" on the survivor).
func TestBoundStackName_PreservesUniquenessUnderTruncation(t *testing.T) {
	a := boundStackName(GenerateStackName("dep-gcpfirestoredatabase-planton-oss-e2e-gcpfsdb-prereq", "run1234abcd"))
	b := boundStackName(GenerateStackName("dep-gcpfirestoredatabase-planton-oss-e2e-gcpfsdb-ent-idx", "run1234abcd"))

	if len(a) > maxStackNameLen || len(b) > maxStackNameLen {
		t.Fatalf("bounded names exceed %d chars: %q (%d), %q (%d)", maxStackNameLen, a, len(a), b, len(b))
	}
	if a == b {
		t.Fatalf("distinct labels collapsed onto one stack name: %q", a)
	}
	// Deterministic: up and destroy must compute the same name.
	if a != boundStackName(GenerateStackName("dep-gcpfirestoredatabase-planton-oss-e2e-gcpfsdb-prereq", "run1234abcd")) {
		t.Fatal("boundStackName is not deterministic")
	}
	// Distinct runs must still produce distinct names for the same label.
	if a == boundStackName(GenerateStackName("dep-gcpfirestoredatabase-planton-oss-e2e-gcpfsdb-prereq", "run5678wxyz")) {
		t.Fatal("run id no longer disambiguates after truncation")
	}
}

func TestBoundStackName_ShortNamesPassThrough(t *testing.T) {
	name := GenerateStackName("dep-gcpvpcnetwork-vpc", "run1234abcd")
	if got := boundStackName(name); got != name {
		t.Fatalf("short name modified: %q -> %q", name, got)
	}
	if strings.Contains(boundStackName(name), "--") {
		t.Fatalf("unexpected separator artifact in %q", boundStackName(name))
	}
}
