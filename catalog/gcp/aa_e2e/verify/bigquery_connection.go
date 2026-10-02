package verify

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// bigQueryConnectionGet reads a BigQuery connection by its full name from
// the https://bigqueryconnection.googleapis.com/v1/ endpoint (the pinned
// client library carries no BigQuery Connection client).
func bigQueryConnectionGet(ctx context.Context, svc *Services, name string) (map[string]interface{}, int, error) {
	obj := map[string]interface{}{}
	status, err := googleRestGet(ctx, svc, "bigquery connection",
		fmt.Sprintf("https://bigqueryconnection.googleapis.com/v1/%s", name), &obj)
	if err != nil {
		return nil, status, err
	}
	return obj, status, nil
}

// bigQueryConnectionVerifier probes the connection: it reads back under its
// name, and for the cloud_resource arm the exported service account is the
// one Google created.
type bigQueryConnectionVerifier struct{}

func (v *bigQueryConnectionVerifier) IDOutputKey() string { return "name" }

func (v *bigQueryConnectionVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	connection, _, err := bigQueryConnectionGet(ctx, svc, name)
	if err != nil {
		return errors.Wrapf(err, "bigquery connection %s not found after deploy", name)
	}
	if exported := outputs["cloud_resource_service_account_id"]; exported != "" {
		cloudResource, _ := connection["cloudResource"].(map[string]interface{})
		if live := stringField(cloudResource, "serviceAccountId"); live != exported {
			return errors.Errorf("bigquery connection %s cloud_resource_service_account_id output %q does not match the live %q", name, exported, live)
		}
	}
	return nil
}

func (v *bigQueryConnectionVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return nil
	}
	_, status, err := bigQueryConnectionGet(ctx, svc, name)
	return restAbsent("bigquery connection", name, status, err)
}
