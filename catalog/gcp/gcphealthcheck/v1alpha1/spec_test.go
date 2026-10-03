package gcphealthcheckv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/plantonhq/planton/shared"
)

func TestSuite(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "GcpHealthCheckSpec Suite")
}

func i32(v int32) *int32 {
	return &v
}

var _ = ginkgo.Describe("GcpHealthCheckSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	// Helper to build a minimal valid GcpHealthCheck (global HTTP probe).
	minimal := func() *GcpHealthCheck {
		return &GcpHealthCheck{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpHealthCheck",
			Metadata: &shared.CatalogObjectMetadata{
				Name: "test-health-check",
			},
			Spec: &GcpHealthCheckSpec{
				Protocol: &GcpHealthCheckSpec_Http{
					Http: &GcpHealthCheckHttp{},
				},
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal valid spec (global HTTP probe)", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a fully-specified HTTPS probe", func() {
		target := minimal()
		target.Spec.HealthCheckName = "web-tier-probe"
		target.Spec.Description = "probes the web tier"
		target.Spec.CheckIntervalSec = i32(10)
		target.Spec.TimeoutSec = i32(8)
		target.Spec.HealthyThreshold = i32(3)
		target.Spec.UnhealthyThreshold = i32(4)
		target.Spec.EnableLogging = true
		target.Spec.Protocol = &GcpHealthCheckSpec_Https{
			Https: &GcpHealthCheckHttps{
				Host:              "app.example.com",
				Port:              8443,
				ProxyHeader:       "PROXY_V1",
				RequestPath:       "/healthz",
				Response:          "ok",
				PortSpecification: "USE_FIXED_PORT",
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a regional TCP probe", func() {
		target := minimal()
		target.Spec.Region = "us-central1"
		target.Spec.Protocol = &GcpHealthCheckSpec_Tcp{
			Tcp: &GcpHealthCheckTcp{Port: 5432},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a gRPC probe with a service name", func() {
		target := minimal()
		target.Spec.Protocol = &GcpHealthCheckSpec_Grpc{
			Grpc: &GcpHealthCheckGrpc{
				GrpcServiceName: "payments.v1.Payments",
				Port:            50051,
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a gRPC-with-TLS probe using USE_FIXED_PORT", func() {
		target := minimal()
		target.Spec.Protocol = &GcpHealthCheckSpec_GrpcTls{
			GrpcTls: &GcpHealthCheckGrpcTls{
				Port:              50052,
				PortSpecification: "USE_FIXED_PORT",
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept USE_NAMED_PORT with a port_name and no port", func() {
		target := minimal()
		target.Spec.Protocol = &GcpHealthCheckSpec_Http{
			Http: &GcpHealthCheckHttp{
				PortSpecification: "USE_NAMED_PORT",
				PortName:          "web",
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept source_regions with exactly 3 regions on a global HTTP check with interval >= 30", func() {
		target := minimal()
		target.Spec.SourceRegions = []string{"us-central1", "us-east1", "europe-west1"}
		target.Spec.CheckIntervalSec = i32(30)
		target.Spec.TimeoutSec = i32(10)
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept USE_SERVING_PORT with neither port nor port_name", func() {
		target := minimal()
		target.Spec.Protocol = &GcpHealthCheckSpec_Http{
			Http: &GcpHealthCheckHttp{
				PortSpecification: "USE_SERVING_PORT",
				RequestPath:       "/healthz",
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject a spec without a protocol block", func() {
		target := minimal()
		target.Spec.Protocol = nil
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid health_check_name", func() {
		target := minimal()
		target.Spec.HealthCheckName = "Invalid_Name"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "RFC1035")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject an uppercase region", func() {
		target := minimal()
		target.Spec.Region = "US-CENTRAL1"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject timeout_sec greater than check_interval_sec", func() {
		target := minimal()
		target.Spec.CheckIntervalSec = i32(5)
		target.Spec.TimeoutSec = i32(10)
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "timeout_sec must not exceed")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject timeout_sec exceeding the default interval when only timeout is set", func() {
		target := minimal()
		target.Spec.TimeoutSec = i32(6)
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject source_regions with fewer than 3 regions", func() {
		target := minimal()
		target.Spec.SourceRegions = []string{"us-central1", "us-east1"}
		target.Spec.CheckIntervalSec = i32(30)
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "exactly 3")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject source_regions on a regional health check", func() {
		target := minimal()
		target.Spec.Region = "us-central1"
		target.Spec.SourceRegions = []string{"us-central1", "us-east1", "europe-west1"}
		target.Spec.CheckIntervalSec = i32(30)
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "global")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject source_regions with an SSL probe", func() {
		target := minimal()
		target.Spec.SourceRegions = []string{"us-central1", "us-east1", "europe-west1"}
		target.Spec.CheckIntervalSec = i32(30)
		target.Spec.Protocol = &GcpHealthCheckSpec_Ssl{
			Ssl: &GcpHealthCheckSsl{Port: 443},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "HTTP, HTTPS, and TCP")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject source_regions with check_interval_sec below 30", func() {
		target := minimal()
		target.Spec.SourceRegions = []string{"us-central1", "us-east1", "europe-west1"}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "at least 30")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject source_regions with a PROXY_V1 proxy header", func() {
		target := minimal()
		target.Spec.SourceRegions = []string{"us-central1", "us-east1", "europe-west1"}
		target.Spec.CheckIntervalSec = i32(30)
		target.Spec.Protocol = &GcpHealthCheckSpec_Http{
			Http: &GcpHealthCheckHttp{ProxyHeader: "PROXY_V1"},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "proxy_header must be NONE")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject source_regions with a TCP request payload", func() {
		target := minimal()
		target.Spec.SourceRegions = []string{"us-central1", "us-east1", "europe-west1"}
		target.Spec.CheckIntervalSec = i32(30)
		target.Spec.Protocol = &GcpHealthCheckSpec_Tcp{
			Tcp: &GcpHealthCheckTcp{Request: "PING"},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a port above 65535", func() {
		target := minimal()
		target.Spec.Protocol = &GcpHealthCheckSpec_Http{
			Http: &GcpHealthCheckHttp{Port: 70000},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "between 1 and 65535")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject an invalid port_specification", func() {
		target := minimal()
		target.Spec.Protocol = &GcpHealthCheckSpec_Http{
			Http: &GcpHealthCheckHttp{PortSpecification: "USE_RANDOM_PORT"},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject an invalid proxy_header", func() {
		target := minimal()
		target.Spec.Protocol = &GcpHealthCheckSpec_Tcp{
			Tcp: &GcpHealthCheckTcp{ProxyHeader: "PROXY_V2"},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "NONE or PROXY_V1")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject USE_NAMED_PORT with a numeric port set", func() {
		target := minimal()
		target.Spec.Protocol = &GcpHealthCheckSpec_Http{
			Http: &GcpHealthCheckHttp{
				PortSpecification: "USE_NAMED_PORT",
				PortName:          "web",
				Port:              8080,
			},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "USE_NAMED_PORT")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject USE_NAMED_PORT without a port_name", func() {
		target := minimal()
		target.Spec.Protocol = &GcpHealthCheckSpec_Tcp{
			Tcp: &GcpHealthCheckTcp{PortSpecification: "USE_NAMED_PORT"},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject USE_SERVING_PORT with a port_name set", func() {
		target := minimal()
		target.Spec.Protocol = &GcpHealthCheckSpec_Http{
			Http: &GcpHealthCheckHttp{
				PortSpecification: "USE_SERVING_PORT",
				PortName:          "web",
			},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "USE_SERVING_PORT")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject USE_NAMED_PORT on a gRPC-with-TLS probe", func() {
		target := minimal()
		target.Spec.Protocol = &GcpHealthCheckSpec_GrpcTls{
			GrpcTls: &GcpHealthCheckGrpcTls{PortSpecification: "USE_NAMED_PORT"},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "gRPC-with-TLS")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a non-positive check_interval_sec", func() {
		target := minimal()
		target.Spec.CheckIntervalSec = i32(0)
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should reject a wrong kind constant", func() {
		target := minimal()
		target.Kind = "GcpHealthChecker"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
	})

	ginkgo.It("should accept every deletion_policy value", func() {
		for _, v := range []string{"DELETE", "PREVENT", "ABANDON", ""} {
			target := minimal()
			target.Spec.DeletionPolicy = v
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should reject an unknown deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "DESTROY"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "deletion_policy")).To(gomega.BeTrue())
	})
})
