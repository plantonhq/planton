package eject

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/shared/catalogkind"
)

// engineModulePath is the Go module every catalog module builds inside — and
// therefore the dependency the ejected copy's synthesized go.mod requires:
// the kind's generated stubs and the engine helpers (iac-input
// loading) resolve from its published releases.
const engineModulePath = "github.com/plantonhq/planton"

// goDirectiveVersion is written into the synthesized go.mod. It matches the
// engine module's own go directive: requiring the engine module means
// building with at least the toolchain the engine builds with.
const goDirectiveVersion = "1.26"

// releaseTagPattern recognizes the plain release tags published to the Go
// module proxy. Anything else (a branch name, "dev", a kind point-tag)
// cannot be a go.mod require version.
var releaseTagPattern = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

// preparePulumiCopy turns the copied catalog module into a standalone,
// user-owned Go module: self-imports rewritten, go.mod synthesized,
// Pulumi.yaml guaranteed, and (unless skipped) dependencies resolved.
func preparePulumiCopy(outputDir string, kind catalogkind.CatalogKind, kindName string, in Input, result *Result) error {
	if in.GoModulePath == "" {
		return errors.New("a Go module path is required to eject a pulumi module")
	}
	if strings.ContainsAny(in.GoModulePath, " \t\"'") {
		return errors.Errorf("the Go module path %q is not a valid module path — use a slash-separated path like github.com/your-org/%s-module", in.GoModulePath, strings.ToLower(kindName))
	}

	if err := rewriteSelfImports(outputDir, kind, kindName, in.GoModulePath); err != nil {
		return err
	}

	if err := writeGoMod(outputDir, in.GoModulePath, result.SourceVersion); err != nil {
		return err
	}

	if err := ensurePulumiYaml(outputDir, kindName); err != nil {
		return err
	}

	if !in.SkipGoModTidy {
		ran, err := runGoModTidy(outputDir)
		if err != nil {
			return err
		}
		result.GoModTidyRan = ran
	}

	return nil
}

// rewriteSelfImports repoints the module's self-imports (the entrypoint
// importing its own `module` subpackage) from the catalog's import path onto
// the user's module path.
//
// Keeping the original path is impossible, not merely inelegant: the
// published engine module contains every catalog package, so a copy that
// declares (or requires) the original path makes each self-import resolve in
// two modules at once and Go rejects the build with an ambiguous-import
// error. The copy must own a distinct module path, which also states the
// truth — the ejected module belongs to the user now.
func rewriteSelfImports(outputDir string, kind catalogkind.CatalogKind, kindName, goModulePath string) error {
	providerSegment := strings.ReplaceAll(catalogkindreflect.GetProvider(kind).String(), "_", "")
	originalPrefix := fmt.Sprintf("%s/catalog/%s/%s/iac/pulumi", engineModulePath, providerSegment, strings.ToLower(kindName))

	return filepath.WalkDir(outputDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".go") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return errors.Wrapf(err, "failed to read %s", path)
		}

		rewritten := strings.ReplaceAll(string(content), originalPrefix, goModulePath)
		if rewritten == string(content) {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return errors.Wrapf(err, "failed to stat %s", path)
		}
		if err := os.WriteFile(path, []byte(rewritten), info.Mode().Perm()); err != nil {
			return errors.Wrapf(err, "failed to write %s", path)
		}
		return nil
	})
}

// writeGoMod synthesizes the ejected module's go.mod. When the source
// version is a published release tag it is required explicitly, pinning the
// stubs and engine helpers to exactly the release the module was ejected
// from; otherwise the require is left for `go mod tidy` to resolve.
func writeGoMod(outputDir, goModulePath, sourceVersion string) error {
	var b strings.Builder
	fmt.Fprintf(&b, "module %s\n\ngo %s\n", goModulePath, goDirectiveVersion)
	if releaseTagPattern.MatchString(sourceVersion) {
		fmt.Fprintf(&b, "\nrequire %s %s\n", engineModulePath, sourceVersion)
	}

	goModPath := filepath.Join(outputDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(b.String()), 0644); err != nil {
		return errors.Wrapf(err, "failed to write %s", goModPath)
	}
	return nil
}

// ensurePulumiYaml guarantees the project file every pulumi module needs.
// A handful of catalog modules ship without one (execution paths synthesize
// it around the compiled binary), but a user-owned source module must carry
// its own — the declared `go` runtime is what module verification and the
// deploy path key on.
func ensurePulumiYaml(outputDir, kindName string) error {
	pulumiYamlPath := filepath.Join(outputDir, "Pulumi.yaml")
	if _, err := os.Stat(pulumiYamlPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return errors.Wrapf(err, "failed to check for %s", pulumiYamlPath)
	}

	content := fmt.Sprintf("name: %s\nruntime: go\ndescription: Customized %s module\n", strings.ToLower(kindName), kindName)
	if err := os.WriteFile(pulumiYamlPath, []byte(content), 0644); err != nil {
		return errors.Wrapf(err, "failed to write %s", pulumiYamlPath)
	}
	cliprint.PrintInfo("Generated Pulumi.yaml (the catalog module did not carry one)")
	return nil
}

// runGoModTidy resolves the synthesized go.mod's dependency graph so the
// copy builds immediately. Returns false without error when no go toolchain
// is on PATH — the contract notes carry the command to run later.
//
// The workspace overlay is disabled for the child process: an eject target
// inside some Go workspace must still resolve from its own go.mod graph,
// exactly as it will everywhere else.
func runGoModTidy(outputDir string) (bool, error) {
	goBinary, err := exec.LookPath("go")
	if err != nil {
		cliprint.PrintInfo("Go toolchain not found on PATH — run 'go mod tidy' in the ejected module before building")
		return false, nil
	}

	cliprint.PrintStep("Resolving module dependencies (go mod tidy)...")
	cmd := exec.Command(goBinary, "mod", "tidy")
	cmd.Dir = outputDir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	if out, err := cmd.CombinedOutput(); err != nil {
		return false, errors.Wrapf(err, "go mod tidy failed in %s:\n%s", outputDir, string(out))
	}
	return true, nil
}
