package runner

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func writeScenario(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "scenario.yaml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// A scenario document with a fixed name sends its kind to run alone, naming
// the file; a tokenized one does not.
func TestLaneAdmissionForFixedAndTokenizedNames(t *testing.T) {
	fixed := writeScenario(t, "apiVersion: gcp.planton.dev/v1\nkind: GcpNoSuchKind\nmetadata:\n  name: fixed\nspec:\n  name: planton-oss-e2e-fixed\n")
	admission, err := LaneAdmissionFor(t.TempDir(), "gcp", "gcpnosuchkind", []string{fixed})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(admission.Alone, "GcpNoSuchKind") {
		t.Fatalf("Alone = %q, want the fixed document named", admission.Alone)
	}

	tokenized := writeScenario(t, "apiVersion: gcp.planton.dev/v1\nkind: GcpNoSuchKind\nmetadata:\n  name: x\n  annotations:\n    planton.dev/e2e-exclusive: rag-engine-config\nspec:\n  name: planton-oss-e2e-${E2E_RUN_ID}\n")
	admission, err = LaneAdmissionFor(t.TempDir(), "gcp", "gcpnosuchkind", []string{tokenized})
	if err != nil {
		t.Fatal(err)
	}
	if admission.Alone != "" || len(admission.Classes) != 1 || admission.Classes[0] != "rag-engine-config" {
		t.Fatalf("admission = %+v, want side by side holding rag-engine-config", admission)
	}
}

// A document that names nothing, and says so with a reason, does not send its
// kind to run alone; the marker with no reason does not count.
func TestLaneAdmissionForADocumentThatNamesNothing(t *testing.T) {
	marked := writeScenario(t, "apiVersion: gcp.planton.dev/v1\nkind: GcpNoSuchKind\nmetadata:\n  name: x\n  annotations:\n    planton.dev/e2e-names-nothing: \"a peering on the run-scoped network\"\nspec:\n  network: ref\n")
	admission, err := LaneAdmissionFor(t.TempDir(), "gcp", "gcpnosuchkind", []string{marked})
	if err != nil {
		t.Fatal(err)
	}
	if admission.Alone != "" {
		t.Fatalf("Alone = %q, want side by side for a document that names nothing", admission.Alone)
	}

	unexplained := writeScenario(t, "apiVersion: gcp.planton.dev/v1\nkind: GcpNoSuchKind\nmetadata:\n  name: x\n  annotations:\n    planton.dev/e2e-names-nothing: \"\"\nspec:\n  network: ref\n")
	admission, err = LaneAdmissionFor(t.TempDir(), "gcp", "gcpnosuchkind", []string{unexplained})
	if err != nil {
		t.Fatal(err)
	}
	if admission.Alone == "" {
		t.Fatal("a names-nothing marker without a reason let the kind run side by side")
	}
}

// The check reads the real prerequisite chain: a catalog scenario whose own
// documents name fixed resources runs alone.
func TestLaneAdmissionForAFixedCatalogScenario(t *testing.T) {
	if os.Getenv("TEST_WORKSPACE") != "" {
		t.Skip("repo-reading test; skipped under Bazel")
	}
	root, err := filepath.Abs(filepath.Join("..", "..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	scenario := filepath.Join(root, "catalog", "gcp", "gcpnetworkfirewallpolicy", "e2e", "scenarios", "regional.yaml")
	raw, err := os.ReadFile(scenario)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), RunIDToken) {
		t.Skip("the scenario is tokenized now; this test pins the fixed-name case and needs another fixed scenario")
	}
	admission, err := LaneAdmissionFor(root, "gcp", "gcpnetworkfirewallpolicy", []string{scenario})
	if err != nil {
		t.Fatal(err)
	}
	if admission.Alone == "" {
		t.Fatal("a scenario with fixed names was admitted side by side")
	}
}

// The scheduler holds side-by-side kinds to the bound, runs an alone kind
// with nothing else running, and runs two kinds of one class one at a time.
func TestLaneSchedulerBoundsAloneAndClasses(t *testing.T) {
	scheduler := NewLaneScheduler(2)
	var running, peak, classRunning, classPeak int32
	var aloneSawOthers atomic.Bool
	var mu sync.Mutex
	track := func(counter, peakOf *int32, delta int32) {
		mu.Lock()
		defer mu.Unlock()
		*counter += delta
		if *counter > *peakOf {
			*peakOf = *counter
		}
	}
	t.Run("group", func(t *testing.T) {
		for i := 0; i < 4; i++ {
			t.Run("side-by-side", func(t *testing.T) {
				release := scheduler.Admit(t, LaneAdmission{})
				defer release()
				track(&running, &peak, 1)
				time.Sleep(30 * time.Millisecond)
				track(&running, &peak, -1)
			})
		}
		for i := 0; i < 2; i++ {
			t.Run("one-class", func(t *testing.T) {
				release := scheduler.Admit(t, LaneAdmission{Classes: []string{"singleton"}})
				defer release()
				track(&classRunning, &classPeak, 1)
				time.Sleep(30 * time.Millisecond)
				track(&classRunning, &classPeak, -1)
			})
		}
		t.Run("alone", func(t *testing.T) {
			release := scheduler.Admit(t, LaneAdmission{Alone: "a fixed name"})
			defer release()
			mu.Lock()
			if running != 0 || classRunning != 0 {
				aloneSawOthers.Store(true)
			}
			mu.Unlock()
			time.Sleep(30 * time.Millisecond)
		})
	})
	if peak > 2 {
		t.Fatalf("%d kinds ran at once under a bound of 2", peak)
	}
	if classPeak > 1 {
		t.Fatalf("%d kinds of one class ran at once", classPeak)
	}
	if aloneSawOthers.Load() {
		t.Fatal("the alone kind ran while others were running")
	}
}

// A kind waiting on a class holds no slot: with one kind running its class
// and others queued behind it, a kind outside the class is admitted at once
// rather than after the queue.
func TestLaneSchedulerClassWaitersHoldNoSlot(t *testing.T) {
	scheduler := NewLaneScheduler(2)
	const classHold = 400 * time.Millisecond
	var freeWaited time.Duration
	t.Run("group", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			t.Run("one-class", func(t *testing.T) {
				release := scheduler.Admit(t, LaneAdmission{Classes: []string{"project-vcpus"}})
				defer release()
				time.Sleep(classHold)
			})
		}
		t.Run("free", func(t *testing.T) {
			// Arrive after the class kinds have queued, so every free slot a
			// waiting class kind could take is already taken.
			time.Sleep(50 * time.Millisecond)
			asked := time.Now()
			release := scheduler.Admit(t, LaneAdmission{})
			defer release()
			freeWaited = time.Since(asked)
		})
	})
	if freeWaited >= classHold/2 {
		t.Fatalf("the unclassed kind waited %s for a slot while kinds queued on a class held the free ones", freeWaited)
	}
}

// LaneRunID keeps the run id's length and gives every kind its own value.
func TestLaneRunID(t *testing.T) {
	a, b := LaneRunID("0a1b2c3d", "gcptagvalue"), LaneRunID("0a1b2c3d", "gcptagbinding")
	if len(a) != 8 || len(b) != 8 {
		t.Fatalf("lengths %d, %d, want 8", len(a), len(b))
	}
	if a == b {
		t.Fatal("two kinds share a lane run id")
	}
	if a != LaneRunID("0a1b2c3d", "gcptagvalue") {
		t.Fatal("LaneRunID is not deterministic")
	}
}
