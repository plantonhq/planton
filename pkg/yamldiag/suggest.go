package yamldiag

import (
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"
)

// suggestField proposes the schema field an unknown key most likely meant,
// computed against the REAL fields of the message at the failure point --
// which is what makes it work for every kind's spec, not just a hardcoded
// handful of envelope words.
func suggestField(key string, md protoreflect.MessageDescriptor) string {
	fields := md.Fields()
	candidates := make([]string, 0, fields.Len())
	for i := 0; i < fields.Len(); i++ {
		candidates = append(candidates, fields.Get(i).JSONName())
	}
	// A pure casing/underscore variant of a real field is the most common
	// near-miss; normalize both sides before measuring edit distance.
	normalizedKey := normalize(key)
	for _, candidate := range candidates {
		if normalize(candidate) == normalizedKey {
			return candidate
		}
	}
	return nearest(key, candidates)
}

// nearest returns the candidate within a small edit distance of the input,
// or "" when nothing is plausibly the intended value.
func nearest(input string, candidates []string) string {
	best, bestDistance := "", 3 // suggestions beyond 2 edits mislead more than they help
	lower := strings.ToLower(input)
	for _, candidate := range candidates {
		if d := levenshtein(lower, strings.ToLower(candidate)); d < bestDistance {
			best, bestDistance = candidate, d
		}
	}
	return best
}

func normalize(s string) string {
	return strings.ToLower(strings.NewReplacer("-", "", "_", "").Replace(s))
}

func levenshtein(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	previous := make([]int, len(b)+1)
	current := make([]int, len(b)+1)
	for j := range previous {
		previous[j] = j
	}
	for i := 1; i <= len(a); i++ {
		current[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			current[j] = minInt(previous[j]+1, current[j-1]+1, previous[j-1]+cost)
		}
		previous, current = current, previous
	}
	return previous[len(b)]
}

func minInt(nums ...int) int {
	m := nums[0]
	for _, n := range nums[1:] {
		if n < m {
			m = n
		}
	}
	return m
}
