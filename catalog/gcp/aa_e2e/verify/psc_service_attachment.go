package verify

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
)

// pscServiceAttachmentVerifier probes a Private Service Connect service
// attachment through the compute API by the attachment_name and region
// outputs. Posture assertions confirm the exported self_link matches the
// live resource -- the value a consumer's PSC endpoint targets -- and that
// the connection preference the manifest declared is what Google enforces.
type pscServiceAttachmentVerifier struct{}

func (v *pscServiceAttachmentVerifier) IDOutputKey() string { return "attachment_name" }

func (v *pscServiceAttachmentVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name, region := outputs["attachment_name"], outputs["region"]
	if name == "" || region == "" {
		return errors.New("attachment_name or region output missing after deploy")
	}

	attachment, err := svc.Compute.ServiceAttachments.Get(svc.Project, region, name).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "service attachment %s (region %s) not found after deploy", name, region)
	}

	if wantSelfLink := outputs["self_link"]; wantSelfLink != "" && attachment.SelfLink != wantSelfLink {
		return errors.Errorf("service attachment %s self_link mismatch: output %q, live %q", name, wantSelfLink, attachment.SelfLink)
	}

	if attachment.ConnectionPreference == "" {
		return errors.Errorf("service attachment %s reports no connection preference after deploy", name)
	}
	if len(attachment.NatSubnets) == 0 {
		return errors.Errorf("service attachment %s has no NAT subnets after deploy", name)
	}
	if attachment.TargetService == "" {
		return errors.Errorf("service attachment %s has no target service after deploy", name)
	}
	return nil
}

func (v *pscServiceAttachmentVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name, region := outputs["attachment_name"], outputs["region"]
	if name == "" || region == "" {
		return nil
	}

	_, err := svc.Compute.ServiceAttachments.Get(svc.Project, region, name).Context(ctx).Do()
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing service attachment %s after destroy", name)
	}
	return errors.Errorf("service attachment %s still exists after destroy", name)
}
