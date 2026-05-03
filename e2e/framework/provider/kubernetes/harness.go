// Package kubernetes implements the E2E provider harness for Kubernetes,
// using kind (Kubernetes IN Docker) as the test cluster substrate.
package kubernetes

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
)

// Harness manages a kind cluster lifecycle for Kubernetes E2E tests.
type Harness struct {
	clusterName    string
	kubeconfigPath string
	tempDir        string
}

// NewHarness creates a Kubernetes test harness with the given cluster name.
func NewHarness(clusterName string) *Harness {
	return &Harness{
		clusterName: clusterName,
	}
}

// Setup creates a kind cluster and stores the kubeconfig path.
func (h *Harness) Setup(ctx context.Context) error {
	tmpDir, err := os.MkdirTemp("", "openmcf-e2e-*")
	if err != nil {
		return errors.Wrap(err, "failed to create temp directory for kubeconfig")
	}
	h.tempDir = tmpDir
	h.kubeconfigPath = filepath.Join(tmpDir, "kubeconfig")

	args := []string{
		"create", "cluster",
		"--name", h.clusterName,
		"--kubeconfig", h.kubeconfigPath,
		"--wait", "120s",
	}

	cmd := exec.CommandContext(ctx, "kind", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Stdout = os.Stdout

	fmt.Printf("  [kind] Creating cluster %q...\n", h.clusterName)
	start := time.Now()

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "kind create cluster failed: %s", stderr.String())
	}

	fmt.Printf("  [kind] Cluster %q ready in %s\n", h.clusterName, time.Since(start).Round(time.Second))

	// Set KUBECONFIG globally so Pulumi's kubernetes provider picks it up
	os.Setenv("KUBECONFIG", h.kubeconfigPath)

	return nil
}

// Teardown deletes the kind cluster and removes temp files.
func (h *Harness) Teardown(ctx context.Context) error {
	fmt.Printf("  [kind] Deleting cluster %q...\n", h.clusterName)

	args := []string{"delete", "cluster", "--name", h.clusterName}
	cmd := exec.CommandContext(ctx, "kind", args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "kind delete cluster failed: %s", stderr.String())
	}

	// Clean up temp directory
	if h.tempDir != "" {
		os.RemoveAll(h.tempDir)
	}

	// Unset KUBECONFIG
	os.Unsetenv("KUBECONFIG")

	return nil
}

// KubeconfigPath returns the path to the kubeconfig for the kind cluster.
func (h *Harness) KubeconfigPath() string {
	return h.kubeconfigPath
}

// ClusterName returns the kind cluster name.
func (h *Harness) ClusterName() string {
	return h.clusterName
}

// VerifyDeployed delegates to resource-type-specific verification based on component.
func (h *Harness) VerifyDeployed(ctx context.Context, component string, outputs map[string]interface{}) error {
	verifier, err := getVerifier(component)
	if err != nil {
		return err
	}
	return verifier.VerifyExists(ctx, h.kubeconfigPath)
}

// VerifyDestroyed delegates to resource-type-specific verification based on component.
func (h *Harness) VerifyDestroyed(ctx context.Context, component string) error {
	verifier, err := getVerifier(component)
	if err != nil {
		return err
	}
	return verifier.VerifyAbsent(ctx, h.kubeconfigPath)
}
