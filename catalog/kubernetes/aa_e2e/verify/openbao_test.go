package verify

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func openBaoSpecFromYAML(t *testing.T, doc string) map[string]interface{} {
	t.Helper()
	var raw map[string]interface{}
	if err := yaml.Unmarshal([]byte(doc), &raw); err != nil {
		t.Fatalf("parsing spec: %v", err)
	}
	spec, _ := raw["spec"].(map[string]interface{})
	return spec
}

// The bootstrap follows the declared seal: an auto_unseal arm (any of the
// four, in either protojson spelling) switches init to recovery shares and
// drops the unseal step; a Shamir manifest declares none.
func TestOpenBaoAutoUnsealDetection(t *testing.T) {
	cases := []struct {
		name string
		doc  string
		want bool
	}{
		{"shamir (no seal block)", "spec:\n  server:\n    ha:\n      replicas: 1\n", false},
		{"transit, camelCase", "spec:\n  autoUnseal:\n    transit:\n      address: http://kms:8200\n      keyName: k\n      token: root\n", true},
		{"gcp kms, snake_case", "spec:\n  auto_unseal:\n    gcp_kms:\n      project:\n        value: p\n", true},
		{"an empty seal block declares nothing", "spec:\n  autoUnseal: {}\n", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := openBaoAutoUnseal(openBaoSpecFromYAML(t, tc.doc)); got != tc.want {
				t.Fatalf("openBaoAutoUnseal = %v, want %v", got, tc.want)
			}
		})
	}
}

// The restore proof places the target's initial root token exactly where
// the spec says the Job will look; both spellings of the field are read,
// and a manifest without a restore block yields nothing.
func TestOpenBaoRestoreRootToken(t *testing.T) {
	cases := []struct {
		name          string
		doc           string
		wantName      string
		wantKey       string
		wantNoRestore bool
	}{
		{"camelCase", "spec:\n  restore:\n    latest: true\n    rootToken:\n      name: init-root\n      key: token\n", "init-root", "token", false},
		{"snake_case", "spec:\n  restore:\n    snapshot_key: a/b.snap\n    root_token:\n      name: s\n      key: k\n", "s", "k", false},
		{"no restore", "spec:\n  backup:\n    retentionDays: 1\n", "", "", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			name, key := openBaoRestoreRootToken(openBaoSpecFromYAML(t, tc.doc))
			if name != tc.wantName || key != tc.wantKey {
				t.Fatalf("openBaoRestoreRootToken = (%q, %q), want (%q, %q)", name, key, tc.wantName, tc.wantKey)
			}
		})
	}
}

// The mode and replica count drive which lifecycle arms run; the keys are
// case-neutral and absent server blocks mean the chart default.
func TestOpenBaoScenarioShape(t *testing.T) {
	cases := []struct {
		name         string
		doc          string
		wantMode     string
		wantReplicas int
	}{
		{"absent = standalone", "spec: {}\n", "standalone", 1},
		{"dev", "spec:\n  server:\n    dev: {}\n", "dev", 1},
		{"ha default replicas", "spec:\n  server:\n    ha: {}\n", "ha", 3},
		{"ha one replica", "spec:\n  server:\n    ha:\n      replicas: 1\n", "ha", 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mode, replicas := openBaoScenarioShape(openBaoSpecFromYAML(t, tc.doc))
			if mode != tc.wantMode || replicas != tc.wantReplicas {
				t.Fatalf("openBaoScenarioShape = (%q, %d), want (%q, %d)", mode, replicas, tc.wantMode, tc.wantReplicas)
			}
		})
	}
}

// An rclone listing is read line by line: only `.snap` names count as
// snapshots, and a seeded stale object is found by its exact name.
func TestOpenBaoListingHelpers(t *testing.T) {
	listing := "a-20260101T000000Z.snap\nnotes.txt\n\na-20010101T000000Z.snap\n"
	if got := snapshotObjects(listing); len(got) != 2 {
		t.Fatalf("snapshotObjects found %d objects, want 2: %v", len(got), got)
	}
	if !containsLine(listing, "a-20010101T000000Z.snap") {
		t.Fatalf("containsLine missed an exact line")
	}
	if containsLine(listing, "a-2001") {
		t.Fatalf("containsLine matched a partial line")
	}
}
