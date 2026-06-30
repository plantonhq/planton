package awsopensearchdomainv1

import (
	"testing"

	"buf.build/go/protovalidate"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	foreignkeyv1 "github.com/plantonhq/planton/apis/dev/planton/shared/foreignkey/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestAwsOpenSearchDomainSpec(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "AwsOpenSearchDomainSpec Validation Suite")
}

// helper to create a StringValueOrRef with a literal value.
func strRef(val string) *foreignkeyv1.StringValueOrRef {
	return &foreignkeyv1.StringValueOrRef{
		LiteralOrRef: &foreignkeyv1.StringValueOrRef_Value{Value: val},
	}
}

// minimalValidSpec returns a minimal valid AwsOpenSearchDomainSpec.
func minimalValidSpec() *AwsOpenSearchDomainSpec {
	return &AwsOpenSearchDomainSpec{
		Region:        "us-west-2",
		EngineVersion: "OpenSearch_2.11",
		ClusterConfig: &AwsOpenSearchDomainClusterConfig{
			InstanceType: "r6g.large.search",
		},
		EbsOptions: &AwsOpenSearchDomainEbsOptions{
			EbsEnabled: true,
			VolumeType: "gp3",
			VolumeSize: 10,
		},
	}
}

var _ = ginkgo.Describe("AwsOpenSearchDomainSpec validations", func() {
	var spec *AwsOpenSearchDomainSpec

	ginkgo.BeforeEach(func() {
		spec = minimalValidSpec()
	})

	// -------------------------------------------------------------------------
	// Happy path
	// -------------------------------------------------------------------------

	ginkgo.It("accepts a minimal valid spec", func() {
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts a full production VPC spec", func() {
		spec.EncryptAtRestEnabled = true
		spec.NodeToNodeEncryptionEnabled = true
		spec.ClusterConfig.DedicatedMasterEnabled = true
		spec.ClusterConfig.DedicatedMasterType = "r6g.large.search"
		spec.ClusterConfig.DedicatedMasterCount = 3
		spec.ClusterConfig.ZoneAwarenessEnabled = true
		spec.ClusterConfig.AvailabilityZoneCount = 3
		spec.ClusterConfig.InstanceCount = proto.Int32(3)
		spec.VpcOptions = &AwsOpenSearchDomainVpcOptions{
			SubnetIds: []*foreignkeyv1.StringValueOrRef{
				strRef("subnet-aaa"), strRef("subnet-bbb"), strRef("subnet-ccc"),
			},
			SecurityGroupIds: []*foreignkeyv1.StringValueOrRef{
				strRef("sg-123"),
			},
		}
		spec.DomainEndpointOptions = &AwsOpenSearchDomainEndpointOptions{
			EnforceHttps:      proto.Bool(true),
			TlsSecurityPolicy: "Policy-Min-TLS-1-2-PFS-2023-10",
		}
		spec.AdvancedSecurityOptions = &AwsOpenSearchDomainAdvancedSecurityOptions{
			Enabled:                     true,
			InternalUserDatabaseEnabled: true,
			MasterUserName:              "admin",
			MasterUserPassword:          strRef("P@ssw0rd!"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts FGAC with IAM master user ARN", func() {
		spec.AdvancedSecurityOptions = &AwsOpenSearchDomainAdvancedSecurityOptions{
			Enabled:       true,
			MasterUserArn: strRef("arn:aws:iam::123456789012:role/opensearch-admin"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts FGAC with internal user database", func() {
		spec.AdvancedSecurityOptions = &AwsOpenSearchDomainAdvancedSecurityOptions{
			Enabled:                     true,
			InternalUserDatabaseEnabled: true,
			MasterUserName:              "admin",
			MasterUserPassword:          strRef("P@ssw0rd!"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts warm storage enabled", func() {
		spec.ClusterConfig.WarmEnabled = true
		spec.ClusterConfig.WarmType = "ultrawarm1.medium.search"
		spec.ClusterConfig.WarmCount = 2
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts cold storage with warm enabled", func() {
		spec.ClusterConfig.WarmEnabled = true
		spec.ClusterConfig.WarmType = "ultrawarm1.medium.search"
		spec.ClusterConfig.WarmCount = 2
		spec.ClusterConfig.ColdStorageEnabled = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts custom endpoint", func() {
		spec.DomainEndpointOptions = &AwsOpenSearchDomainEndpointOptions{
			CustomEndpointEnabled:        true,
			CustomEndpoint:               "search.example.com",
			CustomEndpointCertificateArn: strRef("arn:aws:acm:us-east-1:123456789012:certificate/abc-123"),
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts log publishing all four types with FGAC enabled", func() {
		spec.AdvancedSecurityOptions = &AwsOpenSearchDomainAdvancedSecurityOptions{
			Enabled:       true,
			MasterUserArn: strRef("arn:aws:iam::123456789012:role/admin"),
		}
		spec.LogPublishingOptions = []*AwsOpenSearchDomainLogPublishingOption{
			{LogType: "INDEX_SLOW_LOGS", CloudwatchLogGroupArn: strRef("arn:aws:logs:us-east-1:123456789012:log-group:/os/index-slow")},
			{LogType: "SEARCH_SLOW_LOGS", CloudwatchLogGroupArn: strRef("arn:aws:logs:us-east-1:123456789012:log-group:/os/search-slow")},
			{LogType: "ES_APPLICATION_LOGS", CloudwatchLogGroupArn: strRef("arn:aws:logs:us-east-1:123456789012:log-group:/os/app")},
			{LogType: "AUDIT_LOGS", CloudwatchLogGroupArn: strRef("arn:aws:logs:us-east-1:123456789012:log-group:/os/audit")},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts auto-tune enabled", func() {
		spec.AutoTuneEnabled = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts multi-AZ with standby enabled", func() {
		spec.ClusterConfig.MultiAzWithStandbyEnabled = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts dualstack ip_address_type", func() {
		spec.IpAddressType = "dualstack"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts advanced_options map", func() {
		spec.AdvancedOptions = map[string]string{
			"rest.action.multi.allow_explicit_index": "true",
			"indices.fielddata.cache.size":           "40",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts access_policies as Struct", func() {
		spec.AccessPolicies = &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"Version": structpb.NewStringValue("2012-10-17"),
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts Elasticsearch engine version", func() {
		spec.EngineVersion = "Elasticsearch_7.10"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	ginkgo.It("accepts io1 EBS with IOPS", func() {
		spec.EbsOptions = &AwsOpenSearchDomainEbsOptions{
			EbsEnabled: true,
			VolumeType: "io1",
			VolumeSize: 100,
			Iops:       3000,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).To(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Failure: required fields
	// -------------------------------------------------------------------------

	ginkgo.It("fails when engine_version is missing", func() {
		spec.EngineVersion = ""
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when cluster_config is nil", func() {
		spec.ClusterConfig = nil
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when ebs_options is nil", func() {
		spec.EbsOptions = nil
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Failure: engine_version format CEL
	// -------------------------------------------------------------------------

	ginkgo.It("fails when engine_version has no prefix", func() {
		spec.EngineVersion = "2.11"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when engine_version is missing underscore separator", func() {
		spec.EngineVersion = "OpenSearch2.11"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Failure: cluster_config – dedicated master CEL
	// -------------------------------------------------------------------------

	ginkgo.It("fails when dedicated_master_type set but dedicated_master_enabled false", func() {
		spec.ClusterConfig.DedicatedMasterType = "r6g.large.search"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when dedicated_master_count set but dedicated_master_enabled false", func() {
		spec.ClusterConfig.DedicatedMasterCount = 3
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Failure: cluster_config – zone awareness CEL
	// -------------------------------------------------------------------------

	ginkgo.It("fails when availability_zone_count set but zone_awareness_enabled false", func() {
		spec.ClusterConfig.AvailabilityZoneCount = 3
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when availability_zone_count is 4 (invalid value)", func() {
		spec.ClusterConfig.ZoneAwarenessEnabled = true
		spec.ClusterConfig.AvailabilityZoneCount = 4
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Failure: cluster_config – warm/cold storage CEL
	// -------------------------------------------------------------------------

	ginkgo.It("fails when warm_type set but warm_enabled false", func() {
		spec.ClusterConfig.WarmType = "ultrawarm1.medium.search"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when warm_count set but warm_enabled false", func() {
		spec.ClusterConfig.WarmCount = 3
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when warm_count is 1 (below minimum 2)", func() {
		spec.ClusterConfig.WarmEnabled = true
		spec.ClusterConfig.WarmType = "ultrawarm1.medium.search"
		spec.ClusterConfig.WarmCount = 1
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when cold_storage_enabled without warm_enabled", func() {
		spec.ClusterConfig.ColdStorageEnabled = true
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Failure: EBS CEL validations
	// -------------------------------------------------------------------------

	ginkgo.It("fails when iops set with volume_type gp2", func() {
		spec.EbsOptions = &AwsOpenSearchDomainEbsOptions{
			EbsEnabled: true,
			VolumeType: "gp2",
			VolumeSize: 10,
			Iops:       3000,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when throughput set with volume_type io1", func() {
		spec.EbsOptions = &AwsOpenSearchDomainEbsOptions{
			EbsEnabled: true,
			VolumeType: "io1",
			VolumeSize: 100,
			Iops:       3000,
			Throughput: 250,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when throughput is below minimum 125", func() {
		spec.EbsOptions = &AwsOpenSearchDomainEbsOptions{
			EbsEnabled: true,
			VolumeType: "gp3",
			VolumeSize: 10,
			Throughput: 50,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when EBS fields set but ebs_enabled false", func() {
		spec.EbsOptions = &AwsOpenSearchDomainEbsOptions{
			EbsEnabled: false,
			VolumeType: "gp3",
			VolumeSize: 10,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Failure: domain endpoint CEL validations
	// -------------------------------------------------------------------------

	ginkgo.It("fails when custom_endpoint set but custom_endpoint_enabled false", func() {
		spec.DomainEndpointOptions = &AwsOpenSearchDomainEndpointOptions{
			CustomEndpoint: "search.example.com",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Failure: encryption CEL validations
	// -------------------------------------------------------------------------

	ginkgo.It("fails when kms_key_id set but encrypt_at_rest_enabled false", func() {
		spec.KmsKeyId = strRef("arn:aws:kms:us-east-1:123456789012:key/my-key")
		// encrypt_at_rest_enabled defaults to false
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Failure: AUDIT_LOGS requires FGAC
	// -------------------------------------------------------------------------

	ginkgo.It("fails when AUDIT_LOGS configured without FGAC enabled", func() {
		spec.LogPublishingOptions = []*AwsOpenSearchDomainLogPublishingOption{
			{
				LogType:               "AUDIT_LOGS",
				CloudwatchLogGroupArn: strRef("arn:aws:logs:us-east-1:123456789012:log-group:/os/audit"),
			},
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Failure: FGAC mutual exclusion and requirements
	// -------------------------------------------------------------------------

	ginkgo.It("fails when both master_user_arn and master_user_name are set", func() {
		spec.AdvancedSecurityOptions = &AwsOpenSearchDomainAdvancedSecurityOptions{
			Enabled:        true,
			MasterUserArn:  strRef("arn:aws:iam::123456789012:role/admin"),
			MasterUserName: "admin",
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	ginkgo.It("fails when FGAC enabled without any master user configured", func() {
		spec.AdvancedSecurityOptions = &AwsOpenSearchDomainAdvancedSecurityOptions{
			Enabled: true,
		}
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})

	// -------------------------------------------------------------------------
	// Failure: ip_address_type CEL
	// -------------------------------------------------------------------------

	ginkgo.It("fails when ip_address_type is invalid", func() {
		spec.IpAddressType = "ipv6"
		err := protovalidate.Validate(spec)
		gomega.Expect(err).NotTo(gomega.BeNil())
	})
})
