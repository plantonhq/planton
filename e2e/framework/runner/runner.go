package runner

import (
	"context"
	"fmt"
	"time"

	tt "github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/e2e/framework/provider"
)

// Phase represents a stage in the E2E test lifecycle.
type Phase string

const (
	PhaseFixturesUp Phase = "FIXTURES-UP"
	PhaseValidate   Phase = "VALIDATE"
	PhaseDeploy     Phase = "DEPLOY"
	PhaseVerifyOut  Phase = "VERIFY-OUT"
	PhaseVerifyRes  Phase = "VERIFY-RES"
	PhaseDestroy    Phase = "DESTROY"
	PhaseVerifyCln  Phase = "VERIFY-CLN"
	PhaseFixturesDn Phase = "FIXTURES-DN"
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
// If the component has fixtures (prerequisites), they are deployed first
// and torn down last, wrapping the standard 6-phase lifecycle.
func RunComponentTest(ctx context.Context, tc *provider.ComponentTestContext, harness provider.Harness) *TestResult {
	start := time.Now()
	result := &TestResult{
		Component: tc.Component,
		Engine:    tc.Engine,
		Passed:    true,
	}

	verifyCtx := context.WithValue(ctx, provider.ManifestPathKey{}, tc.ManifestPath)

	// Phase 0: deploy fixtures if the component has a fixtures/ directory
	var fixtureStates []FixtureState
	if tc.RepoRoot != "" {
		fixtureStart := time.Now()
		var err error
		fixtureStates, err = DeployFixtures(ctx, tc.RepoRoot, tc.Provider, tc.Component, tc.BackendURL, tc.RunID, harness)
		pr := PhaseResult{
			Phase:    PhaseFixturesUp,
			Duration: time.Since(fixtureStart),
			Passed:   err == nil,
			Error:    err,
		}
		if len(fixtureStates) > 0 || err != nil {
			result.Phases = append(result.Phases, pr)
		}
		if err != nil {
			result.Passed = false
			TeardownFixtures(fixtureStates)
			result.Duration = time.Since(start)
			return result
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

	// Phase 7: teardown fixtures in reverse order
	if len(fixtureStates) > 0 {
		fixtureStart := time.Now()
		TeardownFixtures(fixtureStates)
		result.Phases = append(result.Phases, PhaseResult{
			Phase:    PhaseFixturesDn,
			Duration: time.Since(fixtureStart),
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
			return nil
		}
		_ = outputJSON
		return nil

	case "terraform":
		opts, ok := tc.TerraformOpts.(*tt.Options)
		if !ok || opts == nil {
			return nil
		}
		outputs, err := TerraformOutputs(tc.T, opts)
		if err != nil {
			return nil
		}
		tc.Outputs = outputs
		return nil

	default:
		return nil
	}
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
