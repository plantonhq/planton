package catalogschema

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/plantonhq/planton/pkg/catalogkindreflect"
	"github.com/plantonhq/planton/shared/catalogkind"
)

const testProviderName = "_test"

// contractFiles are the four versioned-contract protos every kind ships;
// the artifact must carry each one's schema document for the kind's version.
var contractFiles = []string{"spec.proto.json", "api.proto.json", "input.proto.json", "outputs.proto.json"}

// VerifyArtifact proves catalog-schemas.zip serves EXACTLY the user-facing
// catalog before it may ship:
//
//  1. every user-facing registry kind has all four contract schema documents
//     (spec/api/input/outputs) at its declared version directory,
//  2. no _test-provider content is present (the synthetic proving kinds
//     never reach a serving surface),
//  3. every document parses as the published ProtoFile contract with its
//     package name and rawContent present -- an entry that cannot be read
//     back, or that lost its authored source, is unshippable,
//  4. the walk is guarded against vacuous passes.
//
// Extra documents beyond the per-kind contracts (provider.proto, shared
// catalog protos) are legitimate cargo: coverage is one-directional by
// design -- the registry must be served; the tree may carry more.
func VerifyArtifact(zipPath string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("open %s: %w", zipPath, err)
	}
	defer reader.Close()

	var problems []string
	present := map[string]bool{}

	for _, entry := range reader.File {
		if strings.HasPrefix(entry.Name, "catalog/"+testProviderName+"/") {
			problems = append(problems, fmt.Sprintf(
				"%s: the _test proving kinds never ship on a serving surface", entry.Name))
			continue
		}
		if !strings.HasSuffix(entry.Name, ".proto.json") {
			problems = append(problems, fmt.Sprintf(
				"%s: the artifact carries only *.proto.json schema documents", entry.Name))
			continue
		}
		present[entry.Name] = true

		if problem := readbackProblem(entry); problem != "" {
			problems = append(problems, problem)
		}
	}

	checked := 0
	for _, kind := range catalogkindreflect.KindsList() {
		if kind == catalogkind.CatalogKind_unspecified {
			continue
		}
		provider := catalogkindreflect.GetProvider(kind)
		if provider.String() == testProviderName {
			continue
		}
		version, err := catalogkindreflect.KindVersion(kind)
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s: %v", kind, err))
			continue
		}
		providerDir := strings.ReplaceAll(provider.String(), "_", "")
		kindDir := strings.ToLower(catalogkindreflect.ExtractKindNameByKind(kind))
		for _, contract := range contractFiles {
			path := fmt.Sprintf("catalog/%s/%s/%s/%s", providerDir, kindDir, version, contract)
			if !present[path] {
				problems = append(problems, fmt.Sprintf(
					"%s: registry kind %s has no schema document at its declared version",
					path, kind))
			}
		}
		checked++
	}
	if checked == 0 {
		return fmt.Errorf("schema verification checked zero kinds -- the registry walk is broken")
	}

	if len(problems) > 0 {
		sort.Strings(problems)
		return fmt.Errorf("catalog-schemas artifact failed verification (%d kinds checked):\n  %s",
			checked, strings.Join(problems, "\n  "))
	}
	return nil
}

// readbackProblem strict-decodes one schema document against the published
// contract and demands the two facts every consumer relies on: the proto
// package identity and the authored source (rawContent).
func readbackProblem(entry *zip.File) string {
	r, err := entry.Open()
	if err != nil {
		return fmt.Sprintf("%s: %v", entry.Name, err)
	}
	defer r.Close()
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Sprintf("%s: %v", entry.Name, err)
	}

	var schema ProtoFile
	if err := json.Unmarshal(data, &schema); err != nil {
		return fmt.Sprintf("%s: does not parse as the published ProtoFile contract: %v", entry.Name, err)
	}
	if schema.PackageName == "" {
		return fmt.Sprintf("%s: schema document carries no proto package name", entry.Name)
	}
	if schema.RawContent == "" {
		return fmt.Sprintf("%s: schema document lost its authored source (rawContent)", entry.Name)
	}
	return ""
}
