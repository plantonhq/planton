package verify

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// sccGet reads one Security Command Center v2 configuration (a
// notification config, mute config, or BigQuery export) by its full name
// from https://securitycenter.googleapis.com/v2/ -- the pinned client
// library has no securitycenter/v2 package.
func sccGet(ctx context.Context, svc *Services, what, name string) (map[string]interface{}, int, error) {
	obj := map[string]interface{}{}
	status, err := googleRestGet(ctx, svc, what, fmt.Sprintf("https://securitycenter.googleapis.com/v2/%s", name), &obj)
	if err != nil {
		return nil, status, err
	}
	return obj, status, nil
}

// sccAbsent is the destroy verdict every SCC configuration shares: a 404.
func sccAbsent(ctx context.Context, svc *Services, what, name string) error {
	if name == "" {
		return nil
	}
	_, status, err := sccGet(ctx, svc, what, name)
	return restAbsent(what, name, status, err)
}

// sccNotificationConfigVerifier probes a notification config: it reads
// back under its name with a Pub/Sub topic, and its publishing service
// account equals the service_account output.
type sccNotificationConfigVerifier struct{}

func (v *sccNotificationConfigVerifier) IDOutputKey() string { return "name" }

func (v *sccNotificationConfigVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	config, _, err := sccGet(ctx, svc, "scc notification config", name)
	if err != nil {
		return errors.Wrapf(err, "scc notification config %s not found after deploy", name)
	}
	if stringField(config, "pubsubTopic") == "" {
		return errors.Errorf("scc notification config %s has no pubsub topic after deploy", name)
	}
	if account := stringField(config, "serviceAccount"); account == "" || account != outputs["service_account"] {
		return errors.Errorf("scc notification config %s publishes as %q, but the service_account output is %q",
			name, account, outputs["service_account"])
	}
	return nil
}

func (v *sccNotificationConfigVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	return sccAbsent(ctx, svc, "scc notification config", outputs["name"])
}

// sccMuteConfigVerifier probes a mute config: it reads back under its name
// with its filter and a STATIC or DYNAMIC type.
type sccMuteConfigVerifier struct{}

func (v *sccMuteConfigVerifier) IDOutputKey() string { return "name" }

func (v *sccMuteConfigVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	config, _, err := sccGet(ctx, svc, "scc mute config", name)
	if err != nil {
		return errors.Wrapf(err, "scc mute config %s not found after deploy", name)
	}
	if stringField(config, "filter") == "" {
		return errors.Errorf("scc mute config %s has no filter after deploy", name)
	}
	if muteType := stringField(config, "type"); muteType != "STATIC" && muteType != "DYNAMIC" {
		return errors.Errorf("scc mute config %s has type %q after deploy", name, muteType)
	}
	return nil
}

func (v *sccMuteConfigVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	return sccAbsent(ctx, svc, "scc mute config", outputs["name"])
}

// sccBigQueryExportVerifier probes a BigQuery export: it reads back under
// its name with a dataset, and its writing principal equals the principal
// output.
type sccBigQueryExportVerifier struct{}

func (v *sccBigQueryExportVerifier) IDOutputKey() string { return "name" }

func (v *sccBigQueryExportVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	export, _, err := sccGet(ctx, svc, "scc bigquery export", name)
	if err != nil {
		return errors.Wrapf(err, "scc bigquery export %s not found after deploy", name)
	}
	if stringField(export, "dataset") == "" {
		return errors.Errorf("scc bigquery export %s has no dataset after deploy", name)
	}
	if principal := stringField(export, "principal"); principal == "" || principal != outputs["principal"] {
		return errors.Errorf("scc bigquery export %s writes as %q, but the principal output is %q",
			name, principal, outputs["principal"])
	}
	return nil
}

func (v *sccBigQueryExportVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	return sccAbsent(ctx, svc, "scc bigquery export", outputs["name"])
}
