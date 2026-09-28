// Package tekton is the platform's release-pinned Tekton pipeline catalog:
// the service build tracks and the tasks they reference, authored as plain
// Tekton YAML in this directory and embedded in this package so the platform
// compiles builds from a released copy -- no build ever fetches a pipeline
// definition from a live git branch.
//
// One folder, one pin, two consumers. The YAML beside this file is the
// source of truth a user reads; this package is the same bytes as Go, which
// the platform imports through its pin of this module. The platform's
// pipeline compiler reads the tracks and the task catalog; the platform's
// build-readiness probe reads the image set (Images). Nothing else reads it,
// and this package owns nothing else: no compilation, no validation, no
// dispatch -- those live with the compiler and the engine.
//
// Version discipline: Version names exactly one content state, and the
// digest test holds the pair together -- editing any embedded YAML without
// bumping Version (and re-recording the digest) fails the gate. The pin
// stamped on every run record derives from Version, so a silent content
// edit would otherwise corrupt the byte-identical-rerun law. The pin is
// deliberately NOT this module's release version: a release that touches no
// Tekton content must not give one content state a second pin.
package tekton

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
)

// Version names the embedded content state. Bump it with ANY change to the
// embedded YAML, however small: the pin on every run record derives from
// it, and two different content states must never share a pin.
const Version = "v8"

// Pin is the platform-release source pin stamped on run records compiled
// from this content.
func Pin() string {
	return "platform-content/" + Version
}

//go:embed pipelines/*.yaml tasks/*.yaml
var contentFS embed.FS

// Track returns the pipeline YAML for a platform build track (the file
// stem under pipelines/). ok is false for an unknown track -- the compiler
// turns that into an actionable verdict naming the known tracks.
func Track(name string) (yaml []byte, ok bool) {
	data, err := contentFS.ReadFile("pipelines/" + name + ".yaml")
	if err != nil {
		return nil, false
	}
	return data, true
}

// Tracks returns the known track names, sorted, for verdict messages.
func Tracks() []string {
	return listStems("pipelines")
}

// TaskFiles returns every embedded task document, keyed by file stem. The
// compiler indexes these by the Tekton metadata.name inside each document,
// not by stem -- the two coincide for every current task except buildkit
// (file buildkit.yaml, task name buildkit-daemonless), which is exactly why
// the stem is never trusted as the ref name.
func TaskFiles() (map[string][]byte, error) {
	stems := listStems("tasks")
	out := make(map[string][]byte, len(stems))
	for _, stem := range stems {
		data, err := contentFS.ReadFile("tasks/" + stem + ".yaml")
		if err != nil {
			return nil, fmt.Errorf("reading embedded task %s: %w", stem, err)
		}
		out[stem] = data
	}
	return out, nil
}

// allFiles is the digest input: every embedded file in stable order.
func allFiles() (map[string][]byte, error) {
	out := map[string][]byte{}
	err := fs.WalkDir(contentFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, readErr := contentFS.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		out[path] = data
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

func listStems(dir string) []string {
	entries, err := contentFS.ReadDir(dir)
	if err != nil {
		return nil
	}
	stems := make([]string, 0, len(entries))
	for _, e := range entries {
		name := e.Name()
		if len(name) > len(".yaml") && name[len(name)-len(".yaml"):] == ".yaml" {
			stems = append(stems, name[:len(name)-len(".yaml")])
		}
	}
	sort.Strings(stems)
	return stems
}
