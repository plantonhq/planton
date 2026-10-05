package cloudflarezerotrustdeviceposturerulev1alpha1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
	"google.golang.org/protobuf/proto"
)

func TestCloudflareZeroTrustDevicePostureRuleSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "CloudflareZeroTrustDevicePostureRuleSpec Custom Validation Tests")
}

const testAccountID = "0da42c8d2132a9ddaf714f9e7c920711"

func validRule(spec *CloudflareZeroTrustDevicePostureRuleSpec) *CloudflareZeroTrustDevicePostureRule {
	return &CloudflareZeroTrustDevicePostureRule{
		ApiVersion: "cloudflare.planton.dev/v1alpha1",
		Kind:       "CloudflareZeroTrustDevicePostureRule",
		Metadata: &shared.CatalogObjectMetadata{
			Name: "test-posture-rule",
		},
		Spec: spec,
	}
}

func baseSpec() *CloudflareZeroTrustDevicePostureRuleSpec {
	return &CloudflareZeroTrustDevicePostureRuleSpec{
		AccountId: testAccountID,
		Name:      "os-current",
		Type:      "os_version",
	}
}

var _ = ginkgo.Describe("CloudflareZeroTrustDevicePostureRuleSpec Custom Validation Tests", func() {

	ginkgo.Describe("When valid input is passed", func() {

		ginkgo.It("should accept an os_version rule", func() {
			spec := baseSpec()
			spec.Schedule = "5m"
			spec.Match = []*CloudflareZeroTrustDevicePostureRuleMatch{{Platform: "mac"}}
			spec.Input = &CloudflareZeroTrustDevicePostureRuleInput{
				Version:  "14.4.1",
				Operator: ">=",
			}
			gomega.Expect(protovalidate.Validate(validRule(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept a file rule", func() {
			spec := baseSpec()
			spec.Type = "file"
			spec.Match = []*CloudflareZeroTrustDevicePostureRuleMatch{{Platform: "windows"}}
			spec.Input = &CloudflareZeroTrustDevicePostureRuleInput{
				Path:   "C:\\Program Files\\Corp\\agent.exe",
				Exists: proto.Bool(true),
			}
			gomega.Expect(protovalidate.Validate(validRule(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept a disk encryption rule", func() {
			spec := baseSpec()
			spec.Type = "disk_encryption"
			spec.Input = &CloudflareZeroTrustDevicePostureRuleInput{
				RequireAll: proto.Bool(true),
			}
			gomega.Expect(protovalidate.Validate(validRule(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept a client certificate v2 rule with locations", func() {
			spec := baseSpec()
			spec.Type = "client_certificate_v2"
			spec.Input = &CloudflareZeroTrustDevicePostureRuleInput{
				CertificateId:    "f70ff985-a4ef-4643-bbbc-4a0ed4fc8415",
				Cn:               "corp-device",
				CheckPrivateKey:  proto.Bool(true),
				ExtendedKeyUsage: []string{"clientAuth"},
				Locations: &CloudflareZeroTrustDevicePostureRuleCertificateLocations{
					TrustStores: []string{"system"},
				},
			}
			gomega.Expect(protovalidate.Validate(validRule(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept a sentinelone_s2s rule with vendor enums", func() {
			spec := baseSpec()
			spec.Type = "sentinelone_s2s"
			spec.Input = &CloudflareZeroTrustDevicePostureRuleInput{
				ConnectionId:     "f70ff985-a4ef-4643-bbbc-4a0ed4fc8415",
				ActiveThreats:    proto.Int64(0),
				Infected:         proto.Bool(false),
				NetworkStatus:    "connected",
				OperationalState: "na",
			}
			gomega.Expect(protovalidate.Validate(validRule(spec))).To(gomega.BeNil())
		})

		ginkgo.It("should accept a kolide rule with auth states", func() {
			spec := baseSpec()
			spec.Type = "kolide"
			spec.Input = &CloudflareZeroTrustDevicePostureRuleInput{
				AuthState:     []string{"Good", "Notified"},
				CountOperator: "<=",
				IssueCount:    "3",
			}
			gomega.Expect(protovalidate.Validate(validRule(spec))).To(gomega.BeNil())
		})
	})

	ginkgo.Describe("When invalid input is passed", func() {

		ginkgo.It("should reject an unknown rule type", func() {
			spec := baseSpec()
			spec.Type = "edr"
			gomega.Expect(protovalidate.Validate(validRule(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a missing name -- rules must be identifiable", func() {
			spec := baseSpec()
			spec.Name = ""
			gomega.Expect(protovalidate.Validate(validRule(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown match platform", func() {
			spec := baseSpec()
			spec.Match = []*CloudflareZeroTrustDevicePostureRuleMatch{{Platform: "freebsd"}}
			gomega.Expect(protovalidate.Validate(validRule(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown comparison operator", func() {
			spec := baseSpec()
			spec.Input = &CloudflareZeroTrustDevicePostureRuleInput{Operator: "!="}
			gomega.Expect(protovalidate.Validate(validRule(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown compliance status", func() {
			spec := baseSpec()
			spec.Input = &CloudflareZeroTrustDevicePostureRuleInput{ComplianceStatus: "passing"}
			gomega.Expect(protovalidate.Validate(validRule(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown kolide auth state", func() {
			spec := baseSpec()
			spec.Input = &CloudflareZeroTrustDevicePostureRuleInput{AuthState: []string{"good"}}
			gomega.Expect(protovalidate.Validate(validRule(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject an unknown trust store", func() {
			spec := baseSpec()
			spec.Input = &CloudflareZeroTrustDevicePostureRuleInput{
				Locations: &CloudflareZeroTrustDevicePostureRuleCertificateLocations{
					TrustStores: []string{"machine"},
				},
			}
			gomega.Expect(protovalidate.Validate(validRule(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a score above 100", func() {
			spec := baseSpec()
			spec.Input = &CloudflareZeroTrustDevicePostureRuleInput{Score: proto.Float64(101)}
			gomega.Expect(protovalidate.Validate(validRule(spec))).NotTo(gomega.BeNil())
		})

		ginkgo.It("should reject a malformed account_id", func() {
			spec := baseSpec()
			spec.AccountId = "nope"
			gomega.Expect(protovalidate.Validate(validRule(spec))).NotTo(gomega.BeNil())
		})
	})
})
