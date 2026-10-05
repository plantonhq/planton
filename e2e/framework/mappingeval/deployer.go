// Package mappingeval deploys a MappingEvalSuite's fixtures and builds the
// ground truth the scorer grades against. It is the seeded-account half of
// the mapping eval harness; the scoring half lives in
// pkg/iac/mappingeval (kept free of test-infrastructure dependencies --
// this package rides terratest and the E2E runner internals, so it lives
// with them).
//
// Every member deploys through ONE arm -- the terraform path the E2E
// framework already trusts -- in the suite's listed order, with value_from
// references resolved against earlier members' outputs exactly as the E2E
// prerequisite machinery resolves them. One arm, deliberately: the ground
// truth's infra-component identities come from each member's IaC state, and
// a single state format keeps the answer key uniform. (The E2E harness's
// own prerequisite machinery deploys each dependency on its kind's engine,
// Pulumi for most; it is NOT reused here for exactly that reason -- and
// because a suite member is a first-class evaluation subject, never just
// scaffolding.)
package mappingeval

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	stderrors "errors"

	tt "github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/pkg/errors"
	"github.com/plantonhq/planton/e2e/framework/runner"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/pkg/iac/importmap"
	"github.com/plantonhq/planton/pkg/iac/mappingeval"
)

// DeployedMember tracks one deployed suite member for teardown.
type DeployedMember struct {
	Member  mappingeval.SuiteMember
	WorkDir string
	Opts    *tt.Options
	Cleanup func()
}

// DeploySuite deploys every member in listed order and returns the deployed
// states (ALWAYS returned, even on error, so the caller can tear down
// whatever exists) plus the assembled ground truth.
func DeploySuite(t testing.TB, repoRoot, provider string, suite *mappingeval.LoadedSuite) ([]DeployedMember, *mappingeval.GroundTruth, error) {
	catalog, err := importmap.LoadProviderCatalog(repoRoot, provider)
	if err != nil {
		return nil, nil, err
	}
	ccTypeByTerraformType := map[string]string{}
	for _, rt := range catalog.GetSpec().GetResourceTypes() {
		if rt.GetCloudControlTypeName() != "" {
			ccTypeByTerraformType[rt.GetTerraformType()] = rt.GetCloudControlTypeName()
		}
	}

	var deployed []DeployedMember
	groundTruth := &mappingeval.GroundTruth{}
	// Accumulated outputs let later members' value_from references resolve
	// against earlier members -- the same transitive resolution the E2E
	// dependency chain applies.
	accumulated := runner.DependencyOutputs{}

	for _, member := range suite.Members {
		fmt.Printf("  [eval-seed] deploying %s %q\n", member.KindDir, member.Name)
		start := time.Now()

		if _, err := catalogkindreflect.KindVersionDir(member.KindDir); err != nil {
			return deployed, nil, err
		}
		moduleDir := filepath.Join(repoRoot, "catalog", provider, member.KindDir, "iac", "tf")
		workDir, cleanup, err := runner.PrepareWorkDir(moduleDir)
		if err != nil {
			return deployed, nil, errors.Wrapf(err, "preparing workdir for %s", member.KindDir)
		}

		resolvedPath, err := runner.ResolveManifestRefs(member.ManifestPath, accumulated)
		if err != nil {
			cleanup()
			return deployed, nil, errors.Wrapf(err, "resolving references for %s %q", member.KindDir, member.Name)
		}
		// Mapping-eval members deploy with the harness's default posture; the
		// provider-config fixture belongs to the kind E2E lane only.
		input, err := runner.BuildTerraformInput(resolvedPath, workDir, nil)
		if err != nil {
			cleanup()
			return deployed, nil, errors.Wrapf(err, "building terraform input for %s %q", member.KindDir, member.Name)
		}
		// One shared provider plugin cache for the whole suite: every member
		// gets an isolated workdir, and without the cache each init
		// re-downloads the same provider -- slow, and the documented flaky
		// failure mode of these lanes (a mid-suite download hiccup aborts a
		// half-deployed suite). Members deploy sequentially, which is the
		// mode the cache is safe in. An invoker-set cache dir is respected.
		if os.Getenv("TF_PLUGIN_CACHE_DIR") == "" {
			cacheDir := filepath.Join(os.TempDir(), "planton-eval-tf-plugin-cache")
			if err := os.MkdirAll(cacheDir, 0o755); err == nil {
				input.EnvVars["TF_PLUGIN_CACHE_DIR"] = cacheDir
			}
		}
		opts := runner.BuildTerratestOptions(t, workDir, input.TfvarsPath, input.EnvVars)

		state := DeployedMember{Member: member, WorkDir: workDir, Opts: opts, Cleanup: cleanup}
		if _, err := runner.TerraformDeploy(t, opts); err != nil {
			// The failed apply may have created resources; track the member
			// so teardown destroys them.
			deployed = append(deployed, state)
			return deployed, nil, errors.Wrapf(err, "deploying %s %q", member.KindDir, member.Name)
		}
		deployed = append(deployed, state)

		rawOutputs, err := runner.TerraformOutputs(t, opts)
		if err != nil {
			return deployed, nil, errors.Wrapf(err, "reading outputs of %s %q", member.KindDir, member.Name)
		}
		if accumulated[member.Kind] == nil {
			accumulated[member.Kind] = map[string]map[string]interface{}{}
		}
		accumulated[member.Kind][member.Name] = rawOutputs

		claims, invisible, err := stateResourceIdentities(workDir, ccTypeByTerraformType)
		if err != nil {
			return deployed, nil, errors.Wrapf(err, "reading state identities of %s %q", member.KindDir, member.Name)
		}
		groundTruth.Instances = append(groundTruth.Instances, mappingeval.GroundTruthInstance{
			KindDir:                member.KindDir,
			Kind:                   member.Kind,
			Name:                   member.Name,
			Manifest:               member.Manifest,
			Claims:                 claims,
			InvisibleResourceTypes: invisible,
		})
		fmt.Printf("  [eval-seed] %s %q deployed in %s (%d scan-visible resources, %d invisible types)\n",
			member.KindDir, member.Name, time.Since(start).Round(time.Second), len(claims), len(invisible))
	}
	return deployed, groundTruth, nil
}

// TeardownSuite destroys deployed members in reverse order. One member's
// destroy failure never stops the rest (stopping early would leak
// everything deployed before it), but every failure is returned so the
// caller FAILS the run -- a destroy that could not run means real provider
// resources may still exist.
func TeardownSuite(t testing.TB, deployed []DeployedMember) error {
	var failures []error
	for i := len(deployed) - 1; i >= 0; i-- {
		member := deployed[i]
		fmt.Printf("  [eval-seed] destroying %s %q\n", member.Member.KindDir, member.Member.Name)
		if _, err := runner.TerraformDestroy(t, member.Opts); err != nil {
			failures = append(failures, errors.Wrapf(err, "destroying %s %q", member.Member.KindDir, member.Member.Name))
		}
		member.Cleanup()
	}
	return stderrors.Join(failures...)
}

// stateResourceIdentities reads the member's local IaC state and translates
// every managed resource into scan coordinates: terraform type -> Cloud
// Control type via the catalog's declared correspondence, and the state's
// "id" attribute as the identifier (the covered staples' primary
// identifiers coincide with the terraform id; a type where they diverge
// gains a declared exception alongside its catalog entry before it can join
// a suite). Types with no declared scan-side name are returned separately
// as structurally invisible.
func stateResourceIdentities(workDir string, ccTypeByTerraformType map[string]string) ([]mappingeval.AccountResourceRef, []string, error) {
	raw, err := os.ReadFile(filepath.Join(workDir, "terraform.tfstate"))
	if err != nil {
		return nil, nil, errors.Wrap(err, "reading state file")
	}
	var state struct {
		Resources []struct {
			Mode      string `json:"mode"`
			Type      string `json:"type"`
			Instances []struct {
				Attributes map[string]any `json:"attributes"`
			} `json:"instances"`
		} `json:"resources"`
	}
	if err := json.Unmarshal(raw, &state); err != nil {
		return nil, nil, errors.Wrap(err, "parsing state file")
	}

	var claims []mappingeval.AccountResourceRef
	var invisible []string
	seenInvisible := map[string]bool{}
	for _, resource := range state.Resources {
		if resource.Mode != "managed" {
			continue
		}
		ccType, visible := ccTypeByTerraformType[resource.Type]
		if !visible {
			if !seenInvisible[resource.Type] {
				seenInvisible[resource.Type] = true
				invisible = append(invisible, resource.Type)
			}
			continue
		}
		for _, instance := range resource.Instances {
			id, _ := instance.Attributes["id"].(string)
			if id == "" {
				return nil, nil, errors.Errorf("state resource %s has an instance with no id attribute", resource.Type)
			}
			claims = append(claims, mappingeval.AccountResourceRef{TypeName: ccType, Identifier: id})
		}
	}
	return claims, invisible, nil
}

// ScanTypeNames derives the scan's type allowlist: the catalog's declared
// Cloud Control type names restricted to what the suite's member kinds
// can actually create (their modules' resource types). This is the v1 scan
// guardrail -- a single region plus an explicit type allowlist -- derived
// from declared artifacts, never authored per suite.
func ScanTypeNames(repoRoot, provider string, suite *mappingeval.LoadedSuite) ([]string, error) {
	catalog, err := importmap.LoadProviderCatalog(repoRoot, provider)
	if err != nil {
		return nil, err
	}
	ccTypeByTerraformType := map[string]string{}
	for _, rt := range catalog.GetSpec().GetResourceTypes() {
		if rt.GetCloudControlTypeName() != "" {
			ccTypeByTerraformType[rt.GetTerraformType()] = rt.GetCloudControlTypeName()
		}
	}

	seen := map[string]bool{}
	var typeNames []string
	for _, member := range suite.Members {
		if _, err := catalogkindreflect.KindVersionDir(member.KindDir); err != nil {
			return nil, err
		}
		moduleDir := filepath.Join(repoRoot, "catalog", provider, member.KindDir, "iac", "tf")
		moduleTypes, err := moduleResourceTypes(moduleDir)
		if err != nil {
			return nil, errors.Wrapf(err, "module resource types of %s", member.KindDir)
		}
		for _, terraformType := range moduleTypes {
			if ccType, visible := ccTypeByTerraformType[terraformType]; visible && !seen[ccType] {
				seen[ccType] = true
				typeNames = append(typeNames, ccType)
			}
		}
	}
	return typeNames, nil
}
