package gcpvertexaisearchdataconnectorv1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpVertexAiSearchDataConnectorSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpVertexAiSearchDataConnectorSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// A Jira connector syncing projects and issues daily -- the provider's
	// own example, with credentials named as Secret Manager secrets.
	jira := func() *GcpVertexAiSearchDataConnector {
		return &GcpVertexAiSearchDataConnector{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpVertexAiSearchDataConnector",
			Metadata:   &shared.CloudResourceMetadata{Name: "jira-federated"},
			Spec: &GcpVertexAiSearchDataConnectorSpec{
				Location:   "global",
				DataSource: "jira",
				Params: map[string]string{
					"instance_uri":  "https://example.atlassian.net",
					"instance_id":   "projects/ai-project/secrets/jira-instance-id",
					"client_id":     "projects/ai-project/secrets/jira-client-id",
					"client_secret": "projects/ai-project/secrets/jira-client-secret",
					"refresh_token": "projects/ai-project/secrets/jira-refresh-token",
					"auth_type":     "OAUTH",
				},
				RefreshInterval: "86400s",
				Entities: []*GcpVertexAiSearchDataConnectorEntity{
					{EntityName: "project"},
					{EntityName: "issue", Params: `{"inclusion_filters":{"projectKey":["OPS"]}}`},
				},
			},
		}
	}

	ginkgo.It("should accept the Jira connector", func() {
		gomega.Expect(validator.Validate(jira())).To(gomega.Succeed())
	})

	ginkgo.It("should accept every lever: actions, destinations, static IP, CMEK, and json_params", func() {
		msg := jira()
		msg.Spec.ProjectId = litRef("ai-project")
		msg.Spec.CollectionId = "jira-collection"
		msg.Spec.CollectionDisplayName = "Jira Federated"
		msg.Spec.DataSourceVersion = proto.Int32(3)
		msg.Spec.IncrementalRefreshInterval = "21600s"
		msg.Spec.IncrementalSyncDisabled = true
		msg.Spec.AutoRunDisabled = true
		msg.Spec.SyncMode = "PERIODIC"
		msg.Spec.ConnectorModes = []string{"FEDERATED", "ACTIONS"}
		msg.Spec.StaticIpEnabled = true
		msg.Spec.KmsKeyName = litRef("projects/ai-project/locations/global/keyRings/ai/cryptoKeys/search")
		msg.Spec.Entities[0].KeyPropertyMappings = map[string]string{"summary": "title"}
		msg.Spec.DestinationConfigs = []*GcpVertexAiSearchDataConnectorDestinationConfig{{
			Key:          "url",
			Destinations: []*GcpVertexAiSearchDataConnectorDestination{{Host: "https://example.atlassian.net", Port: proto.Int32(443)}},
			Params:       `{"destination_type":"private"}`,
		}}
		msg.Spec.ActionConfig = &GcpVertexAiSearchDataConnectorActionConfig{
			ActionParams:        map[string]string{"instance_uri": "https://example.atlassian.net", "auth_type": "OAUTH"},
			CreateBapConnection: true,
		}
		msg.Spec.BapConfig = &GcpVertexAiSearchDataConnectorBapConfig{
			SupportedConnectorModes: []string{"ACTIONS"},
			EnabledActions:          []string{"create_issue", "create_comment"},
		}
		msg.Spec.DeletionPolicy = "ABANDON"
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())

		msg = jira()
		msg.Spec.Params = nil
		msg.Spec.JsonParams = `{"instance_uri":"https://example.atlassian.net","auth_type":"OAUTH"}`
		gomega.Expect(validator.Validate(msg)).To(gomega.Succeed())
	})

	ginkgo.It("should require location from Google's list, a source, and a refresh interval", func() {
		msg := jira()
		msg.Spec.Location = "us-central1"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = jira()
		msg.Spec.DataSource = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = jira()
		msg.Spec.DataSource = "Jira Cloud"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = jira()
		msg.Spec.RefreshInterval = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = jira()
		msg.Spec.RefreshInterval = "24h"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should require exactly one of params or json_params", func() {
		msg := jira()
		msg.Spec.JsonParams = `{"a":"b"}`
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = jira()
		msg.Spec.Params = nil
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject unknown modes, a bad collection id, an empty entity name, and a bad port", func() {
		msg := jira()
		msg.Spec.ConnectorModes = []string{"REALTIME"}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = jira()
		msg.Spec.SyncMode = "BATCH"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = jira()
		msg.Spec.CollectionId = "Jira_Collection"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = jira()
		msg.Spec.Entities[0].EntityName = ""
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = jira()
		msg.Spec.DestinationConfigs = []*GcpVertexAiSearchDataConnectorDestinationConfig{{
			Destinations: []*GcpVertexAiSearchDataConnectorDestination{{Host: "h", Port: proto.Int32(70000)}},
		}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
		msg = jira()
		msg.Spec.BapConfig = &GcpVertexAiSearchDataConnectorBapConfig{SupportedConnectorModes: []string{"FEDERATED"}}
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an unknown deletion policy", func() {
		msg := jira()
		msg.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(msg)).ToNot(gomega.Succeed())
	})
})
