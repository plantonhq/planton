//go:build !codegen
// +build !codegen

// The E2E-proof join for the public report: read the provider's component
// E2E profiles and reduce them to the renderer's plain proof inputs. Lives
// beside the renderer (not inside it) so RenderPublicReport stays pure and
// the CLI and the drift gate share exactly one join -- two copies of this
// logic would eventually disagree about what "proven" means.

package providerparity

import (
	"os"
	"sort"
	"strings"

	"github.com/plantonhq/planton/pkg/crkreflect"
	"github.com/plantonhq/planton/pkg/e2e/profile"
	componentv1 "github.com/plantonhq/planton/qa/componente2eprofile/v1"
	sharedpb "github.com/plantonhq/planton/shared"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
)

// E2EProof is one kind's live-proof status, reduced from its E2E profile.
type E2EProof struct {
	// Green is the profile's recorded status: the kind's live lanes passed
	// as of the last proof run.
	Green bool
	// Engines are the validated provisioners, sorted (e.g. pulumi,
	// terraform).
	Engines []string
	// RunsOn are the engines the kind declares it runs on
	// (kind_meta.provisioners), sorted; empty for the ordinary kind, which
	// runs on every engine and ships both a Pulumi and an HCL module.
	RunsOn []string
}

// Proven is THE definition of proven for every report surface: green live
// runs that exercised every module the kind ships -- its Pulumi module and
// its HCL module for the ordinary kind, only the HCL module for a kind that
// declares OpenTofu (or Terraform) alone. The HCL module is proven by either
// of the engines that run it. A green profile that leaves a shipped module
// unexercised is progress, not proof.
func (p E2EProof) Proven() bool {
	if !p.Green {
		return false
	}
	validated := moduleFamilies(p.Engines)
	needed := map[string]bool{"pulumi": true, "hcl": true}
	if len(p.RunsOn) > 0 {
		needed = moduleFamilies(p.RunsOn)
	}
	for family := range needed {
		if !validated[family] {
			return false
		}
	}
	return true
}

// moduleFamilies maps engine names to the modules they run: pulumi runs the
// Pulumi module, and tofu and terraform both run the one HCL module.
func moduleFamilies(engines []string) map[string]bool {
	families := map[string]bool{}
	for _, engine := range engines {
		switch engine {
		case "pulumi":
			families["pulumi"] = true
		case "tofu", "terraform":
			families["hcl"] = true
		}
	}
	return families
}

// BuildE2EProofs reads every component E2E profile of one provider and maps
// kind name -> proof status. A provider without an E2E harness (no
// aa_e2e/profile.yaml yet) yields nil proofs, not an error: the report then
// simply shows nothing proven, which is the truth.
func BuildE2EProofs(repoRoot string, provider cloudresourcekind.CloudResourceProvider) (map[string]E2EProof, error) {
	providerName := crkreflect.ProviderDirName(provider)
	if _, err := os.Stat(profile.ProviderProfilePath(repoRoot, providerName)); os.IsNotExist(err) {
		return nil, nil
	}
	result, err := profile.Discover(repoRoot, providerName, profile.FilterOpts{})
	if err != nil {
		return nil, err
	}
	proofs := map[string]E2EProof{}
	for _, ce := range result.Components {
		spec := ce.Profile.Spec
		if spec == nil {
			continue
		}
		kind := crkreflect.KindFromString(ce.Name)
		if kind == cloudresourcekind.CloudResourceKind_unspecified {
			continue
		}
		engines := make([]string, 0, len(spec.ValidatedProvisioners))
		for _, vp := range spec.ValidatedProvisioners {
			engines = append(engines, strings.ToLower(sharedpb.IacProvisioner_name[int32(vp)]))
		}
		sort.Strings(engines)
		declared, err := crkreflect.Provisioners(kind)
		if err != nil {
			return nil, err
		}
		runsOn := make([]string, 0, len(declared))
		for _, p := range declared {
			runsOn = append(runsOn, strings.ToLower(p.String()))
		}
		sort.Strings(runsOn)
		proofs[kind.String()] = E2EProof{
			Green:   spec.Status == componentv1.ComponentE2EProfileSpec_green,
			Engines: engines,
			RunsOn:  runsOn,
		}
	}
	return proofs, nil
}
