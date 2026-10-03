package verify

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/api/googleapi"
)

// monitoringNotificationChannelVerifier probes a Cloud Monitoring
// notification channel by its server-assigned resource name and confirms
// the channel exists with a concrete type. Verifiers are outputs-driven
// and scenario-agnostic, so the disabled-channel scenario's enabled=false
// posture is proven by a live API read during the proof lane, not here —
// enabled is deliberately NOT an output (outputs are composition
// handles, not test hooks). Never assert verification_status: email
// channels report UNVERIFIED until a human verifies, by design.
type monitoringNotificationChannelVerifier struct{}

func (v *monitoringNotificationChannelVerifier) IDOutputKey() string { return "channel_name" }

func (v *monitoringNotificationChannelVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["channel_name"]
	channel, err := svc.Monitoring.Projects.NotificationChannels.Get(name).Context(ctx).Do()
	if err != nil {
		return errors.Wrapf(err, "notification channel %s not found after deploy", name)
	}
	if channel.Type == "" {
		return errors.Errorf("notification channel %s reports no type", name)
	}
	return nil
}

func (v *monitoringNotificationChannelVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["channel_name"]
	_, err := svc.Monitoring.Projects.NotificationChannels.Get(name).Context(ctx).Do()
	if err != nil {
		var apiErr *googleapi.Error
		if errors.As(err, &apiErr) && apiErr.Code == 404 {
			return nil
		}
		return errors.Wrapf(err, "unexpected error probing notification channel %s after destroy", name)
	}
	return errors.Errorf("notification channel %s still exists after destroy", name)
}
