package workloadreplicas

import "testing"

func int32Ptr(v int32) *int32 { return &v }

// The ownership table is pinned offline: who owns the live count decides
// whether the module declares a size, and whether it ever changes it again.
func TestFor(t *testing.T) {
	cases := []struct {
		name             string
		declared         *int32
		inlineAutoscaler bool
		wantCount        *int
		wantKept         bool
	}{
		{name: "declared, no autoscaler: the module owns the size", declared: int32Ptr(3), wantCount: intPtr(3)},
		{name: "declared zero is a size, not unset", declared: int32Ptr(0), wantCount: intPtr(0)},
		{name: "inline autoscaler: created at the declared floor, live count kept", declared: int32Ptr(3), inlineAutoscaler: true, wantCount: intPtr(3), wantKept: true},
		{name: "inline autoscaler, floor unset: created at 1, live count kept", inlineAutoscaler: true, wantCount: intPtr(1), wantKept: true},
		{name: "unset, no autoscaler: left to Kubernetes or an external autoscaler", wantKept: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := For(c.declared, c.inlineAutoscaler)
			if (got.Count == nil) != (c.wantCount == nil) || (got.Count != nil && *got.Count != *c.wantCount) {
				t.Fatalf("Count = %v, want %v", deref(got.Count), deref(c.wantCount))
			}
			if got.LiveSizeKept != c.wantKept {
				t.Fatalf("LiveSizeKept = %t, want %t", got.LiveSizeKept, c.wantKept)
			}
			if (got.Replicas() == nil) != (c.wantCount == nil) {
				t.Fatalf("Replicas() declared = %t, want %t", got.Replicas() != nil, c.wantCount != nil)
			}
			if (len(got.Options()) > 0) != c.wantKept {
				t.Fatalf("Options() ignore the live count = %t, want %t", len(got.Options()) > 0, c.wantKept)
			}
		})
	}
}

func intPtr(v int) *int { return &v }

func deref(p *int) any {
	if p == nil {
		return "unset"
	}
	return *p
}
