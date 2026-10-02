package verify

import (
	"context"

	"github.com/pkg/errors"
)

// cloudDeployDeliveryPipelineVerifier probes a delivery pipeline through the
// Cloud Deploy API by the name output
// (projects/{p}/locations/{l}/deliveryPipelines/{id}). Posture assertions:
// Google's uid matches the exported one, and the platform attribution
// labels landed. The pipeline's automations go with it on destroy (the
// provider deletes with force=true), so absence of the pipeline covers them.
type cloudDeployDeliveryPipelineVerifier struct{}

func (v *cloudDeployDeliveryPipelineVerifier) IDOutputKey() string { return "name" }

func (v *cloudDeployDeliveryPipelineVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	pipeline, err := svc.CloudDeploy.Projects.Locations.DeliveryPipelines.Get(name).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "delivery pipeline %s not found after deploy", name)
	}
	if uid := outputs["uid"]; uid != "" && pipeline.Uid != uid {
		return errors.Errorf("delivery pipeline %s uid output %q does not match live uid %q", name, uid, pipeline.Uid)
	}
	if pipeline.Labels["planton-ai_resource"] != "true" {
		return errors.Errorf("delivery pipeline %s missing the planton-ai_resource attribution label after deploy (labels: %v)", name, pipeline.Labels)
	}
	return nil
}

func (v *cloudDeployDeliveryPipelineVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	_, err := svc.CloudDeploy.Projects.Locations.DeliveryPipelines.Get(name).Context(ctx).Do()
	if err != nil {
		if isGoogleNotFound(err) {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing delivery pipeline %s after destroy", name)
	}
	return errors.Errorf("delivery pipeline %s still exists after destroy", name)
}
