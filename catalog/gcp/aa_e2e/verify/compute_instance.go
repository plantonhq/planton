package verify

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
)

// computeInstanceVerifier probes a Compute Engine VM via the compute API
// using the (project, zone, name) triple from the outputs. Posture
// assertions confirm the platform attribution labels landed (the
// cross-engine label-parity canary — the Terraform module historically
// stamped a different label set than Pulumi, so this is a permanently
// guarded regression) and that the machine-type output matches the live
// instance.
type computeInstanceVerifier struct{}

func (v *computeInstanceVerifier) IDOutputKey() string { return "instance_name" }

func (v *computeInstanceVerifier) project(svc *Services, outputs map[string]string) string {
	if p := outputs["project_id"]; p != "" {
		return p
	}
	return svc.Project
}

func (v *computeInstanceVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["instance_name"]
	zone := outputs["zone"]
	project := v.project(svc, outputs)

	instance, err := svc.Compute.Instances.Get(project, zone, name).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "compute instance %s/%s not found after deploy", zone, name)
	}

	// The platform attribution labels are the cross-engine parity canary:
	// a missing set means one engine stamped labels and the other did not.
	if instance.Labels["planton-ai_resource"] != "true" {
		return errors.Errorf("compute instance %s missing the planton-ai_resource attribution label after deploy (labels: %v)", name, instance.Labels)
	}

	// machine_type is exported as the plain type name; the live attribute
	// is a full URL — suffix-match so both engines' exports verify.
	if mt := outputs["machine_type"]; mt != "" && !hasPathSuffix(instance.MachineType, mt) {
		return errors.Errorf("compute instance %s machine_type output %q does not match live machine type %q", name, mt, instance.MachineType)
	}

	return nil
}

func (v *computeInstanceVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["instance_name"]
	zone := outputs["zone"]
	project := v.project(svc, outputs)

	_, err := svc.Compute.Instances.Get(project, zone, name).Context(ctx).Do()
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing compute instance %s after destroy", name)
	}
	return errors.Errorf("compute instance %s still exists after destroy", name)
}

// hasPathSuffix reports whether the URL-ish live value ends with the plain
// name (…/machineTypes/e2-micro matches "e2-micro"), tolerating exact
// equality too.
func hasPathSuffix(live, plain string) bool {
	if live == plain {
		return true
	}
	suffix := "/" + plain
	return len(live) > len(suffix) && live[len(live)-len(suffix):] == suffix
}
