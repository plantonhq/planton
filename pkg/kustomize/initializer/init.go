//go:build !codegen
// +build !codegen

package initializer

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/pkg/kustomize/schema"
)

const openapiBlock = "\nopenapi:\n  path: ../../" + schema.SchemaFileName + "\n"

// InitResult reports what happened when initializing a single _kustomize directory.
type InitResult struct {
	Dir             string
	SchemaWritten   bool
	OverlaysUpdated []string
}

// InitDir generates the universal Planton kustomize schema and wires it into
// the given _kustomize directory. It writes the schema file at the directory
// root and adds the openapi: reference to every overlay kustomization.yaml.
func InitDir(kustomizeDir string) (*InitResult, error) {
	absDir, err := filepath.Abs(kustomizeDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve absolute path")
	}

	info, err := os.Stat(absDir)
	if err != nil || !info.IsDir() {
		return nil, errors.Errorf("not a valid directory: %s", absDir)
	}

	schemaBytes, err := schema.Generate()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate schema")
	}

	result := &InitResult{Dir: absDir}

	schemaPath := filepath.Join(absDir, schema.SchemaFileName)
	if err := os.WriteFile(schemaPath, append(schemaBytes, '\n'), 0644); err != nil {
		return nil, errors.Wrap(err, "failed to write schema file")
	}
	result.SchemaWritten = true

	overlaysDir := filepath.Join(absDir, "overlays")
	overlayEntries, err := os.ReadDir(overlaysDir)
	if err != nil {
		if os.IsNotExist(err) {
			return result, nil
		}
		return nil, errors.Wrap(err, "failed to read overlays directory")
	}

	for _, entry := range overlayEntries {
		if !entry.IsDir() {
			continue
		}
		kustomizationPath := filepath.Join(overlaysDir, entry.Name(), "kustomization.yaml")
		updated, err := ensureOpenapiRef(kustomizationPath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to update %s", kustomizationPath)
		}
		if updated {
			result.OverlaysUpdated = append(result.OverlaysUpdated, entry.Name())
		}
	}

	return result, nil
}

// ScanAndInit walks rootDir looking for directories named "_kustomize" and
// runs InitDir on each one found.
func ScanAndInit(rootDir string) ([]*InitResult, error) {
	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve absolute path")
	}

	var dirs []string
	err = filepath.WalkDir(absRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() && d.Name() == "_kustomize" {
			dirs = append(dirs, path)
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to walk directory tree")
	}

	var results []*InitResult
	for _, dir := range dirs {
		res, err := InitDir(dir)
		if err != nil {
			return results, errors.Wrapf(err, "failed to init %s", dir)
		}
		results = append(results, res)
	}

	return results, nil
}

// ensureOpenapiRef checks if a kustomization.yaml file already contains an
// openapi: directive. If not, it inserts the block after "kind: Kustomization".
// Returns true if the file was modified.
func ensureOpenapiRef(path string) (bool, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	text := string(content)

	if strings.Contains(text, "openapi:") {
		return false, nil
	}

	const anchor = "kind: Kustomization"
	idx := strings.Index(text, anchor)
	if idx == -1 {
		return false, errors.Errorf("could not find %q in %s", anchor, path)
	}

	insertPos := idx + len(anchor)
	// Skip to end of the anchor line
	if nlPos := strings.Index(text[insertPos:], "\n"); nlPos >= 0 {
		insertPos += nlPos + 1
	}

	updated := text[:insertPos] + openapiBlock + text[insertPos:]

	if err := os.WriteFile(path, []byte(updated), 0644); err != nil {
		return false, errors.Wrap(err, "failed to write updated kustomization.yaml")
	}

	return true, nil
}
