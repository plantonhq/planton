package providerparity

import "testing"

// Proven means every module the kind ships was exercised live: both for the ordinary kind, only
// the HCL module for a kind that declares OpenTofu (or Terraform) alone.
func TestE2EProofProven(t *testing.T) {
	for name, tc := range map[string]struct {
		proof E2EProof
		want  bool
	}{
		"ordinary kind, both engines":            {E2EProof{Green: true, Engines: []string{"pulumi", "terraform"}}, true},
		"ordinary kind, HCL only":                {E2EProof{Green: true, Engines: []string{"terraform"}}, false},
		"ordinary kind, two HCL engines":         {E2EProof{Green: true, Engines: []string{"terraform", "tofu"}}, false},
		"not green":                              {E2EProof{Green: false, Engines: []string{"pulumi", "terraform"}}, false},
		"OpenTofu-only kind, tofu":               {E2EProof{Green: true, Engines: []string{"tofu"}, RunsOn: []string{"tofu"}}, true},
		"OpenTofu-only kind, nothing validated":  {E2EProof{Green: true, RunsOn: []string{"tofu"}}, false},
		"OpenTofu-and-Terraform kind, terraform": {E2EProof{Green: true, Engines: []string{"terraform"}, RunsOn: []string{"terraform", "tofu"}}, true},
	} {
		if got := tc.proof.Proven(); got != tc.want {
			t.Errorf("%s: Proven() = %v, want %v", name, got, tc.want)
		}
	}
}
