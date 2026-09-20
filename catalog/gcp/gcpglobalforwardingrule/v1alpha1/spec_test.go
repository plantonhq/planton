package gcpglobalforwardingrulev1alpha1

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
	ginkgo.RunSpecs(t, "GcpGlobalForwardingRuleSpec Suite")
}

func literalRef(value string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: value},
	}
}

func proxySelfLink() *foreignkeyv1.StringValueOrRef {
	return literalRef("https://www.googleapis.com/compute/v1/projects/p/global/targetHttpsProxies/web-frontend")
}

func schemePtr(scheme string) *string {
	return &scheme
}

var _ = ginkgo.Describe("GcpGlobalForwardingRuleSpec", func() {
	var validator protovalidate.Validator

	ginkgo.BeforeEach(func() {
		var err error
		validator, err = protovalidate.New()
		gomega.Expect(err).ToNot(gomega.HaveOccurred())
	})

	minimal := func() *GcpGlobalForwardingRule {
		return &GcpGlobalForwardingRule{
			ApiVersion: "gcp.planton.dev/v1alpha1",
			Kind:       "GcpGlobalForwardingRule",
			Metadata: &shared.CloudResourceMetadata{
				Name: "test-forwarding-rule",
			},
			Spec: &GcpGlobalForwardingRuleSpec{
				Target:    proxySelfLink(),
				PortRange: "443",
			},
		}
	}

	// ──────────────── Positive Cases ────────────────

	ginkgo.It("should accept a minimal external HTTPS frontend", func() {
		gomega.Expect(validator.Validate(minimal())).To(gomega.Succeed())
	})

	ginkgo.It("should accept a static ip_address reference", func() {
		target := minimal()
		target.Spec.IpAddress = literalRef("34.120.1.2")
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a port range", func() {
		target := minimal()
		target.Spec.PortRange = "8080-8090"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an EXTERNAL_MANAGED frontend", func() {
		target := minimal()
		target.Spec.LoadBalancingScheme = schemePtr("EXTERNAL_MANAGED")
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an INTERNAL_SELF_MANAGED frontend with network and metadata filters", func() {
		target := minimal()
		target.Spec.LoadBalancingScheme = schemePtr("INTERNAL_SELF_MANAGED")
		target.Spec.Network = literalRef("https://www.googleapis.com/compute/v1/projects/p/global/networks/mesh-vpc")
		target.Spec.MetadataFilters = []*GcpGlobalForwardingRuleMetadataFilter{
			{
				FilterMatchCriteria: "MATCH_ANY",
				FilterLabels: []*GcpGlobalForwardingRuleMetadataFilterLabel{
					{Name: "env", Value: "prod"},
				},
			},
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a Private Service Connect Google-APIs frontend", func() {
		target := minimal()
		target.Spec.Target = literalRef("all-apis")
		target.Spec.LoadBalancingScheme = schemePtr("NONE")
		target.Spec.PortRange = ""
		target.Spec.IpAddress = literalRef("10.10.0.5")
		target.Spec.Network = literalRef("https://www.googleapis.com/compute/v1/projects/p/global/networks/main-vpc")
		target.Spec.NoAutomateDnsZone = true
		target.Spec.ServiceDirectoryRegistration = &GcpGlobalForwardingRuleServiceDirectoryRegistration{
			Namespace:              "psc-endpoints",
			ServiceDirectoryRegion: "us-central1",
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept the PREMIUM network tier", func() {
		target := minimal()
		target.Spec.NetworkTier = "PREMIUM"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept labels", func() {
		target := minimal()
		target.Spec.Labels = map[string]string{"env": "prod", "team": "platform"}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept the backend-bucket migration canary states", func() {
		for _, state := range []string{"PREPARE", "TEST_BY_PERCENTAGE", "TEST_ALL_TRAFFIC"} {
			target := minimal()
			target.Spec.ExternalManagedBackendBucketMigrationState = state
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should accept a migration testing percentage with TEST_BY_PERCENTAGE", func() {
		target := minimal()
		target.Spec.ExternalManagedBackendBucketMigrationState = "TEST_BY_PERCENTAGE"
		target.Spec.ExternalManagedBackendBucketMigrationTestingPercentage = 25.5
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept each ip_protocol", func() {
		for _, protocol := range []string{"TCP", "UDP", "ESP", "AH", "SCTP", "ICMP"} {
			target := minimal()
			target.Spec.IpProtocol = &protocol
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	ginkgo.It("should accept IPV6 for an auto-assigned address", func() {
		target := minimal()
		target.Spec.IpVersion = "IPV6"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept each deletion_policy value", func() {
		for _, v := range []string{"DELETE", "PREVENT", "ABANDON"} {
			target := minimal()
			target.Spec.DeletionPolicy = v
			gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
		}
	})

	// ──────────────── Negative Cases ────────────────

	ginkgo.It("should reject an invalid deletion_policy", func() {
		target := minimal()
		target.Spec.DeletionPolicy = "KEEP"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a rule with neither target nor backend_service", func() {
		target := minimal()
		target.Spec.Target = nil
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "exactly one of target")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject an invalid forwarding_rule_name", func() {
		target := minimal()
		target.Spec.ForwardingRuleName = "Frontend-1"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "RFC1035")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject an invalid ip_protocol", func() {
		target := minimal()
		invalid := "GRE"
		target.Spec.IpProtocol = &invalid
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an invalid ip_version", func() {
		target := minimal()
		target.Spec.IpVersion = "IPV5"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an invalid load_balancing_scheme", func() {
		target := minimal()
		target.Spec.LoadBalancingScheme = schemePtr("EXTERNAL_PASSTHROUGH")
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject the INTERNAL scheme on a global rule", func() {
		target := minimal()
		target.Spec.LoadBalancingScheme = schemePtr("INTERNAL")
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "INTERNAL scheme")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a malformed port_range", func() {
		target := minimal()
		target.Spec.PortRange = "443,80"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "contiguous range")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject the STANDARD network tier on a global rule", func() {
		target := minimal()
		target.Spec.NetworkTier = "STANDARD"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "PREMIUM")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject network on an external frontend", func() {
		target := minimal()
		target.Spec.Network = literalRef("https://www.googleapis.com/compute/v1/projects/p/global/networks/main-vpc")
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "internal and Private Service Connect frontends")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject metadata_filters without Traffic Director", func() {
		target := minimal()
		target.Spec.MetadataFilters = []*GcpGlobalForwardingRuleMetadataFilter{
			{
				FilterMatchCriteria: "MATCH_ALL",
				FilterLabels: []*GcpGlobalForwardingRuleMetadataFilterLabel{
					{Name: "env", Value: "prod"},
				},
			},
		}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "Traffic Director")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject service_directory_registration without the PSC scheme", func() {
		target := minimal()
		target.Spec.ServiceDirectoryRegistration = &GcpGlobalForwardingRuleServiceDirectoryRegistration{
			Namespace: "psc-endpoints",
		}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject no_automate_dns_zone without the PSC scheme", func() {
		target := minimal()
		target.Spec.NoAutomateDnsZone = true
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a migration percentage without TEST_BY_PERCENTAGE", func() {
		target := minimal()
		target.Spec.ExternalManagedBackendBucketMigrationTestingPercentage = 50
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "TEST_BY_PERCENTAGE")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a migration percentage above 100", func() {
		target := minimal()
		target.Spec.ExternalManagedBackendBucketMigrationState = "TEST_BY_PERCENTAGE"
		target.Spec.ExternalManagedBackendBucketMigrationTestingPercentage = 101
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject an invalid migration state", func() {
		target := minimal()
		target.Spec.ExternalManagedBackendBucketMigrationState = "FINISH"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a metadata filter with an invalid criteria", func() {
		target := minimal()
		target.Spec.LoadBalancingScheme = schemePtr("INTERNAL_SELF_MANAGED")
		target.Spec.MetadataFilters = []*GcpGlobalForwardingRuleMetadataFilter{
			{
				FilterMatchCriteria: "MATCH_SOME",
				FilterLabels: []*GcpGlobalForwardingRuleMetadataFilterLabel{
					{Name: "env", Value: "prod"},
				},
			},
		}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a metadata filter with no labels", func() {
		target := minimal()
		target.Spec.LoadBalancingScheme = schemePtr("INTERNAL_SELF_MANAGED")
		target.Spec.MetadataFilters = []*GcpGlobalForwardingRuleMetadataFilter{
			{FilterMatchCriteria: "MATCH_ALL"},
		}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a wrong kind literal", func() {
		target := minimal()
		target.Kind = "GcpForwardingRule"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject missing spec", func() {
		target := minimal()
		target.Spec = nil
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	// ──────────────── Regional arm ────────────────

	regionalProxy := func() *foreignkeyv1.StringValueOrRef {
		return literalRef("https://www.googleapis.com/compute/v1/projects/p/regions/us-central1/targetHttpProxies/web")
	}
	regionalBackend := func() *foreignkeyv1.StringValueOrRef {
		return literalRef("https://www.googleapis.com/compute/v1/projects/p/regions/us-central1/backendServices/ilb")
	}
	network := func() *foreignkeyv1.StringValueOrRef {
		return literalRef("https://www.googleapis.com/compute/v1/projects/p/global/networks/vpc")
	}

	// A regional external Application Load Balancer frontend: the proxy form
	// with the regional-only network and STANDARD tier.
	regionalAlb := func() *GcpGlobalForwardingRule {
		target := minimal()
		target.Spec.Region = "us-central1"
		target.Spec.Target = regionalProxy()
		target.Spec.LoadBalancingScheme = schemePtr("EXTERNAL_MANAGED")
		target.Spec.PortRange = "80"
		target.Spec.Network = network()
		target.Spec.NetworkTier = "STANDARD"
		return target
	}

	// An internal passthrough Network Load Balancer frontend: the
	// backend-service form.
	internalNlb := func() *GcpGlobalForwardingRule {
		target := minimal()
		target.Spec.Region = "us-central1"
		target.Spec.Target = nil
		target.Spec.BackendService = regionalBackend()
		target.Spec.LoadBalancingScheme = schemePtr("INTERNAL")
		target.Spec.PortRange = ""
		target.Spec.Ports = []string{"80", "443", "8080-8090"}
		target.Spec.Network = network()
		target.Spec.Subnetwork = literalRef("https://www.googleapis.com/compute/v1/projects/p/regions/us-central1/subnetworks/apps")
		return target
	}

	ginkgo.It("should accept a regional external managed ALB frontend", func() {
		gomega.Expect(validator.Validate(regionalAlb())).To(gomega.Succeed())
	})

	ginkgo.It("should accept an internal passthrough NLB frontend with its internal levers", func() {
		target := internalNlb()
		target.Spec.AllowGlobalAccess = true
		target.Spec.ServiceLabel = "orders"
		target.Spec.IsMirroringCollector = true
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an L3_DEFAULT rule forwarding every port", func() {
		target := internalNlb()
		target.Spec.Ports = nil
		target.Spec.AllPorts = true
		target.Spec.IpProtocol = schemePtr("L3_DEFAULT")
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept an external passthrough NLB with source_ip_ranges and a BYOIP collection", func() {
		target := internalNlb()
		target.Spec.LoadBalancingScheme = schemePtr("EXTERNAL")
		target.Spec.Network = nil
		target.Spec.Subnetwork = nil
		target.Spec.SourceIpRanges = []string{"203.0.113.0/24", "198.51.100.7"}
		target.Spec.IpCollection = "projects/p/regions/us-central1/publicDelegatedPrefixes/byoip-v6"
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should accept a regional Private Service Connect consumer endpoint", func() {
		target := minimal()
		target.Spec.Region = "us-central1"
		target.Spec.Target = literalRef("https://www.googleapis.com/compute/v1/projects/producer/regions/us-central1/serviceAttachments/db")
		target.Spec.LoadBalancingScheme = schemePtr("NONE")
		target.Spec.PortRange = ""
		target.Spec.Network = network()
		target.Spec.AllowPscGlobalAccess = true
		target.Spec.RecreateClosedPsc = true
		target.Spec.ServiceDirectoryRegistration = &GcpGlobalForwardingRuleServiceDirectoryRegistration{
			Namespace: "psc", Service: "db",
		}
		gomega.Expect(validator.Validate(target)).To(gomega.Succeed())
	})

	ginkgo.It("should reject a malformed region", func() {
		target := regionalAlb()
		target.Spec.Region = "us central1"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject target and backend_service together", func() {
		target := internalNlb()
		target.Spec.Target = regionalProxy()
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "exactly one of target")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject backend_service on a global rule", func() {
		target := minimal()
		target.Spec.Target = nil
		target.Spec.BackendService = regionalBackend()
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "backend_service is the passthrough")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject the regional-only levers on a global rule", func() {
		for _, mutate := range []func(*GcpGlobalForwardingRuleSpec){
			func(s *GcpGlobalForwardingRuleSpec) { s.Ports = []string{"80"}; s.PortRange = "" },
			func(s *GcpGlobalForwardingRuleSpec) { s.AllPorts = true; s.PortRange = "" },
			func(s *GcpGlobalForwardingRuleSpec) { s.SourceIpRanges = []string{"10.0.0.0/8"} },
			func(s *GcpGlobalForwardingRuleSpec) { s.ServiceLabel = "orders" },
			func(s *GcpGlobalForwardingRuleSpec) { s.AllowGlobalAccess = true },
			func(s *GcpGlobalForwardingRuleSpec) { s.AllowPscGlobalAccess = true },
			func(s *GcpGlobalForwardingRuleSpec) { s.IsMirroringCollector = true },
			func(s *GcpGlobalForwardingRuleSpec) { s.IpCollection = "regions/us-central1/publicDelegatedPrefixes/p" },
			func(s *GcpGlobalForwardingRuleSpec) { s.RecreateClosedPsc = true },
		} {
			target := minimal()
			mutate(target.Spec)
			err := validator.Validate(target)
			gomega.Expect(err).To(gomega.HaveOccurred())
			gomega.Expect(strings.Contains(err.Error(), "exist only on a regional forwarding rule")).To(gomega.BeTrue())
		}
	})

	ginkgo.It("should reject two port forms together", func() {
		target := internalNlb()
		target.Spec.AllPorts = true
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "mutually exclusive")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject a malformed ports entry and more than five ports", func() {
		target := internalNlb()
		target.Spec.Ports = []string{"80,443"}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
		target.Spec.Ports = []string{"1", "2", "3", "4", "5", "6"}
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject L3_DEFAULT on a global rule and without all_ports", func() {
		global := minimal()
		global.Spec.IpProtocol = schemePtr("L3_DEFAULT")
		global.Spec.PortRange = ""
		gomega.Expect(validator.Validate(global)).ToNot(gomega.Succeed())

		regional := internalNlb()
		regional.Spec.IpProtocol = schemePtr("L3_DEFAULT")
		err := validator.Validate(regional)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "L3_DEFAULT protocol exists only")).To(gomega.BeTrue())
	})

	ginkgo.It("should accept the STANDARD tier on a regional rule", func() {
		gomega.Expect(validator.Validate(regionalAlb())).To(gomega.Succeed())
	})

	ginkgo.It("should reject INTERNAL_SELF_MANAGED and metadata_filters on a regional rule", func() {
		target := regionalAlb()
		target.Spec.LoadBalancingScheme = schemePtr("INTERNAL_SELF_MANAGED")
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "INTERNAL_SELF_MANAGED scheme (Traffic Director) exists only on a global")).To(gomega.BeTrue())
	})

	ginkgo.It("should reject the backend-bucket migration canary on a regional rule", func() {
		target := regionalAlb()
		target.Spec.LoadBalancingScheme = schemePtr("EXTERNAL")
		target.Spec.Network = nil
		target.Spec.ExternalManagedBackendBucketMigrationState = "PREPARE"
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "regional rule has no backend buckets")).To(gomega.BeTrue())
	})

	ginkgo.It("should bind the Service Directory fields to their scope", func() {
		global := minimal()
		global.Spec.LoadBalancingScheme = schemePtr("NONE")
		global.Spec.Target = literalRef("all-apis")
		global.Spec.PortRange = ""
		global.Spec.Network = network()
		global.Spec.IpAddress = literalRef("10.10.0.5")
		global.Spec.ServiceDirectoryRegistration = &GcpGlobalForwardingRuleServiceDirectoryRegistration{Service: "apis"}
		err := validator.Validate(global)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "service_directory_region exists only on a global rule")).To(gomega.BeTrue())

		regional := minimal()
		regional.Spec.Region = "us-central1"
		regional.Spec.LoadBalancingScheme = schemePtr("NONE")
		regional.Spec.Target = literalRef("projects/producer/regions/us-central1/serviceAttachments/db")
		regional.Spec.PortRange = ""
		regional.Spec.Network = network()
		regional.Spec.ServiceDirectoryRegistration = &GcpGlobalForwardingRuleServiceDirectoryRegistration{ServiceDirectoryRegion: "us-central1"}
		gomega.Expect(validator.Validate(regional)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject network on an external passthrough NLB but accept it on the regional external ALB", func() {
		nlb := internalNlb()
		nlb.Spec.LoadBalancingScheme = schemePtr("EXTERNAL")
		nlb.Spec.Subnetwork = nil
		err := validator.Validate(nlb)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "network applies to")).To(gomega.BeTrue())

		gomega.Expect(validator.Validate(regionalAlb())).To(gomega.Succeed())

		globalManaged := minimal()
		globalManaged.Spec.LoadBalancingScheme = schemePtr("EXTERNAL_MANAGED")
		globalManaged.Spec.Network = network()
		gomega.Expect(validator.Validate(globalManaged)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should bind the internal-only levers to the INTERNAL scheme", func() {
		target := regionalAlb()
		target.Spec.AllowGlobalAccess = true
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "belong to the internal passthrough")).To(gomega.BeTrue())
	})

	ginkgo.It("should bind the PSC consumer levers to the NONE scheme", func() {
		target := regionalAlb()
		target.Spec.AllowPscGlobalAccess = true
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "Private Service Connect consumer endpoint")).To(gomega.BeTrue())
	})

	ginkgo.It("should bind source_ip_ranges to the EXTERNAL scheme and validate each entry", func() {
		target := internalNlb()
		target.Spec.SourceIpRanges = []string{"10.0.0.0/8"}
		err := validator.Validate(target)
		gomega.Expect(err).To(gomega.HaveOccurred())
		gomega.Expect(strings.Contains(err.Error(), "source_ip_ranges filters the external")).To(gomega.BeTrue())

		malformed := internalNlb()
		malformed.Spec.LoadBalancingScheme = schemePtr("EXTERNAL")
		malformed.Spec.Network = nil
		malformed.Spec.Subnetwork = nil
		malformed.Spec.SourceIpRanges = []string{"not-an-ip"}
		gomega.Expect(validator.Validate(malformed)).ToNot(gomega.Succeed())
	})

	ginkgo.It("should reject a malformed service_label", func() {
		target := internalNlb()
		target.Spec.ServiceLabel = "Orders"
		gomega.Expect(validator.Validate(target)).ToNot(gomega.Succeed())
	})
})
