package setdeploy

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/cli/workspace"
	"github.com/plantonhq/planton/pkg/iac/iacinput/iacinputproviderconfig"
	"github.com/plantonhq/planton/pkg/iac/provisioner"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumistack"
	"github.com/plantonhq/planton/pkg/iac/tofu/tofumodule"
	"github.com/plantonhq/planton/pkg/outputs"
	"github.com/plantonhq/planton/shared/iac/pulumi"
	"github.com/plantonhq/planton/shared/iac/terraform"
)

// EngineDeployer is the production NodeDeployer: tofu/terraform nodes run
// through the engine's single-manifest command path inside their identity-
// keyed workspaces; pulumi nodes run through the pulumi stack path (their
// state lives behind the backend URL, never in the working directory, so they
// need no workspace copy).
//
// Modules resolve once per (kind, provisioner) and sync into each node's
// workspace; Close releases whatever the resolution staged. Ambient
// credentials are the deploy posture (the wall probed them): no per-node
// provider-config files ride a set deploy in v1.
type EngineDeployer struct {
	Flags Flags

	// ResolveModuleDir, when set, answers module resolution directly — the
	// seam integration tests (and library consumers that already hold a
	// module checkout) use to bypass download/staging. Returning ok=false
	// falls through to the normal resolution ladder.
	ResolveModuleDir func(kindName string, prov provisioner.ProvisionerType) (dir string, ok bool)

	resolved map[string]resolvedModule
}

type resolvedModule struct {
	dir     string
	cleanup func() error
}

// Deploy hands one node to its engine and returns the captured outputs.
func (d *EngineDeployer) Deploy(node NodePlan, manifestPath string) (*outputs.CaptureResult, error) {
	// Sequential inits in fresh workspaces re-download providers; the
	// engine's own shared plugin cache turns that into a hardlink. Safe here
	// because set execution is sequential by design.
	ensureSharedPluginCache()

	switch node.Provisioner {
	case provisioner.ProvisionerTypePulumi:
		return d.deployPulumi(node, manifestPath)
	default:
		return d.deployHcl(node, manifestPath)
	}
}

func (d *EngineDeployer) deployHcl(node NodePlan, manifestPath string) (*outputs.CaptureResult, error) {
	binaryName := "tofu"
	if node.Provisioner == provisioner.ProvisionerTypeTerraform {
		binaryName = "terraform"
	}

	moduleDir, err := d.moduleDirFor(node.KindName, node.Provisioner)
	if err != nil {
		return nil, err
	}
	nodeDir, err := nodeWorkspaceDir(node.Identity)
	if err != nil {
		return nil, err
	}
	if err := syncModuleToWorkspace(moduleDir, nodeDir); err != nil {
		return nil, errors.Wrapf(err, "failed to prepare workspace for %s", node.Identity)
	}

	captured := &outputs.CaptureResult{}
	err = tofumodule.RunCommand(
		binaryName,
		nodeDir,
		manifestPath,
		terraform.TerraformOperationType_apply,
		nil,   // value overrides are single-resource surface; refused on sets
		true,  // the set-level approval already happened (one decision per set)
		false, // not a destroy plan
		false, // each node owns its workspace; its backend config never changes underfoot
		d.Flags.ModuleVersion,
		true, // the workspace is deliberately persistent — never cleaned up
		node.KubeContext,
		&iacinputproviderconfig.ProviderConfig{Path: "", Provider: node.Provider},
		node.TofuBackend,
		tofumodule.WithOutputCapture(captured),
	)
	if err != nil {
		return nil, err
	}
	return captured, nil
}

func (d *EngineDeployer) deployPulumi(node NodePlan, manifestPath string) (*outputs.CaptureResult, error) {
	moduleDir := ""
	if d.ResolveModuleDir != nil {
		if dir, ok := d.ResolveModuleDir(node.KindName, provisioner.ProvisionerTypePulumi); ok {
			moduleDir = dir
		}
	}
	captured := &outputs.CaptureResult{}
	err := pulumistack.Run(
		moduleDir,
		node.PulumiStackFqdn,
		manifestPath,
		pulumi.PulumiOperationType_update,
		false, // not a preview
		true,  // the set-level approval already happened
		nil,   // value overrides are single-resource surface; refused on sets
		false, // diff display is the single lane's flag
		d.Flags.ModuleVersion,
		false,
		node.KubeContext,
		"", // no iac-input file: the manifest is the input
		&iacinputproviderconfig.ProviderConfig{Path: "", Provider: node.Provider},
		pulumistack.WithBackendURL(node.PulumiBackendURL),
		pulumistack.WithOutputCapture(captured),
	)
	if err != nil {
		return nil, err
	}
	return captured, nil
}

// moduleDirFor resolves the module once per (kind, provisioner) and caches
// the result for the set's other nodes of the same kind.
func (d *EngineDeployer) moduleDirFor(kindName string, prov provisioner.ProvisionerType) (string, error) {
	if d.ResolveModuleDir != nil {
		if dir, ok := d.ResolveModuleDir(kindName, prov); ok {
			return dir, nil
		}
	}
	key := kindName + "/" + prov.String()
	if d.resolved == nil {
		d.resolved = map[string]resolvedModule{}
	}
	if m, ok := d.resolved[key]; ok {
		return m.dir, nil
	}
	pathResult, err := tofumodule.GetModulePath("", kindName, d.Flags.ModuleVersion, false)
	if err != nil {
		return "", errors.Wrapf(err, "failed to resolve the %s module for %s", prov, kindName)
	}
	m := resolvedModule{dir: pathResult.ModulePath, cleanup: func() error { return nil }}
	if pathResult.ShouldCleanup {
		m.cleanup = pathResult.CleanupFunc
	}
	d.resolved[key] = m
	return m.dir, nil
}

// Close releases every staged module resolution. Node workspaces are
// deliberately NOT removed — they hold local state and warm provider caches.
func (d *EngineDeployer) Close() {
	for _, m := range d.resolved {
		_ = m.cleanup()
	}
	d.resolved = nil
}

// ensureSharedPluginCache points the engines at one plugin cache under the
// CLI's workspace root (when the user has not chosen their own), so per-node
// workspaces hardlink providers instead of re-downloading them.
func ensureSharedPluginCache() {
	if os.Getenv("TF_PLUGIN_CACHE_DIR") != "" {
		return
	}
	root, err := workspace.GetWorkspaceDir()
	if err != nil {
		return
	}
	cacheDir := filepath.Join(root, "plugin-cache")
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return
	}
	os.Setenv("TF_PLUGIN_CACHE_DIR", cacheDir)
}
