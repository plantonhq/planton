package confluentkafkav1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/apis/dev/planton/shared"
)

func TestConfluentKafka(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "ConfluentKafka Suite")
}

var _ = ginkgo.Describe("ConfluentKafka Custom Validation Tests", func() {
	var input *ConfluentKafka

	ginkgo.BeforeEach(func() {
		input = &ConfluentKafka{
			ApiVersion: "confluent.planton.dev/v1",
			Kind:       "ConfluentKafka",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-kafka-cluster",
			},
			Spec: &ConfluentKafkaSpec{
				Cloud:         "AWS",
				Region:        "us-east-2",
				Availability:  "MULTI_ZONE",
				EnvironmentId: "env-12345",
				ClusterType:   "STANDARD",
			},
		}
	})

	ginkgo.Describe("When valid input is passed", func() {
		ginkgo.Context("AWS Standard Cluster", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Cloud = "AWS"
				input.Spec.Region = "us-east-2"
				input.Spec.Availability = "MULTI_ZONE"
				input.Spec.ClusterType = "STANDARD"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("GCP Basic Cluster", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Cloud = "GCP"
				input.Spec.Region = "us-central1"
				input.Spec.Availability = "SINGLE_ZONE"
				input.Spec.ClusterType = "BASIC"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("AZURE Enterprise Cluster", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Cloud = "AZURE"
				input.Spec.Region = "eastus"
				input.Spec.Availability = "MULTI_ZONE"
				input.Spec.ClusterType = "ENTERPRISE"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("Dedicated Cluster with CKU", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Cloud = "AWS"
				input.Spec.Region = "us-west-2"
				input.Spec.Availability = "MULTI_ZONE"
				input.Spec.ClusterType = "DEDICATED"
				input.Spec.DedicatedConfig = &ConfluentKafkaDedicatedConfig{
					Cku: 2,
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("Cluster with Network Config", func() {
			ginkgo.It("should not return a validation error", func() {
				input.Spec.Cloud = "AWS"
				input.Spec.Region = "us-east-1"
				input.Spec.Availability = "MULTI_ZONE"
				input.Spec.ClusterType = "ENTERPRISE"
				input.Spec.NetworkConfig = &ConfluentKafkaNetworkConfig{
					NetworkId: "n-abc123",
				}
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})

		ginkgo.Context("Cluster with legacy availability value", func() {
			ginkgo.It("should not return a validation error for LOW", func() {
				input.Spec.Availability = "LOW"
				input.Spec.ClusterType = "BASIC"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})

			ginkgo.It("should not return a validation error for HIGH", func() {
				input.Spec.Availability = "HIGH"
				input.Spec.ClusterType = "BASIC"
				err := protovalidate.Validate(input)
				gomega.Expect(err).To(gomega.BeNil())
			})
		})
	})

	ginkgo.Describe("Cloud Field Validation", func() {
		ginkgo.It("should fail validation if the cloud field is invalid", func() {
			input.Spec.Cloud = "IBM"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail validation if the cloud field is empty", func() {
			input.Spec.Cloud = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("Region Field Validation", func() {
		ginkgo.It("should fail validation if the region field is empty", func() {
			input.Spec.Region = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept any non-empty region string", func() {
			input.Spec.Region = "custom-region-1"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("Availability Field Validation", func() {
		ginkgo.It("should fail validation if the availability field is invalid", func() {
			input.Spec.Availability = "UNSUPPORTED_ZONE"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should fail validation if the availability field is empty", func() {
			input.Spec.Availability = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})
	})

	ginkgo.Describe("Environment ID Field Validation", func() {
		ginkgo.It("should fail validation if the environment_id field is empty", func() {
			input.Spec.EnvironmentId = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept any non-empty environment_id string", func() {
			input.Spec.EnvironmentId = "env-prod-xyz-789"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("Cluster Type Field Validation", func() {
		ginkgo.It("should fail validation if the cluster_type field is invalid", func() {
			input.Spec.ClusterType = "INVALID_TYPE"
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should allow the cluster_type field to be omitted", func() {
			input.Spec.ClusterType = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept BASIC cluster type", func() {
			input.Spec.ClusterType = "BASIC"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept STANDARD cluster type", func() {
			input.Spec.ClusterType = "STANDARD"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept ENTERPRISE cluster type", func() {
			input.Spec.ClusterType = "ENTERPRISE"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should accept DEDICATED cluster type", func() {
			input.Spec.ClusterType = "DEDICATED"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("Dedicated Config Validation", func() {
		ginkgo.It("should fail validation if CKU is less than 1", func() {
			input.Spec.ClusterType = "DEDICATED"
			input.Spec.DedicatedConfig = &ConfluentKafkaDedicatedConfig{
				Cku: 0,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept valid CKU values", func() {
			input.Spec.ClusterType = "DEDICATED"
			input.Spec.DedicatedConfig = &ConfluentKafkaDedicatedConfig{
				Cku: 4,
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("Network Config Validation", func() {
		ginkgo.It("should fail validation if network_id is empty", func() {
			input.Spec.NetworkConfig = &ConfluentKafkaNetworkConfig{
				NetworkId: "",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).NotTo(gomega.BeNil())
		})

		ginkgo.It("should accept valid network_id", func() {
			input.Spec.NetworkConfig = &ConfluentKafkaNetworkConfig{
				NetworkId: "n-private-123",
			}
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("Display Name Validation", func() {
		ginkgo.It("should accept a custom display name", func() {
			input.Spec.DisplayName = "My Production Kafka Cluster"
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})

		ginkgo.It("should allow display_name to be omitted", func() {
			input.Spec.DisplayName = ""
			err := protovalidate.Validate(input)
			gomega.Expect(err).To(gomega.BeNil())
		})
	})
})
