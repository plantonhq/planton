package verify

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// managedKafkaGet reads one Managed Service for Apache Kafka resource by its
// full name (clusters, topics, ACLs, Connect clusters, connectors) from the
// global https://managedkafka.googleapis.com/v1/ endpoint. The decoded body
// is a generic map; the status code tells a 404 from any other error.
func managedKafkaGet(ctx context.Context, svc *Services, name string) (map[string]interface{}, int, error) {
	obj := map[string]interface{}{}
	status, err := googleRestGet(ctx, svc, "managed kafka resource",
		fmt.Sprintf("https://managedkafka.googleapis.com/v1/%s", name), &obj)
	if err != nil {
		return nil, status, err
	}
	return obj, status, nil
}

// managedKafkaAbsent is the destroy verdict every Managed Kafka verifier
// shares.
func managedKafkaAbsent(ctx context.Context, svc *Services, what, name string) error {
	if name == "" {
		return nil
	}
	_, status, err := managedKafkaGet(ctx, svc, name)
	return restAbsent(what, name, status, err)
}

// managedKafkaClusterVerifier probes the cluster: it reads back ACTIVE under
// its name, carries the platform attribution label (the cross-engine label
// canary), and matches the exported id.
type managedKafkaClusterVerifier struct{}

func (v *managedKafkaClusterVerifier) IDOutputKey() string { return "name" }

func (v *managedKafkaClusterVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	cluster, _, err := managedKafkaGet(ctx, svc, name)
	if err != nil {
		return errors.Wrapf(err, "managed kafka cluster %s not found after deploy", name)
	}
	if state := stringField(cluster, "state"); state != "ACTIVE" {
		return errors.Errorf("managed kafka cluster %s is %q after deploy, want ACTIVE", name, state)
	}
	labels, _ := cluster["labels"].(map[string]interface{})
	if labels["planton-ai_resource"] != "true" {
		return errors.Errorf("managed kafka cluster %s missing the planton-ai_resource attribution label after deploy", name)
	}
	if got := outputs["cluster_id"]; got != lastPathSegment(name) {
		return errors.Errorf("managed kafka cluster %s cluster_id output %q does not match its name", name, got)
	}
	return nil
}

func (v *managedKafkaClusterVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	return managedKafkaAbsent(ctx, svc, "managed kafka cluster", outputs["name"])
}

// managedKafkaTopicVerifier probes the topic under its name.
type managedKafkaTopicVerifier struct{}

func (v *managedKafkaTopicVerifier) IDOutputKey() string { return "name" }

func (v *managedKafkaTopicVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	if _, _, err := managedKafkaGet(ctx, svc, name); err != nil {
		return errors.Wrapf(err, "managed kafka topic %s not found after deploy", name)
	}
	if got := outputs["topic_id"]; got != lastPathSegment(name) {
		return errors.Errorf("managed kafka topic %s topic_id output %q does not match its name", name, got)
	}
	return nil
}

func (v *managedKafkaTopicVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	return managedKafkaAbsent(ctx, svc, "managed kafka topic", outputs["name"])
}

// managedKafkaAclVerifier probes the ACL: it reads back under its name with
// the resource type the modules exported.
type managedKafkaAclVerifier struct{}

func (v *managedKafkaAclVerifier) IDOutputKey() string { return "name" }

func (v *managedKafkaAclVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	acl, _, err := managedKafkaGet(ctx, svc, name)
	if err != nil {
		return errors.Wrapf(err, "managed kafka acl %s not found after deploy", name)
	}
	if got, want := outputs["resource_type"], stringField(acl, "resourceType"); got != want {
		return errors.Errorf("managed kafka acl %s resource_type output %q does not match the live %q", name, got, want)
	}
	return nil
}

func (v *managedKafkaAclVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	return managedKafkaAbsent(ctx, svc, "managed kafka acl", outputs["name"])
}

// managedKafkaConnectClusterVerifier probes the Connect cluster: ACTIVE
// under its name with the platform attribution label.
type managedKafkaConnectClusterVerifier struct{}

func (v *managedKafkaConnectClusterVerifier) IDOutputKey() string { return "name" }

func (v *managedKafkaConnectClusterVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	connectCluster, _, err := managedKafkaGet(ctx, svc, name)
	if err != nil {
		return errors.Wrapf(err, "managed kafka connect cluster %s not found after deploy", name)
	}
	if state := stringField(connectCluster, "state"); state != "ACTIVE" {
		return errors.Errorf("managed kafka connect cluster %s is %q after deploy, want ACTIVE", name, state)
	}
	labels, _ := connectCluster["labels"].(map[string]interface{})
	if labels["planton-ai_resource"] != "true" {
		return errors.Errorf("managed kafka connect cluster %s missing the planton-ai_resource attribution label after deploy", name)
	}
	return nil
}

func (v *managedKafkaConnectClusterVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	return managedKafkaAbsent(ctx, svc, "managed kafka connect cluster", outputs["name"])
}

// managedKafkaConnectorVerifier probes the connector under its name.
type managedKafkaConnectorVerifier struct{}

func (v *managedKafkaConnectorVerifier) IDOutputKey() string { return "name" }

func (v *managedKafkaConnectorVerifier) VerifyExists(ctx context.Context, svc *Services, outputs map[string]string) error {
	name := outputs["name"]
	if name == "" {
		return errors.New("name output missing after deploy")
	}
	if _, _, err := managedKafkaGet(ctx, svc, name); err != nil {
		return errors.Wrapf(err, "managed kafka connector %s not found after deploy", name)
	}
	if got := outputs["connector_id"]; got != lastPathSegment(name) {
		return errors.Errorf("managed kafka connector %s connector_id output %q does not match its name", name, got)
	}
	return nil
}

func (v *managedKafkaConnectorVerifier) VerifyAbsent(ctx context.Context, svc *Services, outputs map[string]string) error {
	return managedKafkaAbsent(ctx, svc, "managed kafka connector", outputs["name"])
}
