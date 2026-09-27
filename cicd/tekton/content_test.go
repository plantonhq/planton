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
)

// The recorded digest holds Version and the embedded tree together: editing
// any embedded YAML without bumping Version (and re-recording the digest)
// fails the gate. Two different content states must never share a pin --
// the pin is what makes stamped provenance and byte-identical reruns honest.
func TestContent_versionNamesExactlyOneContentState(t *testing.T) {
	computed := computeDigest(t)
	digestPath := filepath.Join("testdata", "content-digest.txt")
	recorded, readErr := os.ReadFile(digestPath)
	want := strings.TrimSpace(string(recorded))
	if readErr != nil || computed != want {
		if os.Getenv("UPDATE_CONTENT_DIGEST") != "" {
			if err := os.WriteFile(digestPath, []byte(computed+"\n"), 0o644); err != nil {
				t.Fatalf("recording the new digest: %v", err)
			}
			t.Logf("content digest recorded as %s -- verify Version was bumped for this content change", computed)
			return
		}
		t.Fatalf("the embedded content changed but the recorded digest did not: computed %s, recorded %q.\nBump Version in content.go (two content states must never share a pin), then re-record with UPDATE_CONTENT_DIGEST=1 go test ./cicd/tekton/", computed, want)
	}
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
	// The BuildKit track hands its step the sign-in's expiry, the fact the
	// runner supplies to a pipeline that declares it.
	if dockerfile, _ := Track("dockerfile"); !strings.Contains(string(dockerfile), "$(params.registry-sign-in-expires-at)") {
		t.Error("the dockerfile track does not hand the registry sign-in's expiry to its build task")
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
