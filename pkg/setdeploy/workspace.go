package setdeploy

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/cli/workspace"
	"github.com/plantonhq/planton/pkg/iac/tofu/tfoverride"
	"github.com/plantonhq/planton/pkg/manifestgraph"
)

// Identity-keyed node workspaces. Module cache directories are shared per
// kind (the zip cache) or disposable per run (the staging copy) — running set
// nodes in them directly would collide same-kind local state files and leak
// one node's backend.tf into the next node's init. Every tofu/terraform node
// therefore executes in its own stable copy of the module, keyed by the
// node's graph identity under the CLI's workspace root:
//
//	~/.planton/setdeploy/<env or "default">/<kind>/<slug>/
//
// Stability is the point: local backend state written here persists across
// runs, so "run the same command again" is a true recovery story, and remote
// -backend nodes keep their .terraform provider dir warm between runs.

// nodeWorkspaceDir returns (creating parents as needed) the node's stable
// workspace path. The slug and env are joined into a filesystem path, so an
// unlawful one (a dot, a slash, "..") is refused rather than joined: a
// lawful slug is lowercase letters and digits joined by single hyphens and
// can never climb out of the workspace root.
func nodeWorkspaceDir(id manifestgraph.Identity) (string, error) {
	if !manifestgraph.IsLawfulSlug(id.Slug) {
		return "", errors.Errorf("%s slug %q is not a lawful slug: %s", id.Kind, id.Slug, manifestgraph.SlugRule)
	}
	if id.Env != "" && !manifestgraph.IsLawfulSlug(id.Env) {
		return "", errors.Errorf("%s %q env %q is not a lawful slug: %s", id.Kind, id.Slug, id.Env, manifestgraph.SlugRule)
	}
	root, err := workspace.GetWorkspaceDir()
	if err != nil {
		return "", err
	}
	env := id.Env
	if env == "" {
		env = "default"
	}
	dir := filepath.Join(root, "setdeploy", env, strings.ToLower(id.Kind.String()), id.Slug)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", errors.Wrapf(err, "failed to create node workspace %s", dir)
	}
	return dir, nil
}

// syncModuleToWorkspace copies the resolved module's files into the node
// workspace, overwriting module files while PRESERVING the node's own runtime
// artifacts: state files, the .terraform directory, and any leftover override
// file are never copied FROM the source (a shared cache may hold another
// run's leavings) and never deleted in the destination (they are this node's
// state and warm cache).
func syncModuleToWorkspace(moduleDir, nodeDir string) error {
	return filepath.Walk(moduleDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(moduleDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		if skipInModuleSync(rel, info.IsDir()) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		target := filepath.Join(nodeDir, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target, info.Mode())
	})
}

// skipInModuleSync names what never travels from a module source into a node
// workspace: engine runtime state (another consumer's, or a previous run's in
// a shared cache) and the per-run provider-override file.
func skipInModuleSync(rel string, isDir bool) bool {
	base := filepath.Base(rel)
	if isDir && base == ".terraform" {
		return true
	}
	if strings.HasSuffix(base, ".tfstate") || strings.Contains(base, ".tfstate.") {
		return true
	}
	if base == tfoverride.OverrideFileName {
		return true
	}
	return false
}

func copyFile(src, dst string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode.Perm())
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
