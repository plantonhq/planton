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
		{"shamir (no seal block)", "spec:\n  server:\n    raft: {}\n    replicas: 1\n", false},
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

// Dev or the storage engine, and the replica count, drive which lifecycle
// arms run; the keys are case-neutral, an absent server block or an unset
// engine means Raft at one replica (the spec's defaults), and the
// PostgreSQL arm yields the cluster and the database the psql proof opens
// — from the reference as authored, or from the `<cluster>-rw` literal
// the harness resolves the reference into before the verifier reads it.
func TestOpenBaoScenarioShape(t *testing.T) {
	cases := []struct {
		name string
		doc  string
		want openBaoShape
	}{
		{"absent = raft at one", "spec: {}\n", openBaoShape{Storage: storageRaft, Replicas: 1}},
		{"dev", "spec:\n  server:\n    dev: {}\n", openBaoShape{Dev: true, Storage: storageRaft, Replicas: 1}},
		{"raft with a volume at three", "spec:\n  server:\n    raft:\n      dataStorage:\n        size: 1Gi\n    replicas: 3\n", openBaoShape{Storage: storageRaft, Replicas: 3}},
		{"unset engine with a count", "spec:\n  server:\n    replicas: 3\n", openBaoShape{Storage: storageRaft, Replicas: 3}},
		{"postgresql by reference, camelCase",
			"spec:\n  server:\n    postgresql:\n      host:\n        valueFrom:\n          name: e2e-bao-pg\n      database: openbao\n      username: openbao\n      passwordSecret:\n        secretName:\n          valueFrom:\n            name: e2e-bao-pg\n    replicas: 1\n",
			openBaoShape{Storage: storagePostgresql, Replicas: 1, PgCluster: "e2e-bao-pg", PgDatabase: "openbao"}},
		{"postgresql by reference, snake_case",
			"spec:\n  server:\n    postgresql:\n      host:\n        value_from:\n          name: pg\n      database: vault\n    replicas: 2\n",
			openBaoShape{Storage: storagePostgresql, Replicas: 2, PgCluster: "pg", PgDatabase: "vault"}},
		{"postgresql with the reference resolved to the rw Service (what the harness hands the verifier)",
			"spec:\n  server:\n    postgresql:\n      host:\n        value: e2e-bao-pg-rw\n      database: openbao\n      username: openbao\n      passwordSecret:\n        secretName:\n          value: e2e-bao-pg-app\n    replicas: 1\n",
			openBaoShape{Storage: storagePostgresql, Replicas: 1, PgCluster: "e2e-bao-pg", PgDatabase: "openbao"}},
		{"postgresql with a qualified rw Service host",
			"spec:\n  server:\n    postgresql:\n      host:\n        value: pg-rw.data.svc.cluster.local\n      database: vault\n",
			openBaoShape{Storage: storagePostgresql, Replicas: 1, PgCluster: "pg", PgDatabase: "vault"}},
		{"postgresql with a literal host that is no rw Service has no cluster to open",
			"spec:\n  server:\n    postgresql:\n      host:\n        value: db.example.internal\n      database: openbao\n",
			openBaoShape{Storage: storagePostgresql, Replicas: 1, PgDatabase: "openbao"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := openBaoScenarioShape(openBaoSpecFromYAML(t, tc.doc))
			if got != tc.want {
				t.Fatalf("openBaoScenarioShape = %+v, want %+v", got, tc.want)
			}
		})
	}
}

// A host names a CloudNativePG cluster only when it is that cluster's
// read-write Service, in any DNS form; anything else has no cluster.
func TestCnpgClusterFromRwHost(t *testing.T) {
	cases := map[string]string{
		"pg-rw":                        "pg",
		"e2e-bao-pg-rw":                "e2e-bao-pg",
		"pg-rw.data":                   "pg",
		"pg-rw.data.svc.cluster.local": "pg",
		"pg-ro":                        "",
		"pg":                           "",
		"-rw":                          "",
		"db.example.internal":          "",
		"":                             "",
	}
	for host, want := range cases {
		if got := cnpgClusterFromRwHost(host); got != want {
			t.Errorf("cnpgClusterFromRwHost(%q) = %q, want %q", host, got, want)
		}
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
