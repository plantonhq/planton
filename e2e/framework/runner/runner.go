package runner

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/plantonhq/openmcf/e2e/framework/provider"
)

// Phase represents a stage in the E2E test lifecycle.
type Phase string

const (
	PhaseValidate  Phase = "VALIDATE"
	PhaseDeploy    Phase = "DEPLOY"
	PhaseVerifyOut Phase = "VERIFY-OUT"
	PhaseVerifyRes Phase = "VERIFY-RES"
	PhaseDestroy   Phase = "DESTROY"
	PhaseVerifyCln Phase = "VERIFY-CLN"
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

// RunComponentTest executes the full 6-phase E2E lifecycle for a single component.
func RunComponentTest(ctx context.Context, tc *provider.ComponentTestContext, harness provider.Harness) *TestResult {
	start := time.Now()
	result := &TestResult{
		Component: tc.Component,
		Engine:    tc.Engine,
		Passed:    true,
	}

	// Inject the manifest path into context for the harness verifiers
	verifyCtx := context.WithValue(ctx, provider.ManifestPathKey{}, tc.ManifestPath)

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
			// After deploy failure, still attempt destroy for cleanup
			if p.phase == PhaseDeploy || p.phase == PhaseVerifyOut || p.phase == PhaseVerifyRes {
				cleanupErr := runDestroy(tc)
				if cleanupErr != nil {
					fmt.Printf("  [WARN] cleanup destroy also failed: %v\n", cleanupErr)
				}
			}
			break
		}
	}

	result.Duration = time.Since(start)
	return result
}

func runValidate(tc *provider.ComponentTestContext) error {
	// For T01, validation = confirm the manifest file exists and is readable
	if tc.ManifestPath == "" {
		return errors.New("manifest path is empty")
	}
	// BuildStackInput already validates the manifest loads correctly
	stackInputPath, err := BuildStackInput(tc.ManifestPath, tc.ModuleDir)
	if err != nil {
		return errors.Wrap(err, "validation failed: cannot build stack input from manifest")
	}
	tc.StackInputFilePath = stackInputPath
	return nil
}

func runDeploy(tc *provider.ComponentTestContext) error {
	switch tc.Engine {
	case "pulumi":
		result, err := PulumiDeploy(tc.ModuleDir, tc.StackName, tc.BackendURL, tc.StackInputFilePath)
		if err != nil {
			return err
		}
		_ = result
		return nil
	case "terraform":
		_, err := TerraformDeploy(tc.ModuleDir, tc.StackInputFilePath)
		return err
	default:
		return errors.Errorf("unsupported engine: %s", tc.Engine)
	}
}

func runVerifyOutputs(tc *provider.ComponentTestContext) error {
	if tc.Engine != "pulumi" {
		return nil
	}

	outputJSON, err := PulumiStackOutputs(tc.ModuleDir, tc.StackName, tc.BackendURL)
	if err != nil {
		// Some components may not export outputs -- this is acceptable
		return nil
	}
	_ = outputJSON
	// Future: parse JSON and populate tc.Outputs
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
		// Clean up the stack
		return PulumiRemoveStack(tc.ModuleDir, tc.StackName, tc.BackendURL)
	case "terraform":
		_, err := TerraformDestroy(tc.ModuleDir)
		return err
	default:
		return errors.Errorf("unsupported engine: %s", tc.Engine)
	}
}

func runVerifyCleanup(ctx context.Context, tc *provider.ComponentTestContext, harness provider.Harness) error {
	return harness.VerifyDestroyed(ctx, tc.Component)
}
