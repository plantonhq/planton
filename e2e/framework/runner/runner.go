package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	tt "github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"
	"github.com/plantonhq/openmcf/e2e/framework/provider"
	"github.com/plantonhq/openmcf/pkg/crkreflect"
)

// Phase represents a stage in the E2E test lifecycle.
type Phase string

const (
	PhaseDepsUp    Phase = "DEPENDENCIES-UP"
	PhaseValidate  Phase = "VALIDATE"
	PhaseDeploy    Phase = "DEPLOY"
	PhaseVerifyOut Phase = "VERIFY-OUT"
	PhaseVerifyRes Phase = "VERIFY-RES"
	PhaseDestroy   Phase = "DESTROY"
	PhaseVerifyCln Phase = "VERIFY-CLN"
	PhaseDepsDn    Phase = "DEPENDENCIES-DOWN"
)

// PhaseResult captures the outcome of a single lifecycle phase.
type PhaseResult struct {
	Phase    Phase
	Passed   bool
	Duration time.Duration
	Error    error
}

// TestResult captures the full 6-phase lifecycle outcome for a component.
type TestResult struct {
	Component string
	Engine    string
	Phases    []PhaseResult
	Passed    bool
	Duration  time.Duration
}

// RunComponentTest executes the E2E lifecycle for a single component.
// If the component has dependencies (registry prerequisites), they are deployed
// first and torn down last, wrapping the standard 6-phase lifecycle.
func RunComponentTest(ctx context.Context, tc *provider.ComponentTestContext, harness provider.Harness) *TestResult {
	start := time.Now()
	result := &TestResult{
		Component: tc.Component,
		Engine:    tc.Engine,
		Passed:    true,
	}

	verifyCtx := context.WithValue(ctx, provider.ManifestPathKey{}, tc.ManifestPath)

	// Phase 0: deploy dependencies (registry prerequisites)
	var dependencyStates []DependencyState
	if tc.RepoRoot != "" {
		depStart := time.Now()
		var err error
		dependencyStates, err = DeployDependencies(ctx, tc.RepoRoot, tc.Provider, tc.Component, tc.BackendURL, tc.RunID, harness)
		pr := PhaseResult{
			Phase:    PhaseDepsUp,
			Duration: time.Since(depStart),
			Passed:   err == nil,
			Error:    err,
		}
		if len(dependencyStates) > 0 || err != nil {
			result.Phases = append(result.Phases, pr)
		}
		if err != nil {
			result.Passed = false
			TeardownDependencies(dependencyStates)
			result.Duration = time.Since(start)
			return result
		}

		// Resolve the component manifest's value_from refs against the deployed
		// prerequisites' outputs -- the orchestrator's resolution step, performed
		// here so a composed topology (e.g. subnet -> vpc) can be tested standalone.
		if len(dependencyStates) > 0 {
			depOutputs := make(map[cloudresourcekind.CloudResourceKind]map[string]interface{}, len(dependencyStates))
			for _, depState := range dependencyStates {
				depOutputs[crkreflect.KindFromString(depState.Dependency.KindSlug)] = depState.Outputs
			}
			resolvedPath, resolveErr := ResolveManifestRefs(tc.ManifestPath, depOutputs)
			if resolveErr != nil {
				result.Passed = false
				result.Phases = append(result.Phases, PhaseResult{
					Phase:    PhaseValidate,
					Duration: 0,
					Passed:   false,
					Error:    errors.Wrap(resolveErr, "failed to resolve manifest references from dependency outputs"),
				})
				TeardownDependencies(dependencyStates)
				result.Duration = time.Since(start)
				return result
			}
			tc.ManifestPath = resolvedPath
			verifyCtx = context.WithValue(ctx, provider.ManifestPathKey{}, tc.ManifestPath)
		}
	}

	// Phases 1-6: standard lifecycle
	phases := []struct {
		phase Phase
		fn    func() error
	}{
		{PhaseValidate, func() error { return runValidate(tc) }},
		{PhaseDeploy, func() error { return runDeploy(tc) }},
		{PhaseVerifyOut, func() error { return runVerifyOutputs(tc) }},
		{PhaseVerifyRes, func() error { return runVerifyResources(verifyCtx, tc, harness) }},
		{PhaseDestroy, func() error { return runDestroy(tc) }},
		{PhaseVerifyCln, func() error { return runVerifyCleanup(verifyCtx, tc, harness) }},
	}

	for _, p := range phases {
		phaseStart := time.Now()
		err := p.fn()
		pr := PhaseResult{
			Phase:    p.phase,
			Duration: time.Since(phaseStart),
			Passed:   err == nil,
			Error:    err,
		}
		result.Phases = append(result.Phases, pr)

		if err != nil {
			result.Passed = false
			if p.phase == PhaseDeploy || p.phase == PhaseVerifyOut || p.phase == PhaseVerifyRes {
				cleanupErr := runDestroy(tc)
				if cleanupErr != nil {
					fmt.Printf("  [WARN] cleanup destroy also failed: %v\n", cleanupErr)
				}
			}
			break
		}
	}

	// Phase 7: teardown dependencies in reverse order
	if len(dependencyStates) > 0 {
		depStart := time.Now()
		TeardownDependencies(dependencyStates)
		result.Phases = append(result.Phases, PhaseResult{
			Phase:    PhaseDepsDn,
			Duration: time.Since(depStart),
			Passed:   true,
		})
	}

	// Clean up Terraform working directory if one was created
	if tc.TerraformCleanup != nil {
		tc.TerraformCleanup()
	}

	result.Duration = time.Since(start)
	return result
}

func runValidate(tc *provider.ComponentTestContext) error {
	if tc.ManifestPath == "" {
		return errors.New("manifest path is empty")
	}

	switch tc.Engine {
	case "pulumi":
		stackInputPath, err := BuildStackInput(tc.ManifestPath, tc.ModuleDir)
		if err != nil {
			return errors.Wrap(err, "validation failed: cannot build stack input from manifest")
		}
		tc.StackInputFilePath = stackInputPath

	case "terraform":
		workDir, cleanup, err := PrepareWorkDir(tc.ModuleDir)
		if err != nil {
			return errors.Wrap(err, "validation failed: cannot prepare terraform working directory")
		}
		tc.TerraformWorkDir = workDir
		tc.TerraformCleanup = cleanup

		input, err := BuildTerraformInput(tc.ManifestPath, workDir)
		if err != nil {
			cleanup()
			return errors.Wrap(err, "validation failed: cannot build terraform input from manifest")
		}

		tc.TerraformOpts = BuildTerratestOptions(tc.T, workDir, input.TfvarsPath, input.EnvVars)

	default:
		return errors.Errorf("unsupported engine for validation: %s", tc.Engine)
	}

	return nil
}

func runDeploy(tc *provider.ComponentTestContext) error {
	switch tc.Engine {
	case "pulumi":
		_, err := PulumiDeploy(tc.ModuleDir, tc.StackName, tc.BackendURL, tc.StackInputFilePath)
		return err
	case "terraform":
		opts, ok := tc.TerraformOpts.(*tt.Options)
		if !ok || opts == nil {
			return errors.New("terraform options not initialized (runValidate must run first)")
		}
		_, err := TerraformDeploy(tc.T, opts)
		return err
	default:
		return errors.Errorf("unsupported engine: %s", tc.Engine)
	}
}

func runVerifyOutputs(tc *provider.ComponentTestContext) error {
	switch tc.Engine {
	case "pulumi":
		outputJSON, err := PulumiStackOutputs(tc.ModuleDir, tc.StackName, tc.BackendURL)
		if err != nil {
			return errors.Wrap(err, "failed to retrieve pulumi stack outputs")
		}
		rawOutputs, parseErr := parsePulumiOutputs(outputJSON)
		if parseErr != nil {
			return errors.Wrap(parseErr, "failed to parse pulumi stack outputs JSON")
		}
		tc.Outputs = rawOutputs

	case "terraform":
		opts, ok := tc.TerraformOpts.(*tt.Options)
		if !ok || opts == nil {
			return errors.New("terraform options not initialized (runValidate must run first)")
		}
		rawOutputs, err := TerraformOutputs(tc.T, opts)
		if err != nil {
			return errors.Wrap(err, "failed to retrieve terraform outputs")
		}
		tc.Outputs = rawOutputs

	default:
		return nil
	}

	if len(tc.Outputs) == 0 {
		fmt.Printf("  [outputs] %s: no outputs captured, skipping transformation validation\n", tc.Component)
		return nil
	}

	msg, flatOutputs, err := VerifyOutputTransformation(tc.Component, tc.Outputs, tc.ModuleDir)
	if err != nil {
		return err
	}
	tc.FlatOutputs = flatOutputs
	tc.TransformedOutputs = msg
	return nil
}

func runVerifyResources(ctx context.Context, tc *provider.ComponentTestContext, harness provider.Harness) error {
	return harness.VerifyDeployed(ctx, tc.Component, tc.Outputs)
}

func runDestroy(tc *provider.ComponentTestContext) error {
	switch tc.Engine {
	case "pulumi":
		_, err := PulumiDestroy(tc.ModuleDir, tc.StackName, tc.BackendURL, tc.StackInputFilePath)
		if err != nil {
			return err
		}
		return PulumiRemoveStack(tc.ModuleDir, tc.StackName, tc.BackendURL)
	case "terraform":
		opts, ok := tc.TerraformOpts.(*tt.Options)
		if !ok || opts == nil {
			return errors.New("terraform options not initialized")
		}
		_, err := TerraformDestroy(tc.T, opts)
		return err
	default:
		return errors.Errorf("unsupported engine: %s", tc.Engine)
	}
}

func runVerifyCleanup(ctx context.Context, tc *provider.ComponentTestContext, harness provider.Harness) error {
	return harness.VerifyDestroyed(ctx, tc.Component)
}

// parsePulumiOutputs converts the JSON string from `pulumi stack output --json`
// into a map[string]interface{} compatible with tc.Outputs.
func parsePulumiOutputs(outputJSON string) (map[string]interface{}, error) {
	if outputJSON == "" {
		return nil, nil
	}
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(outputJSON), &raw); err != nil {
		return nil, err
	}
	return raw, nil
}
