package awsdynamodbv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
)

func TestAwsDynamodbSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsDynamodbSpec Validation Suite")
}

var _ = ginkgo.Describe("AwsDynamodbSpec validations", func() {
	var spec *AwsDynamodbSpec

	ginkgo.BeforeEach(func() {
		spec = &AwsDynamodbSpec{
			Region:      "us-west-2",
			BillingMode: AwsDynamodbSpec_PAY_PER_REQUEST,
			AttributeDefinitions: []*AwsDynamodbSpec_AttributeDefinition{
				{Name: "pk", Type: AwsDynamodbSpec_S},
			},
			KeySchema: []*AwsDynamodbSpec_KeySchemaElement{
				{AttributeName: "pk", KeyType: AwsDynamodbSpec_KeySchemaElement_HASH},
			},
		}
	})

	ginkgo.It("accepts a minimal valid PAY_PER_REQUEST table", func() {
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a valid PROVISIONED table with throughput", func() {
		spec.BillingMode = AwsDynamodbSpec_PROVISIONED
		spec.ProvisionedThroughput = &AwsDynamodbSpec_ProvisionedThroughput{ReadCapacityUnits: 5, WriteCapacityUnits: 5}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// Field-level validations
	ginkgo.It("fails when attribute_definitions is empty", func() {
		spec.AttributeDefinitions = []*AwsDynamodbSpec_AttributeDefinition{}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when key_schema is empty", func() {
		spec.AttributeDefinitions = []*AwsDynamodbSpec_AttributeDefinition{{Name: "pk", Type: AwsDynamodbSpec_S}}
		spec.KeySchema = []*AwsDynamodbSpec_KeySchemaElement{}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when attribute name is empty", func() {
		spec.AttributeDefinitions = []*AwsDynamodbSpec_AttributeDefinition{{Name: "", Type: AwsDynamodbSpec_S}}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when enum values are undefined", func() {
		spec.AttributeDefinitions = []*AwsDynamodbSpec_AttributeDefinition{{Name: "pk", Type: AwsDynamodbSpec_AttributeType(99)}}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// CEL: billing mode must be non-zero
	ginkgo.It("fails when billing_mode is UNSPECIFIED", func() {
		spec.BillingMode = AwsDynamodbSpec_BILLING_MODE_UNSPECIFIED
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// CEL: PROVISIONED requires throughput > 0
	ginkgo.It("fails when PROVISIONED without throughput", func() {
		spec.BillingMode = AwsDynamodbSpec_PROVISIONED
		spec.ProvisionedThroughput = nil
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when PAY_PER_REQUEST has throughput set > 0", func() {
		spec.BillingMode = AwsDynamodbSpec_PAY_PER_REQUEST
		spec.ProvisionedThroughput = &AwsDynamodbSpec_ProvisionedThroughput{ReadCapacityUnits: 1, WriteCapacityUnits: 1}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// CEL: key schema shape
	ginkgo.It("fails when key_schema has two HASH keys", func() {
		spec.KeySchema = []*AwsDynamodbSpec_KeySchemaElement{
			{AttributeName: "pk", KeyType: AwsDynamodbSpec_KeySchemaElement_HASH},
			{AttributeName: "sk", KeyType: AwsDynamodbSpec_KeySchemaElement_HASH},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// Projection INCLUDE requires non_key_attributes
	ginkgo.It("fails when projection type is INCLUDE without attributes", func() {
		gsi := &AwsDynamodbSpec_GlobalSecondaryIndex{
			Name:       "gsi1",
			KeySchema:  []*AwsDynamodbSpec_KeySchemaElement{{AttributeName: "pk", KeyType: AwsDynamodbSpec_KeySchemaElement_HASH}},
			Projection: &AwsDynamodbSpec_Projection{Type: AwsDynamodbSpec_INCLUDE, NonKeyAttributes: []string{}},
		}
		spec.GlobalSecondaryIndexes = []*AwsDynamodbSpec_GlobalSecondaryIndex{gsi}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// Streams rule
	ginkgo.It("fails when stream_enabled is true but view type is UNSPECIFIED", func() {
		spec.StreamEnabled = true
		spec.StreamViewType = AwsDynamodbSpec_STREAM_VIEW_TYPE_UNSPECIFIED
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when stream_enabled is false but view type is set", func() {
		spec.StreamEnabled = false
		spec.StreamViewType = AwsDynamodbSpec_KEYS_ONLY
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// TTL rule
	ginkgo.It("fails when TTL enabled without attribute_name", func() {
		spec.Ttl = &AwsDynamodbSpec_TimeToLive{Enabled: true, AttributeName: ""}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when TTL disabled but attribute_name is set", func() {
		spec.Ttl = &AwsDynamodbSpec_TimeToLive{Enabled: false, AttributeName: "expiresAt"}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// GSI throughput rule matches billing mode
	ginkgo.It("fails when PROVISIONED table has GSI without throughput", func() {
		spec.BillingMode = AwsDynamodbSpec_PROVISIONED
		spec.ProvisionedThroughput = &AwsDynamodbSpec_ProvisionedThroughput{ReadCapacityUnits: 5, WriteCapacityUnits: 5}
		gsi := &AwsDynamodbSpec_GlobalSecondaryIndex{
			Name:                  "gsi1",
			KeySchema:             []*AwsDynamodbSpec_KeySchemaElement{{AttributeName: "pk", KeyType: AwsDynamodbSpec_KeySchemaElement_HASH}},
			Projection:            &AwsDynamodbSpec_Projection{Type: AwsDynamodbSpec_KEYS_ONLY_PROJECTION},
			ProvisionedThroughput: nil,
		}
		spec.GlobalSecondaryIndexes = []*AwsDynamodbSpec_GlobalSecondaryIndex{gsi}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})
