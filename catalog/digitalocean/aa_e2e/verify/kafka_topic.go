package verify

import (
	"context"

	"github.com/digitalocean/godo"
	pkgerrors "github.com/pkg/errors"
)

// kafkaTopicVerifier verifies a DigitalOceanDatabaseKafkaTopic via
// GET /v2/databases/{cluster_id}/topics/{name}. The API has no standalone
// topic id -- the (cluster, topic name) pair is the identity -- so the
// verifier reads both from the stack outputs.
type kafkaTopicVerifier struct{}

func (*kafkaTopicVerifier) IDOutputKey() string { return "topic_name" }

func (*kafkaTopicVerifier) VerifyExists(ctx context.Context, client *godo.Client, id string) error {
	return pkgerrors.New("digitaloceandatabasekafkatopic requires the full outputs map (cluster_id + topic_name); " +
		"the harness dispatches through VerifyExistsFromOutputs")
}

func (*kafkaTopicVerifier) VerifyAbsent(ctx context.Context, client *godo.Client, id string) error {
	return pkgerrors.New("digitaloceandatabasekafkatopic requires the full outputs map (cluster_id + topic_name); " +
		"the harness dispatches through VerifyAbsentFromOutputs")
}

func (v *kafkaTopicVerifier) VerifyExistsFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	exists, err := v.existsFromOutputs(ctx, client, outputs)
	if err != nil {
		return pkgerrors.Wrap(err, "digitaloceandatabasekafkatopic verify-exists failed")
	}
	if !exists {
		return pkgerrors.Errorf("digitaloceandatabasekafkatopic %q not found after deploy", StringOutput(outputs, "topic_name"))
	}
	return nil
}

func (v *kafkaTopicVerifier) VerifyAbsentFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) error {
	exists, err := v.existsFromOutputs(ctx, client, outputs)
	if err != nil {
		return pkgerrors.Wrap(err, "digitaloceandatabasekafkatopic verify-absent failed")
	}
	if exists {
		return &StillExistsError{Component: "digitaloceandatabasekafkatopic", ID: StringOutput(outputs, "topic_name")}
	}
	return nil
}

// existsFromOutputs reports the topic's live existence. The cluster
// vanishing is also a valid absence signal on composed teardowns.
func (*kafkaTopicVerifier) existsFromOutputs(ctx context.Context, client *godo.Client, outputs map[string]interface{}) (bool, error) {
	clusterID := StringOutput(outputs, "cluster_id")
	name := StringOutput(outputs, "topic_name")
	if clusterID == "" || name == "" {
		return false, pkgerrors.Errorf("outputs must carry cluster_id and topic_name (got cluster_id=%q, topic_name=%q)", clusterID, name)
	}
	_, _, err := client.Databases.GetTopic(ctx, clusterID, name)
	if err != nil {
		if isNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
