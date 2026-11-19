// Package workloadreplicas decides how a workload module writes spec.replicas,
// so an apply never fights the controller that owns a workload's size.
//
// A Deployment's or StatefulSet's live replica count has exactly one owner.
// When the spec declares a count and no autoscaler runs, the module owns it and
// every apply holds the workload to it. When an autoscaler owns it (the kind's
// inline HPA, or an external HorizontalPodAutoscaler or KEDA ScaledObject
// targeting the workload while the spec leaves replicas unset), an apply must
// never move the live count.
//
// Leaving the field out of the applied object is not enough under server-side
// apply, which the catalog's Kubernetes provider always uses: omitting a field
// tells the API server the module no longer owns it, and when the module was its
// only owner the field is removed and Kubernetes re-defaults it to 1. An HPA owns
// nothing until it writes a different size, so a workload running at its floor
// would drop to 1 pod on the apply that turns autoscaling on. Re-sending the
// declared count instead resets the autoscaler's choice on every apply (and,
// without forced ownership, conflicts with it).
//
// The provider's ignoreChanges is the mechanism that keeps the live count in
// both cases: for an ignored path it drops the field from the patch when another
// manager owns it, and re-sends the live value when the module still does. The
// declared count is used only on create (and on a replacement), so an
// autoscaled workload starts at its floor.
//
// The OpenTofu twin writes `replicas = null` in the same cases: the
// hashicorp/kubernetes provider models replicas as optional and computed on
// kubernetes_deployment_v1 and kubernetes_stateful_set_v1, and does not use
// server-side apply, so null keeps the live count.
package workloadreplicas

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// replicasPath is the property path ignoreChanges keeps on its live value.
const replicasPath = "spec.replicas"

// Plan is how a module writes a workload's replica count.
type Plan struct {
	// Count is the replicas value the module declares, or nil to leave the field
	// out. When LiveSizeKept is set it is used only on create.
	Count *int
	// LiveSizeKept means another controller owns the live count, so the module
	// never changes it after create.
	LiveSizeKept bool
}

// For returns the plan for a workload whose spec declares `declared` replicas
// (nil when unset) and whose inline autoscaler is on or off.
//
//   - declared, no inline autoscaler: the module owns the size and declares it.
//   - inline autoscaler on: the declared count (1 when unset) is the floor the
//     workload is created at; the autoscaler owns the live count after that.
//   - unset, no inline autoscaler: Kubernetes or an external autoscaler owns the
//     size; the workload is created with Kubernetes' default of 1 pod.
func For(declared *int32, inlineAutoscaler bool) Plan {
	switch {
	case inlineAutoscaler:
		floor := 1
		if declared != nil {
			floor = int(*declared)
		}
		return Plan{Count: &floor, LiveSizeKept: true}
	case declared != nil:
		count := int(*declared)
		return Plan{Count: &count}
	default:
		return Plan{LiveSizeKept: true}
	}
}

// Replicas is the value for the workload spec's Replicas field.
func (p Plan) Replicas() pulumi.IntPtrInput {
	if p.Count == nil {
		return nil
	}
	return pulumi.IntPtr(*p.Count)
}

// Options are the resource options the workload resource is registered with.
func (p Plan) Options() []pulumi.ResourceOption {
	if !p.LiveSizeKept {
		return nil
	}
	return []pulumi.ResourceOption{pulumi.IgnoreChanges([]string{replicasPath})}
}
