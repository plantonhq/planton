package gcpregionnetworkendpointgroupv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpRegionNetworkEndpointGroupSpec Suite")
}

func strPtr(v string) *string { return &v }

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpRegionNetworkEndpointGroupSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Minimal valid NEG: a regional SERVERLESS NEG fronting a Cloud Run service.
	minimal := func() *GcpRegionNetworkEndpointGroup {
		return &GcpRegionNetworkEndpointGroup{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpRegionNetworkEndpointGroup",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-neg",
			},
			Spec: &GcpRegionNetworkEndpointGroupSpec{
				Region: "us-central1",
				ServerlessTarget: &GcpRegionNetworkEndpointGroupSpec_CloudRun{
					CloudRun: &GcpRegionNetworkEndpointGroupCloudRun{
						Service: litRef("my-service"),
					},
				},
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal serverless Cloud Run NEG", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a Cloud Run NEG using url_mask instead of service", func() {
		target := minimal()
		target.Spec.ServerlessTarget = &GcpRegionNetworkEndpointGroupSpec_CloudRun{CloudRun: &GcpRegionNetworkEndpointGroupCloudRun{
			UrlMask: "<service>.example.com",
		}}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a Cloud Functions NEG", func() {
		target := minimal()
		target.Spec.ServerlessTarget = nil
		target.Spec.ServerlessTarget = &GcpRegionNetworkEndpointGroupSpec_CloudFunction{CloudFunction: &GcpRegionNetworkEndpointGroupCloudFunction{
			Function: litRef("my-func"),
		}}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an empty App Engine block (default app)", func() {
		target := minimal()
		target.Spec.ServerlessTarget = nil
		target.Spec.ServerlessTarget = &GcpRegionNetworkEndpointGroupSpec_AppEngine{AppEngine: &GcpRegionNetworkEndpointGroupAppEngine{}}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an explicit SERVERLESS type with a name", func() {
		target := minimal()
		target.Spec.NetworkEndpointType = strPtr("SERVERLESS")
		target.Spec.NetworkEndpointGroupName = "web-neg"
		target.Spec.Description = "fronts the web Cloud Run service"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a PSC NEG with target service, network, and subnetwork", func() {
		target := minimal()
		target.Spec.ServerlessTarget = nil
		target.Spec.NetworkEndpointType = strPtr("PRIVATE_SERVICE_CONNECT")
		target.Spec.PscTargetService = "asia-northeast3-cloudkms.googleapis.com"
		target.Spec.Network = litRef("projects/p/global/networks/default")
		target.Spec.Subnetwork = litRef("projects/p/regions/us-central1/subnetworks/sub")
		target.Spec.PscData = &GcpRegionNetworkEndpointGroupPscData{ProducerPort: "80"}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an INTERNET_FQDN_PORT NEG with network", func() {
		target := minimal()
		target.Spec.ServerlessTarget = nil
		target.Spec.NetworkEndpointType = strPtr("INTERNET_FQDN_PORT")
		target.Spec.Network = litRef("projects/p/global/networks/default")
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a spec without a region", func() {
		target := minimal()
		target.Spec.Region = ""
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an uppercase region", func() {
		target := minimal()
		target.Spec.Region = "US-CENTRAL1"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid network_endpoint_group_name", func() {
		target := minimal()
		target.Spec.NetworkEndpointGroupName = "Invalid_Name"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "RFC1035")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject an invalid network_endpoint_type", func() {
		target := minimal()
		target.Spec.NetworkEndpointType = strPtr("SERVERFUL")
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a SERVERLESS NEG with no serverless block", func() {
		target := minimal()
		target.Spec.ServerlessTarget = nil
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "exactly one")).To(gomega.BeTrue())
	})

	ginkgo.It("carries exactly one serverless target by construction (setting a second arm replaces the first)", func() {
		target := minimal()
		target.Spec.ServerlessTarget = &GcpRegionNetworkEndpointGroupSpec_AppEngine{AppEngine: &GcpRegionNetworkEndpointGroupAppEngine{}}
		gomega.Expect(target.Spec.GetCloudRun()).To(gomega.BeNil())
		gomega.Expect(target.Spec.GetAppEngine()).ToNot(gomega.BeNil())
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should reject a serverless block on a PSC NEG", func() {
		target := minimal()
		target.Spec.NetworkEndpointType = strPtr("PRIVATE_SERVICE_CONNECT")
		target.Spec.PscTargetService = "asia-northeast3-cloudkms.googleapis.com"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "SERVERLESS")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a PSC NEG without psc_target_service", func() {
		target := minimal()
		target.Spec.ServerlessTarget = nil
		target.Spec.NetworkEndpointType = strPtr("PRIVATE_SERVICE_CONNECT")
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "psc_target_service")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a cloud_run block with neither service nor url_mask", func() {
		target := minimal()
		target.Spec.ServerlessTarget = &GcpRegionNetworkEndpointGroupSpec_CloudRun{CloudRun: &GcpRegionNetworkEndpointGroupCloudRun{}}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "service or url_mask")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject subnetwork on a serverless NEG", func() {
		target := minimal()
		target.Spec.Subnetwork = litRef("projects/p/regions/us-central1/subnetworks/sub")
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a wrong kind constant", func() {
		target := minimal()
		target.Kind = "GcpRegionNeg"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept each deletion_policy value", func() {
		for _, v := range []string{"DELETE", "PREVENT", "ABANDON"} {
			target := minimal()
			target.Spec.DeletionPolicy = v
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "KEEP"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})
})
