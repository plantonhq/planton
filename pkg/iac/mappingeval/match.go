package mappingeval

import (
	"sort"
)

// InstanceMatch pairs one ground-truth instance with the proposal instance
// that best accounts for the same provider resources.
type InstanceMatch struct {
	GroundTruth *GroundTruthInstance
	Proposed    *ProposedInstance
	// Overlap is how many of the ground-truth instance's claims the
	// proposed instance also claims -- the evidence the match rests on.
	Overlap int
}

// matchInstances pairs ground-truth and proposal instances by kind plus
// claim overlap -- never by name, because a blind proposer cannot know the
// seeded fixtures' internal names. Greedy on descending overlap (ties break
// on ground-truth name, then proposal name, for determinism); a pair needs
// at least one shared claim, and each instance matches at most once.
//
// Greedy is sufficient here, not a compromise: claims are near-disjoint by
// construction (an provider resource has one owner), so the assignment problem
// degenerates -- a wrong greedy pick can only happen when instances
// genuinely share claims, which the grouping axis then reports as the
// duplicate-claim defect it is.
func matchInstances(gt *GroundTruth, proposal *LoadedProposal) []InstanceMatch {
	type candidate struct {
		gtIndex, propIndex, overlap int
	}
	var candidates []candidate
	for gi := range gt.Instances {
		gtInstance := &gt.Instances[gi]
		gtClaims := map[AccountResourceRef]bool{}
		for _, claim := range gtInstance.Claims {
			gtClaims[claim] = true
		}
		for pi := range proposal.Instances {
			propInstance := &proposal.Instances[pi]
			if propInstance.Kind != gtInstance.Kind {
				continue
			}
			overlap := 0
			for _, claim := range propInstance.Claims {
				if gtClaims[claim] {
					overlap++
				}
			}
			if overlap > 0 {
				candidates = append(candidates, candidate{gi, pi, overlap})
			}
		}
	}
	sort.Slice(candidates, func(a, b int) bool {
		ca, cb := candidates[a], candidates[b]
		if ca.overlap != cb.overlap {
			return ca.overlap > cb.overlap
		}
		if gt.Instances[ca.gtIndex].Name != gt.Instances[cb.gtIndex].Name {
			return gt.Instances[ca.gtIndex].Name < gt.Instances[cb.gtIndex].Name
		}
		return proposal.Instances[ca.propIndex].Name < proposal.Instances[cb.propIndex].Name
	})

	usedGT := map[int]bool{}
	usedProp := map[int]bool{}
	var matches []InstanceMatch
	for _, c := range candidates {
		if usedGT[c.gtIndex] || usedProp[c.propIndex] {
			continue
		}
		usedGT[c.gtIndex] = true
		usedProp[c.propIndex] = true
		matches = append(matches, InstanceMatch{
			GroundTruth: &gt.Instances[c.gtIndex],
			Proposed:    &proposal.Instances[c.propIndex],
			Overlap:     c.overlap,
		})
	}
	return matches
}
