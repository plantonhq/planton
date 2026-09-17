//go:build e2e

// Package gcp contains end-to-end tests that provision real GCP resources via
// Planton IaC modules and verify them through the Google Cloud APIs.
// Credentials come from the ambient ADC chain (local
// `gcloud auth application-default login` or workload identity federation in
// CI -- never a stored secret); see the aa_e2e harness package. The test
// project resolves from E2E_GCP_PROJECT / GOOGLE_PROJECT / the ADC credential.
//
// Run with: go test -tags=e2e -timeout=120m -v ./e2e/gcp/...
package gcp

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	gcpe2e "github.com/plantonhq/planton/catalog/gcp/aa_e2e"
	"github.com/plantonhq/planton/e2e/framework/discovery"
	"github.com/plantonhq/planton/e2e/framework/provider"
	"strings"

	"github.com/plantonhq/planton/e2e/framework/runner"
	"github.com/plantonhq/planton/pkg/e2e/profile"
	componentv1 "github.com/plantonhq/planton/qa/componente2eprofile/v1"
	gcpstorage "google.golang.org/api/storage/v1"
)

var (
	testHarness      *gcpe2e.Harness
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

	backendDir, err := os.MkdirTemp("", "planton-e2e-gcp-pulumi-*")
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

	providerProfile, err := profile.LoadProviderProfile(repoRoot, "gcp")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load GCP provider E2E profile: %v\n", err)
		os.Exit(1)
	}
	assertApplyIdempotency = providerProfile.GetSpec().GetAssertApplyIdempotency()

	testHarness = gcpe2e.NewHarness()
	ctx := context.Background()
	if err := testHarness.Setup(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup GCP harness: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	if err := testHarness.Teardown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to teardown GCP harness: %v\n", err)
	}

	os.Exit(code)
}

// --- GCP Service Account (the identity leaf everything references) ---

func TestGcpServiceAccount_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpserviceaccount", "pulumi")
}
func TestGcpServiceAccount_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpserviceaccount", "terraform")
}

// --- GCP IAM Custom Role (least-privilege permission bundle) ---

func TestGcpIamCustomRole_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpiamcustomrole", "pulumi")
}
func TestGcpIamCustomRole_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpiamcustomrole", "terraform")
}

// --- GCP Project IAM Member (composed grant: deploys GcpServiceAccount + GcpIamCustomRole prerequisites) ---

func TestGcpProjectIamMember_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpprojectiammember", "pulumi")
}
func TestGcpProjectIamMember_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpprojectiammember", "terraform")
}

// --- GCP Service Account IAM Member (composed grant ON the account: deploys the GcpServiceAccount prerequisite) ---

func TestGcpServiceAccountIamMember_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpserviceaccountiammember", "pulumi")
}
func TestGcpServiceAccountIamMember_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpserviceaccountiammember", "terraform")
}

// --- GCP Workload Identity Pool (the keyless-auth trust boundary) ---

func TestGcpWorkloadIdentityPool_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpworkloadidentitypool", "pulumi")
}
func TestGcpWorkloadIdentityPool_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpworkloadidentitypool", "terraform")
}

// --- GCP Workload Identity Pool Provider (composed issuer: deploys a GcpWorkloadIdentityPool prerequisite) ---

func TestGcpWorkloadIdentityPoolProvider_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpworkloadidentitypoolprovider", "pulumi")
}
func TestGcpWorkloadIdentityPoolProvider_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpworkloadidentitypoolprovider", "terraform")
}

// --- GCP Health Check (the load-balancing family's leaf; global + regional scenarios) ---

func TestGcpHealthCheck_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcphealthcheck", "pulumi")
}
func TestGcpHealthCheck_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcphealthcheck", "terraform")
}

// --- GCP Backend Bucket (composed static origin: deploys a GcpGcsBucket prerequisite) ---

func TestGcpBackendBucket_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpbackendbucket", "pulumi")
}
func TestGcpBackendBucket_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpbackendbucket", "terraform")
}

// --- GCP Backend Service (composed LB hub: deploys a GcpHealthCheck prerequisite) ---

func TestGcpBackendService_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpbackendservice", "pulumi")
}
func TestGcpBackendService_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpbackendservice", "terraform")
}

// --- GCP Subnetwork (composed address plan: deploys a GcpVpcNetwork prerequisite) ---

func TestGcpSubnetwork_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpsubnetwork", "pulumi")
}
func TestGcpSubnetwork_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpsubnetwork", "terraform")
}

// --- GCP Firewall Rule (composed traffic policy: deploys a GcpVpcNetwork fixture via the scenario annotation) ---

func TestGcpFirewallRule_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpfirewallrule", "pulumi")
}
func TestGcpFirewallRule_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpfirewallrule", "terraform")
}

// --- GCP Region Network Endpoint Group (serverless NEG bridge) ---

func TestGcpRegionNetworkEndpointGroup_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpregionnetworkendpointgroup", "pulumi")
}
func TestGcpRegionNetworkEndpointGroup_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpregionnetworkendpointgroup", "terraform")
}

// --- GCP URL Map (composed routing: deploys GcpBackendService + GcpHealthCheck prerequisites) ---

func TestGcpUrlMap_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpurlmap", "pulumi")
}
func TestGcpUrlMap_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpurlmap", "terraform")
}

// --- GCP Managed SSL Certificate (global managed cert leaf) ---

func TestGcpManagedSslCertificate_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpmanagedsslcertificate", "pulumi")
}
func TestGcpManagedSslCertificate_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpmanagedsslcertificate", "terraform")
}

// --- GCP Target HTTP Proxy (composed frontend adapter: deploys the GcpUrlMap prerequisite chain) ---

func TestGcpTargetHttpProxy_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcptargethttpproxy", "pulumi")
}
func TestGcpTargetHttpProxy_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcptargethttpproxy", "terraform")
}

// --- GCP Target HTTPS Proxy (composed TLS frontend: deploys GcpUrlMap + GcpManagedSslCertificate prerequisites) ---

func TestGcpTargetHttpsProxy_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcptargethttpsproxy", "pulumi")
}
func TestGcpTargetHttpsProxy_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcptargethttpsproxy", "terraform")
}

// --- GCP Global Forwarding Rule (the VIP node; the deepest composed chain in the GCP harness) ---

func TestGcpGlobalForwardingRule_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpglobalforwardingrule", "pulumi")
}
func TestGcpGlobalForwardingRule_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpglobalforwardingrule", "terraform")
}

// --- GCP SSL Policy (TLS version/cipher hardening; leaf with global + regional scenarios) ---

func TestGcpSslPolicy_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpsslpolicy", "pulumi")
}
func TestGcpSslPolicy_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpsslpolicy", "terraform")
}

// --- GCP SSL Certificate (self-managed cert upload; leaf with a checked-in throwaway keypair) ---

func TestGcpSslCertificate_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpsslcertificate", "pulumi")
}
func TestGcpSslCertificate_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpsslcertificate", "terraform")
}

// --- GCP Service Networking Connection (private services access peering) ---

func TestGcpServiceNetworkingConnection_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpservicenetworkingconnection", "pulumi")
}
func TestGcpServiceNetworkingConnection_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpservicenetworkingconnection", "terraform")
}

// --- GCP Address (regional static IP reservation) ---

func TestGcpAddress_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpaddress", "pulumi")
}
func TestGcpAddress_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpaddress", "terraform")
}

// --- GCP Global Address (global static IP reservation; the LB VIP class) ---

func TestGcpGlobalAddress_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpglobaladdress", "pulumi")
}
func TestGcpGlobalAddress_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpglobaladdress", "terraform")
}

// --- GCP VPC (deep-rebuilt network; leaf scenario) ---

func TestGcpVpcNetwork_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpvpcnetwork", "pulumi")
}
func TestGcpVpcNetwork_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpvpcnetwork", "terraform")
}

// --- GCP Cloud SQL (composed PSA chain + public/private instance scenarios) ---

func TestGcpCloudSql_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcloudsql", "pulumi")
}
func TestGcpCloudSql_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcloudsql", "terraform")
}

// --- GCP Cloud SQL Database (composed instance prerequisite) ---

func TestGcpCloudSqlDatabase_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcloudsqldatabase", "pulumi")
}
func TestGcpCloudSqlDatabase_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcloudsqldatabase", "terraform")
}

// --- GCP Cloud SQL User (composed instance prerequisite) ---

func TestGcpCloudSqlUser_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcloudsqluser", "pulumi")
}
func TestGcpCloudSqlUser_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcloudsqluser", "terraform")
}

// GcpRedisInstance scenarios (Memorystore for Redis; the PSA scenario rides
// the same service networking chain Cloud SQL private IP uses).
func TestGcpRedisInstance_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpredisinstance", "pulumi")
}
func TestGcpRedisInstance_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpredisinstance", "terraform")
}

// GcpRouterNat scenarios (Cloud Router + NAT gateway; the manual scenario
// proves NAT IPs referencing GcpAddress reservations).
func TestGcpRouterNat_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcprouternat", "pulumi")
}
func TestGcpRouterNat_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcprouternat", "terraform")
}

// GcpGkeCluster scenarios (the slowest resources in the harness — creates run
// 10-25 minutes each; batch runs need -timeout >= 120m). The minimal zonal
// scenario exercises GKE-managed pod ranges; the private regional scenario
// exercises named secondary ranges + the private control plane.
func TestGcpGkeCluster_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpgkecluster", "pulumi")
}
func TestGcpGkeCluster_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpgkecluster", "terraform")
}

// GcpGkeNodePool scenarios. Every scenario deploys a full GKE cluster
// prerequisite chain (VPC -> subnetwork -> zonal cluster, 10-25 minutes)
// before the pool itself; batch runs need -timeout >= 180m. The minimal
// scenario exercises fixed-count sizing; the autoscaling-spot scenario
// exercises scale-to-zero autoscaling, Spot capacity, taints, and kubelet
// hardening.
func TestGcpGkeNodePool_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpgkenodepool", "pulumi")
}
func TestGcpGkeNodePool_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpgkenodepool", "terraform")
}

// GcpCloudRun scenarios. The minimal scenario is a public hello-service leaf
// (~3-5 minutes). Direct VPC egress has no live scenario: GCP holds a
// serverless address reservation in the subnetwork for 1-2 hours after the
// service is destroyed, which no ephemeral teardown can wait out (recorded
// exclusion in the component's e2e/profile.yaml).
func TestGcpCloudRun_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcloudrun", "pulumi")
}
func TestGcpCloudRun_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcloudrun", "terraform")
}

// GcpAlloydbCluster scenarios (PSA chain + primary cluster; ~10–20 minutes
// each). Batch runs need -timeout >= 120m.
func TestGcpAlloydbCluster_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpalloydbcluster", "pulumi")
}
func TestGcpAlloydbCluster_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpalloydbcluster", "terraform")
}

// GcpAlloydbInstance scenarios (cluster prerequisite + read pool instance).
func TestGcpAlloydbInstance_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpalloydbinstance", "pulumi")
}
func TestGcpAlloydbInstance_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpalloydbinstance", "terraform")
}

// GcpAlloydbUser scenarios (cluster prerequisite + BUILT_IN user).
func TestGcpAlloydbUser_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpalloydbuser", "pulumi")
}
func TestGcpAlloydbUser_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpalloydbuser", "terraform")
}

// GcpDnsZone scenarios (public and private managed zones).
func TestGcpDnsZone_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpdnszone", "pulumi")
}
func TestGcpDnsZone_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpdnszone", "terraform")
}

// GcpCloudRunJob scenarios (minimal single-task batch job).
func TestGcpCloudRunJob_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcloudrunjob", "pulumi")
}
func TestGcpCloudRunJob_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcloudrunjob", "terraform")
}

// GcpSpannerInstance scenarios (100-PU minimal + autoscaling; ~1-2 minutes each).
func TestGcpSpannerInstance_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpspannerinstance", "pulumi")
}
func TestGcpSpannerInstance_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpspannerinstance", "terraform")
}

// GcpSpannerDatabase scenarios (instance prerequisite + database with DDL).
func TestGcpSpannerDatabase_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpspannerdatabase", "pulumi")
}
func TestGcpSpannerDatabase_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpspannerdatabase", "terraform")
}

// GcpSpannerBackupSchedule scenarios (instance + database prerequisites +
// daily full-backup schedule).
func TestGcpSpannerBackupSchedule_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpspannerbackupschedule", "pulumi")
}
func TestGcpSpannerBackupSchedule_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpspannerbackupschedule", "terraform")
}

// GcpBigQueryDataset scenarios (minimal US dataset + explicit access ACL).
func TestGcpBigQueryDataset_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpbigquerydataset", "pulumi")
}
func TestGcpBigQueryDataset_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpbigquerydataset", "terraform")
}

// GcpBigQueryTable scenarios (dataset prerequisite + partitioned native table
// + literal-SELECT view).
func TestGcpBigQueryTable_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpbigquerytable", "pulumi")
}
func TestGcpBigQueryTable_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpbigquerytable", "terraform")
}

// GcpServerlessVpcConnector scenarios (network-placement /28 carve +
// subnet-placement against a dedicated /28 subnetwork prerequisite).
func TestGcpServerlessVpcConnector_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpserverlessvpcconnector", "pulumi")
}
func TestGcpServerlessVpcConnector_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpserverlessvpcconnector", "terraform")
}

// GcpCloudFunction scenarios (public HTTP function + connector-egress
// composition). Gen 2 functions need real source in GCS before the apply,
// and the runner has no hook between prerequisite deploy and the scenario
// apply — so the source zip is staged by the test entrypoint first (see
// stageCloudFunctionSource).
func TestGcpCloudFunction_Pulumi(t *testing.T) {
	stageCloudFunctionSource(t, "pulumi")
	runAllScenariosForComponent(t, "gcpcloudfunction", "pulumi")
}
func TestGcpCloudFunction_Terraform(t *testing.T) {
	stageCloudFunctionSource(t, "terraform")
	runAllScenariosForComponent(t, "gcpcloudfunction", "terraform")
}

func TestGcpServiceConnectionPolicy_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpserviceconnectionpolicy", "pulumi")
}
func TestGcpServiceConnectionPolicy_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpserviceconnectionpolicy", "terraform")
}

func TestGcpMemorystoreInstance_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpmemorystoreinstance", "pulumi")
}
func TestGcpMemorystoreInstance_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpmemorystoreinstance", "terraform")
}

func TestGcpBigtableInstance_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpbigtableinstance", "pulumi")
}
func TestGcpBigtableInstance_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpbigtableinstance", "terraform")
}

func TestGcpBigtableTable_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpbigtabletable", "pulumi")
}
func TestGcpBigtableTable_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpbigtabletable", "terraform")
}

func TestGcpFirestoreDatabase_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpfirestoredatabase", "pulumi")
}
func TestGcpFirestoreDatabase_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpfirestoredatabase", "terraform")
}

func TestGcpFirestoreBackupSchedule_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpfirestorebackupschedule", "pulumi")
}
func TestGcpFirestoreBackupSchedule_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpfirestorebackupschedule", "terraform")
}

func TestGcpFirestoreIndex_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpfirestoreindex", "pulumi")
}
func TestGcpFirestoreIndex_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpfirestoreindex", "terraform")
}

// --- GCP Dataproc Autoscaling Policy (the shareable scaling contract clusters attach by reference) ---

func TestGcpDataprocAutoscalingPolicy_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpdataprocautoscalingpolicy", "pulumi")
}
func TestGcpDataprocAutoscalingPolicy_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpdataprocautoscalingpolicy", "terraform")
}

// --- GCP Dataproc Cluster (composed: VPC -> subnetwork -> autoscaling policy chain; ~2-4m per create) ---

func TestGcpDataprocCluster_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpdataproccluster", "pulumi")
}
func TestGcpDataprocCluster_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpdataproccluster", "terraform")
}

// --- GCP Cloud Composer Environment (composed: VPC -> subnetwork chain; 25-45m per create — batch >=240m) ---

func TestGcpCloudComposerEnvironment_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcloudcomposerenvironment", "pulumi")
}
func TestGcpCloudComposerEnvironment_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcloudcomposerenvironment", "terraform")
}

// --- GCP Cloud Composer user workloads Secret (composed: full environment chain — batch >=240m) ---

func TestGcpCloudComposerUserWorkloadsSecret_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcloudcomposeruserworkloadssecret", "pulumi")
}
func TestGcpCloudComposerUserWorkloadsSecret_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcloudcomposeruserworkloadssecret", "terraform")
}

// --- GCP Cloud Composer user workloads ConfigMap (composed: full environment chain — batch >=240m) ---

func TestGcpCloudComposerUserWorkloadsConfigMap_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcloudcomposeruserworkloadsconfigmap", "pulumi")
}
func TestGcpCloudComposerUserWorkloadsConfigMap_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcloudcomposeruserworkloadsconfigmap", "terraform")
}

// --- GCP Pub/Sub Schema (the message contract topics attach by reference) ---

func TestGcpPubSubSchema_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcppubsubschema", "pulumi")
}
func TestGcpPubSubSchema_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcppubsubschema", "terraform")
}

// --- GCP Pub/Sub Topic (minimal + schema-validated on the schema prerequisite) ---

func TestGcpPubSubTopic_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcppubsubtopic", "pulumi")
}
func TestGcpPubSubTopic_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcppubsubtopic", "terraform")
}

// --- GCP Pub/Sub Subscription (composed: schema -> topic -> subscription chain) ---

func TestGcpPubSubSubscription_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcppubsubsubscription", "pulumi")
}
func TestGcpPubSubSubscription_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcppubsubsubscription", "terraform")
}

// --- GCP KMS Key Ring (the permanent container for crypto keys; destroy is
// state-only by GCP design — run-scoped names keep the leftover rings inert) ---

func TestGcpKmsKeyRing_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpkmskeyring", "pulumi")
}
func TestGcpKmsKeyRing_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpkmskeyring", "terraform")
}

// --- GCP KMS Key (composed on the ring prerequisite; destroy destroys key
// VERSIONS and disables rotation — the verifier asserts exactly that) ---

func TestGcpKmsKey_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpkmskey", "pulumi")
}
func TestGcpKmsKey_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpkmskey", "terraform")
}

// --- GCP KMS Key IAM Member (composed key-scoped grant: deploys the ring → key → service account prerequisite chain) ---

func TestGcpKmsKeyIamMember_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpkmskeyiammember", "pulumi")
}
func TestGcpKmsKeyIamMember_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpkmskeyiammember", "terraform")
}

// GcpCloudTasksQueue scenarios: minimal + http-target-oidc (SA chain).
func TestGcpCloudTasksQueue_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcloudtasksqueue", "pulumi")
}
func TestGcpCloudTasksQueue_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcloudtasksqueue", "terraform")
}

// GcpCloudSchedulerJob scenarios: pubsub-target (topic chain) +
// http-target-oidc (SA chain).
func TestGcpCloudSchedulerJob_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcloudschedulerjob", "pulumi")
}
func TestGcpCloudSchedulerJob_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcloudschedulerjob", "terraform")
}

// --- GCP Vertex AI Endpoint (minimal + dedicated-endpoint; fast deploy) ---

func TestGcpVertexAiEndpoint_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpvertexaiendpoint", "pulumi")
}
func TestGcpVertexAiEndpoint_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpvertexaiendpoint", "terraform")
}

// --- GCP Vertex AI Index (Vector Search; empty STREAM_UPDATE index, ~10-60m per create) ---

func TestGcpVertexAiIndex_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpvertexaiindex", "pulumi")
}
func TestGcpVertexAiIndex_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpvertexaiindex", "terraform")
}

// --- GCP Vertex AI Index Endpoint (Vector Search serving surface; public arm, fast) ---

func TestGcpVertexAiIndexEndpoint_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpvertexaiindexendpoint", "pulumi")
}
func TestGcpVertexAiIndexEndpoint_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpvertexaiindexendpoint", "terraform")
}

// --- GCP Vertex AI Deployed Index (composed: index + index endpoint chain; deploy up to 45m) ---

func TestGcpVertexAiDeployedIndex_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpvertexaideployedindex", "pulumi")
}
func TestGcpVertexAiDeployedIndex_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpvertexaideployedindex", "terraform")
}

// --- GCP Vertex AI Notebook (composed: VPC -> subnetwork chain; ~5-15m per create) ---

func TestGcpVertexAiNotebook_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpvertexainotebook", "pulumi")
}
func TestGcpVertexAiNotebook_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpvertexainotebook", "terraform")
}

// GcpArtifactRegistryRepo scenarios: docker-standard (immutable tags +
// both cleanup-policy arms + additive IAM member on the SA prerequisite)
// + remote-docker-hub (the pull-through REMOTE_REPOSITORY arm).
func TestGcpArtifactRegistryRepo_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpartifactregistryrepo", "pulumi")
}
func TestGcpArtifactRegistryRepo_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpartifactregistryrepo", "terraform")
}

// GcpGcsBucket scenarios: minimal (private posture) + features
// (versioning, explicit-zero lifecycle contract, autoclass, soft delete,
// additive IAM member on the SA prerequisite).
func TestGcpGcsBucket_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpgcsbucket", "pulumi")
}
func TestGcpGcsBucket_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpgcsbucket", "terraform")
}

// GcpComputeDisk scenario: the minimal empty data volume (leaf; creates in
// seconds). Its attachment role is proven by the compute instance chain.
func TestGcpComputeDisk_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcomputedisk", "pulumi")
}
func TestGcpComputeDisk_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcomputedisk", "terraform")
}

// GcpComputeInstance scenarios: minimal (subnetwork chain + ephemeral
// external IP) + features (Spot + shielded + SA identity + a pre-created
// GcpComputeDisk attached by self-link reference).
func TestGcpComputeInstance_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcomputeinstance", "pulumi")
}
func TestGcpComputeInstance_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcomputeinstance", "terraform")
}

// GcpFilestoreInstance scenario: BASIC_HDD share on the chained VPC via
// DIRECT_PEERING (~5-12m per create — under the plan-only threshold).
func TestGcpFilestoreInstance_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpfilestoreinstance", "pulumi")
}
func TestGcpFilestoreInstance_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpfilestoreinstance", "terraform")
}

// GcpDnsRecord scenario: static A record in a chained GcpDnsZone.
func TestGcpDnsRecord_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpdnsrecord", "pulumi")
}
func TestGcpDnsRecord_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpdnsrecord", "terraform")
}

// GcpGkeWorkloadIdentityBinding scenario: IAM grant on a chained
// GcpServiceAccount.
func TestGcpGkeWorkloadIdentityBinding_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpgkeworkloadidentitybinding", "pulumi")
}
func TestGcpGkeWorkloadIdentityBinding_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpgkeworkloadidentitybinding", "terraform")
}

// GcpCloudArmorPolicy scenario: leaf allowlist policy with explicit default deny.
func TestGcpCloudArmorPolicy_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcloudarmorpolicy", "pulumi")
}
func TestGcpCloudArmorPolicy_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcloudarmorpolicy", "terraform")
}

// GcpCertManagerDnsAuthorization scenario: DNS authorization in a chained zone.
func TestGcpCertManagerDnsAuthorization_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcertmanagerdnsauthorization", "pulumi")
}
func TestGcpCertManagerDnsAuthorization_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcertmanagerdnsauthorization", "terraform")
}

// GcpCertManagerCert scenario: managed cert composed from zone→auth→record→cert.
func TestGcpCertManagerCert_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcertmanagercert", "pulumi")
}
func TestGcpCertManagerCert_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcertmanagercert", "terraform")
}

// GcpProject is plan-only unless org-level Project Creator credentials are available.
func TestGcpProject_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpproject", "pulumi")
}
func TestGcpProject_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpproject", "terraform")
}

// --- GCP Monitoring Notification Channel (alerting delivery endpoint; free, second-scale) ---

func TestGcpMonitoringNotificationChannel_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpmonitoringnotificationchannel", "pulumi")
}
func TestGcpMonitoringNotificationChannel_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpmonitoringnotificationchannel", "terraform")
}

// --- GCP Monitoring Alert Policy (composes the channel prerequisite fixture) ---

func TestGcpMonitoringAlertPolicy_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpmonitoringalertpolicy", "pulumi")
}
func TestGcpMonitoringAlertPolicy_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpmonitoringalertpolicy", "terraform")
}

// --- GCP Monitoring Uptime Check (probes example.com; free, second-scale) ---

func TestGcpMonitoringUptimeCheck_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpmonitoringuptimecheck", "pulumi")
}
func TestGcpMonitoringUptimeCheck_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpmonitoringuptimecheck", "terraform")
}

// --- GCP Logging Sink (composes the GCS bucket prerequisite; pubsub arm composes a topic) ---

func TestGcpLoggingSink_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcploggingsink", "pulumi")
}
func TestGcpLoggingSink_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcploggingsink", "terraform")
}

// --- GCP Secret Manager Secret (composes the service-account prerequisite for the scoped grant) ---

func TestGcpSecretManagerSecret_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpsecretmanagersecret", "pulumi")
}
func TestGcpSecretManagerSecret_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpsecretmanagersecret", "terraform")
}

func TestGcpIdentityPlatformConfig_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpidentityplatformconfig", "pulumi")
}

func TestGcpIdentityPlatformConfig_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpidentityplatformconfig", "terraform")
}

func TestGcpIdentityPlatformTenant_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpidentityplatformtenant", "pulumi")
}

func TestGcpIdentityPlatformTenant_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpidentityplatformtenant", "terraform")
}

func TestGcpIamOauthClient_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpiamoauthclient", "pulumi")
}

func TestGcpIamOauthClient_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpiamoauthclient", "terraform")
}

func TestGcpIamDenyPolicy_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpiamdenypolicy", "pulumi")
}

func TestGcpIamDenyPolicy_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpiamdenypolicy", "terraform")
}

func TestGcpCloudRunDomainMapping_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcloudrundomainmapping", "pulumi")
}

func TestGcpCloudRunDomainMapping_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcloudrundomainmapping", "terraform")
}

func TestGcpMonitoringDashboard_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpmonitoringdashboard", "pulumi")
}

func TestGcpMonitoringDashboard_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpmonitoringdashboard", "terraform")
}

func TestGcpMonitoringSlo_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpmonitoringslo", "pulumi")
}

func TestGcpMonitoringSlo_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpmonitoringslo", "terraform")
}

// --- GCP Log Bucket (the views-analytics arm links a BigQuery dataset) ---

func TestGcpLogBucket_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcplogbucket", "pulumi")
}

func TestGcpLogBucket_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcplogbucket", "terraform")
}

// --- GCP Log Metric (the bucket-scoped arm composes the log-bucket prerequisite) ---

func TestGcpLogMetric_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcplogmetric", "pulumi")
}

func TestGcpLogMetric_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcplogmetric", "terraform")
}

// --- GCP Workflow (serverless orchestrator; the Eventarc destination target) ---

func TestGcpWorkflow_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpworkflow", "pulumi")
}
func TestGcpWorkflow_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpworkflow", "terraform")
}

// --- GCP Compute MIG (managed instance group: deploys the GcpVpcNetwork prerequisite; the regional scenario adds the GcpHealthCheck fixture) ---

func TestGcpComputeMig_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcomputemig", "pulumi")
}
func TestGcpComputeMig_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcomputemig", "terraform")
}

// --- GCP Eventarc Trigger (event routing: deploys the GcpCloudRun prerequisite) ---

func TestGcpEventarcTrigger_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpeventarctrigger", "pulumi")
}
func TestGcpEventarcTrigger_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpeventarctrigger", "terraform")
}

// --- GCP Eventarc Message Bus (Eventarc Advanced: bus + sources + pipelines + enrollments) ---

func TestGcpEventarcMessageBus_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpeventarcmessagebus", "pulumi")
}
func TestGcpEventarcMessageBus_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpeventarcmessagebus", "terraform")
}

func TestGcpPlantonRunner_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpplantonrunner", "pulumi")
}
func TestGcpPlantonRunner_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpplantonrunner", "terraform")
}

// --- GCP API Key (soft-deleted for 30 days after destroy, and the key id
// stays reserved for that window — scenarios carry the run id in key_id) ---

func TestGcpApiKey_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpapikey", "pulumi")
}

func TestGcpApiKey_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpapikey", "terraform")
}

// --- GCP Firebase Project (enabling Firebase is permanent by Google's
// design: create adopts an already-enabled project, destroy detaches, and
// the test project stays Firebase-enabled — the verifier asserts the
// detach contract, not absence) ---

func TestGcpFirebaseProject_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpfirebaseproject", "pulumi")
}

func TestGcpFirebaseProject_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpfirebaseproject", "terraform")
}

// --- GCP Firebase app registrations (Android, Apple, Web). Each deploys
// the GcpFirebaseProject prerequisite, which ADOPTS the test project's
// permanent enablement; identities carry the run id (a project accepts each
// package name / bundle id once); deletion_policy DELETE removes an app
// immediately and permanently, so the verifier accepts a 404 or a DELETED
// state as gone. The App Check scenarios are the instrument for Firebase's
// documented propagation delay between a new app and its first App Check
// configuration ---

func TestGcpFirebaseAndroidApp_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpfirebaseandroidapp", "pulumi")
}

func TestGcpFirebaseAndroidApp_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpfirebaseandroidapp", "terraform")
}

func TestGcpFirebaseAppleApp_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpfirebaseappleapp", "pulumi")
}

func TestGcpFirebaseAppleApp_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpfirebaseappleapp", "terraform")
}

func TestGcpFirebaseWebApp_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpfirebasewebapp", "pulumi")
}

func TestGcpFirebaseWebApp_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpfirebasewebapp", "terraform")
}

// --- GCP Certificate Map (SNI routing table: deploys the GcpCertManagerCert prerequisite chain) ---

func TestGcpCertificateMap_Pulumi(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcertificatemap", "pulumi")
}
func TestGcpCertificateMap_Terraform(t *testing.T) {
	runAllScenariosForComponent(t, "gcpcertificatemap", "terraform")
}

// runAllScenariosForComponent discovers and runs all E2E scenarios for a GCP component.
func runAllScenariosForComponent(t *testing.T, component, engine string) {
	t.Helper()

	if cp, err := profile.LoadComponentProfile(repoRoot, "gcp", component); err == nil && cp.Spec != nil {
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

	moduleDir, err := discovery.ModuleDir(repoRoot, "gcp", component, engine)
	if err != nil {
		t.Fatalf("failed to locate %s %s module: %v", component, engine, err)
	}

	if !fileExists(moduleDir) {
		t.Skipf("component %s %s module not found at %s", component, engine, moduleDir)
	}

	scenarios, err := discovery.DiscoverTestScenarios(repoRoot, "gcp", component)
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

	// Scenarios needing owner-arranged external credentials (the
	// e2e-required-env annotation) skip honestly where the environment
	// does not carry the arrangement — unset ${E2E_ENV:...} tokens would
	// otherwise fail expansion loudly, turning a deferral into a false
	// failure on every lane without the tokens (CI included).
	if missing, err := runner.ScenarioMissingRequiredEnv(scenario.ManifestPath); err != nil {
		t.Fatalf("reading required-env declaration for scenario %s/%s: %v", component, scenario.Name, err)
	} else if len(missing) > 0 {
		t.Skipf("scenario %s/%s needs owner-arranged environment variables that are unset: %s (per %s)",
			component, scenario.Name, strings.Join(missing, ", "), runner.ScenarioRequiredEnvAnnotation)
	}

	tc := &provider.ComponentTestContext{
		Component:    component,
		Provider:     "gcp",
		Engine:       engine,
		ModuleDir:    moduleDir,
		ManifestPath: scenario.ManifestPath,
		RepoRoot:     repoRoot,
		RunID:        runID,
		T:            t,
		// Dependencies always deploy via Pulumi — even for Terraform
		// scenarios — so the backend URL must be set unconditionally.
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

// stageCloudFunctionSource zips the checked-in function source fixture and
// uploads it to a run-scoped GCS bucket BEFORE the runner starts — Gen 2
// functions cannot deploy without a real source archive, object bytes cannot
// be expressed as IaC, and the runner has no hook between prerequisite deploy
// and the scenario apply. The bucket name reproduces the runner's
// engine-scoped run-id expansion so the scenarios' ${E2E_RUN_ID} tokens land
// exactly on the staged path. Cleanup is registered on the test (the harness
// Teardown is a no-op by design).
func stageCloudFunctionSource(t *testing.T, engine string) {
	t.Helper()
	ctx := context.Background()

	// Must mirror the values in gcpcloudfunction's e2e/scenarios/*.yaml.
	bucketName := "planton-oss-e2e-cldfunc-src-" + runner.EngineScopedRunID(runID, engine)
	const objectName = "functions/hello.zip"

	// GOOGLE_PROJECT is exported by the harness Setup (which TestMain runs
	// before any test), so it is always the project the scenarios deploy into.
	project := os.Getenv("GOOGLE_PROJECT")
	if project == "" {
		t.Fatal("GOOGLE_PROJECT not set — harness Setup should have exported it")
	}

	fixtureDir := filepath.Join(repoRoot, "catalog", "gcp",
		"gcpcloudfunction", "e2e", "fixtures", "function-source")
	zipBytes, err := zipDirectory(fixtureDir)
	if err != nil {
		t.Fatalf("failed to zip function source fixture: %v", err)
	}

	storageService, err := gcpstorage.NewService(ctx)
	if err != nil {
		t.Fatalf("failed to create storage client for source staging: %v", err)
	}

	if _, err := storageService.Buckets.Insert(project, &gcpstorage.Bucket{Name: bucketName}).Context(ctx).Do(); err != nil {
		t.Fatalf("failed to create staging bucket %s: %v", bucketName, err)
	}
	t.Cleanup(func() {
		// Best-effort: the object must go before the bucket can.
		if err := storageService.Objects.Delete(bucketName, objectName).Do(); err != nil {
			t.Logf("warning: failed to delete staged object %s/%s: %v", bucketName, objectName, err)
		}
		if err := storageService.Buckets.Delete(bucketName).Do(); err != nil {
			t.Logf("warning: failed to delete staging bucket %s: %v", bucketName, err)
		}
	})

	if _, err := storageService.Objects.Insert(bucketName, &gcpstorage.Object{Name: objectName}).
		Media(bytes.NewReader(zipBytes)).Context(ctx).Do(); err != nil {
		t.Fatalf("failed to upload source zip to %s/%s: %v", bucketName, objectName, err)
	}

	t.Logf("staged function source at gs://%s/%s (%d bytes)", bucketName, objectName, len(zipBytes))
}

// zipDirectory zips every regular file in dir at the archive root — the shape
// buildpacks expect (main.py / requirements.txt at the top level).
func zipDirectory(dir string) ([]byte, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, err
		}
		fileWriter, err := zipWriter.Create(entry.Name())
		if err != nil {
			return nil, err
		}
		if _, err := fileWriter.Write(content); err != nil {
			return nil, err
		}
	}
	if err := zipWriter.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
