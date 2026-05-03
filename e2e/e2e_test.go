//go:build e2e

// Package e2e contains end-to-end tests that deploy real infrastructure
// using OpenMCF IaC modules and verify the results.
//
// These tests require:
//   - kind CLI installed
//   - pulumi CLI installed
//   - kubectl CLI installed
//   - Docker running
//
// Run with: go test -tags=e2e -timeout=30m -v ./e2e/...
package e2e

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/plantonhq/openmcf/e2e/framework/provider/kubernetes"
	"github.com/plantonhq/openmcf/e2e/framework/runner"
)

var (
	// testHarness is the shared kind cluster harness for all Kubernetes tests.
	testHarness *kubernetes.Harness

	// repoRoot is the absolute path to the openmcf repository root.
	repoRoot string

	// runID is a unique identifier for this test run.
	runID string

	// pulumiBackendURL is the file-based backend for Pulumi stacks.
	pulumiBackendURL string
)

func TestMain(m *testing.M) {
	// Resolve repo root (this file lives at e2e/e2e_test.go)
	var err error
	repoRoot, err = filepath.Abs(filepath.Join(".."))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve repo root: %v\n", err)
		os.Exit(1)
	}

	// Generate unique run ID
	runID = uuid.New().String()[:8]

	// Set up local Pulumi backend
	backendDir, err := os.MkdirTemp("", "openmcf-e2e-pulumi-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp backend dir: %v\n", err)
		os.Exit(1)
	}
	pulumiBackendURL = "file://" + backendDir
	defer os.RemoveAll(backendDir)

	// Log into local backend
	if err := runner.PulumiLogin(pulumiBackendURL); err != nil {
		fmt.Fprintf(os.Stderr, "failed to login to pulumi backend: %v\n", err)
		os.Exit(1)
	}

	// Create kind cluster
	clusterName := fmt.Sprintf("openmcf-e2e-%s", runID)
	testHarness = kubernetes.NewHarness(clusterName)

	ctx := context.Background()
	if err := testHarness.Setup(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "failed to create kind cluster: %v\n", err)
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Teardown kind cluster
	if err := testHarness.Teardown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to delete kind cluster: %v\n", err)
	}

	os.Exit(code)
}
