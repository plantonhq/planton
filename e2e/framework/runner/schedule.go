package runner

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/pkg/errors"
)

// ExclusiveAnnotation serializes a scenario's kind against every other kind
// that names the same class, while lanes otherwise run side by side. The
// value is a class name the scenario chooses: a singleton the cloud allows
// once per project (a location's RAG Engine configuration, a project's Model
// Armor floor), or a quota too small to share (a project's few vCPUs). Two
// kinds naming one class never run at once; unrelated kinds still do.
const ExclusiveAnnotation = "planton.dev/e2e-exclusive"

// LaneAdmission is how a kind's lanes may run beside other kinds' in one test
// process: alone (nothing else running), or side by side under the bound,
// holding the exclusive classes its scenarios name.
type LaneAdmission struct {
	// Alone, when set, says why the kind must run with nothing else running.
	Alone   string
	Classes []string
}

// LaneAdmissionFor decides a kind's admission from its files. A kind runs
// alone when any file its lanes deploy -- a scenario, or a prerequisite in a
// scenario's chain -- holds a document with no ${E2E_RUN_ID} in it: such a
// document builds a name that does not change from lane to lane, so two
// lanes running at once would build the same thing. A document that is
// tokenized gets a value unique to its kind and engine (LaneRunID), however
// many kinds compose the file. The check is per document and textual: a
// document that tokenizes one name but fixes another still passes it, so a
// session that makes a kind's files run side by side reads them, rather than
// trusting the check alone.
func LaneAdmissionFor(repoRoot, kindProvider, kindDir string, scenarioPaths []string) (LaneAdmission, error) {
	var admission LaneAdmission
	classes := map[string]bool{}
	checked := map[string]bool{}
	for _, scenario := range scenarioPaths {
		class, err := ManifestAnnotation(scenario, ExclusiveAnnotation)
		if err != nil {
			return admission, err
		}
		if class != "" {
			classes[class] = true
		}
		files := []string{scenario}
		deps, err := ResolveDependencies(repoRoot, kindProvider, kindDir, scenario)
		if err != nil {
			return admission, errors.Wrapf(err, "resolving %s's prerequisites to decide whether it may run beside other kinds", scenario)
		}
		for _, dep := range deps {
			files = append(files, dep.ManifestPath)
		}
		for _, file := range files {
			if checked[file] || admission.Alone != "" {
				continue
			}
			checked[file] = true
			fixed, err := documentWithoutRunID(file)
			if err != nil {
				return admission, err
			}
			if fixed != "" {
				admission.Alone = fmt.Sprintf("%s holds a document with no %s (%s), so its names are the same in every lane", file, RunIDToken, fixed)
			}
		}
	}
	for class := range classes {
		admission.Classes = append(admission.Classes, class)
	}
	sort.Strings(admission.Classes)
	return admission, nil
}

// documentWithoutRunID returns the kind of the first document in the file
// that never mentions the run-id token, or "" when every document does.
func documentWithoutRunID(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", errors.Wrapf(err, "reading %s", path)
	}
	for _, document := range strings.Split(string(raw), "\n---") {
		kind := documentKind(document)
		if kind == "" {
			continue
		}
		if !strings.Contains(document, "${E2E_RUN_ID") {
			return kind, nil
		}
	}
	return "", nil
}

func documentKind(document string) string {
	for _, line := range strings.Split(document, "\n") {
		if strings.HasPrefix(line, "kind:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "kind:"))
		}
	}
	return ""
}

// LaneScheduler admits kinds' tests to run side by side in one test process:
// at most bound at once, a kind whose admission says Alone with nothing else
// running, and kinds that share an exclusive class one at a time. The one
// test process per provider account rule (the e2e README) still holds; this
// is concurrency inside that process.
type LaneScheduler struct {
	slots   chan struct{}
	whole   sync.RWMutex
	mu      sync.Mutex
	classes map[string]*sync.Mutex
}

// NewLaneScheduler bounds side-by-side kinds at bound (at least one).
func NewLaneScheduler(bound int) *LaneScheduler {
	if bound < 1 {
		bound = 1
	}
	return &LaneScheduler{slots: make(chan struct{}, bound), classes: map[string]*sync.Mutex{}}
}

// Admit marks the kind's test parallel, waits for its turn, and returns the
// release the test defers. Call it after every skip decision, so a skipped
// kind never holds a place.
func (s *LaneScheduler) Admit(t *testing.T, admission LaneAdmission) func() {
	t.Helper()
	t.Parallel()
	if admission.Alone != "" {
		t.Logf("runs alone: %s", admission.Alone)
		s.whole.Lock()
		return s.whole.Unlock
	}
	s.whole.RLock()
	s.slots <- struct{}{}
	locks := make([]*sync.Mutex, 0, len(admission.Classes))
	for _, class := range admission.Classes {
		s.mu.Lock()
		lock, ok := s.classes[class]
		if !ok {
			lock = &sync.Mutex{}
			s.classes[class] = lock
		}
		s.mu.Unlock()
		locks = append(locks, lock)
	}
	// Classes are sorted (LaneAdmissionFor), so every kind takes its locks in
	// one order and two kinds can never wait on each other.
	for _, lock := range locks {
		lock.Lock()
	}
	return func() {
		for i := len(locks) - 1; i >= 0; i-- {
			locks[i].Unlock()
		}
		<-s.slots
		s.whole.RUnlock()
	}
}
