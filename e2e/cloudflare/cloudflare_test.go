//go:build e2e

package cloudflare

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	cloudflaree2e "github.com/plantonhq/planton/catalog/cloudflare/aa_e2e"
	"github.com/plantonhq/planton/e2e/framework/discovery"
	"github.com/plantonhq/planton/e2e/framework/provider"
	"github.com/plantonhq/planton/e2e/framework/runner"
	profilepkg "github.com/plantonhq/planton/pkg/e2e/profile"
	kindv1 "github.com/plantonhq/planton/qa/catalogkinde2eprofile/v1"
)

var (
	testHarness            *cloudflaree2e.Harness
	repoRoot               string
	runID                  string
	pulumiBackendURL       string
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

	backendDir, err := os.MkdirTemp("", "planton-e2e-cloudflare-pulumi-*")
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

	// The provider profile owns lane-wide policy (apply-idempotency);
	// reading it here keeps the policy in one committed file instead of
	// scattered test constants.
	providerProfile, err := profilepkg.LoadProviderProfile(repoRoot, "cloudflare")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load Cloudflare provider E2E profile: %v\n", err)
		os.Exit(1)
	}
	assertApplyIdempotency = providerProfile.GetSpec().GetAssertApplyIdempotency()

	testHarness = cloudflaree2e.NewHarness()
	ctx := context.Background()
	if err := testHarness.Setup(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup Cloudflare harness: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	if err := testHarness.Teardown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to teardown Cloudflare harness: %v\n", err)
	}

	os.Exit(code)
}

// Entrypoint pairs for every Cloudflare kind. Names follow the discover
// contract (Test + registry kind name + engine suffix); kinds without an
// e2e/profile.yaml or scenarios skip cleanly, so the full set costs nothing
// while the catalog enrolls kind by kind.

func TestCloudflareDnsZone_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarednszone", "pulumi")
}
func TestCloudflareDnsZone_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarednszone", "terraform")
}

func TestCloudflareKvNamespace_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarekvnamespace", "pulumi")
}
func TestCloudflareKvNamespace_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarekvnamespace", "terraform")
}

func TestCloudflareR2Bucket_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarer2bucket", "pulumi")
}
func TestCloudflareR2Bucket_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarer2bucket", "terraform")
}

func TestCloudflareWorker_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareworker", "pulumi")
}
func TestCloudflareWorker_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareworker", "terraform")
}

func TestCloudflareLoadBalancer_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareloadbalancer", "pulumi")
}
func TestCloudflareLoadBalancer_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareloadbalancer", "terraform")
}

func TestCloudflareD1Database_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflared1database", "pulumi")
}
func TestCloudflareD1Database_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflared1database", "terraform")
}

func TestCloudflareZeroTrustAccessApplication_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustaccessapplication", "pulumi")
}
func TestCloudflareZeroTrustAccessApplication_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustaccessapplication", "terraform")
}

func TestCloudflareDnsRecord_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarednsrecord", "pulumi")
}
func TestCloudflareDnsRecord_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarednsrecord", "terraform")
}

func TestCloudflareRuleset_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareruleset", "pulumi")
}
func TestCloudflareRuleset_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareruleset", "terraform")
}

func TestCloudflareWorkersKvPair_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareworkerskvpair", "pulumi")
}
func TestCloudflareWorkersKvPair_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareworkerskvpair", "terraform")
}

func TestCloudflareHyperdriveConfig_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarehyperdriveconfig", "pulumi")
}
func TestCloudflareHyperdriveConfig_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarehyperdriveconfig", "terraform")
}

func TestCloudflareLoadBalancerPool_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareloadbalancerpool", "pulumi")
}
func TestCloudflareLoadBalancerPool_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareloadbalancerpool", "terraform")
}

func TestCloudflareLoadBalancerMonitor_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareloadbalancermonitor", "pulumi")
}
func TestCloudflareLoadBalancerMonitor_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareloadbalancermonitor", "terraform")
}

func TestCloudflareZeroTrustAccessPolicy_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustaccesspolicy", "pulumi")
}
func TestCloudflareZeroTrustAccessPolicy_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustaccesspolicy", "terraform")
}

func TestCloudflareZeroTrustAccessGroup_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustaccessgroup", "pulumi")
}
func TestCloudflareZeroTrustAccessGroup_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustaccessgroup", "terraform")
}

func TestCloudflareQueue_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarequeue", "pulumi")
}
func TestCloudflareQueue_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarequeue", "terraform")
}

func TestCloudflarePagesProject_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarepagesproject", "pulumi")
}
func TestCloudflarePagesProject_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarepagesproject", "terraform")
}

func TestCloudflareZeroTrustTunnel_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrusttunnel", "pulumi")
}
func TestCloudflareZeroTrustTunnel_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrusttunnel", "terraform")
}

func TestCloudflareZeroTrustTunnelVirtualNetwork_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrusttunnelvirtualnetwork", "pulumi")
}
func TestCloudflareZeroTrustTunnelVirtualNetwork_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrusttunnelvirtualnetwork", "terraform")
}

func TestCloudflareZeroTrustTunnelRoute_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrusttunnelroute", "pulumi")
}
func TestCloudflareZeroTrustTunnelRoute_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrusttunnelroute", "terraform")
}

func TestCloudflareList_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarelist", "pulumi")
}
func TestCloudflareList_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarelist", "terraform")
}

func TestCloudflareListItem_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarelistitem", "pulumi")
}
func TestCloudflareListItem_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarelistitem", "terraform")
}

func TestCloudflareTurnstileWidget_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareturnstilewidget", "pulumi")
}
func TestCloudflareTurnstileWidget_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareturnstilewidget", "terraform")
}

func TestCloudflareEmailRoutingZone_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareemailroutingzone", "pulumi")
}
func TestCloudflareEmailRoutingZone_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareemailroutingzone", "terraform")
}

func TestCloudflareEmailRoutingRule_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareemailroutingrule", "pulumi")
}
func TestCloudflareEmailRoutingRule_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareemailroutingrule", "terraform")
}

func TestCloudflareEmailRoutingAddress_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareemailroutingaddress", "pulumi")
}
func TestCloudflareEmailRoutingAddress_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareemailroutingaddress", "terraform")
}

func TestCloudflareOriginCaCertificate_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareorigincacertificate", "pulumi")
}
func TestCloudflareOriginCaCertificate_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareorigincacertificate", "terraform")
}

func TestCloudflareCertificatePack_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarecertificatepack", "pulumi")
}
func TestCloudflareCertificatePack_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarecertificatepack", "terraform")
}

func TestCloudflareCustomHostname_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarecustomhostname", "pulumi")
}
func TestCloudflareCustomHostname_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarecustomhostname", "terraform")
}

func TestCloudflareCustomHostnameFallbackOrigin_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarecustomhostnamefallbackorigin", "pulumi")
}
func TestCloudflareCustomHostnameFallbackOrigin_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarecustomhostnamefallbackorigin", "terraform")
}

func TestCloudflareZoneSettings_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezonesettings", "pulumi")
}
func TestCloudflareZoneSettings_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezonesettings", "terraform")
}

func TestCloudflareCacheSettings_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarecachesettings", "pulumi")
}
func TestCloudflareCacheSettings_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarecachesettings", "terraform")
}

func TestCloudflareZoneTlsSettings_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezonetlssettings", "pulumi")
}
func TestCloudflareZoneTlsSettings_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezonetlssettings", "terraform")
}

func TestCloudflareCustomSslCertificate_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarecustomsslcertificate", "pulumi")
}
func TestCloudflareCustomSslCertificate_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarecustomsslcertificate", "terraform")
}

func TestCloudflareMtlsCertificate_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflaremtlscertificate", "pulumi")
}
func TestCloudflareMtlsCertificate_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflaremtlscertificate", "terraform")
}

func TestCloudflareAuthenticatedOriginPulls_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareauthenticatedoriginpulls", "pulumi")
}
func TestCloudflareAuthenticatedOriginPulls_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareauthenticatedoriginpulls", "terraform")
}

func TestCloudflareAuthenticatedOriginPullsCertificate_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareauthenticatedoriginpullscertificate", "pulumi")
}
func TestCloudflareAuthenticatedOriginPullsCertificate_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareauthenticatedoriginpullscertificate", "terraform")
}

func TestCloudflareWorkflow_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareworkflow", "pulumi")
}
func TestCloudflareWorkflow_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareworkflow", "terraform")
}

func TestCloudflareSecretsStore_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflaresecretsstore", "pulumi")
}
func TestCloudflareSecretsStore_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflaresecretsstore", "terraform")
}

func TestCloudflareSecretsStoreSecret_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflaresecretsstoresecret", "pulumi")
}
func TestCloudflareSecretsStoreSecret_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflaresecretsstoresecret", "terraform")
}

func TestCloudflareAiGateway_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareaigateway", "pulumi")
}
func TestCloudflareAiGateway_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareaigateway", "terraform")
}

func TestCloudflareZeroTrustAccessIdentityProvider_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustaccessidentityprovider", "pulumi")
}
func TestCloudflareZeroTrustAccessIdentityProvider_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustaccessidentityprovider", "terraform")
}

func TestCloudflareZeroTrustAccessServiceToken_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustaccessservicetoken", "pulumi")
}

func TestCloudflareZeroTrustOrganization_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustorganization", "pulumi")
}
func TestCloudflareZeroTrustOrganization_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustorganization", "terraform")
}

func TestCloudflareZeroTrustAccessInfrastructureTarget_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustaccessinfrastructuretarget", "pulumi")
}
func TestCloudflareZeroTrustAccessInfrastructureTarget_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustaccessinfrastructuretarget", "terraform")
}

func TestCloudflareZeroTrustMcpPortal_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustmcpportal", "pulumi")
}
func TestCloudflareZeroTrustMcpPortal_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustmcpportal", "terraform")
}

func TestCloudflareZeroTrustMcpServer_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustmcpserver", "pulumi")
}
func TestCloudflareZeroTrustMcpServer_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustmcpserver", "terraform")
}

func TestCloudflareZeroTrustGatewaySettings_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustgatewaysettings", "pulumi")
}
func TestCloudflareZeroTrustGatewaySettings_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustgatewaysettings", "terraform")
}

func TestCloudflareZeroTrustDnsLocation_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustdnslocation", "pulumi")
}
func TestCloudflareZeroTrustDnsLocation_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustdnslocation", "terraform")
}

func TestCloudflareZeroTrustDeviceDefaultProfile_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustdevicedefaultprofile", "pulumi")
}
func TestCloudflareZeroTrustDeviceDefaultProfile_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustdevicedefaultprofile", "terraform")
}

func TestCloudflareZeroTrustDeviceCustomProfile_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustdevicecustomprofile", "pulumi")
}
func TestCloudflareZeroTrustDeviceCustomProfile_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustdevicecustomprofile", "terraform")
}

func TestCloudflareZeroTrustDevicePostureRule_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustdeviceposturerule", "pulumi")
}
func TestCloudflareZeroTrustDevicePostureRule_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustdeviceposturerule", "terraform")
}
func TestCloudflareLogpushJob_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarelogpushjob", "pulumi")
}
func TestCloudflareLogpushJob_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarelogpushjob", "terraform")
}
func TestCloudflareNotificationPolicy_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarenotificationpolicy", "pulumi")
}
func TestCloudflareNotificationPolicy_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarenotificationpolicy", "terraform")
}
func TestCloudflareNotificationWebhook_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarenotificationwebhook", "pulumi")
}
func TestCloudflareNotificationWebhook_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarenotificationwebhook", "terraform")
}
func TestCloudflareWebAnalyticsSite_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarewebanalyticssite", "pulumi")
}
func TestCloudflareWebAnalyticsSite_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarewebanalyticssite", "terraform")
}
func TestCloudflareAccountApiToken_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareaccountapitoken", "pulumi")
}
func TestCloudflareAccountApiToken_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareaccountapitoken", "terraform")
}
func TestCloudflareZeroTrustAccessServiceToken_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustaccessservicetoken", "terraform")
}

func TestCloudflareZeroTrustGatewayPolicy_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustgatewaypolicy", "pulumi")
}
func TestCloudflareZeroTrustGatewayPolicy_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustgatewaypolicy", "terraform")
}

func TestCloudflareZeroTrustList_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustlist", "pulumi")
}
func TestCloudflareZeroTrustList_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarezerotrustlist", "terraform")
}

func TestCloudflareIpAccessRule_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareipaccessrule", "pulumi")
}
func TestCloudflareIpAccessRule_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflareipaccessrule", "terraform")
}

func TestCloudflareBotManagement_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarebotmanagement", "pulumi")
}
func TestCloudflareBotManagement_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarebotmanagement", "terraform")
}

func TestCloudflareSnippet_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflaresnippet", "pulumi")
}
func TestCloudflareSnippet_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflaresnippet", "terraform")
}

func TestCloudflareSnippetRules_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflaresnippetrules", "pulumi")
}
func TestCloudflareSnippetRules_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflaresnippetrules", "terraform")
}

func TestCloudflareHealthcheck_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarehealthcheck", "pulumi")
}
func TestCloudflareHealthcheck_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarehealthcheck", "terraform")
}

func TestCloudflareWaitingRoom_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarewaitingroom", "pulumi")
}
func TestCloudflareWaitingRoom_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarewaitingroom", "terraform")
}

func TestCloudflareWaitingRoomEvent_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarewaitingroomevent", "pulumi")
}
func TestCloudflareWaitingRoomEvent_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "cloudflarewaitingroomevent", "terraform")
}

// runAllScenariosForKind discovers and runs all E2E scenarios for a
// Cloudflare kind.
func runAllScenariosForKind(t *testing.T, kindDir, engine string) {
	t.Helper()

	if cp, err := profilepkg.LoadKindProfile(repoRoot, "cloudflare", kindDir); err == nil && cp.Spec != nil {
		switch cp.Spec.Status {
		case kindv1.CatalogKindE2EProfileSpec_deferred,
			kindv1.CatalogKindE2EProfileSpec_skip,
			kindv1.CatalogKindE2EProfileSpec_stub,
			// pending_proof: fully authored, offline-validated, awaiting its
			// first live proof. The proving session flips the profile to green
			// immediately before executing the lanes; until then a sweep must
			// never run it.
			kindv1.CatalogKindE2EProfileSpec_pending_proof,
			// real_cluster has no meaning for a SaaS provider; a profile
			// carrying it is a mistake that must skip loudly, never run.
			kindv1.CatalogKindE2EProfileSpec_real_cluster:
			reason := cp.Spec.DeferredReason
			if reason == "" {
				reason = cp.Spec.Status.String()
			}
			t.Skipf("kind %s E2E profile status is %s: %s", kindDir, cp.Spec.Status, reason)
		}
	}

	moduleDir, err := discovery.ModuleDir(repoRoot, "cloudflare", kindDir, engine)
	if err != nil {
		t.Fatalf("failed to locate %s %s module: %v", kindDir, engine, err)
	}

	if !fileExists(moduleDir) {
		t.Skipf("kind %s %s module not found at %s", kindDir, engine, moduleDir)
	}

	scenarios, err := discovery.DiscoverTestScenarios(repoRoot, "cloudflare", kindDir)
	if err != nil {
		t.Fatalf("failed to discover test scenarios for %s: %v", kindDir, err)
	}

	if len(scenarios) == 0 {
		t.Skipf("no test scenarios found for %s", kindDir)
	}

	t.Logf("Discovered %d scenarios for %s [%s]", len(scenarios), kindDir, engine)

	for _, scenario := range scenarios {
		scenario := scenario
		t.Run(scenario.Name, func(t *testing.T) {
			runSingleScenario(t, kindDir, moduleDir, engine, scenario)
		})
	}
}

func runSingleScenario(t *testing.T, kindDir, moduleDir, engine string, scenario discovery.TestScenario) {
	t.Helper()

	// Scenarios needing owner-arranged external context (the
	// e2e-required-env annotation -- for Cloudflare, typically the real
	// ACTIVE zone injected as ${E2E_ENV:PLANTON_E2E_CLOUDFLARE_ZONE_ID})
	// skip honestly where the environment does not carry the arrangement --
	// unset tokens would otherwise fail expansion loudly, turning a
	// deferral into a false failure.
	if missing, err := runner.ScenarioMissingRequiredEnv(scenario.ManifestPath); err != nil {
		t.Fatalf("reading required-env declaration for scenario %s/%s: %v", kindDir, scenario.Name, err)
	} else if len(missing) > 0 {
		t.Skipf("scenario %s/%s needs owner-arranged environment variables that are unset: %s (per %s)",
			kindDir, scenario.Name, strings.Join(missing, ", "), runner.ScenarioRequiredEnvAnnotation)
	}

	tc := &provider.KindTestContext{
		Kind:         kindDir,
		Provider:     "cloudflare",
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
		tc.StackName = runner.GenerateStackName(kindDir+"-"+scenario.Name, runID)
	}

	ctx := context.Background()
	result := runner.RunKindTest(ctx, tc, testHarness)

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
		t.Fatalf("scenario %s/%s [%s] failed (total: %s)", kindDir, scenario.Name, engine, result.Duration)
	}

	t.Logf("scenario %s/%s [%s] passed (total: %s)", kindDir, scenario.Name, engine, result.Duration)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
