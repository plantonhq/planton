package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/plantonhq/planton/pkg/downloads/modulemanifest"
	"sigs.k8s.io/yaml"
)

// repoRoot is the repository root, or "" in a sandbox that carries only the
// package's declared inputs (Bazel); the workflow mirror gates run from
// `go test` at the repository root, like the other release-lane gates.
func repoRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".github", "workflows", "release.yaml")); err != nil {
		t.Skipf("repository tree not present at %s (sandboxed run)", root)
	}
	return root
}

type workflow struct {
	Jobs map[string]struct {
		Needs    needs  `json:"needs"`
		Steps    []step `json:"steps"`
		Strategy struct {
			Matrix struct {
				Provider []string `json:"provider"`
				Platform []struct {
					Goos   string `json:"goos"`
					Goarch string `json:"goarch"`
				} `json:"platform"`
			} `json:"matrix"`
		} `json:"strategy"`
	} `json:"jobs"`
}

// needs is a job's needs, which a workflow writes as one job's name or as a
// list.
type needs []string

func (n *needs) UnmarshalJSON(data []byte) error {
	var one string
	if err := json.Unmarshal(data, &one); err == nil {
		*n = needs{one}
		return nil
	}
	return json.Unmarshal(data, (*[]string)(n))
}

type step struct {
	Uses string         `json:"uses"`
	With map[string]any `json:"with"`
	Run  string         `json:"run"`
}

// stepUsing is the job's first step that uses action (actions/upload-artifact,
// ...), whatever its version.
func stepUsing(steps []step, action string) (step, bool) {
	for _, s := range steps {
		if strings.HasPrefix(s.Uses, action+"@") {
			return s, true
		}
	}
	return step{}, false
}

func readWorkflow(t *testing.T, root, name string) workflow {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, ".github", "workflows", name))
	if err != nil {
		t.Fatal(err)
	}
	var w workflow
	if err := yaml.Unmarshal(data, &w); err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return w
}

// TestPulumiMatrixBuildsEveryManifestPlatform holds the release's Pulumi
// matrix to the platforms the module manifest requires: a platform the matrix
// dropped would leave every Pulumi module incomplete, and the release would
// publish no Pulumi manifest at all.
func TestPulumiMatrixBuildsEveryManifestPlatform(t *testing.T) {
	root := repoRoot(t)
	var built []string
	for _, p := range readWorkflow(t, root, "release.pulumi-modules.yaml").Jobs["build"].Strategy.Matrix.Platform {
		built = append(built, p.Goos+"_"+p.Goarch)
	}
	sort.Strings(built)
	want := append([]string(nil), modulemanifest.PulumiPlatforms...)
	sort.Strings(want)
	if strings.Join(built, ",") != strings.Join(want, ",") {
		t.Fatalf("release.pulumi-modules.yaml builds %v; the module manifest requires %v", built, want)
	}
}

// TestModuleMatricesCoverEveryProvider holds each module lane's provider
// matrix to the providers that ship modules of its engine, so a new provider
// fails here, on its pull request, rather than at release time.
func TestModuleMatricesCoverEveryProvider(t *testing.T) {
	root := repoRoot(t)
	modules, err := ReleasedModules(root)
	if err != nil {
		t.Fatal(err)
	}
	for engine, file := range map[modulemanifest.Engine]string{modulemanifest.Pulumi: "release.pulumi-modules.yaml", modulemanifest.Terraform: "release.terraform-modules.yaml"} {
		matrix := map[string]bool{}
		for _, p := range readWorkflow(t, root, file).Jobs["build"].Strategy.Matrix.Provider {
			matrix[p] = true
		}
		for _, m := range modules {
			if m.Engine == engine && !matrix[m.Provider] {
				t.Errorf("%s ships %s modules, but %s's matrix has no %s leg", m.Provider, engine, file, m.Provider)
			}
		}
	}
}

// TestEachEngineManifestWaitsOnlyForItsOwnLane holds release.yaml's manifest
// jobs to the reason there are two: each engine's manifest is published when
// its own lane is done. A manifest job that also needed the other lane would
// make a platform's pin wait for a build it does not run; one that read the
// other lane's parts would be refused for listing modules of another engine.
func TestEachEngineManifestWaitsOnlyForItsOwnLane(t *testing.T) {
	root := repoRoot(t)
	script, err := os.ReadFile(filepath.Join(root, "tools", "ci", "release", "publish_module_manifest.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(script), `modulebuild manifest -engine "${engine}"`) {
		t.Error("publish_module_manifest.sh does not assemble the manifest of the engine it is given with modulebuild manifest")
	}
	release := readWorkflow(t, root, "release.yaml")
	for _, engine := range modulemanifest.Engines {
		lane, manifestJob := string(engine)+"-modules", string(engine)+"-module-manifest"
		job, ok := release.Jobs[manifestJob]
		if !ok {
			t.Errorf("release.yaml has no %s job", manifestJob)
			continue
		}
		if strings.Join(job.Needs, ",") != lane {
			t.Errorf("%s needs %v; it must need only %s", manifestJob, job.Needs, lane)
		}

		partPrefix := "module-part-" + string(engine) + "-"
		download, ok := stepUsing(job.Steps, "actions/download-artifact")
		if pattern, _ := download.With["pattern"].(string); !ok || pattern != partPrefix+"*" {
			t.Errorf("%s collects parts matching %q; it must collect only %q", manifestJob, pattern, partPrefix+"*")
		}
		assembles := false
		for _, s := range job.Steps {
			assembles = assembles || (strings.Contains(s.Run, "tools/ci/release/publish_module_manifest.sh") && strings.Contains(s.Run, " "+string(engine)+" "))
		}
		if !assembles {
			t.Errorf("%s does not assemble its manifest with tools/ci/release/publish_module_manifest.sh <tag> %s <parts>", manifestJob, engine)
		}

		upload, ok := stepUsing(readWorkflow(t, root, "release."+lane+".yaml").Jobs["build"].Steps, "actions/upload-artifact")
		if name, _ := upload.With["name"].(string); !ok || !strings.HasPrefix(name, partPrefix) {
			t.Errorf("release.%s.yaml uploads its parts as %q; %s collects only %s*", lane, name, manifestJob, partPrefix)
		}
	}
}
