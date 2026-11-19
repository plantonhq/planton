package aa_e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/plantonhq/planton/e2e/framework/provider"
)

// The rule reads only Google's own record: a name the first act created and
// the upgrade deleted or created again is a replacement; every other write is
// a change in place, whatever the method is called.
func TestUpgradeVerdictReadsTheAuditLog(t *testing.T) {
	t0 := time.Date(2026, 11, 19, 10, 0, 0, 0, time.UTC)
	firstActDone, upgradeDone := t0.Add(5*time.Minute), t0.Add(10*time.Minute)
	network := "projects/p/global/networks/n"
	policy := "projects/p/global/securityPolicies/sp"
	entries := []auditEntry{
		{t0.Add(time.Minute), "v1.compute.securityPolicies.insert", policy},
		{t0.Add(2 * time.Minute), "v1.compute.securityPolicies.addRule", policy},
		{t0.Add(6 * time.Minute), "v1.compute.securityPolicies.patch", policy},
		{t0.Add(7 * time.Minute), "v1.compute.networks.updatePeering", network},
		{t0.Add(11 * time.Minute), "v1.compute.securityPolicies.delete", policy}, // the destroy, after the upgrade
	}
	created, upgrade := splitAuditWindows(entries, firstActDone, upgradeDone)
	if !created[policy] || len(created) != 1 || len(upgrade) != 2 {
		t.Fatalf("created %v, upgrade %v", created, upgrade)
	}
	if verdict, evidence := upgradeVerdict(created, upgrade); verdict != provider.UpgradeInPlace || !strings.Contains(evidence, "patch") {
		t.Fatalf("a patch and a peering update: verdict %q (%s), want in place", verdict, evidence)
	}

	replacedEntries := append(entries[:2:2],
		auditEntry{t0.Add(6 * time.Minute), "v1.compute.securityPolicies.delete", policy},
		auditEntry{t0.Add(7 * time.Minute), "v1.compute.securityPolicies.insert", policy},
	)
	created, upgrade = splitAuditWindows(replacedEntries, firstActDone, upgradeDone)
	if verdict, evidence := upgradeVerdict(created, upgrade); verdict != provider.UpgradeReplaced || !strings.Contains(evidence, "delete") {
		t.Fatalf("a delete and an insert of the created policy: verdict %q (%s), want replaced", verdict, evidence)
	}

	discovery := []auditEntry{
		{t0.Add(time.Minute), "google.cloud.discoveryengine.v1.EngineService.CreateEngine", "projects/1/locations/global/collections/default_collection/engines/e"},
		{t0.Add(6 * time.Minute), "google.cloud.discoveryengine.v1.EngineService.UpdateEngine", "projects/1/locations/global/collections/default_collection/engines/e"},
	}
	created, upgrade = splitAuditWindows(discovery, firstActDone, upgradeDone)
	if verdict, _ := upgradeVerdict(created, upgrade); verdict != provider.UpgradeInPlace {
		t.Fatalf("an UpdateEngine: verdict %q, want in place", verdict)
	}
}

// A declared lane's second act must say what the upgrade does; the harness
// refuses it before minting anything otherwise.
func TestUpgradeExpectationIsRequired(t *testing.T) {
	write := func(annotation string) string {
		body := "apiVersion: gcp.planton.dev/v1alpha1\nkind: GcpTagKey\nmetadata:\n  name: x\n"
		if annotation != "" {
			body += "  annotations:\n    " + provider.ExpectUpgradeAnnotation + ": " + annotation + "\n"
		}
		path := filepath.Join(t.TempDir(), "second.yaml")
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	if err := requireUpgradeExpectation(write(provider.UpgradeInPlace)); err != nil {
		t.Fatalf("in-place: %v", err)
	}
	if err := requireUpgradeExpectation(write("")); err == nil {
		t.Fatal("a second act with no expectation must be refused")
	}
	if err := requireUpgradeExpectation(write("sideways")); err == nil {
		t.Fatal("an unknown expectation must be refused")
	}
}
