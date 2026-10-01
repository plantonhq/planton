package awsroute53dnsrecordv1alpha1

import (
	"strings"
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
)

func TestAwsRoute53DnsRecordSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsRoute53DnsRecordSpec Validation Suite")
}

// helper to create a StringValueOrRef with a literal value.
func strRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

// helper to build a valid alias target (an ALB pair).
func albAlias() *AwsRoute53AliasTarget {
	return &AwsRoute53AliasTarget{
		DnsName: strRef("my-alb-1234567890.us-east-1.elb.amazonaws.com"),
		ZoneId:  strRef("Z35SXDOTRQ7X7K"),
	}
}

var _ = ginkgo.Describe("AwsRoute53DnsRecordSpec validations", func() {
	var spec *AwsRoute53DnsRecordSpec

	ginkgo.BeforeEach(func() {
		// Minimal valid spec: a simple A record.
		spec = &AwsRoute53DnsRecordSpec{
			Region: "us-west-2",
			ZoneId: strRef("Z1234567890ABC"),
			Name:   "www.example.com",
			Type:   "A",
			Ttl:    proto.Int32(300),
			Values: []*foreignkeyv1.StringValueOrRef{strRef("192.0.2.1")},
		}
	})

	// -------------------------------------------------------------------------
	// Happy path: standard records
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a simple A record", func() {
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.Succeed())
	})

	ginkgo.It("accepts every supported record type", func() {
		for _, t := range []string{"A", "AAAA", "CAA", "CNAME", "DS", "HTTPS", "MX", "NAPTR", "NS", "PTR", "SOA", "SPF", "SRV", "SSHFP", "SVCB", "TLSA", "TXT"} {
			spec.Type = t
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.Succeed(), "type %s should be valid", t)
		}
	})

	ginkgo.It("accepts a multi-value A record and a wildcard name", func() {
		spec.Name = "*.example.com"
		spec.Values = []*foreignkeyv1.StringValueOrRef{strRef("192.0.2.1"), strRef("192.0.2.2")}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.Succeed())
	})

	ginkgo.It("accepts underscore-prefixed service-record names", func() {
		for _, n := range []string{"_dmarc.example.com", "token._domainkey.example.com", "_sip._tcp.example.com"} {
			spec.Name = n
			spec.Type = "TXT"
			spec.Values = []*foreignkeyv1.StringValueOrRef{strRef("v=DMARC1; p=none")}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.Succeed(), "name %s should be valid", n)
		}
	})

	ginkgo.It("accepts allow_overwrite", func() {
		spec.AllowOverwrite = true
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.Succeed())
	})

	// -------------------------------------------------------------------------
	// Happy path: alias records
	// -------------------------------------------------------------------------

	ginkgo.It("accepts an alias record to an ALB", func() {
		spec.Ttl = nil
		spec.Values = nil
		spec.AliasTarget = albAlias()
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.Succeed())
	})

	ginkgo.It("accepts an alias record with target-health evaluation", func() {
		spec.Ttl = nil
		spec.Values = nil
		spec.AliasTarget = albAlias()
		spec.AliasTarget.EvaluateTargetHealth = true
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.Succeed())
	})

	// -------------------------------------------------------------------------
	// Happy path: routing policies
	// -------------------------------------------------------------------------

	ginkgo.It("accepts weighted routing with a set identifier", func() {
		spec.SetIdentifier = "weight-70"
		spec.RoutingPolicy = &AwsRoute53RoutingPolicy{
			Policy: &AwsRoute53RoutingPolicy_Weighted{Weighted: &AwsRoute53WeightedPolicy{Weight: 70}},
		}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.Succeed())
	})

	ginkgo.It("accepts latency routing", func() {
		spec.SetIdentifier = "us-east-1"
		spec.RoutingPolicy = &AwsRoute53RoutingPolicy{
			Policy: &AwsRoute53RoutingPolicy_Latency{Latency: &AwsRoute53LatencyPolicy{Region: "us-east-1"}},
		}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.Succeed())
	})

	ginkgo.It("accepts failover routing with a health check ref", func() {
		spec.SetIdentifier = "primary"
		spec.HealthCheckId = strRef("abcdef11-2222-3333-4444-555555fedcba")
		spec.RoutingPolicy = &AwsRoute53RoutingPolicy{
			Policy: &AwsRoute53RoutingPolicy_Failover{Failover: &AwsRoute53FailoverPolicy{FailoverType: "PRIMARY"}},
		}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.Succeed())
	})

	ginkgo.It("accepts geolocation routing by country", func() {
		spec.SetIdentifier = "eu-users"
		spec.RoutingPolicy = &AwsRoute53RoutingPolicy{
			Policy: &AwsRoute53RoutingPolicy_Geolocation{Geolocation: &AwsRoute53GeolocationPolicy{Country: "DE"}},
		}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.Succeed())
	})

	ginkgo.It("accepts geoproximity routing by region with bias", func() {
		spec.SetIdentifier = "us-east"
		spec.RoutingPolicy = &AwsRoute53RoutingPolicy{
			Policy: &AwsRoute53RoutingPolicy_Geoproximity{Geoproximity: &AwsRoute53GeoproximityPolicy{
				AwsRegion: "us-east-1",
				Bias:      25,
			}},
		}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.Succeed())
	})

	ginkgo.It("accepts geoproximity routing by coordinates", func() {
		spec.SetIdentifier = "nyc-dc"
		spec.RoutingPolicy = &AwsRoute53RoutingPolicy{
			Policy: &AwsRoute53RoutingPolicy_Geoproximity{Geoproximity: &AwsRoute53GeoproximityPolicy{
				Coordinates: &AwsRoute53Coordinates{Latitude: "40.71", Longitude: "-74.01"},
			}},
		}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.Succeed())
	})

	ginkgo.It("accepts CIDR routing", func() {
		spec.SetIdentifier = "office-network"
		spec.RoutingPolicy = &AwsRoute53RoutingPolicy{
			Policy: &AwsRoute53RoutingPolicy_Cidr{Cidr: &AwsRoute53CidrPolicy{
				CollectionId: "cf1234ab-cdef-5678-90ab-cdef12345678",
				LocationName: "office",
			}},
		}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.Succeed())
	})

	ginkgo.It("accepts multivalue answer routing on a standard record", func() {
		spec.SetIdentifier = "server-1"
		spec.RoutingPolicy = &AwsRoute53RoutingPolicy{
			Policy: &AwsRoute53RoutingPolicy_MultivalueAnswer{MultivalueAnswer: &AwsRoute53MultivalueAnswerPolicy{}},
		}
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.Succeed())
	})

	// -------------------------------------------------------------------------
	// Required fields
	// -------------------------------------------------------------------------

	ginkgo.It("rejects a missing zone_id", func() {
		spec.ZoneId = nil
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.Succeed())
	})

	ginkgo.It("rejects a missing name", func() {
		spec.Name = ""
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.Succeed())
	})

	ginkgo.It("rejects a missing type", func() {
		spec.Type = ""
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.Succeed())
	})

	ginkgo.It("rejects an unknown record type", func() {
		spec.Type = "ALIAS"
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.Succeed())
	})

	ginkgo.It("rejects an invalid record name", func() {
		spec.Name = "not a domain!"
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.Succeed())
	})

	// -------------------------------------------------------------------------
	// Standard XOR alias
	// -------------------------------------------------------------------------

	ginkgo.It("rejects a record with neither values nor alias", func() {
		spec.Values = nil
		spec.Ttl = nil
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.Succeed())
	})

	ginkgo.It("rejects a record with both values and alias", func() {
		spec.AliasTarget = albAlias()
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.Succeed())
	})

	ginkgo.It("rejects a ttl on an alias record", func() {
		spec.Values = nil
		spec.AliasTarget = albAlias()
		spec.Ttl = proto.Int32(300)
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.Succeed())
	})

	ginkgo.It("rejects an explicit zero ttl on an alias record", func() {
		// Presence is what conflicts: even ttl 0 is a set TTL on an alias.
		spec.Values = nil
		spec.AliasTarget = albAlias()
		spec.Ttl = proto.Int32(0)
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.Succeed())
	})

	ginkgo.It("rejects a standard record without a ttl", func() {
		spec.Ttl = nil
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.Succeed())
	})

	ginkgo.It("accepts an explicit zero ttl on a standard record", func() {
		// AWS's contract: TTL 0 is valid and means "never cache".
		spec.Ttl = proto.Int32(0)
		gomega.Expect(protovalidate.Validate(spec)).To(gomega.Succeed())
	})

	ginkgo.It("accepts a ttl above one week up to AWS's maximum", func() {
		for _, ttl := range []int32{604801, 2147483647} {
			spec.Ttl = proto.Int32(ttl)
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.Succeed(), "ttl %d should be valid", ttl)
		}
	})

	ginkgo.It("rejects a negative ttl", func() {
		spec.Ttl = proto.Int32(-1)
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.Succeed())
	})

	ginkgo.It("rejects a value above AWS's 4000-character limit", func() {
		spec.Type = "TXT"
		spec.Values = []*foreignkeyv1.StringValueOrRef{strRef(strings.Repeat("a", 4001))}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.Succeed())
	})

	ginkgo.It("rejects an alias missing its target zone id", func() {
		spec.Values = nil
		spec.Ttl = nil
		spec.AliasTarget = &AwsRoute53AliasTarget{
			DnsName: strRef("d1234abcd.cloudfront.net"),
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.Succeed())
	})

	// -------------------------------------------------------------------------
	// Routing policy contracts
	// -------------------------------------------------------------------------

	ginkgo.It("rejects a routing policy without a set identifier", func() {
		spec.RoutingPolicy = &AwsRoute53RoutingPolicy{
			Policy: &AwsRoute53RoutingPolicy_Weighted{Weighted: &AwsRoute53WeightedPolicy{Weight: 50}},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.Succeed())
	})

	ginkgo.It("rejects a set identifier without a routing policy", func() {
		// AWS scopes SetIdentifier to non-simple routing and rejects it on
		// simple records — dead configuration guarded both ways.
		spec.SetIdentifier = "orphaned"
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.Succeed())
	})

	ginkgo.It("rejects a malformed latency-routing region", func() {
		spec.SetIdentifier = "l"
		for _, region := range []string{"US-EAST-1", "useast1", "us-east", "eu_west_1"} {
			spec.RoutingPolicy = &AwsRoute53RoutingPolicy{
				Policy: &AwsRoute53RoutingPolicy_Latency{Latency: &AwsRoute53LatencyPolicy{Region: region}},
			}
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.Succeed(), "region %q should be rejected", region)
		}
	})

	ginkgo.It("accepts extended-partition latency regions", func() {
		spec.SetIdentifier = "l"
		for _, region := range []string{"us-gov-west-1", "cn-north-1", "eusc-de-east-1", "ap-southeast-7"} {
			spec.RoutingPolicy = &AwsRoute53RoutingPolicy{
				Policy: &AwsRoute53RoutingPolicy_Latency{Latency: &AwsRoute53LatencyPolicy{Region: region}},
			}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.Succeed(), "region %q should be valid", region)
		}
	})

	ginkgo.It("accepts geolocation by continent and the default-record country", func() {
		spec.SetIdentifier = "g"
		for _, geo := range []*AwsRoute53GeolocationPolicy{
			{Continent: "EU"},
			{Country: "*"},
		} {
			spec.RoutingPolicy = &AwsRoute53RoutingPolicy{
				Policy: &AwsRoute53RoutingPolicy_Geolocation{Geolocation: geo},
			}
			gomega.Expect(protovalidate.Validate(spec)).To(gomega.Succeed())
		}
	})

	ginkgo.It("rejects malformed geolocation codes", func() {
		spec.SetIdentifier = "g"
		for _, geo := range []*AwsRoute53GeolocationPolicy{
			{Continent: "XX"},                    // not a continent code
			{Country: "de"},                      // ISO codes are uppercase
			{Country: "US", Subdivision: "ABCD"}, // subdivision is at most 3 chars
		} {
			spec.RoutingPolicy = &AwsRoute53RoutingPolicy{
				Policy: &AwsRoute53RoutingPolicy_Geolocation{Geolocation: geo},
			}
			gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.Succeed())
		}
	})

	ginkgo.It("rejects a weight above 255", func() {
		spec.SetIdentifier = "w"
		spec.RoutingPolicy = &AwsRoute53RoutingPolicy{
			Policy: &AwsRoute53RoutingPolicy_Weighted{Weighted: &AwsRoute53WeightedPolicy{Weight: 256}},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.Succeed())
	})

	ginkgo.It("rejects latency routing without a region", func() {
		spec.SetIdentifier = "l"
		spec.RoutingPolicy = &AwsRoute53RoutingPolicy{
			Policy: &AwsRoute53RoutingPolicy_Latency{Latency: &AwsRoute53LatencyPolicy{}},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.Succeed())
	})

	ginkgo.It("rejects an invalid failover type", func() {
		spec.SetIdentifier = "p"
		spec.RoutingPolicy = &AwsRoute53RoutingPolicy{
			Policy: &AwsRoute53RoutingPolicy_Failover{Failover: &AwsRoute53FailoverPolicy{FailoverType: "ACTIVE"}},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.Succeed())
	})

	ginkgo.It("rejects geolocation routing without any location", func() {
		spec.SetIdentifier = "g"
		spec.RoutingPolicy = &AwsRoute53RoutingPolicy{
			Policy: &AwsRoute53RoutingPolicy_Geolocation{Geolocation: &AwsRoute53GeolocationPolicy{}},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.Succeed())
	})

	ginkgo.It("rejects geoproximity with two location determinants", func() {
		spec.SetIdentifier = "gp"
		spec.RoutingPolicy = &AwsRoute53RoutingPolicy{
			Policy: &AwsRoute53RoutingPolicy_Geoproximity{Geoproximity: &AwsRoute53GeoproximityPolicy{
				AwsRegion:      "us-east-1",
				LocalZoneGroup: "us-east-1-bue-1",
			}},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.Succeed())
	})

	ginkgo.It("rejects geoproximity with no location determinant", func() {
		spec.SetIdentifier = "gp"
		spec.RoutingPolicy = &AwsRoute53RoutingPolicy{
			Policy: &AwsRoute53RoutingPolicy_Geoproximity{Geoproximity: &AwsRoute53GeoproximityPolicy{Bias: 10}},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.Succeed())
	})

	ginkgo.It("rejects a geoproximity bias outside -99..99", func() {
		spec.SetIdentifier = "gp"
		spec.RoutingPolicy = &AwsRoute53RoutingPolicy{
			Policy: &AwsRoute53RoutingPolicy_Geoproximity{Geoproximity: &AwsRoute53GeoproximityPolicy{
				AwsRegion: "us-east-1",
				Bias:      100,
			}},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.Succeed())
	})

	ginkgo.It("rejects CIDR routing without a location name", func() {
		spec.SetIdentifier = "c"
		spec.RoutingPolicy = &AwsRoute53RoutingPolicy{
			Policy: &AwsRoute53RoutingPolicy_Cidr{Cidr: &AwsRoute53CidrPolicy{CollectionId: "cf1234"}},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.Succeed())
	})

	ginkgo.It("rejects multivalue answer routing on an alias record", func() {
		spec.Values = nil
		spec.Ttl = nil
		spec.AliasTarget = albAlias()
		spec.SetIdentifier = "server-1"
		spec.RoutingPolicy = &AwsRoute53RoutingPolicy{
			Policy: &AwsRoute53RoutingPolicy_MultivalueAnswer{MultivalueAnswer: &AwsRoute53MultivalueAnswerPolicy{}},
		}
		gomega.Expect(protovalidate.Validate(spec)).NotTo(gomega.Succeed())
	})
})
