package tekton

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"sigs.k8s.io/yaml"
)

// The ledger holds Version and the embedded tree together: every
// (Version, content digest) pair ever recorded, one per line, append-only.
// Two different content states must never share a pin, and one content
// state must never carry two -- the pin is what makes stamped provenance and
// byte-identical reruns honest.
//
// A single recorded digest could not hold that line: re-recording it after a
// content edit passed whether or not Version moved, so the bump was on the
// honour system (v7 named two content states that way). The ledger remembers
// what each Version already names, so re-recording cannot launder either
// mistake: a Version already bound to other content, or content already bound
// to another Version, is refused even under UPDATE_CONTENT_DIGEST.
func TestContent_versionNamesExactlyOneContentState(t *testing.T) {
	computed := computeDigest(t)
	ledgerPath := filepath.Join("testdata", "content-digest.txt")
	ledger := readContentLedger(t, ledgerPath)

	if recorded, ok := ledger.digestOf[Version]; ok {
		if recorded == computed {
			return
		}
		t.Fatalf("Version %s already names another content state (recorded %s), but the embedded content now hashes to %s.\n"+
			"Bump Version in content.go for this content change, then record the new pair with UPDATE_CONTENT_DIGEST=1 go test ./cicd/tekton/",
			Version, recorded, computed)
	}
	if earlier, ok := ledger.versionOf[computed]; ok {
		t.Fatalf("Version is %s, but this exact content was already recorded as %s: a bump with no content change would give one content state two pins.\n"+
			"Set Version back to %s, or make the content change the bump is for.", Version, earlier, earlier)
	}
	if os.Getenv("UPDATE_CONTENT_DIGEST") == "" {
		t.Fatalf("the content state %s %s is not in %s.\n"+
			"If you changed the embedded YAML and bumped Version for it, record the pair with UPDATE_CONTENT_DIGEST=1 go test ./cicd/tekton/",
			Version, computed, ledgerPath)
	}
	f, err := os.OpenFile(ledgerPath, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatalf("opening the content ledger to record %s: %v", Version, err)
	}
	defer f.Close()
	if _, err := f.WriteString(Version + " " + computed + "\n"); err != nil {
		t.Fatalf("recording %s in the content ledger: %v", Version, err)
	}
	t.Logf("recorded %s %s in %s -- add the line to reviewedLedger in content_test.go in the same change", Version, computed, ledgerPath)
}

// The build step waits for its BuildKit daemon long enough for a busy node:
// buildctl-daemonless.sh's default of 10 connection tries is about 2.2
// seconds, and a daemon still starting then fails the build at random. The
// value is read from the build-and-push step's own env, so a copy of the
// line elsewhere in the task cannot satisfy it.
func TestContent_theBuildKitStepWaitsForItsDaemonToStart(t *testing.T) {
	tasks, err := TaskFiles()
	if err != nil {
		t.Fatal(err)
	}
	var task struct {
		Spec struct {
			Steps []struct {
				Name string `json:"name"`
				Env  []struct {
					Name  string `json:"name"`
					Value string `json:"value"`
				} `json:"env"`
			} `json:"steps"`
		} `json:"spec"`
	}
	if err := yaml.Unmarshal(tasks["buildkit"], &task); err != nil {
		t.Fatalf("parsing the BuildKit task: %v", err)
	}
	for _, step := range task.Spec.Steps {
		if step.Name != "build-and-push" {
			continue
		}
		for _, env := range step.Env {
			if env.Name == "BUILDCTL_CONNECT_RETRIES_MAX" {
				if env.Value != "70" {
					t.Errorf("build-and-push sets BUILDCTL_CONNECT_RETRIES_MAX to %q, want \"70\" (about 57 seconds for the daemon to start)", env.Value)
				}
				return
			}
		}
		t.Fatal("build-and-push does not set BUILDCTL_CONNECT_RETRIES_MAX -- without it the build gives its daemon about 2.2 seconds to start")
	}
	t.Fatal("the BuildKit task has no build-and-push step")
}

func TestContent_tracksAndTasksArePresent(t *testing.T) {
	wantTracks := []string{"buildpacks", "dockerfile"}
	if got := Tracks(); !reflect.DeepEqual(got, wantTracks) {
		t.Fatalf("expected exactly the service build tracks %v, got %v", wantTracks, got)
	}
	tasks, err := TaskFiles()
	if err != nil {
		t.Fatal(err)
	}
	for _, stem := range []string{"git-clone", "kustomize-build", "buildkit", "buildpacks"} {
		if _, ok := tasks[stem]; !ok {
			t.Fatalf("embedded task %s.yaml is missing", stem)
		}
	}
	if _, ok := Track("no-such-track"); ok {
		t.Fatal("an unknown track must not resolve")
	}
}

// Every track builds for the deploy target's platform, not the build
// machine's: each declares the optional target-platform fact (the platform
// supplies it only to pipelines that declare it), and each build task
// reports the machine it ran on as the build-node-architecture result -- the
// two facts a run needs to say "built for linux/amd64 on an arm64 machine,
// emulated" as a fact rather than a guess.
func TestContent_everyTrackBuildsForTheTargetPlatformAndReportsItsMachine(t *testing.T) {
	for _, track := range Tracks() {
		yaml, _ := Track(track)
		if !strings.Contains(string(yaml), "- name: target-platform") {
			t.Errorf("track %s does not declare the target-platform fact", track)
		}
		if !strings.Contains(string(yaml), "$(params.target-platform)") {
			t.Errorf("track %s declares target-platform but never hands it to its build task", track)
		}
	}
	tasks, err := TaskFiles()
	if err != nil {
		t.Fatal(err)
	}
	for _, stem := range []string{"buildkit", "buildpacks"} {
		task := string(tasks[stem])
		if !strings.Contains(task, "- name: build-node-architecture") {
			t.Errorf("build task %s does not declare the build-node-architecture result", stem)
		}
		if !strings.Contains(task, "$(results.build-node-architecture.path)") {
			t.Errorf("build task %s declares the result but never writes it", stem)
		}
	}
}

// Every build task names what it pushed by its digest -- the image-digest
// result a deployment pins, so re-pushing the tag never changes what runs --
// and fails its own step when it pushed and cannot name it. Every track
// declares the optional cache-image fact and hands it to its build task, so a
// build tagged per commit still reads the last build's layers.
func TestContent_everyBuildNamesItsImageByDigestAndReusesOneCache(t *testing.T) {
	tasks, err := TaskFiles()
	if err != nil {
		t.Fatal(err)
	}
	for _, stem := range []string{"buildkit", "buildpacks"} {
		task := string(tasks[stem])
		if !strings.Contains(task, "- name: image-digest") {
			t.Errorf("build task %s does not declare the image-digest result", stem)
		}
		if !strings.Contains(task, "$(results.image-digest.path)") {
			t.Errorf("build task %s declares image-digest but never writes it", stem)
		}
		if !strings.Contains(task, "cannot name the image it produced") {
			t.Errorf("build task %s does not fail, in words, a push it cannot name", stem)
		}
	}
	buildkit := string(tasks["buildkit"])
	for _, want := range []string{"--metadata-file", "type=registry,ref=$(params.cacheImage)", "ignore-error=true", "image-manifest=true",
		"$(params.registrySignInExpiresAt)", "The registry sign-in Planton made when this build started expired at"} {
		if !strings.Contains(buildkit, want) {
			t.Errorf("the BuildKit task lacks %q", want)
		}
	}
	for _, track := range Tracks() {
		yaml, _ := Track(track)
		if !strings.Contains(string(yaml), "- name: cache-image") || !strings.Contains(string(yaml), "$(params.cache-image)") {
			t.Errorf("track %s does not declare the cache-image fact and hand it to its build task", track)
		}
	}
	// Every track hands its build the sign-in's expiry, the fact the runner
	// supplies to a pipeline that declares it, and every build task says it
	// when a push was refused after that moment.
	for _, track := range Tracks() {
		yaml, _ := Track(track)
		if !strings.Contains(string(yaml), "$(params.registry-sign-in-expires-at)") {
			t.Errorf("track %s does not hand the registry sign-in's expiry to its build task", track)
		}
	}
	for _, stem := range []string{"buildkit", "buildpacks"} {
		if !strings.Contains(string(tasks[stem]), "The registry sign-in Planton made when this build started expired at") {
			t.Errorf("build task %s does not say when a push was refused after its sign-in expired", stem)
		}
	}
}

// A tag release names its image by the release too, on the same digest, so
// the registry answers "which image is v1.4.0" without the run history: every
// track declares the optional git-tag fact and hands it to its build task,
// and every build task pushes the second name only for a tag an image can
// carry, says so when it cannot, and reports the tag it pushed as the
// release-tag result the run records exactly or not at all.
func TestContent_aTagReleaseAlsoNamesItsImageByItsTag(t *testing.T) {
	for _, track := range Tracks() {
		yaml, _ := Track(track)
		if !strings.Contains(string(yaml), "- name: git-tag") || !strings.Contains(string(yaml), "$(params.git-tag)") {
			t.Errorf("track %s does not declare the git-tag fact and hand it to its build task", track)
		}
	}
	tasks, err := TaskFiles()
	if err != nil {
		t.Fatal(err)
	}
	for _, stem := range []string{"buildkit", "buildpacks"} {
		task := string(tasks[stem])
		for _, want := range []string{"- name: release-tag", "$(results.release-tag.path)",
			"^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$", "isn't a valid image tag"} {
			if !strings.Contains(task, want) {
				t.Errorf("build task %s lacks %q", stem, want)
			}
		}
	}
	if !strings.Contains(string(tasks["buildkit"]), `"type=image,\"name=$names\",push=$(params.push)"`) {
		t.Error("the BuildKit task does not push every name the build carries in one output")
	}
	if !strings.Contains(string(tasks["buildpacks"]), `"-tag=$repository:$(params.RELEASE_TAG)"`) {
		t.Error("the Buildpacks task does not hand the lifecycle the release's name")
	}
}

// Every image the content can run is digest-pinned: the tag documents
// intent, the digest is what the cluster pulls, so a build is reproducible
// and immune to upstream tag mutation -- and the derived allowlist an
// air-gapped cluster mirrors is immutable.
func TestImages_everyImageIsDigestPinned(t *testing.T) {
	images, err := Images()
	if err != nil {
		t.Fatal(err)
	}
	if len(images) == 0 {
		t.Fatal("the content runs no images at all -- the derivation is broken")
	}
	for _, image := range images {
		if !strings.Contains(image, "@sha256:") {
			t.Errorf("image %q is not digest-pinned", image)
		}
		if strings.Contains(image, "$(") {
			t.Errorf("image %q is an unresolved param reference", image)
		}
	}
}

// The derived image set is recorded so an image can never enter or leave
// silently: a content change that alters what the cluster pulls shows up
// here as a reviewed diff, beside the Version bump the digest test forces.
func TestImages_matchTheRecordedSet(t *testing.T) {
	images, err := Images()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"bitnamilegacy/kubectl:latest@sha256:cd354d5b25562b195b277125439c23e4046902d7f1abc0dc3c75aad04d298c17",
		"docker.io/library/bash:5.1.4@sha256:c523c636b722339f41b6a431b44588ab2f762c5de5ec3bd7964420ff982fb1d9",
		"ghcr.io/tektoncd-catalog/git-clone:v1.1.0@sha256:b7fe6c370322586feb555c807f3fae7ca5d62c20ebbcca987114e69366151957",
		"ghcr.io/tektoncd/github.com/tektoncd/pipeline/cmd/git-init:v0.45.0@sha256:8ab0f58d8381b0b71f5b2bae1f63522989d739e3154d8cab1bacfa0ef5317214",
		"moby/buildkit:v0.23.2-rootless@sha256:cab936745de5d673465948f1e93ff4d6e372bbe33f218afd3314eba45a6f85a9",
		"paketobuildpacks/builder-jammy-base:latest@sha256:f93da4e8abc73ab3555793d3a992b724ba7d0baafffeddb1219cf5c433fcf3b2",
	}
	if !reflect.DeepEqual(images, want) {
		t.Fatalf("the derived image set changed.\n got: %s\nwant: %s\nIf the content change is intended, update this list (and Version).", strings.Join(images, "\n      "), strings.Join(want, "\n      "))
	}
}

// reviewedLedger is the ledger as last reviewed, line for line. The ledger is
// append-only, and a hand edit of an old line would quietly rename a content
// state a run record already carries; this list is a second copy that edit
// would also have to change, in plain sight. When you record a new Version,
// add its line here in the same change -- the gate allows exactly one ledger
// line beyond this list, the one being recorded.
var reviewedLedger = []string{
	"v3 a4c48a698cfbefe3ae509a8885f918dcdbc3fb141bdaccf7ef19f15fe6bbc3be",
	"v4 1e77e9bfd8b375fe1192efec9ee765c8f6cb69deb12dfdaf2a8420894479e78e",
	"v5 5e175757dcfe53f2267365d7bdff725182bfc41146b147437cd1f5538cf085eb",
	"v6 5bfbc81dc8380efea1a4e2fdc3e592c09181bae7289dc76b902a184e97e698fa",
	"v7 fd738ac9d4f81d74692cd7a5705ecea18db28c6915161dda372857eb0baccf56",
	"v8 edbafddc965a05a343ba736924d633949932861b0ead3d4417d11f395392d646",
	"v9 674e42c1d0dbf3b624459631467f09ee6efdb7f19e34147c4ab41a213affb7b4",
}

func TestContent_theLedgerNeverRewritesARecordedState(t *testing.T) {
	ledger := readContentLedger(t, filepath.Join("testdata", "content-digest.txt"))
	for i, want := range reviewedLedger {
		if i >= len(ledger.lines) {
			t.Fatalf("the ledger lost %q -- a recorded content state is never removed", want)
		}
		if ledger.lines[i] != want {
			t.Fatalf("ledger line %d reads %q, but it was recorded as %q -- a recorded content state is never rewritten; restore the line and record a new Version instead", i+1, ledger.lines[i], want)
		}
	}
	if extra := len(ledger.lines) - len(reviewedLedger); extra > 1 {
		t.Fatalf("the ledger has %d lines beyond reviewedLedger (%v) -- add each recorded line to reviewedLedger in the change that records it", extra, ledger.lines[len(reviewedLedger):])
	}
}

// contentLedger is testdata/content-digest.txt read both ways: what each
// Version names, and which Version names each content state. Lines starting
// with '#' are commentary.
type contentLedger struct {
	digestOf  map[string]string
	versionOf map[string]string
	lines     []string // the pairs in ledger order, "<Version> <digest>"
}

func readContentLedger(t *testing.T, path string) contentLedger {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading the content ledger: %v", err)
	}
	ledger := contentLedger{digestOf: map[string]string{}, versionOf: map[string]string{}}
	for i, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			t.Fatalf("%s:%d: want \"<Version> <digest>\", got %q", path, i+1, line)
		}
		version, digest := fields[0], fields[1]
		if earlier, ok := ledger.digestOf[version]; ok {
			t.Fatalf("%s:%d: %s is recorded twice (%s and %s) -- one Version names one content state", path, i+1, version, earlier, digest)
		}
		if earlier, ok := ledger.versionOf[digest]; ok {
			t.Fatalf("%s:%d: content %s is recorded as both %s and %s -- one content state carries one Version", path, i+1, digest, earlier, version)
		}
		ledger.digestOf[version] = digest
		ledger.versionOf[digest] = version
		ledger.lines = append(ledger.lines, version+" "+digest)
	}
	return ledger
}

func computeDigest(t *testing.T) string {
	t.Helper()
	files, err := allFiles()
	if err != nil {
		t.Fatalf("walking embedded content: %v", err)
	}
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	h := sha256.New()
	for _, path := range paths {
		h.Write([]byte(path))
		h.Write([]byte{0})
		h.Write(files[path])
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}
