package verify

import (
	"context"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// kafkaSchemaVerifier verifies a DigitalOceanDatabaseKafkaSchema via
// GET /v2/databases/{cluster_id}/schema-registry/{subject_name}. The
// (cluster, subject name) pair is the identity -- the registry's internal
// numeric schema id is discarded by the provider -- so the verifier reads
// both from the stack outputs. This live lookup matters doubly for this
// kind: the provider's own importer and destroy check are broken at the
// pin (both address an empty subject name), so the harness's verification
// is the only trustworthy existence signal.
type kafkaSchemaVerifier struct{}

func (*kafkaSchemaVerifier) IDOutputKey() string { return "subject_name" }

func (*kafkaSchemaVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	return pkgerrors.New("digitaloceandatabasekafkaschema requires the full outputs map (cluster_id + subject_name); " +
		"the harness dispatches through VerifyExistsFromOutputs")
}

func (*kafkaSchemaVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	return pkgerrors.New("digitaloceandatabasekafkaschema requires the full outputs map (cluster_id + subject_name); " +
		"the harness dispatches through VerifyAbsentFromOutputs")
}

func (v *kafkaSchemaVerifier) VerifyExistsFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	exists, err := v.existsFromOutputs(ctx, client, outputs)
	if err != nil {
		return pkgerrors.Wrap(err, "digitaloceandatabasekafkaschema verify-exists failed")
	}
	if !exists {
		return pkgerrors.Errorf("digitaloceandatabasekafkaschema %q not found after deploy", StringOutput(outputs, "subject_name"))
	}
	return nil
}

func (v *kafkaSchemaVerifier) VerifyAbsentFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	exists, err := v.existsFromOutputs(ctx, client, outputs)
	if err != nil {
		return pkgerrors.Wrap(err, "digitaloceandatabasekafkaschema verify-absent failed")
	}
	if exists {
		return &StillExistsError{Component: "digitaloceandatabasekafkaschema", ID: StringOutput(outputs, "subject_name")}
	}
	return nil
}

// existsFromOutputs reports the subject's live existence. The cluster
// vanishing is also a valid absence signal on composed teardowns.
func (*kafkaSchemaVerifier) existsFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) (bool, error) {
	clusterID := StringOutput(outputs, "cluster_id")
	subject := StringOutput(outputs, "subject_name")
	if clusterID == "" || subject == "" {
		return false, pkgerrors.Errorf("outputs must carry cluster_id and subject_name (got cluster_id=%q, subject_name=%q)", clusterID, subject)
	}
	_, _, err := client.Databases.GetKafkaSchemaRegistry(ctx, clusterID, subject)
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
