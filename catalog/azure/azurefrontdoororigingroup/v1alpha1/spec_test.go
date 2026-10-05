package azurefrontdoororigingroupv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAzureFrontDoorOriginGroupSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AzureFrontDoorOriginGroupSpec Validation Tests")
}

func literal(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

const profileId = "/subscriptions/s/resourceGroups/rg/providers/Microsoft.Cdn/profiles/planton-fd"

// minimal valid spec: a group deploying Azure's load-balancing defaults
// with probing disabled.
func minimalSpec() *AzureFrontDoorOriginGroup {
	return &AzureFrontDoorOriginGroup{
		ApiVersion: "azure.planton.dev/v1alpha1",
		Kind:       "AzureFrontDoorOriginGroup",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-front-door-origin-group",
		},
		Spec: &AzureFrontDoorOriginGroupSpec{
			ProfileId:       literal(profileId),
			OriginGroupName: "test-origin-group",
		},
	}
}

func validHealthProbe() *AzureFrontDoorOriginGroupHealthProbe {
	return &AzureFrontDoorOriginGroupHealthProbe{
		Protocol:          AzureFrontDoorOriginGroupHealthProbeProtocol_HTTPS,
		IntervalInSeconds: 30,
	}
}

var _ = ginkgo.Describe("AzureFrontDoorOriginGroupSpec Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept a minimal origin group", func() {
			gomega.Expect(protovalidate.Validate(minimalSpec())).To(gomega.BeNil())
		})

		ginkgo.It("should accept full load-balancing settings at their boundaries", func() {
			input := minimalSpec()
			input.Spec.LoadBalancing = &AzureFrontDoorOriginGroupLoadBalancing{
				SampleSize:                      proto.Int32(255),
				SuccessfulSamplesRequired:       proto.Int32(0),
				AdditionalLatencyInMilliseconds: proto.Int32(1000),
			}
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept a full health probe", func() {
			input := minimalSpec()
			probe := validHealthProbe()
			probe.RequestType = AzureFrontDoorOriginGroupHealthProbeRequestType_GET
			probe.Path = proto.String("/healthz")
			input.Spec.HealthProbe = probe
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept probe interval boundaries 1 and 255", func() {
			for _, interval := range []int32{1, 255} {
				input := minimalSpec()
				probe := validHealthProbe()
				probe.IntervalInSeconds = interval
				input.Spec.HealthProbe = probe
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "interval %d must be accepted", interval)
			}
		})

		ginkgo.It("should accept the Http probe protocol", func() {
			input := minimalSpec()
			probe := validHealthProbe()
			probe.Protocol = AzureFrontDoorOriginGroupHealthProbeProtocol_HTTP
			input.Spec.HealthProbe = probe
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept session affinity disabled", func() {
			input := minimalSpec()
			input.Spec.SessionAffinityEnabled = proto.Bool(false)
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})

		ginkgo.It("should accept traffic-restore boundaries 0 and 50", func() {
			for _, minutes := range []int32{0, 50} {
				input := minimalSpec()
				input.Spec.RestoreTrafficTimeToHealedOrNewEndpointInMinutes = proto.Int32(minutes)
				gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil(), "restore time %d must be accepted", minutes)
			}
		})

		ginkgo.It("should accept origin group name boundaries (2 and 90 characters)", func() {
			input := minimalSpec()
			input.Spec.OriginGroupName = "ab"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
			input.Spec.OriginGroupName = "a" + strings.Repeat("b", 88) + "c"
			gomega.Expect(protovalidate.Validate(input)).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject a missing profile reference", func() {
			input := minimalSpec()
			input.Spec.ProfileId = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing origin group name", func() {
			input := minimalSpec()
			input.Spec.OriginGroupName = ""
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an origin group name over 90 characters", func() {
			input := minimalSpec()
			input.Spec.OriginGroupName = "a" + strings.Repeat("b", 89) + "c"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an origin group name with a trailing hyphen", func() {
			input := minimalSpec()
			input.Spec.OriginGroupName = "group-"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject sample size above 255", func() {
			input := minimalSpec()
			input.Spec.LoadBalancing = &AzureFrontDoorOriginGroupLoadBalancing{
				SampleSize: proto.Int32(256),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject successful samples above 255", func() {
			input := minimalSpec()
			input.Spec.LoadBalancing = &AzureFrontDoorOriginGroupLoadBalancing{
				SuccessfulSamplesRequired: proto.Int32(256),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject additional latency above 1000 ms", func() {
			input := minimalSpec()
			input.Spec.LoadBalancing = &AzureFrontDoorOriginGroupLoadBalancing{
				AdditionalLatencyInMilliseconds: proto.Int32(1001),
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a health probe without a protocol", func() {
			input := minimalSpec()
			input.Spec.HealthProbe = &AzureFrontDoorOriginGroupHealthProbe{
				IntervalInSeconds: 30,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a health probe without an interval", func() {
			input := minimalSpec()
			input.Spec.HealthProbe = &AzureFrontDoorOriginGroupHealthProbe{
				Protocol: AzureFrontDoorOriginGroupHealthProbeProtocol_HTTPS,
			}
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a probe interval above 255 seconds", func() {
			input := minimalSpec()
			probe := validHealthProbe()
			probe.IntervalInSeconds = 256
			input.Spec.HealthProbe = probe
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a probe path not starting with '/'", func() {
			input := minimalSpec()
			probe := validHealthProbe()
			probe.Path = proto.String("healthz")
			input.Spec.HealthProbe = probe
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an undefined probe request type", func() {
			input := minimalSpec()
			probe := validHealthProbe()
			probe.RequestType = AzureFrontDoorOriginGroupHealthProbeRequestType(99)
			input.Spec.HealthProbe = probe
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject traffic-restore time above 50 minutes", func() {
			input := minimalSpec()
			input.Spec.RestoreTrafficTimeToHealedOrNewEndpointInMinutes = proto.Int32(51)
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a wrong kind", func() {
			input := minimalSpec()
			input.Kind = "AzureFrontDoorOriginGroups"
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject missing metadata", func() {
			input := minimalSpec()
			input.Metadata = nil
			gomega.Expect(protovalidate.Validate(input)).NotTo(gomega.BeNil())
		})
	})
})
