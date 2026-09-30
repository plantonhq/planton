//go:build e2e

// Package aws contains end-to-end tests that provision real AWS resources via
// Planton IaC modules and verify them through the AWS SDK. Credentials come from
// the ambient chain (local AWS SSO or GitHub Actions OIDC -- never a stored
// secret); see the aa_e2e harness package.
//
// Run with: go test -tags=e2e -timeout=30m -v ./e2e/aws/...
package aws

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	awse2e "github.com/plantonhq/planton/catalog/aws/aa_e2e"
	"github.com/plantonhq/planton/e2e/framework/discovery"
	"github.com/plantonhq/planton/e2e/framework/provider"
	"github.com/plantonhq/planton/e2e/framework/runner"
	profilepkg "github.com/plantonhq/planton/pkg/e2e/profile"
	componentv1 "github.com/plantonhq/planton/qa/componente2eprofile/v1"
)

var (
	testHarness      *awse2e.Harness
	repoRoot         string
	runID            string
	pulumiBackendURL string
	// assertApplyIdempotency mirrors the provider profile's
	// assert_apply_idempotency field: when armed, every scenario lifecycle
	// gains the IDEMPOTENCY phase (re-plan after apply must be empty).
	assertApplyIdempotency bool
)

func TestMain(m *testing.M) {
	var err error
	repoRoot, err = filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve repo root: %v\n", err)
		os.Exit(1)
	}

	runID = uuid.New().String()[:8]

	backendDir, err := os.MkdirTemp("", "planton-e2e-aws-pulumi-*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create temp backend dir: %v\n", err)
		os.Exit(1)
	}
	pulumiBackendURL = "file://" + backendDir
	defer os.RemoveAll(backendDir)

	if err := runner.PulumiLogin(pulumiBackendURL); err != nil {
		fmt.Fprintf(os.Stderr, "failed to login to pulumi backend: %v\n", err)
		os.Exit(1)
	}

	providerProfile, err := profilepkg.LoadProviderProfile(repoRoot, "aws")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load AWS provider E2E profile: %v\n", err)
		os.Exit(1)
	}
	assertApplyIdempotency = providerProfile.GetSpec().GetAssertApplyIdempotency()

	testHarness = awse2e.NewHarness()
	ctx := context.Background()
	if err := testHarness.Setup(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup AWS harness: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	if err := testHarness.Teardown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to teardown AWS harness: %v\n", err)
	}

	os.Exit(code)
}

// --- AWS S3 Bucket (walking skeleton) ---

func TestAwsS3Bucket_Pulumi(t *testing.T) { runAllScenariosForComponent(t, "awss3bucket", "pulumi") }
func TestAwsS3Bucket_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awss3bucket", "terraform")
}

// --- AWS VPC (thin root of the networking graph) ---

func TestAwsVpc_Pulumi(t *testing.T)    { runAllScenariosForComponent(t, "awsvpc", "pulumi") }
func TestAwsVpc_Terraform(t *testing.T) { runAllScenariosForComponent(t, "awsvpc", "terraform") }

// --- AWS Subnet (first composed topology: deploys an AwsVpc prerequisite) ---

func TestAwsSubnet_Pulumi(t *testing.T)    { runAllScenariosForComponent(t, "awssubnet", "pulumi") }
func TestAwsSubnet_Terraform(t *testing.T) { runAllScenariosForComponent(t, "awssubnet", "terraform") }

// --- AWS Elastic IP (standalone allocation; also serves as the NAT Gateway prerequisite) ---

func TestAwsElasticIp_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awselasticip", "pulumi")
}
func TestAwsElasticIp_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awselasticip", "terraform")
}

// --- AWS NAT Gateway (deep composed topology: AwsVpc -> AwsSubnet -> AwsElasticIp) ---

func TestAwsNatGateway_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsnatgateway", "pulumi")
}
func TestAwsNatGateway_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsnatgateway", "terraform")
}

// --- AWS Internet Gateway (attaches to a gateway-free AwsVpc prerequisite) ---

func TestAwsInternetGateway_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsinternetgateway", "pulumi")
}
func TestAwsInternetGateway_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsinternetgateway", "terraform")
}

// --- AWS Egress-Only Internet Gateway (IPv6 outbound-only; attaches to an AwsVpc prerequisite) ---

func TestAwsEgressOnlyInternetGateway_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsegressonlyinternetgateway", "pulumi")
}
func TestAwsEgressOnlyInternetGateway_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsegressonlyinternetgateway", "terraform")
}

// --- AWS VPC Endpoint (private service access: gateway on the VPC's default route table + interface on the subnet pair) ---

func TestAwsVpcEndpoint_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsvpcendpoint", "pulumi")
}
func TestAwsVpcEndpoint_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsvpcendpoint", "terraform")
}

// --- AWS IAM Policy (leaf of the identity graph: a standalone managed policy) ---

func TestAwsIamPolicy_Pulumi(t *testing.T) { runAllScenariosForComponent(t, "awsiampolicy", "pulumi") }
func TestAwsIamPolicy_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsiampolicy", "terraform")
}

// --- AWS IAM Role (trust policy + attachments + inline policies) ---

func TestAwsIamRole_Pulumi(t *testing.T) { runAllScenariosForComponent(t, "awsiamrole", "pulumi") }
func TestAwsIamRole_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsiamrole", "terraform")
}

// --- AWS IAM Instance Profile (composed identity chain: deploys an AwsIamRole prerequisite) ---

func TestAwsIamInstanceProfile_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsiaminstanceprofile", "pulumi")
}
func TestAwsIamInstanceProfile_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsiaminstanceprofile", "terraform")
}

// --- AWS IAM User (long-lived identity with access key + force-destroy teardown) ---

func TestAwsIamUser_Pulumi(t *testing.T) { runAllScenariosForComponent(t, "awsiamuser", "pulumi") }
func TestAwsIamUser_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsiamuser", "terraform")
}

// --- AWS IAM OIDC Provider (keyless-federation trust anchor; synthetic issuer) ---

func TestAwsIamOidcProvider_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsiamoidcprovider", "pulumi")
}
func TestAwsIamOidcProvider_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsiamoidcprovider", "terraform")
}

// --- AWS LB Target Group (routing destination; deploys an AwsVpc prerequisite) ---

func TestAwsLbTargetGroup_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awslbtargetgroup", "pulumi")
}
func TestAwsLbTargetGroup_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awslbtargetgroup", "terraform")
}

// --- AWS ALB (two-AZ placement: deploys the VPC + two-subnet prerequisite pair) ---

func TestAwsAlb_Pulumi(t *testing.T)    { runAllScenariosForComponent(t, "awsalb", "pulumi") }
func TestAwsAlb_Terraform(t *testing.T) { runAllScenariosForComponent(t, "awsalb", "terraform") }

// --- AWS NLB (internal single-mapping smoke; deploys the VPC + subnet prerequisites) ---

func TestAwsNlb_Pulumi(t *testing.T)    { runAllScenariosForComponent(t, "awsnlb", "pulumi") }
func TestAwsNlb_Terraform(t *testing.T) { runAllScenariosForComponent(t, "awsnlb", "terraform") }

// --- AWS LB Listener (deepest chain: VPC -> subnets -> ALB -> target group -> listener) ---

func TestAwsLbListener_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awslblistener", "pulumi")
}
func TestAwsLbListener_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awslblistener", "terraform")
}

// --- AWS LB Listener Rule (per-service routing; rides the full family chain) ---

func TestAwsLbListenerRule_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awslblistenerrule", "pulumi")
}
func TestAwsLbListenerRule_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awslblistenerrule", "terraform")
}

func TestAwsLaunchTemplate_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awslaunchtemplate", "pulumi")
}
func TestAwsLaunchTemplate_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awslaunchtemplate", "terraform")
}

func TestAwsAutoScalingGroup_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsautoscalinggroup", "pulumi")
}
func TestAwsAutoScalingGroup_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsautoscalinggroup", "terraform")
}

// --- AWS RDS Cluster (Aurora Serverless v2 with one folded db.serverless instance; managed master password) ---

func TestAwsRdsCluster_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsrdscluster", "pulumi")
}
func TestAwsRdsCluster_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsrdscluster", "terraform")
}

// --- AWS RDS Instance (single-AZ postgres db.t4g.micro on gp3; managed master password) ---

func TestAwsRdsInstance_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsrdsinstance", "pulumi")
}
func TestAwsRdsInstance_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsrdsinstance", "terraform")
}

// --- AWS DocumentDB (one db.t4g.medium folded instance; managed master password) ---

func TestAwsDocumentDb_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsdocumentdb", "pulumi")
}
func TestAwsDocumentDb_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsdocumentdb", "terraform")
}

// --- AWS Neptune Cluster (Serverless 1-2 NCU with one folded db.serverless instance) ---

func TestAwsNeptuneCluster_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsneptunecluster", "pulumi")
}
func TestAwsNeptuneCluster_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsneptunecluster", "terraform")
}

// --- AWS Redshift Cluster (single ra3.large node; managed admin password) ---

func TestAwsRedshiftCluster_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsredshiftcluster", "pulumi")
}
func TestAwsRedshiftCluster_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsredshiftcluster", "terraform")
}

// --- AWS Redshift Serverless Namespace (account-level data plane; managed admin password) ---

func TestAwsRedshiftServerlessNamespace_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsredshiftserverlessnamespace", "pulumi")
}
func TestAwsRedshiftServerlessNamespace_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsredshiftserverlessnamespace", "terraform")
}

// --- AWS Redshift Serverless Workgroup (8-RPU compute plane on the namespace + three-AZ subnet trio) ---

func TestAwsRedshiftServerlessWorkgroup_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsredshiftserverlessworkgroup", "pulumi")
}
func TestAwsRedshiftServerlessWorkgroup_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsredshiftserverlessworkgroup", "terraform")
}

// --- AWS DynamoDB (true leaf; on-demand table with GSI, streams, PITR, folded satellites) ---

func TestAwsDynamodb_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsdynamodb", "pulumi")
}
func TestAwsDynamodb_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsdynamodb", "terraform")
}

// --- AWS Kinesis Data Stream (true leaf; ON_DEMAND full surface with folded resource policy) ---

func TestAwsKinesisStream_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awskinesisstream", "pulumi")
}
func TestAwsKinesisStream_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awskinesisstream", "terraform")
}

// --- AWS Kinesis Stream Consumer (enhanced fan-out; AwsKinesisStream prerequisite chain) ---

func TestAwsKinesisStreamConsumer_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awskinesisstreamconsumer", "pulumi")
}
func TestAwsKinesisStreamConsumer_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awskinesisstreamconsumer", "terraform")
}

// --- AWS Kinesis Firehose (Direct PUT -> extended_s3 + Splunk-with-processors; S3 bucket + IAM role prerequisite chain) ---

func TestAwsKinesisFirehose_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awskinesisfirehose", "pulumi")
}
func TestAwsKinesisFirehose_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awskinesisfirehose", "terraform")
}

// --- AWS ECR repository (true leaf; folded lifecycle rules + repository policy) ---

func TestAwsEcrRepo_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsecrrepo", "pulumi")
}
func TestAwsEcrRepo_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsecrrepo", "terraform")
}

// --- AWS Route 53 hosted zone (public leaf + private-VPC composed arm) ---

func TestAwsRoute53Zone_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsroute53zone", "pulumi")
}
func TestAwsRoute53Zone_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsroute53zone", "terraform")
}

// --- AWS Route 53 DNS record (AwsRoute53Zone prerequisite chain) ---

func TestAwsRoute53DnsRecord_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsroute53dnsrecord", "pulumi")
}
func TestAwsRoute53DnsRecord_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsroute53dnsrecord", "terraform")
}

// --- AWS Route 53 health check (true leaf; disabled probe, zero external traffic) ---

func TestAwsRoute53HealthCheck_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsroute53healthcheck", "pulumi")
}
func TestAwsRoute53HealthCheck_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsroute53healthcheck", "terraform")
}

// --- AWS ElastiCache RBAC (account-level; no VPC prerequisite) ---

func TestAwsElasticacheUser_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awselasticacheuser", "pulumi")
}
func TestAwsElasticacheUser_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awselasticacheuser", "terraform")
}

func TestAwsElasticacheUserGroup_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awselasticacheusergroup", "pulumi")
}
func TestAwsElasticacheUserGroup_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awselasticacheusergroup", "terraform")
}

// --- AWS ElastiCache caches (AwsSubnet two-AZ prerequisite chain) ---

func TestAwsRedisElasticache_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsrediselasticache", "pulumi")
}
func TestAwsRedisElasticache_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsrediselasticache", "terraform")
}

func TestAwsMemcachedElasticache_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsmemcachedelasticache", "pulumi")
}
func TestAwsMemcachedElasticache_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsmemcachedelasticache", "terraform")
}

func TestAwsServerlessElasticache_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsserverlesselasticache", "pulumi")
}
func TestAwsServerlessElasticache_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsserverlesselasticache", "terraform")
}

// --- AWS MSK Cluster (two kafka.t3.small brokers, SASL/IAM; subnet + security-group prerequisite chain) ---

func TestAwsMskCluster_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsmskcluster", "pulumi")
}
func TestAwsMskCluster_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsmskcluster", "terraform")
}

// --- AWS MSK Serverless Cluster (capacity-managed Kafka, SASL/IAM only; subnet prerequisite chain) ---

func TestAwsMskServerlessCluster_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsmskserverlesscluster", "pulumi")
}
func TestAwsMskServerlessCluster_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsmskserverlesscluster", "terraform")
}

// --- AWS MWAA Environment (managed Airflow; S3 + IAM-role fixtures via the scenario's e2e-prerequisites annotation) ---

func TestAwsMwaaEnvironment_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsmwaaenvironment", "pulumi")
}
func TestAwsMwaaEnvironment_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsmwaaenvironment", "terraform")
}

// --- AWS OpenSearch Domain (true leaf; public single-node t3.small.search domain) ---

func TestAwsOpenSearchDomain_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsopensearchdomain", "pulumi")
}
func TestAwsOpenSearchDomain_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsopensearchdomain", "terraform")
}

// --- AWS EKS Cluster (the control plane; slowest single resource in the suite, ~25 min/engine) ---

func TestAwsEksCluster_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsekscluster", "pulumi")
}
func TestAwsEksCluster_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsekscluster", "terraform")
}

// --- AWS EKS Node Group (zero-capacity pool on a full control-plane prerequisite chain) ---

func TestAwsEksNodeGroup_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awseksnodegroup", "pulumi")
}
func TestAwsEksNodeGroup_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awseksnodegroup", "terraform")
}

// --- AWS EKS Add-on (kube-proxy adopted with OVERWRITE on the control-plane prerequisite chain) ---

func TestAwsEksAddon_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awseksaddon", "pulumi")
}
func TestAwsEksAddon_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awseksaddon", "terraform")
}

// --- AWS EKS Fargate Profile (namespace selector; no pods scheduled; full control-plane prerequisite chain) ---

func TestAwsEksFargateProfile_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awseksfargateprofile", "pulumi")
}
func TestAwsEksFargateProfile_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awseksfargateprofile", "terraform")
}

// --- AWS EKS Access Entry (STANDARD entry + view-policy association on an API-mode cluster) ---

func TestAwsEksAccessEntry_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awseksaccessentry", "pulumi")
}
func TestAwsEksAccessEntry_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awseksaccessentry", "terraform")
}

// --- AWS ECS Task Definition (revision registration only; no task launches; execution-role prerequisite) ---

func TestAwsEcsTaskDefinition_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsecstaskdefinition", "pulumi")
}
func TestAwsEcsTaskDefinition_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsecstaskdefinition", "terraform")
}

// --- AWS ECS Service (desiredCount 0 on the cluster + task-definition + subnet prerequisite chain) ---

func TestAwsEcsService_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsecsservice", "pulumi")
}
func TestAwsEcsService_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsecsservice", "terraform")
}

func TestAwsPlantonRunner_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsplantonrunner", "pulumi")
}

func TestAwsPlantonRunner_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsplantonrunner", "terraform")
}

// --- AWS ECS Cluster (Fargate leaf + the EC2-capacity chain via the scenario's e2e-prerequisites annotation) ---

func TestAwsEcsCluster_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsecscluster", "pulumi")
}
func TestAwsEcsCluster_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsecscluster", "terraform")
}

// --- AWS EC2 Instance (one t3.micro; subnet + security-group fixtures via the scenario's e2e-prerequisites annotation) ---

func TestAwsEc2Instance_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsec2instance", "pulumi")
}
func TestAwsEc2Instance_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsec2instance", "terraform")
}

// --- AWS Security Group (rules-rich group on the VPC prerequisite; prefix lists, self-reference, SG-to-SG refs) ---

func TestAwsSecurityGroup_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awssecuritygroup", "pulumi")
}
func TestAwsSecurityGroup_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awssecuritygroup", "terraform")
}

// --- AWS CloudFront (placeholder-origin distribution on the default certificate; true leaf, slow edge propagation) ---

func TestAwsCloudFront_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awscloudfront", "pulumi")
}
func TestAwsCloudFront_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awscloudfront", "terraform")
}

// --- AWS Certificate Manager (no-zone requested certificate resting in PENDING_VALIDATION; true leaf) ---

func TestAwsCertManagerCert_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awscertmanagercert", "pulumi")
}
func TestAwsCertManagerCert_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awscertmanagercert", "terraform")
}

// --- AWS KMS Key (symmetric key with one alias; true leaf) ---

func TestAwsKmsKey_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awskmskey", "pulumi")
}
func TestAwsKmsKey_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awskmskey", "terraform")
}

// --- AWS Lambda (zip-backed function; S3 bucket + object-set + execution-role chain) ---

func TestAwsLambda_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awslambda", "pulumi")
}
func TestAwsLambda_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awslambda", "terraform")
}

// --- AWS Lambda Event Source Mapping (SQS queue -> mapping -> function chain) ---

func TestAwsLambdaEventSourceMapping_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awslambdaeventsourcemapping", "pulumi")
}
func TestAwsLambdaEventSourceMapping_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awslambdaeventsourcemapping", "terraform")
}

// --- AWS SQS Queue (FIFO queue, full FIFO delivery surface) ---

func TestAwsSqsQueue_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awssqsqueue", "pulumi")
}
func TestAwsSqsQueue_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awssqsqueue", "terraform")
}

// --- AWS SNS Topic (standard topic with tracing + policy struct) ---

func TestAwsSnsTopic_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awssnstopic", "pulumi")
}
func TestAwsSnsTopic_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awssnstopic", "terraform")
}

// --- AWS SNS Subscription (topic registry prerequisite + composed SQS queue) ---

func TestAwsSnsSubscription_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awssnssubscription", "pulumi")
}
func TestAwsSnsSubscription_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awssnssubscription", "terraform")
}

// --- AWS EventBridge Bus (custom bus with log_config + composed DLQ) ---

func TestAwsEventBridgeBus_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awseventbridgebus", "pulumi")
}
func TestAwsEventBridgeBus_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awseventbridgebus", "terraform")
}

// --- AWS EventBridge Rule (event-pattern rule on custom bus + SQS target) ---

func TestAwsEventBridgeRule_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awseventbridgerule", "pulumi")
}
func TestAwsEventBridgeRule_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awseventbridgerule", "terraform")
}

// --- AWS CloudWatch observability family ---

func TestAwsCloudwatchLogGroup_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awscloudwatchloggroup", "pulumi")
}
func TestAwsCloudwatchLogGroup_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awscloudwatchloggroup", "terraform")
}

func TestAwsCloudwatchAlarm_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awscloudwatchalarm", "pulumi")
}
func TestAwsCloudwatchAlarm_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awscloudwatchalarm", "terraform")
}

func TestAwsCloudwatchCompositeAlarm_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awscloudwatchcompositealarm", "pulumi")
}
func TestAwsCloudwatchCompositeAlarm_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awscloudwatchcompositealarm", "terraform")
}

// --- AWS serverless front door (Step Functions + HTTP API family) ---

func TestAwsStepFunction_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsstepfunction", "pulumi")
}
func TestAwsStepFunction_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsstepfunction", "terraform")
}

func TestAwsHttpApiGateway_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awshttpapigateway", "pulumi")
}
func TestAwsHttpApiGateway_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awshttpapigateway", "terraform")
}

func TestAwsHttpApiVpcLink_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awshttpapivpclink", "pulumi")
}
func TestAwsHttpApiVpcLink_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awshttpapivpclink", "terraform")
}

func TestAwsHttpApiDomain_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awshttpapidomain", "pulumi")
}
func TestAwsHttpApiDomain_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awshttpapidomain", "terraform")
}

// --- AWS Cognito family (pool root + pool-scoped client/IdP/resource server) ---

func TestAwsCognitoUserPool_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awscognitouserpool", "pulumi")
}
func TestAwsCognitoUserPool_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awscognitouserpool", "terraform")
}

func TestAwsCognitoUserPoolClient_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awscognitouserpoolclient", "pulumi")
}
func TestAwsCognitoUserPoolClient_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awscognitouserpoolclient", "terraform")
}

func TestAwsCognitoIdentityProvider_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awscognitoidentityprovider", "pulumi")
}
func TestAwsCognitoIdentityProvider_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awscognitoidentityprovider", "terraform")
}

func TestAwsCognitoResourceServer_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awscognitoresourceserver", "pulumi")
}
func TestAwsCognitoResourceServer_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awscognitoresourceserver", "terraform")
}

// --- AWS EFS family (file system root + file-system-scoped access point) ---

func TestAwsElasticFileSystem_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awselasticfilesystem", "pulumi")
}
func TestAwsElasticFileSystem_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awselasticfilesystem", "terraform")
}

func TestAwsEfsAccessPoint_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsefsaccesspoint", "pulumi")
}
func TestAwsEfsAccessPoint_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsefsaccesspoint", "terraform")
}

// --- AWS WAF family (leaf sets + the web ACL composing them by reference) ---

func TestAwsWafIpSet_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awswafipset", "pulumi")
}
func TestAwsWafIpSet_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awswafipset", "terraform")
}

func TestAwsWafRegexPatternSet_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awswafregexpatternset", "pulumi")
}
func TestAwsWafRegexPatternSet_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awswafregexpatternset", "terraform")
}

func TestAwsWafWebAcl_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awswafwebacl", "pulumi")
}
func TestAwsWafWebAcl_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awswafwebacl", "terraform")
}

func TestAwsBatchComputeEnvironment_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsbatchcomputeenvironment", "pulumi")
}
func TestAwsBatchComputeEnvironment_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsbatchcomputeenvironment", "terraform")
}

func TestAwsBatchJobQueue_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsbatchjobqueue", "pulumi")
}
func TestAwsBatchJobQueue_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsbatchjobqueue", "terraform")
}

func TestAwsBatchSchedulingPolicy_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsbatchschedulingpolicy", "pulumi")
}
func TestAwsBatchSchedulingPolicy_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsbatchschedulingpolicy", "terraform")
}

func TestAwsBatchJobDefinition_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsbatchjobdefinition", "pulumi")
}
func TestAwsBatchJobDefinition_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsbatchjobdefinition", "terraform")
}

func TestAwsSesConfigurationSet_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awssesconfigurationset", "pulumi")
}
func TestAwsSesConfigurationSet_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awssesconfigurationset", "terraform")
}

func TestAwsSesEmailIdentity_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awssesemailidentity", "pulumi")
}
func TestAwsSesEmailIdentity_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awssesemailidentity", "terraform")
}

func TestAwsAppRunnerService_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsapprunnerservice", "pulumi")
}

func TestAwsAppRunnerService_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsapprunnerservice", "terraform")
}

func TestAwsAppRunnerAutoScalingConfiguration_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsapprunnerautoscalingconfiguration", "pulumi")
}

func TestAwsAppRunnerAutoScalingConfiguration_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsapprunnerautoscalingconfiguration", "terraform")
}

func TestAwsAppRunnerVpcConnector_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsapprunnervpcconnector", "pulumi")
}

func TestAwsAppRunnerVpcConnector_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsapprunnervpcconnector", "terraform")
}

func TestAwsAppRunnerObservabilityConfiguration_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsapprunnerobservabilityconfiguration", "pulumi")
}

func TestAwsAppRunnerObservabilityConfiguration_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsapprunnerobservabilityconfiguration", "terraform")
}

// --- AWS Transit Gateway family (hub, VPC attachment, route table) ---

func TestAwsTransitGateway_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awstransitgateway", "pulumi")
}

func TestAwsTransitGateway_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awstransitgateway", "terraform")
}

func TestAwsTransitGatewayVpcAttachment_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awstransitgatewayvpcattachment", "pulumi")
}

func TestAwsTransitGatewayVpcAttachment_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awstransitgatewayvpcattachment", "terraform")
}

func TestAwsTransitGatewayRouteTable_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awstransitgatewayroutetable", "pulumi")
}

func TestAwsTransitGatewayRouteTable_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awstransitgatewayroutetable", "terraform")
}

// --- AWS analytics pair (Athena workgroup, Glue Data Catalog database) ---

func TestAwsAthenaWorkgroup_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsathenaworkgroup", "pulumi")
}

func TestAwsAthenaWorkgroup_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsathenaworkgroup", "terraform")
}

func TestAwsGlueCatalogDatabase_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsgluecatalogdatabase", "pulumi")
}

func TestAwsGlueCatalogDatabase_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsgluecatalogdatabase", "terraform")
}

// --- AWS CI/CD pair (CodeBuild project, CodePipeline) ---

func TestAwsCodeBuildProject_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awscodebuildproject", "pulumi")
}

func TestAwsCodeBuildProject_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awscodebuildproject", "terraform")
}

func TestAwsCodePipeline_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awscodepipeline", "pulumi")
}

func TestAwsCodePipeline_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awscodepipeline", "terraform")
}

// --- AWS MemoryDB family (user, ACL, cluster) ---

func TestAwsMemorydbUser_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsmemorydbuser", "pulumi")
}

func TestAwsMemorydbUser_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsmemorydbuser", "terraform")
}

func TestAwsMemorydbAcl_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsmemorydbacl", "pulumi")
}

func TestAwsMemorydbAcl_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsmemorydbacl", "terraform")
}

func TestAwsMemorydbCluster_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsmemorydbcluster", "pulumi")
}

func TestAwsMemorydbCluster_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsmemorydbcluster", "terraform")
}

// --- AWS Client VPN ---

func TestAwsClientVpn_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsclientvpn", "pulumi")
}

func TestAwsClientVpn_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsclientvpn", "terraform")
}

// AwsGlobalAccelerator: a dependency-free minimal lane plus an Elastic
// IP-composed lane (the EIP fixture resolves into the polymorphic endpoint
// reference).
func TestAwsGlobalAccelerator_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsglobalaccelerator", "pulumi")
}
func TestAwsGlobalAccelerator_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsglobalaccelerator", "terraform")
}

// --- AWS FSx family (Lustre, OpenZFS, Windows, data repository association) ---

func TestAwsFsxLustreFileSystem_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsfsxlustrefilesystem", "pulumi")
}

func TestAwsFsxLustreFileSystem_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsfsxlustrefilesystem", "terraform")
}

func TestAwsFsxOpenzfsFileSystem_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsfsxopenzfsfilesystem", "pulumi")
}

func TestAwsFsxOpenzfsFileSystem_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsfsxopenzfsfilesystem", "terraform")
}

func TestAwsFsxWindowsFileSystem_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsfsxwindowsfilesystem", "pulumi")
}

func TestAwsFsxWindowsFileSystem_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsfsxwindowsfilesystem", "terraform")
}

func TestAwsFsxDataRepositoryAssociation_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsfsxdatarepositoryassociation", "pulumi")
}

func TestAwsFsxDataRepositoryAssociation_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsfsxdatarepositoryassociation", "terraform")
}

// --- AWS FSx for NetApp ONTAP (file system → SVM → volume) ---

func TestAwsFsxOntapFileSystem_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsfsxontapfilesystem", "pulumi")
}

func TestAwsFsxOntapFileSystem_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsfsxontapfilesystem", "terraform")
}

func TestAwsFsxOntapStorageVirtualMachine_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsfsxontapstoragevirtualmachine", "pulumi")
}

func TestAwsFsxOntapStorageVirtualMachine_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsfsxontapstoragevirtualmachine", "terraform")
}

func TestAwsFsxOntapVolume_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsfsxontapvolume", "pulumi")
}

func TestAwsFsxOntapVolume_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsfsxontapvolume", "terraform")
}

// AwsS3ObjectSet: a minimal single-object lane plus a full-surface lane
// (metadata/header breadth, checksum, SSE-S3 override, STANDARD_IA, website
// redirect, force_destroy versioned purge), both riding the shared versioned
// S3 bucket fixture.
func TestAwsS3ObjectSet_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awss3objectset", "pulumi")
}

func TestAwsS3ObjectSet_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awss3objectset", "terraform")
}

// AwsSagemakerDomain: a required-only minimal lane plus a full-surface lane
// (tag propagation, CloudTrail attribution, Docker, all four IAM-compatible
// app baselines with idle shutdown, S3 sharing, both inheritance planes,
// POSIX identity, Studio UI hiding, role-free Canvas governance), both with
// homeEfsRetentionPolicy Delete for zero-orphan teardown.
func TestAwsSagemakerDomain_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awssagemakerdomain", "pulumi")
}

func TestAwsSagemakerDomain_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awssagemakerdomain", "terraform")
}

func TestAwsSecretsManagerSecret_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awssecretsmanagersecret", "pulumi")
}

func TestAwsSecretsManagerSecret_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awssecretsmanagersecret", "terraform")
}

func TestAwsBedrockGuardrail_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsbedrockguardrail", "pulumi")
}

func TestAwsBedrockGuardrail_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsbedrockguardrail", "terraform")
}

func TestAwsBedrockCustomModel_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsbedrockcustommodel", "pulumi")
}

func TestAwsBedrockCustomModel_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsbedrockcustommodel", "terraform")
}

func TestAwsBedrockInferenceProfile_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsbedrockinferenceprofile", "pulumi")
}

func TestAwsBedrockInferenceProfile_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsbedrockinferenceprofile", "terraform")
}

func TestAwsBedrockProvisionedThroughput_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsbedrockprovisionedthroughput", "pulumi")
}

func TestAwsBedrockProvisionedThroughput_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsbedrockprovisionedthroughput", "terraform")
}

func TestAwsBedrockModelAccess_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsbedrockmodelaccess", "pulumi")
}

func TestAwsBedrockModelAccess_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsbedrockmodelaccess", "terraform")
}

func TestAwsOpenSearchServerlessCollection_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsopensearchserverlesscollection", "pulumi")
}

func TestAwsOpenSearchServerlessCollection_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsopensearchserverlesscollection", "terraform")
}

func TestAwsBedrockAgent_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsbedrockagent", "pulumi")
}

func TestAwsBedrockAgent_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsbedrockagent", "terraform")
}

// --- AWS Bedrock Knowledge Base (the VECTOR lane stores vectors in the
// standing S3 Vectors fixture the harness ensures below; see
// aa_e2e.EnsureS3VectorsKnowledgeBaseFixture) ---

func ensureS3VectorsFixture(t *testing.T) {
	t.Helper()
	if err := awse2e.EnsureS3VectorsKnowledgeBaseFixture(context.Background()); err != nil {
		t.Fatalf("S3 Vectors knowledge-base fixture: %v", err)
	}
}

func TestAwsBedrockKnowledgeBase_Pulumi(t *testing.T) {
	ensureS3VectorsFixture(t)
	runAllScenariosForComponent(t, "awsbedrockknowledgebase", "pulumi")
}

func TestAwsBedrockKnowledgeBase_Terraform(t *testing.T) {
	ensureS3VectorsFixture(t)
	runAllScenariosForComponent(t, "awsbedrockknowledgebase", "terraform")
}

func TestAwsBedrockFlow_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsbedrockflow", "pulumi")
}

func TestAwsBedrockFlow_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsbedrockflow", "terraform")
}

func TestAwsBedrockPrompt_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsbedrockprompt", "pulumi")
}

func TestAwsBedrockPrompt_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsbedrockprompt", "terraform")
}

// --- AWS Bedrock AgentCore (the runtime lanes execute the code bundle the
// harness seeds below; see aa_e2e.EnsureAgentCoreCodeBundleFixture) ---

func ensureAgentCoreCodeFixture(t *testing.T) {
	t.Helper()
	if err := awse2e.EnsureAgentCoreCodeBundleFixture(context.Background()); err != nil {
		t.Fatalf("AgentCore code-bundle fixture: %v", err)
	}
}

func TestAwsBedrockAgentCoreRuntime_Pulumi(t *testing.T) {
	ensureAgentCoreCodeFixture(t)
	runAllScenariosForComponent(t, "awsbedrockagentcoreruntime", "pulumi")
}

func TestAwsBedrockAgentCoreRuntime_Terraform(t *testing.T) {
	ensureAgentCoreCodeFixture(t)
	runAllScenariosForComponent(t, "awsbedrockagentcoreruntime", "terraform")
}

func TestAwsBedrockAgentCoreGateway_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsbedrockagentcoregateway", "pulumi")
}

func TestAwsBedrockAgentCoreGateway_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsbedrockagentcoregateway", "terraform")
}

func TestAwsBedrockAgentCoreMemory_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsbedrockagentcorememory", "pulumi")
}

func TestAwsBedrockAgentCoreMemory_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsbedrockagentcorememory", "terraform")
}

func TestAwsBedrockAgentCoreIdentity_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsbedrockagentcoreidentity", "pulumi")
}

func TestAwsBedrockAgentCoreIdentity_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsbedrockagentcoreidentity", "terraform")
}

func TestAwsBedrockAgentCoreTools_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsbedrockagentcoretools", "pulumi")
}

func TestAwsBedrockAgentCoreTools_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsbedrockagentcoretools", "terraform")
}

// --- AWS SageMaker (models, serving, MLOps) ---

func TestAwsSagemakerModel_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awssagemakermodel", "pulumi")
}

func TestAwsSagemakerModel_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awssagemakermodel", "terraform")
}

func TestAwsSagemakerEndpoint_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awssagemakerendpoint", "pulumi")
}

func TestAwsSagemakerEndpoint_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awssagemakerendpoint", "terraform")
}

func TestAwsSagemakerNotebookInstance_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awssagemakernotebookinstance", "pulumi")
}

func TestAwsSagemakerNotebookInstance_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awssagemakernotebookinstance", "terraform")
}

func TestAwsSagemakerFeatureGroup_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awssagemakerfeaturegroup", "pulumi")
}

func TestAwsSagemakerFeatureGroup_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awssagemakerfeaturegroup", "terraform")
}

func TestAwsSagemakerModelRegistry_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awssagemakermodelregistry", "pulumi")
}

func TestAwsSagemakerModelRegistry_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awssagemakermodelregistry", "terraform")
}

func TestAwsSagemakerPipeline_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awssagemakerpipeline", "pulumi")
}

func TestAwsSagemakerPipeline_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awssagemakerpipeline", "terraform")
}

func TestAwsSagemakerImage_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awssagemakerimage", "pulumi")
}

func TestAwsSagemakerImage_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awssagemakerimage", "terraform")
}

func TestAwsSagemakerMlflowServer_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awssagemakermlflowserver", "pulumi")
}

func TestAwsSagemakerMlflowServer_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awssagemakermlflowserver", "terraform")
}

func TestAwsSagemakerMlflowApp_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awssagemakermlflowapp", "pulumi")
}

func TestAwsSagemakerMlflowApp_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awssagemakermlflowapp", "terraform")
}

// --- AWS API Gateway REST (v1) family + AgentCore Evaluations ---

func TestAwsRestApiGateway_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsrestapigateway", "pulumi")
}

func TestAwsRestApiGateway_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsrestapigateway", "terraform")
}

func TestAwsRestApiDomain_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsrestapidomain", "pulumi")
}

func TestAwsRestApiDomain_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsrestapidomain", "terraform")
}

func TestAwsRestApiUsagePlan_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsrestapiusageplan", "pulumi")
}

func TestAwsRestApiUsagePlan_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsrestapiusageplan", "terraform")
}

func TestAwsRestApiVpcLink_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsrestapivpclink", "pulumi")
}

func TestAwsRestApiVpcLink_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsrestapivpclink", "terraform")
}

func TestAwsBedrockAgentCoreEvaluation_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsbedrockagentcoreevaluation", "pulumi")
}

func TestAwsBedrockAgentCoreEvaluation_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsbedrockagentcoreevaluation", "terraform")
}

func TestAwsApiGatewayAccountSettings_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsapigatewayaccountsettings", "pulumi")
}

func TestAwsApiGatewayAccountSettings_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsapigatewayaccountsettings", "terraform")
}

func TestAwsBedrockInvocationLogging_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsbedrockinvocationlogging", "pulumi")
}

func TestAwsBedrockInvocationLogging_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsbedrockinvocationlogging", "terraform")
}

// --- AWS Bedrock AgentCore Token Vault (the CMK lane binds the standing
// KMS key the harness ensures below -- SetTokenVaultCMK validates the
// PREVIOUS key is ENABLED on every write, so a chain-deployed key whose
// teardown precedes the service-managed revert wedges the vault; see
// aa_e2e.EnsureTokenVaultKMSKeyFixture) ---

func ensureTokenVaultKMSKeyFixture(t *testing.T) {
	t.Helper()
	if err := awse2e.EnsureTokenVaultKMSKeyFixture(context.Background()); err != nil {
		t.Fatalf("token-vault KMS key fixture: %v", err)
	}
}

func TestAwsBedrockAgentCoreTokenVault_Pulumi(t *testing.T) {
	ensureTokenVaultKMSKeyFixture(t)
	runAllScenariosForComponent(t, "awsbedrockagentcoretokenvault", "pulumi")
}

func TestAwsBedrockAgentCoreTokenVault_Terraform(t *testing.T) {
	ensureTokenVaultKMSKeyFixture(t)
	runAllScenariosForComponent(t, "awsbedrockagentcoretokenvault", "terraform")
}

func TestAwsSesAccountSettings_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awssesaccountsettings", "pulumi")
}

func TestAwsSesAccountSettings_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awssesaccountsettings", "terraform")
}

func TestAwsCloudTrail_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awscloudtrail", "pulumi")
}

func TestAwsCloudTrail_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awscloudtrail", "terraform")
}

func TestAwsConfigRecorder_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsconfigrecorder", "pulumi")
}

func TestAwsConfigRecorder_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsconfigrecorder", "terraform")
}

func TestAwsConfigRule_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsconfigrule", "pulumi")
}

func TestAwsConfigRule_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsconfigrule", "terraform")
}

func TestAwsGuardDuty_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsguardduty", "pulumi")
}

func TestAwsGuardDuty_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsguardduty", "terraform")
}

func TestAwsCloudTrailEventDataStore_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awscloudtraileventdatastore", "pulumi")
}

func TestAwsCloudTrailEventDataStore_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awscloudtraileventdatastore", "terraform")
}

func TestAwsConfigAggregator_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsconfigaggregator", "pulumi")
}

func TestAwsConfigAggregator_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsconfigaggregator", "terraform")
}

func TestAwsConfigConformancePack_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsconfigconformancepack", "pulumi")
}

func TestAwsConfigConformancePack_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsconfigconformancepack", "terraform")
}

func TestAwsGuardDutyMalwareProtectionPlan_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsguarddutymalwareprotectionplan", "pulumi")
}

func TestAwsGuardDutyMalwareProtectionPlan_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsguarddutymalwareprotectionplan", "terraform")
}

func TestAwsBackupVault_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsbackupvault", "pulumi")
}

func TestAwsBackupVault_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsbackupvault", "terraform")
}

func TestAwsBackupPlan_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsbackupplan", "pulumi")
}

func TestAwsBackupPlan_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsbackupplan", "terraform")
}

func TestAwsBackupFramework_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsbackupframework", "pulumi")
}

func TestAwsBackupFramework_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsbackupframework", "terraform")
}

func TestAwsBackupReportPlan_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsbackupreportplan", "pulumi")
}

func TestAwsBackupReportPlan_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsbackupreportplan", "terraform")
}

func TestAwsBackupRestoreTestingPlan_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsbackuprestoretestingplan", "pulumi")
}

func TestAwsBackupRestoreTestingPlan_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsbackuprestoretestingplan", "terraform")
}

func TestAwsBackupSettings_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsbackupsettings", "pulumi")
}

func TestAwsBackupSettings_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsbackupsettings", "terraform")
}

func TestAwsSsmParameter_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsssmparameter", "pulumi")
}

func TestAwsSsmParameter_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsssmparameter", "terraform")
}

func TestAwsSsmDocument_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsssmdocument", "pulumi")
}

func TestAwsSsmDocument_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsssmdocument", "terraform")
}

func TestAwsSsmAssociation_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsssmassociation", "pulumi")
}

func TestAwsSsmAssociation_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsssmassociation", "terraform")
}

func TestAwsSsmMaintenanceWindow_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsssmmaintenancewindow", "pulumi")
}

func TestAwsSsmMaintenanceWindow_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsssmmaintenancewindow", "terraform")
}

func TestAwsSsmPatchBaseline_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsssmpatchbaseline", "pulumi")
}

func TestAwsSsmPatchBaseline_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsssmpatchbaseline", "terraform")
}

func TestAwsOrganization_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsorganization", "pulumi")
}

func TestAwsOrganization_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsorganization", "terraform")
}

func TestAwsOrganizationalUnit_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsorganizationalunit", "pulumi")
}

func TestAwsOrganizationalUnit_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsorganizationalunit", "terraform")
}

func TestAwsOrganizationAccount_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsorganizationaccount", "pulumi")
}

func TestAwsOrganizationAccount_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsorganizationaccount", "terraform")
}

func TestAwsOrganizationPolicy_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsorganizationpolicy", "pulumi")
}

func TestAwsOrganizationPolicy_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsorganizationpolicy", "terraform")
}

func TestAwsBudget_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsbudget", "pulumi")
}

func TestAwsBudget_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsbudget", "terraform")
}

func TestAwsCostAnomalyMonitor_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awscostanomalymonitor", "pulumi")
}

func TestAwsCostAnomalyMonitor_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awscostanomalymonitor", "terraform")
}

func TestAwsCostCategory_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awscostcategory", "pulumi")
}

func TestAwsCostCategory_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awscostcategory", "terraform")
}

func TestAwsIamGroup_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsiamgroup", "pulumi")
}

func TestAwsIamGroup_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsiamgroup", "terraform")
}

func TestAwsIamSamlProvider_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsiamsamlprovider", "pulumi")
}

func TestAwsIamSamlProvider_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsiamsamlprovider", "terraform")
}

func TestAwsIamAccountSettings_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsiamaccountsettings", "pulumi")
}

func TestAwsIamAccountSettings_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsiamaccountsettings", "terraform")
}

func TestAwsCloudwatchDashboard_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awscloudwatchdashboard", "pulumi")
}

func TestAwsCloudwatchDashboard_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awscloudwatchdashboard", "terraform")
}

func TestAwsCloudwatchSynthetics_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awscloudwatchsynthetics", "pulumi")
}

func TestAwsCloudwatchSynthetics_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awscloudwatchsynthetics", "terraform")
}

func TestAwsCloudwatchLogDelivery_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awscloudwatchlogdelivery", "pulumi")
}

func TestAwsCloudwatchLogDelivery_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awscloudwatchlogdelivery", "terraform")
}

func TestAwsCloudwatchLogAccountPolicy_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awscloudwatchlogaccountpolicy", "pulumi")
}

func TestAwsCloudwatchLogAccountPolicy_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awscloudwatchlogaccountpolicy", "terraform")
}

func TestAwsCloudwatchLogAnomalyDetector_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awscloudwatchloganomalydetector", "pulumi")
}

func TestAwsCloudwatchLogAnomalyDetector_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awscloudwatchloganomalydetector", "terraform")
}

func TestAwsCloudwatchLogResourcePolicy_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awscloudwatchlogresourcepolicy", "pulumi")
}

func TestAwsCloudwatchLogResourcePolicy_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awscloudwatchlogresourcepolicy", "terraform")
}

func TestAwsManagedPrometheus_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsmanagedprometheus", "pulumi")
}

func TestAwsManagedPrometheus_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsmanagedprometheus", "terraform")
}

func TestAwsManagedPrometheusScraper_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsmanagedprometheusscraper", "pulumi")
}

func TestAwsManagedPrometheusScraper_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsmanagedprometheusscraper", "terraform")
}

func TestAwsEventBridgePipe_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awseventbridgepipe", "pulumi")
}

func TestAwsEventBridgePipe_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awseventbridgepipe", "terraform")
}

func TestAwsEventBridgeScheduler_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awseventbridgescheduler", "pulumi")
}

func TestAwsEventBridgeScheduler_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awseventbridgescheduler", "terraform")
}

func TestAwsEventBridgeApiDestination_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awseventbridgeapidestination", "pulumi")
}

func TestAwsEventBridgeApiDestination_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awseventbridgeapidestination", "terraform")
}

func TestAwsVpcPeering_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsvpcpeering", "pulumi")
}

func TestAwsVpcPeering_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsvpcpeering", "terraform")
}

func TestAwsNetworkAcl_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsnetworkacl", "pulumi")
}

func TestAwsNetworkAcl_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsnetworkacl", "terraform")
}

func TestAwsManagedPrefixList_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsmanagedprefixlist", "pulumi")
}

func TestAwsManagedPrefixList_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsmanagedprefixlist", "terraform")
}

func TestAwsEbsVolume_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsebsvolume", "pulumi")
}

func TestAwsEbsVolume_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsebsvolume", "terraform")
}

func TestAwsEbsSnapshot_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsebssnapshot", "pulumi")
}

func TestAwsEbsSnapshot_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsebssnapshot", "terraform")
}

func TestAwsDlmLifecyclePolicy_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsdlmlifecyclepolicy", "pulumi")
}

func TestAwsDlmLifecyclePolicy_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsdlmlifecyclepolicy", "terraform")
}

func TestAwsS3DirectoryBucket_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awss3directorybucket", "pulumi")
}

func TestAwsS3DirectoryBucket_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awss3directorybucket", "terraform")
}

func TestAwsS3TableBucket_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awss3tablebucket", "pulumi")
}

func TestAwsS3TableBucket_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awss3tablebucket", "terraform")
}

func TestAwsS3VectorBucket_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awss3vectorbucket", "pulumi")
}

func TestAwsS3VectorBucket_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awss3vectorbucket", "terraform")
}

func TestAwsRoute53ResolverEndpoint_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsroute53resolverendpoint", "pulumi")
}

func TestAwsRoute53ResolverEndpoint_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsroute53resolverendpoint", "terraform")
}

func TestAwsRoute53ResolverFirewall_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsroute53resolverfirewall", "pulumi")
}

func TestAwsRoute53ResolverFirewall_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsroute53resolverfirewall", "terraform")
}

func TestAwsRoute53ResolverQueryLog_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsroute53resolverquerylog", "pulumi")
}

func TestAwsRoute53ResolverQueryLog_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsroute53resolverquerylog", "terraform")
}

func TestAwsCloudMapNamespace_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awscloudmapnamespace", "pulumi")
}

func TestAwsCloudMapNamespace_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awscloudmapnamespace", "terraform")
}

func TestAwsLambdaLayer_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awslambdalayer", "pulumi")
}

func TestAwsLambdaLayer_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awslambdalayer", "terraform")
}

func TestAwsRdsProxy_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsrdsproxy", "pulumi")
}

func TestAwsRdsProxy_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsrdsproxy", "terraform")
}

func TestAwsAppSyncApi_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsappsyncapi", "pulumi")
}

func TestAwsAppSyncApi_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsappsyncapi", "terraform")
}

func TestAwsAuroraDsql_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsauroradsql", "pulumi")
}

func TestAwsAuroraDsql_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsauroradsql", "terraform")
}

func TestAwsEcrRegistrySettings_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsecrregistrysettings", "pulumi")
}

func TestAwsEcrRegistrySettings_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsecrregistrysettings", "terraform")
}

func TestAwsPrivateCa_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "awsprivateca", "pulumi")
}

func TestAwsPrivateCa_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "awsprivateca", "terraform")
}

// runAllScenariosForComponent discovers and runs all E2E scenarios for an AWS component.
func runAllScenariosForComponent(t *testing.T, component, engine string) {
	t.Helper()

	if cp, err := profilepkg.LoadComponentProfile(repoRoot, "aws", component); err == nil && cp.Spec != nil {
		switch cp.Spec.Status {
		case componentv1.ComponentE2EProfileSpec_deferred,
			componentv1.ComponentE2EProfileSpec_skip,
			componentv1.ComponentE2EProfileSpec_stub,
			// pending_proof: fully authored, offline-validated, awaiting its
			// first live proof. The proving session flips the profile to green
			// immediately before executing the lanes; until then a sweep must
			// never run it.
			componentv1.ComponentE2EProfileSpec_pending_proof:
			reason := cp.Spec.DeferredReason
			if reason == "" {
				reason = cp.Spec.Status.String()
			}
			t.Skipf("component %s E2E profile status is %s: %s", component, cp.Spec.Status, reason)
		}
	}

	moduleDir, err := discovery.ModuleDir(repoRoot, "aws", component, engine)
	if err != nil {
		t.Fatalf("failed to locate %s %s module: %v", component, engine, err)
	}

	if !fileExists(moduleDir) {
		t.Skipf("component %s %s module not found at %s", component, engine, moduleDir)
	}

	scenarios, err := discovery.DiscoverTestScenarios(repoRoot, "aws", component)
	if err != nil {
		t.Fatalf("failed to discover test scenarios for %s: %v", component, err)
	}

	if len(scenarios) == 0 {
		t.Skipf("no test scenarios found for %s", component)
	}

	t.Logf("Discovered %d scenarios for %s [%s]", len(scenarios), component, engine)

	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.Name, func(t *testing.T) {
			runSingleScenario(t, component, moduleDir, engine, scenario)
		})
	}
}

func runSingleScenario(t *testing.T, component, moduleDir, engine string, scenario discovery.TestScenario) {
	t.Helper()

	tc := &provider.ComponentTestContext{
		Component:    component,
		Provider:     "aws",
		Engine:       engine,
		ModuleDir:    moduleDir,
		ManifestPath: scenario.ManifestPath,
		RepoRoot:     repoRoot,
		RunID:        runID,
		T:            t,
		// A prerequisite deploys on Pulumi whenever its kind has a Pulumi
		// module, as these kinds do — even for Terraform scenarios — so the
		// backend URL must be set unconditionally.
		// Leaving it empty makes the dependency stacks fall back to the
		// machine's ambient `pulumi login` backend, coupling the run to
		// stale developer state.
		BackendURL:             pulumiBackendURL,
		AssertApplyIdempotency: assertApplyIdempotency,
	}

	if engine == "pulumi" {
		// GenerateStackName enforces the length cap uniqueness-preservingly
		// (blind truncation here would collide long kind names' scenarios).
		tc.StackName = runner.GenerateStackName(component+"-"+scenario.Name, runID)
	}

	ctx := context.Background()
	result := runner.RunComponentTest(ctx, tc, testHarness)

	for _, phase := range result.Phases {
		status := "PASS"
		if !phase.Passed {
			status = "FAIL"
		}
		t.Logf("  %s: %s (%s)", phase.Phase, status, phase.Duration)
		if phase.Error != nil {
			t.Logf("    Error: %v", phase.Error)
		}
	}

	if !result.Passed {
		t.Fatalf("scenario %s/%s [%s] failed (total: %s)", component, scenario.Name, engine, result.Duration)
	}

	t.Logf("scenario %s/%s [%s] passed (total: %s)", component, scenario.Name, engine, result.Duration)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
