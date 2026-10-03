package gcpmonitoringuptimecheckv1alpha1

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
	ginkgo.RunSpecs(t, "GcpMonitoringUptimeCheckSpec Suite")
}

func litRef(v string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: v},
	}
}

var _ = ginkgo.Describe("GcpMonitoringUptimeCheckSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// The canonical public-URL HTTPS check.
	minimal := func() *GcpMonitoringUptimeCheck {
		return &GcpMonitoringUptimeCheck{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpMonitoringUptimeCheck",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-uptime-check",
			},
			Spec: &GcpMonitoringUptimeCheckSpec{
				Timeout: "10s",
				MonitoredResource: &GcpMonitoringUptimeCheckMonitoredResource{
					Type:   "uptime_url",
					Labels: map[string]string{"host": "example.com"},
				},
				HttpCheck: &GcpMonitoringUptimeCheckHttpCheck{
					Path:   "/",
					UseSsl: true,
				},
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept the canonical public-URL HTTPS check", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept each documented period", func() {
		for _, v := range []string{"60s", "300s", "600s", "900s"} {
			target := minimal()
			target.Spec.Period = v
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should accept a tcp check", func() {
		target := minimal()
		target.Spec.HttpCheck = nil
		target.Spec.TcpCheck = &GcpMonitoringUptimeCheckTcpCheck{Port: 5432}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a resource_group target", func() {
		target := minimal()
		target.Spec.MonitoredResource = nil
		target.Spec.ResourceGroup = &GcpMonitoringUptimeCheckResourceGroup{
			GroupId:      "my-group",
			ResourceType: "INSTANCE",
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a synthetic monitor without any check block", func() {
		target := minimal()
		target.Spec.MonitoredResource = nil
		target.Spec.HttpCheck = nil
		target.Spec.SyntheticMonitor = &GcpMonitoringUptimeCheckSyntheticMonitor{
			CloudFunction: litRef("projects/p/locations/us-central1/functions/probe"),
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept basic auth and accepted status codes", func() {
		target := minimal()
		target.Spec.HttpCheck.AuthInfo = &GcpMonitoringUptimeCheckHttpAuthInfo{
			Username: "probe",
			Password: "s3cret",
		}
		target.Spec.HttpCheck.AcceptedResponseStatusCodes = []*GcpMonitoringUptimeCheckStatusCode{
			{StatusClass: "STATUS_CLASS_2XX"},
			{StatusValue: 401},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a POST probe with URL_ENCODED body", func() {
		target := minimal()
		target.Spec.HttpCheck.RequestMethod = "POST"
		target.Spec.HttpCheck.ContentType = "URL_ENCODED"
		target.Spec.HttpCheck.Body = "Zm9vPWJhcg==" // base64("foo=bar")
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept USER_PROVIDED content type with its custom value", func() {
		target := minimal()
		target.Spec.HttpCheck.RequestMethod = "POST"
		target.Spec.HttpCheck.ContentType = "USER_PROVIDED"
		target.Spec.HttpCheck.CustomContentType = "application/json"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a JSON-path content matcher with its sub-matcher", func() {
		target := minimal()
		target.Spec.ContentMatchers = []*GcpMonitoringUptimeCheckContentMatcher{{
			Content: "ok",
			Matcher: "MATCHES_JSON_PATH",
			JsonPathMatcher: &GcpMonitoringUptimeCheckJsonPathMatcher{
				JsonPath:    "$.status",
				JsonMatcher: "EXACT_MATCH",
			},
		}}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a plain string matcher without a JSON-path sub-matcher", func() {
		target := minimal()
		target.Spec.ContentMatchers = []*GcpMonitoringUptimeCheckContentMatcher{{
			Content: "healthy",
			Matcher: "CONTAINS_STRING",
		}}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept checker_type, selected_regions, and deletion_policy", func() {
		target := minimal()
		target.Spec.CheckerType = "STATIC_IP_CHECKERS"
		target.Spec.SelectedRegions = []string{"USA", "EUROPE", "ASIA_PACIFIC"}
		target.Spec.DeletionPolicy = "PREVENT"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a missing timeout", func() {
		target := minimal()
		target.Spec.Timeout = ""
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an out-of-bounds timeout", func() {
		for _, v := range []string{"0s", "61s", "600s"} {
			target := minimal()
			target.Spec.Timeout = v
			gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
		}
	})

	ginkgo.It("should reject an unsupported period", func() {
		target := minimal()
		target.Spec.Period = "120s"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "900s")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject zero targets and multiple targets", func() {
		none := minimal()
		none.Spec.MonitoredResource = nil
		gomega.Expect(validator.Validate(none)).ToNot(gomega.Succeed())

		both := minimal()
		both.Spec.ResourceGroup = &GcpMonitoringUptimeCheckResourceGroup{GroupId: "g"}
		gomega.Expect(validator.Validate(both)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject both http and tcp checks", func() {
		target := minimal()
		target.Spec.TcpCheck = &GcpMonitoringUptimeCheckTcpCheck{Port: 80}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a check block alongside a synthetic monitor", func() {
		target := minimal()
		target.Spec.MonitoredResource = nil
		target.Spec.SyntheticMonitor = &GcpMonitoringUptimeCheckSyntheticMonitor{
			CloudFunction: litRef("projects/p/locations/us-central1/functions/probe"),
		}
		// http_check stays set from minimal() — forbidden with a synthetic monitor.
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject USER_PROVIDED content type without the custom value", func() {
		target := minimal()
		target.Spec.HttpCheck.ContentType = "USER_PROVIDED"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "custom_content_type")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a custom content type without opting in", func() {
		target := minimal()
		target.Spec.HttpCheck.CustomContentType = "application/json"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a status code entry setting both class and value", func() {
		target := minimal()
		target.Spec.HttpCheck.AcceptedResponseStatusCodes = []*GcpMonitoringUptimeCheckStatusCode{
			{StatusClass: "STATUS_CLASS_2XX", StatusValue: 200},
		}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a JSON-path matcher without its sub-matcher (and vice versa)", func() {
		missing := minimal()
		missing.Spec.ContentMatchers = []*GcpMonitoringUptimeCheckContentMatcher{{
			Content: "ok",
			Matcher: "MATCHES_JSON_PATH",
		}}
		gomega.Expect(validator.Validate(missing)).ToNot(gomega.Succeed())

		extra := minimal()
		extra.Spec.ContentMatchers = []*GcpMonitoringUptimeCheckContentMatcher{{
			Content:         "ok",
			Matcher:         "CONTAINS_STRING",
			JsonPathMatcher: &GcpMonitoringUptimeCheckJsonPathMatcher{JsonPath: "$.x"},
		}}
		gomega.Expect(validator.Validate(extra)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an out-of-range pings_count", func() {
		target := minimal()
		target.Spec.HttpCheck.PingConfig = &GcpMonitoringUptimeCheckPingConfig{PingsCount: 4}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an invalid deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})
})
