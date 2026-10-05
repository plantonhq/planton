//go:build e2e

// Package azure contains end-to-end tests that provision real Azure resources via
// Planton IaC modules and verify them through the Azure SDK. Credentials come from
// the ambient chain (local `az` login or GitHub Actions OIDC federation -- never a
// stored secret); see the aa_e2e harness package.
//
// Run with: go test -tags=e2e -timeout=30m -v ./e2e/azure/...
package azure

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	azuree2e "github.com/plantonhq/planton/catalog/azure/aa_e2e"
	"github.com/plantonhq/planton/e2e/framework/discovery"
	"github.com/plantonhq/planton/e2e/framework/provider"
	"github.com/plantonhq/planton/e2e/framework/runner"
	profilepkg "github.com/plantonhq/planton/pkg/e2e/profile"
	kindv1 "github.com/plantonhq/planton/qa/catalogkinde2eprofile/v1"
)

var (
	testHarness      *azuree2e.Harness
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

	backendDir, err := os.MkdirTemp("", "planton-e2e-azure-pulumi-*")
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

	providerProfile, err := profilepkg.LoadProviderProfile(repoRoot, "azure")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load Azure provider E2E profile: %v\n", err)
		os.Exit(1)
	}
	assertApplyIdempotency = providerProfile.GetSpec().GetAssertApplyIdempotency()

	testHarness = azuree2e.NewHarness()
	ctx := context.Background()
	if err := testHarness.Setup(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "failed to setup Azure harness: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	if err := testHarness.Teardown(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to teardown Azure harness: %v\n", err)
	}

	os.Exit(code)
}

// --- Azure Resource Group (walking skeleton: the Layer-0 container every other kind references) ---

func TestAzureResourceGroup_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureresourcegroup", "pulumi")
}
func TestAzureResourceGroup_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureresourcegroup", "terraform")
}

// --- Azure Role Assignment (composed: fixture RG + fixture identity -> Reader grant at RG scope) ---

func TestAzureRoleAssignment_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureroleassignment", "pulumi")
}
func TestAzureRoleAssignment_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureroleassignment", "terraform")
}

// --- Azure Role Definition (composed: custom role scoped at the fixture RG) ---

func TestAzureRoleDefinition_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureroledefinition", "pulumi")
}
func TestAzureRoleDefinition_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureroledefinition", "terraform")
}

// --- Azure Federated Identity Credential (composed: GitHub-shaped trust rule on the fixture identity) ---

func TestAzureFederatedIdentityCredential_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurefederatedidentitycredential", "pulumi")
}
func TestAzureFederatedIdentityCredential_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurefederatedidentitycredential", "terraform")
}

// --- Azure User Assigned Identity (the identity workloads act as; tags exercised) ---

func TestAzureUserAssignedIdentity_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureuserassignedidentity", "pulumi")
}
func TestAzureUserAssignedIdentity_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureuserassignedidentity", "terraform")
}

// --- Azure Virtual Network (composed: multi-CIDR network in the fixture RG) ---

func TestAzureVirtualNetwork_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurevirtualnetwork", "pulumi")
}
func TestAzureVirtualNetwork_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurevirtualnetwork", "terraform")
}

// --- Azure Route Table (composed: black-hole + service-tag routes in the fixture RG) ---

func TestAzureRouteTable_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureroutetable", "pulumi")
}
func TestAzureRouteTable_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureroutetable", "terraform")
}

// --- Azure Private DNS Zone (composed: custom internal zone in the fixture RG) ---

func TestAzurePrivateDnsZone_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureprivatednszone", "pulumi")
}
func TestAzurePrivateDnsZone_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureprivatednszone", "terraform")
}

// --- Azure Private DNS Zone Virtual Network Link (composed two-parent chain: fixture zone + fixture network -> link) ---

func TestAzurePrivateDnsZoneVirtualNetworkLink_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureprivatednszonevirtualnetworklink", "pulumi")
}
func TestAzurePrivateDnsZoneVirtualNetworkLink_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureprivatednszonevirtualnetworklink", "terraform")
}

// --- Azure Subnet (the attach-model showcase: fixture network + extra route-table/NSG/NAT fixtures -> subnet with all three seams) ---

func TestAzureSubnet_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuresubnet", "pulumi")
}
func TestAzureSubnet_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuresubnet", "terraform")
}

// --- Azure Network Security Group (singular + plural rule forms in the fixture RG) ---

func TestAzureNetworkSecurityGroup_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurenetworksecuritygroup", "pulumi")
}
func TestAzureNetworkSecurityGroup_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurenetworksecuritygroup", "terraform")
}

// --- Azure Public IP (zone-redundant static address with a scope-hashed DNS label) ---

func TestAzurePublicIp_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurepublicip", "pulumi")
}
func TestAzurePublicIp_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurepublicip", "terraform")
}

// --- Azure Public IP Prefix (smallest reservable range, /31) ---

func TestAzurePublicIpPrefix_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurepublicipprefix", "pulumi")
}
func TestAzurePublicIpPrefix_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurepublicipprefix", "terraform")
}

// --- Azure Bastion Host (composed: fixture AzureBastionSubnet + the lane's own fixture public IP -> Basic host; ~10-minute create class, ~$0.19/h while up) ---

func TestAzureBastionHost_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurebastionhost", "pulumi")
}
func TestAzureBastionHost_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurebastionhost", "terraform")
}

// --- Azure Network Watcher Flow Log (composed: fixture VNet's flows -> fixture storage account + Traffic Analytics on the fixture workspace; attaches to the AUTO-CREATED regional Network Watcher) ---

func TestAzureNetworkWatcherFlowLog_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurenetworkwatcherflowlog", "pulumi")
}
func TestAzureNetworkWatcherFlowLog_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurenetworkwatcherflowlog", "terraform")
}

// --- Azure DNS Private Resolver (composed: fixture VNet + two delegated fixture subnets -> one inbound + one outbound endpoint; ONE resolver per network, so this family's lanes never run in parallel with each other) ---

func TestAzurePrivateDnsResolver_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureprivatednsresolver", "pulumi")
}
func TestAzurePrivateDnsResolver_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureprivatednsresolver", "terraform")
}

// --- Azure DNS Forwarding Ruleset (composed: one rule -> RFC-5737 target, bound to the fixture resolver's outbound endpoint; chains the fixture resolver in the lane) ---

func TestAzurePrivateDnsResolverForwardingRuleset_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureprivatednsresolverforwardingruleset", "pulumi")
}
func TestAzurePrivateDnsResolverForwardingRuleset_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureprivatednsresolverforwardingruleset", "terraform")
}

// --- Azure DNS Resolver Virtual Network Link (composed: links the fixture ruleset to the fixture network -- the hub-link pattern; chains fixture resolver -> fixture ruleset in the lane) ---

func TestAzurePrivateDnsResolverVirtualNetworkLink_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureprivatednsresolvervirtualnetworklink", "pulumi")
}
func TestAzurePrivateDnsResolverVirtualNetworkLink_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureprivatednsresolvervirtualnetworklink", "terraform")
}

// --- Azure Private DNS Record (one A record in the fixture private zone -- the seven-payload union's live lane; the other six types are pure property writes on the same create path) ---

func TestAzurePrivateDnsRecord_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureprivatednsrecord", "pulumi")
}
func TestAzurePrivateDnsRecord_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureprivatednsrecord", "terraform")
}

// --- Azure Traffic Manager Profile (global DNS traffic director; only a resource group beneath it -- fast, free at rest) ---

func TestAzureTrafficManagerProfile_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuretrafficmanagerprofile", "pulumi")
}
func TestAzureTrafficManagerProfile_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuretrafficmanagerprofile", "terraform")
}

// --- Azure Traffic Manager Endpoint (composed: one external endpoint on the fixture profile -- the reference edge is the lane's proof point; endpoints are free at rest) ---

func TestAzureTrafficManagerEndpoint_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuretrafficmanagerendpoint", "pulumi")
}
func TestAzureTrafficManagerEndpoint_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuretrafficmanagerendpoint", "terraform")
}

// --- Azure Application Security Group (empty micro-segmentation anchor in the fixture RG) ---

func TestAzureApplicationSecurityGroup_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureapplicationsecuritygroup", "pulumi")
}
func TestAzureApplicationSecurityGroup_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureapplicationsecuritygroup", "terraform")
}

// --- Azure Private Endpoint (composed: fixture subnet + scenario-local storage target + blob privatelink zone) ---

func TestAzurePrivateEndpoint_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureprivateendpoint", "pulumi")
}
func TestAzurePrivateEndpoint_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureprivateendpoint", "terraform")
}

// --- Azure Disk Encryption Set (profile-deferred: purge-protected vault cannot teardown to zero orphans) ---

func TestAzureDiskEncryptionSet_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurediskencryptionset", "pulumi")
}
func TestAzureDiskEncryptionSet_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurediskencryptionset", "terraform")
}

// --- Azure MSSQL Failover Group (composed: shared primary + scenario-local partner + database) ---

func TestAzureMssqlFailoverGroup_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuremssqlfailovergroup", "pulumi")
}
func TestAzureMssqlFailoverGroup_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuremssqlfailovergroup", "terraform")
}

// --- Azure Monitor Activity Log Alert (composed: fixture action group -> administrative alert) ---

func TestAzureMonitorActivityLogAlert_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuremonitoractivitylogalert", "pulumi")
}
func TestAzureMonitorActivityLogAlert_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuremonitoractivitylogalert", "terraform")
}

// --- Azure Application Insights Standard Web Test (composed: fixture App Insights -> web test) ---

func TestAzureApplicationInsightsStandardWebTest_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureapplicationinsightsstandardwebtest", "pulumi")
}
func TestAzureApplicationInsightsStandardWebTest_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureapplicationinsightsstandardwebtest", "terraform")
}

// --- Azure NAT Gateway (composed: extra public-IP + prefix fixtures -> gateway with both association forms) ---

func TestAzureNatGateway_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurenatgateway", "pulumi")
}
func TestAzureNatGateway_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurenatgateway", "terraform")
}

// --- Azure Virtual Network Peering (composed: fixture network + path-declared second network -> one peering direction) ---

func TestAzureVirtualNetworkPeering_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurevirtualnetworkpeering", "pulumi")
}
func TestAzureVirtualNetworkPeering_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurevirtualnetworkpeering", "terraform")
}

// --- Azure Virtual Network Gateway (THE slow lane: fixture GatewaySubnet + dedicated PIP -> VpnGw1 gateway + NAT rule; 25-45 min create) ---

func TestAzureVirtualNetworkGateway_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurevirtualnetworkgateway", "pulumi")
}
func TestAzureVirtualNetworkGateway_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurevirtualnetworkgateway", "terraform")
}

// --- Azure Local Network Gateway (pure ARM metadata: the on-premises site description; seconds) ---

func TestAzureLocalNetworkGateway_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurelocalnetworkgateway", "pulumi")
}
func TestAzureLocalNetworkGateway_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurelocalnetworkgateway", "terraform")
}

// --- Azure Virtual Network Gateway Connection (composed: fixture gateway + fixture site -> IPsec tunnel; the fixture gateway's slow cycle runs INSIDE this lane) ---

func TestAzureVirtualNetworkGatewayConnection_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurevirtualnetworkgatewayconnection", "pulumi")
}
func TestAzureVirtualNetworkGatewayConnection_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurevirtualnetworkgatewayconnection", "terraform")
}

// --- Azure Private Link Service (composed: policies-disabled subnet fixture + annotation-declared fixture load balancer -> the provider side of Private Link; ~1-2 min) ---

func TestAzurePrivateLinkService_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureprivatelinkservice", "pulumi")
}
func TestAzurePrivateLinkService_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureprivatelinkservice", "terraform")
}

// --- Azure ExpressRoute Circuit (fixture RG -> a NotProvisioned 50 Mbps circuit + one authorization; minutes, bills from creation) ---

func TestAzureExpressRouteCircuit_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureexpressroutecircuit", "pulumi")
}
func TestAzureExpressRouteCircuit_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureexpressroutecircuit", "terraform")
}

// --- Azure ExpressRoute Circuit Peering (composed: fixture circuit -> private peering; ARM gates peering config on provider provisioning -- see the scenario header) ---

func TestAzureExpressRouteCircuitPeering_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureexpressroutecircuitpeering", "pulumi")
}
func TestAzureExpressRouteCircuitPeering_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureexpressroutecircuitpeering", "terraform")
}

// --- Azure ExpressRoute Port (fixture RG -> a 10 Gbps Dot1Q port + one authorization; bills its full monthly rate from creation -- pro-rated cents-to-dollars for a lane-length life -- and the create may reject on subscriptions not enrolled for ExpressRoute Direct: that outcome IS the lane's proof point, never a module defect) ---

func TestAzureExpressRoutePort_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureexpressrouteport", "pulumi")
}
func TestAzureExpressRoutePort_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureexpressrouteport", "terraform")
}

// --- Azure Virtual WAN (fixture RG -> a free Standard WAN policy object; minutes) ---

func TestAzureVirtualWan_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurevirtualwan", "pulumi")
}
func TestAzureVirtualWan_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurevirtualwan", "terraform")
}

// --- Azure Virtual Hub (composed: fixture WAN -> a Standard hub in eastus2 + one route table + one route map; the hub router provisions in 15-30 minutes and bills from creation; the scenario region deliberately differs from the fixture hub's eastus -- ARM allows one hub per region per WAN) ---

func TestAzureVirtualHub_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurevirtualhub", "pulumi")
}
func TestAzureVirtualHub_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurevirtualhub", "terraform")
}

// --- Azure Virtual Hub Connection (composed: fixture hub + fixture VNet -> the attachment with a propagation-only routing block; the chain pays a full fixture-hub cycle inside the lane) ---

func TestAzureVirtualHubConnection_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurevirtualhubconnection", "pulumi")
}
func TestAzureVirtualHubConnection_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurevirtualhubconnection", "terraform")
}

// --- Azure ExpressRoute Gateway (composed: fixture hub -> a one-scale-unit gateway, NO connections -- ARM accepts a connection only against a provider-PROVISIONED circuit peering, the same recorded boundary as the peering's own lane; ~30-minute create billing ~$0.42/hr) ---

func TestAzureExpressRouteGateway_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureexpressroutegateway", "pulumi")
}
func TestAzureExpressRouteGateway_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureexpressroutegateway", "terraform")
}

// --- Azure VPN Site (fixture WAN -> a free two-link branch description; seconds -- the flagship is the name-keyed link_ids output connections pin to) ---

func TestAzureVpnSite_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurevpnsite", "pulumi")
}
func TestAzureVpnSite_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurevpnsite", "terraform")
}

// --- Azure VPN Gateway (composed: fixture hub -> a one-scale-unit gateway + one NAT rule; 30-45 minute create billing ~$0.36/hr per scale unit, 10-20 minute delete -- ARM allows one VPN gateway per hub) ---

func TestAzureVpnGateway_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurevpngateway", "pulumi")
}
func TestAzureVpnGateway_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurevpngateway", "terraform")
}

// --- Azure VPN Gateway Connection (composed: fixture gateway + fixture site -> one tunnel pinned to the site's primary link; the chain pays the whole VPN family -- hub AND gateway cycles -- and the tunnel provisions but never reaches Connected: no real branch device exists behind the fixture site's TEST-NET endpoint) ---

func TestAzureVpnGatewayConnection_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurevpngatewayconnection", "pulumi")
}
func TestAzureVpnGatewayConnection_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurevpngatewayconnection", "terraform")
}

// --- Azure VPN Server Configuration (fixture RG -> a free certificate-auth policy + one composed policy group; seconds -- the flagship is the name-keyed policy_group_ids output) ---

func TestAzureVpnServerConfiguration_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurevpnserverconfiguration", "pulumi")
}
func TestAzureVpnServerConfiguration_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurevpnserverconfiguration", "terraform")
}

// --- Azure Point-to-Site VPN Gateway (composed: fixture hub + fixture server configuration -> a one-scale-unit gateway with one client pool; 30-45 minute create billing from creation -- no client ever connects: the fixture root's private key does not exist) ---

func TestAzurePointToSiteVpnGateway_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurepointtositevpngateway", "pulumi")
}
func TestAzurePointToSiteVpnGateway_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurepointtositevpngateway", "terraform")
}

// --- Azure Cognitive Account (fixture RG -> an OpenAI-kind S0 account with composed responsible-AI children; 1-3 minutes, no idle cost -- deletion soft-deletes and the module purges the ghost) ---

func TestAzureCognitiveAccount_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurecognitiveaccount", "pulumi")
}
func TestAzureCognitiveAccount_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurecognitiveaccount", "terraform")
}

// --- Azure Cognitive Deployment (composed: fixture account -> gpt-4o-mini on GlobalStandard; seconds-to-minutes, per-token billing -- the quota/model-availability boundary is the lane's proof point) ---

func TestAzureCognitiveDeployment_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurecognitivedeployment", "pulumi")
}
func TestAzureCognitiveDeployment_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurecognitivedeployment", "terraform")
}

// --- Azure Cognitive Account Project (composed: fixture AIServices account -> a free system-identity AI Foundry project; seconds -- no agents or data-plane assets are ever created) ---

func TestAzureCognitiveAccountProject_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurecognitiveaccountproject", "pulumi")
}
func TestAzureCognitiveAccountProject_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurecognitiveaccountproject", "terraform")
}

// --- Azure Machine Learning Workspace (composed: fixture storage account + key vault + application insights -> a Basic system-identity workspace on the DEFAULT network; create ~1 min, destroy ~2 min; deletes purge the soft-delete ghost via the modules' machine_learning features flag; ghosts have no list API, portal "Recently deleted" only. The managed-VNet + outbound-rule flagship stayed offline-proven after two live failures.) ---

func TestAzureMachineLearningWorkspace_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuremachinelearningworkspace", "pulumi")
}
func TestAzureMachineLearningWorkspace_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuremachinelearningworkspace", "terraform")
}

// --- Azure Machine Learning Datastore (composed: fixture workspace -> a blob-variant datastore on the scenario-declared fixture container under workspace-identity auth; seconds, free -- the provider skips service-side validation on create, so the lane proves registration, not data access) ---

func TestAzureMachineLearningDatastore_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuremachinelearningdatastore", "pulumi")
}
func TestAzureMachineLearningDatastore_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuremachinelearningdatastore", "terraform")
}

// --- Azure Machine Learning Compute Cluster (composed: fixture workspace -> a scale-to-zero dedicated cluster with a system identity; the compute OBJECT is the proof -- no node ever provisions, zero VM cost) ---

func TestAzureMachineLearningComputeCluster_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuremachinelearningcomputecluster", "pulumi")
}
func TestAzureMachineLearningComputeCluster_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuremachinelearningcomputecluster", "terraform")
}

// --- Azure Machine Learning Compute Instance (composed: fixture workspace -> one small workstation VM with a system identity; the cloud-side name carries the run-id token -- instance names are reserved region-wide per subscription) ---

func TestAzureMachineLearningComputeInstance_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuremachinelearningcomputeinstance", "pulumi")
}
func TestAzureMachineLearningComputeInstance_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuremachinelearningcomputeinstance", "terraform")
}

// --- Azure AI Foundry (composed: fixture RG + key vault + HNS-off storage account -> a system-identity hub; minutes -- hubs are ML workspaces at ARM, soft-delete ghosts hold the name, the orphan sweep checks `az ml workspace list --archived`) ---

func TestAzureAiFoundry_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureaifoundry", "pulumi")
}
func TestAzureAiFoundry_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureaifoundry", "terraform")
}

// --- Azure AI Foundry Project (composed: fixture hub chain -> a system-identity project INSIDE the hub; the proof is the cross-kind wiring -- ARM returns kind "Project", and the project lands in the hub's resource group) ---

func TestAzureAiFoundryProject_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureaifoundryproject", "pulumi")
}
func TestAzureAiFoundryProject_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureaifoundryproject", "terraform")
}

// --- Azure AI Search Service (composed: fixture RG + storage account -> a standard-SKU service with one shared private link to the fixture storage's blob subresource; the link sits PENDING by design -- nothing approves the target side, the verifier asserts ARM state; the service name is GLOBALLY unique endpoint DNS) ---

func TestAzureSearchService_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuresearchservice", "pulumi")
}
func TestAzureSearchService_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuresearchservice", "terraform")
}

// --- Azure ML Online Endpoint (composed: the fixture workspace chain -> a key-auth endpoint with a system identity; minutes, no instances -- the endpoint object is free at rest. FIRST raw-API kind: Terraform deploys through azapi, Pulumi through azure-native; scenario endpoint names carry the run-id token, endpoint names are reserved region-wide) ---

func TestAzureMachineLearningOnlineEndpoint_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuremachinelearningonlineendpoint", "pulumi")
}
func TestAzureMachineLearningOnlineEndpoint_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuremachinelearningonlineendpoint", "terraform")
}

// --- Azure ML Online Deployment (composed: workspace chain -> fixture endpoint -> one Standard_F2s_v2 managed instance serving a SETUP-seeded MLflow model; 10-20 min provisioning, bills until destroy -- no scale-to-zero. A model is required live: the service rejects a model-less managed create with a bare 400 despite the ARM schema marking model/environment/code optional; deployment names are endpoint-scoped, so scenarios carry the run-id token because both engines attach to the SAME fixture endpoint) ---

func TestAzureMachineLearningOnlineDeployment_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuremachinelearningonlinedeployment", "pulumi")
}
func TestAzureMachineLearningOnlineDeployment_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuremachinelearningonlinedeployment", "terraform")
}

// --- Azure ML Batch Endpoint (composed: the fixture workspace chain -> an Entra-auth endpoint with a system identity; a pure routing object -- no compute, free at rest. Raw-API kind riding the online pair's machinery; scenario names carry the run-id token -- batch names are expected to reserve region-wide like the online sibling's, unproven) ---

func TestAzureMachineLearningBatchEndpoint_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuremachinelearningbatchendpoint", "pulumi")
}
func TestAzureMachineLearningBatchEndpoint_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuremachinelearningbatchendpoint", "terraform")
}

// --- Azure ML Batch Deployment (composed: workspace chain -> fixture batch endpoint + fixture scale-to-zero compute cluster -> a Model-type recipe serving a SETUP-seeded MLflow model; nothing provisions or bills -- compute materializes per job and no lane submits one. A model AND a compute are required live: the service rejects a bare recipe with 400 UserError naming both properties (now a spec CEL); deployment names are endpoint-scoped, so scenarios carry the run-id token because both engines attach to the SAME fixture endpoint) ---

func TestAzureMachineLearningBatchDeployment_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuremachinelearningbatchdeployment", "pulumi")
}
func TestAzureMachineLearningBatchDeployment_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuremachinelearningbatchdeployment", "terraform")
}

// --- Azure Recovery Services Vault (composed: fixture RG -> a locally-redundant vault with the monitoring block; FREE at rest, creates in minutes. Immutability/CMK are one-way doors deliberately left to offline proof) ---

func TestAzureRecoveryServicesVault_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurerecoveryservicesvault", "pulumi")
}
func TestAzureRecoveryServicesVault_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurerecoveryservicesvault", "terraform")
}

// --- Azure Backup Policy VM (composed: fixture RG -> fixture vault -> a V2 daily policy with the full retention ladder; a FREE configuration object) ---

func TestAzureBackupPolicyVm_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurebackuppolicyvm", "pulumi")
}
func TestAzureBackupPolicyVm_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurebackuppolicyvm", "terraform")
}

// --- Azure Backup Protected VM (composed: the trio's deepest chain -- policy -> fixture vault, plus the fixture VM's NIC/subnet/VNet/disk chain; creation only REGISTERS protection, so no backup storage accrues and the cost is the fixture VM's hour. The smoke also proves the LUN-exclusion seam against the VM's data disk) ---

func TestAzureBackupProtectedVm_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurebackupprotectedvm", "pulumi")
}
func TestAzureBackupProtectedVm_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurebackupprotectedvm", "terraform")
}

// --- Azure Backup Policy File Share (composed: fixture RG -> fixture vault -> two FREE configuration objects; the minimal daily-snapshot smoke doubles as the protected-share lane's install manifest, and the hourly-vault-standard scenario proves the window dials and the vaulted tier in isolation) ---

func TestAzureBackupPolicyFileShare_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurebackuppolicyfileshare", "pulumi")
}
func TestAzureBackupPolicyFileShare_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurebackuppolicyfileshare", "terraform")
}

// --- Azure Backup Container Storage Account (composed: fixture RG -> fixture vault + fixture storage account -> a FREE registration binding; the smoke and the protected-share lane's fixture registration target the SAME ARM identity -- sequential lanes only) ---

func TestAzureBackupContainerStorageAccount_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurebackupcontainerstorageaccount", "pulumi")
}
func TestAzureBackupContainerStorageAccount_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurebackupcontainerstorageaccount", "terraform")
}

// --- Azure Backup Protected File Share (composed: the group's deepest chain -- fixture RG -> vault + storage account -> fixture share + the container REGISTRATION + the policy smoke; creation only REGISTERS protection. Protect/unprotect sit in the provider's 80-minute class, and destroy leaves a soft-delete ghost that can delay the fixture registration's unregister -- the queue watch-list leads with it) ---

func TestAzureBackupProtectedFileShare_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurebackupprotectedfileshare", "pulumi")
}
func TestAzureBackupProtectedFileShare_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurebackupprotectedfileshare", "terraform")
}

// --- Azure Data Protection Backup Vault (composed: fixture RG -> a FREE locally-redundant vault with a system identity; the one-way doors -- Locked immutability, AlwaysOn soft delete, CMK -- stay offline-proven by design, and destroy runs past Azure's early "deleted" answer while the provider polls the name free) ---

func TestAzureDataProtectionBackupVault_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuredataprotectionbackupvault", "pulumi")
}
func TestAzureDataProtectionBackupVault_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuredataprotectionbackupvault", "terraform")
}

// --- Azure Data Protection Backup Policy (composed: fixture RG -> fixture Data Protection vault -> THREE free configuration objects covering all three retention dialects; the minimal disk smoke doubles as future backup-instance lanes' install manifest, the kubernetes scenario proves the nested life-cycle dialect plus the vault name+RG addressing seam, and the blob scenario proves the dual-store dialect) ---

func TestAzureDataProtectionBackupPolicy_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuredataprotectionbackuppolicy", "pulumi")
}
func TestAzureDataProtectionBackupPolicy_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuredataprotectionbackuppolicy", "terraform")
}

// --- Azure Data Protection Backup Instance (composed: fixture RG -> fixture Data Protection vault + disk policy + managed disk + the vault identity's TWO RBAC grants -> the disk-variant binding; the vault identity's grants are validated by Azure at instance create, and grant propagation lag is the lane's expected retry class. REQUIRES the fixture vault to carry a system-assigned identity -- the grants' principal -- and soft delete Off, so a destroyed instance cannot ghost and wedge the vault's teardown) ---

func TestAzureDataProtectionBackupInstance_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuredataprotectionbackupinstance", "pulumi")
}
func TestAzureDataProtectionBackupInstance_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuredataprotectionbackupinstance", "terraform")
}

// --- Azure Data Protection Resource Guard (composed: fixture RG -> a FREE governance object that creates in seconds; no vault references it in the lane, so create and delete are unconstrained) ---

func TestAzureDataProtectionResourceGuard_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuredataprotectionresourceguard", "pulumi")
}
func TestAzureDataProtectionResourceGuard_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuredataprotectionresourceguard", "terraform")
}

// --- Azure AKS Cluster (composed: fixture RG -> managed-networking cluster with a single-node default pool) ---

func TestAzureAksCluster_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureakscluster", "pulumi")
}
func TestAzureAksCluster_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureakscluster", "terraform")
}

// --- Azure AKS Node Pool (composed: fixture RG -> cluster prerequisite -> zero-node user pool) ---

func TestAzureAksNodePool_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureaksnodepool", "pulumi")
}
func TestAzureAksNodePool_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureaksnodepool", "terraform")
}

// --- Azure Managed Disk (fixture RG -> empty zonal Standard SSD data disk) ---

func TestAzureManagedDisk_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuremanageddisk", "pulumi")
}
func TestAzureManagedDisk_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuremanageddisk", "terraform")
}

// --- Azure Network Interface (composed: fixture RG -> VNet -> subnet -> NIC with an extra-fixture NSG attached) ---

func TestAzureNetworkInterface_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurenetworkinterface", "pulumi")
}
func TestAzureNetworkInterface_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurenetworkinterface", "terraform")
}

// --- Azure Virtual Machine (composed: fixture RG -> VNet -> subnet -> NIC -> Linux VM with an extra-fixture disk at LUN 0) ---

func TestAzureVirtualMachine_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurevirtualmachine", "pulumi")
}
func TestAzureVirtualMachine_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurevirtualmachine", "terraform")
}

// --- Azure Container Registry (fixture RG -> Standard registry with the admin account on) ---

func TestAzureContainerRegistry_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurecontainerregistry", "pulumi")
}
func TestAzureContainerRegistry_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurecontainerregistry", "terraform")
}

// --- Azure Load Balancer (composed: fixture RG + extra-fixture public IP -> public LB with pool, probe, rule, NAT rule, and outbound rule) ---

func TestAzureLoadBalancer_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureloadbalancer", "pulumi")
}
func TestAzureLoadBalancer_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureloadbalancer", "terraform")
}

// --- Azure Virtual Machine Scale Set (composed: fixture RG -> VNet -> subnet -> one-instance fleet, BOTH orchestration modes) ---

func TestAzureVirtualMachineScaleSet_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurevirtualmachinescaleset", "pulumi")
}
func TestAzureVirtualMachineScaleSet_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurevirtualmachinescaleset", "terraform")
}

// --- Azure Key Vault (fixture RG -> Standard RBAC vault with network rules; purge-on-destroy frees the global name) ---

func TestAzureKeyVault_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurekeyvault", "pulumi")
}
func TestAzureKeyVault_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurekeyvault", "terraform")
}

// --- Azure Key Vault Key (composed: fixture RG -> vault -> RSA CMK with a rotation policy; data-plane create) ---

func TestAzureKeyVaultKey_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurekeyvaultkey", "pulumi")
}
func TestAzureKeyVaultKey_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurekeyvaultkey", "terraform")
}

// --- Azure Key Vault Secret (composed: fixture RG -> vault -> one versioned secret with content type and expiry; data-plane create -- the ambient principal's subscription-scoped Key Vault Administrator grant covers it) ---

func TestAzureKeyVaultSecret_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurekeyvaultsecret", "pulumi")
}
func TestAzureKeyVaultSecret_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurekeyvaultsecret", "terraform")
}

// --- Azure Key Vault Certificate (composed: fixture RG -> vault -> self-signed auto-renewing certificate; data-plane create + verify) ---

func TestAzureKeyVaultCertificate_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurekeyvaultcertificate", "pulumi")
}
func TestAzureKeyVaultCertificate_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurekeyvaultcertificate", "terraform")
}

// --- Azure Application Gateway (composed: fixture RG -> VNet -> dedicated subnet + extra-fixture public IP -> Standard_v2 gateway; the waf-attach scenario adds the fixture WAF policy) ---

func TestAzureApplicationGateway_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureapplicationgateway", "pulumi")
}
func TestAzureApplicationGateway_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureapplicationgateway", "terraform")
}

// --- Azure Web Application Firewall Policy (fixture RG -> OWASP 3.2 policy with a rate-limit rule, override, and log scrubbing) ---

func TestAzureWebApplicationFirewallPolicy_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurewebapplicationfirewallpolicy", "pulumi")
}
func TestAzureWebApplicationFirewallPolicy_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurewebapplicationfirewallpolicy", "terraform")
}

// --- Azure PostgreSQL Flexible Server (fixture RG -> burstable public server with database, firewall rule, and server parameter) ---

func TestAzurePostgresqlFlexibleServer_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurepostgresqlflexibleserver", "pulumi")
}
func TestAzurePostgresqlFlexibleServer_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurepostgresqlflexibleserver", "terraform")
}

// --- Azure MySQL Flexible Server (fixture RG -> burstable public server with database, firewall rule, and server parameter) ---

func TestAzureMysqlFlexibleServer_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuremysqlflexibleserver", "pulumi")
}
func TestAzureMysqlFlexibleServer_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuremysqlflexibleserver", "terraform")
}

// --- Azure Mongo Cluster (fixture RG -> Free-tier Mongo vCore sandbox with Entra + native auth and one composed firewall rule; the cloud-side name carries the run-id token -- cluster names are global hostname labels) ---

func TestAzureMongoCluster_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuremongocluster", "pulumi")
}
func TestAzureMongoCluster_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuremongocluster", "terraform")
}

// --- Azure Mongo Cluster User (composed: cluster smoke chain + fixture identity -> the Entra access grant, the three-kind reference chain) ---

func TestAzureMongoClusterUser_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuremongoclusteruser", "pulumi")
}
func TestAzureMongoClusterUser_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuremongoclusteruser", "terraform")
}

// --- Azure SQL logical server (fixture RG -> SQL-auth server with firewall rule and Defender alert policy) ---

func TestAzureMssqlServer_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuremssqlserver", "pulumi")
}
func TestAzureMssqlServer_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuremssqlserver", "terraform")
}

// --- Azure SQL Database (fixture RG -> fixture server -> Basic database; pool-attach joins the fixture elastic pool) ---

func TestAzureMssqlDatabase_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuremssqldatabase", "pulumi")
}
func TestAzureMssqlDatabase_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuremssqldatabase", "terraform")
}

// --- Azure SQL elastic pool (fixture RG -> fixture server -> BasicPool) ---

func TestAzureMssqlElasticPool_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuremssqlelasticpool", "pulumi")
}
func TestAzureMssqlElasticPool_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuremssqlelasticpool", "terraform")
}

// --- Azure storage account (fixture RG -> StorageV2 with firewall + lifecycle policy) ---

func TestAzureStorageAccount_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurestorageaccount", "pulumi")
}
func TestAzureStorageAccount_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurestorageaccount", "terraform")
}

// --- Azure storage container (fixture RG -> scenario-local account -> private container;
// plus the composed encryption-scope chain) ---

func TestAzureStorageContainer_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurestoragecontainer", "pulumi")
}
func TestAzureStorageContainer_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurestoragecontainer", "terraform")
}

// --- Azure storage share (fixture RG -> scenario-local account -> SMB share) ---

func TestAzureStorageShare_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurestorageshare", "pulumi")
}
func TestAzureStorageShare_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurestorageshare", "terraform")
}

// --- Azure storage queue (fixture RG -> scenario-local account -> work queue) ---

func TestAzureStorageQueue_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurestoragequeue", "pulumi")
}
func TestAzureStorageQueue_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurestoragequeue", "terraform")
}

// --- Azure storage table (fixture RG -> scenario-local account -> entities table with ACL) ---

func TestAzureStorageTable_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurestoragetable", "pulumi")
}
func TestAzureStorageTable_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurestoragetable", "terraform")
}

// --- Azure storage encryption scope (fixture RG -> scenario-local account ->
// platform-managed scope; destroy verified state-aware: ARM soft-disables scopes) ---

func TestAzureStorageEncryptionScope_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurestorageencryptionscope", "pulumi")
}
func TestAzureStorageEncryptionScope_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurestorageencryptionscope", "terraform")
}

// --- Azure Data Lake Gen2 filesystem (fixture RG -> scenario-local HNS
// account -> filesystem with a root POSIX ACL; verified via the ARM
// blob-container proxy) ---

func TestAzureStorageDataLakeGen2Filesystem_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurestoragedatalakegen2filesystem", "pulumi")
}
func TestAzureStorageDataLakeGen2Filesystem_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurestoragedatalakegen2filesystem", "terraform")
}

// --- Azure storage local user (fixture RG -> scenario-local SFTP account +
// container -> user with both auth methods and a container-scoped grant) ---

func TestAzureStorageLocalUser_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurestoragelocaluser", "pulumi")
}
func TestAzureStorageLocalUser_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurestoragelocaluser", "terraform")
}

// --- Azure storage object replication (fixture RG -> two scenario-local
// versioned accounts + a container on each -> the two-sided policy) ---

func TestAzureStorageObjectReplication_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurestorageobjectreplication", "pulumi")
}
func TestAzureStorageObjectReplication_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurestorageobjectreplication", "terraform")
}

// --- Azure Redis cache (fixture RG -> Standard C0 cache with Entra auth,
// firewall rule, patch window; Redis provisioning runs 15-40 min per cache) ---

func TestAzureRedisCache_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurerediscache", "pulumi")
}
func TestAzureRedisCache_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurerediscache", "terraform")
}

// --- Azure Redis cache access policy (fixture RG -> scenario-local Entra
// cache -> prefix-scoped custom policy) ---

func TestAzureRedisCacheAccessPolicy_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurerediscacheaccesspolicy", "pulumi")
}
func TestAzureRedisCacheAccessPolicy_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurerediscacheaccesspolicy", "terraform")
}

// --- Azure Redis cache access policy assignment (fixture RG + identity ->
// scenario-local Entra cache -> custom policy -> the grant; proves both FK
// seams: policy-name and identity-principal references) ---

func TestAzureRedisCacheAccessPolicyAssignment_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurerediscacheaccesspolicyassignment", "pulumi")
}
func TestAzureRedisCacheAccessPolicyAssignment_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurerediscacheaccesspolicyassignment", "terraform")
}

// --- Azure Redis linked server (fixture RG -> two scenario-local Premium P1
// caches in different regions -> the geo-replication link; the most expensive
// scenario in the suite) ---

func TestAzureRedisLinkedServer_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureredislinkedserver", "pulumi")
}
func TestAzureRedisLinkedServer_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureredislinkedserver", "terraform")
}

// --- Azure Managed Redis (fixture RG -> Balanced_B0 instance with access
// keys enabled; Managed Redis clusters provision and delete in tens of
// minutes) ---

func TestAzureManagedRedis_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuremanagedredis", "pulumi")
}
func TestAzureManagedRedis_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuremanagedredis", "terraform")
}

// --- Azure Managed Redis geo-replication (fixture RG -> two scenario-local
// Balanced_B3 clusters in different regions declaring the same group name ->
// the group link; the most expensive fixture pair in the suite) ---

func TestAzureManagedRedisGeoReplication_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuremanagedredisgeoreplication", "pulumi")
}
func TestAzureManagedRedisGeoReplication_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuremanagedredisgeoreplication", "terraform")
}

// --- Azure Managed Redis access policy assignment (fixture RG + identity ->
// scenario-local KEYLESS cluster -> the data-plane grant; proves the
// cluster-reference and identity-principal FK seams) ---

func TestAzureManagedRedisAccessPolicyAssignment_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuremanagedredisaccesspolicyassignment", "pulumi")
}
func TestAzureManagedRedisAccessPolicyAssignment_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuremanagedredisaccesspolicyassignment", "terraform")
}

// --- Azure Cosmos DB account (fixture RG -> single-region SQL API account;
// Cosmos accounts provision in ~5-10 min each) ---

func TestAzureCosmosdbAccount_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurecosmosdbaccount", "pulumi")
}
func TestAzureCosmosdbAccount_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurecosmosdbaccount", "terraform")
}

// --- Azure Cosmos DB SQL database (fixture RG -> scenario-local account ->
// database with shared throughput) ---

func TestAzureCosmosdbSqlDatabase_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurecosmosdbsqldatabase", "pulumi")
}
func TestAzureCosmosdbSqlDatabase_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurecosmosdbsqldatabase", "terraform")
}

// --- Azure Cosmos DB SQL container (composed chain: scenario-local account
// -> database -> container with partition key) ---

func TestAzureCosmosdbSqlContainer_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurecosmosdbsqlcontainer", "pulumi")
}
func TestAzureCosmosdbSqlContainer_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurecosmosdbsqlcontainer", "terraform")
}

// --- Azure Cosmos DB Mongo database (fixture RG -> scenario-local Mongo
// account -> database with shared throughput) ---

func TestAzureCosmosdbMongoDatabase_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurecosmosdbmongodatabase", "pulumi")
}
func TestAzureCosmosdbMongoDatabase_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurecosmosdbmongodatabase", "terraform")
}

// --- Azure Cosmos DB Mongo collection (composed chain: scenario-local Mongo
// account -> database -> collection with shard key) ---

func TestAzureCosmosdbMongoCollection_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurecosmosdbmongocollection", "pulumi")
}
func TestAzureCosmosdbMongoCollection_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurecosmosdbmongocollection", "terraform")
}

// --- Azure Cosmos DB SQL role definition (fixture RG -> scenario-local
// SQL account -> custom data-plane role) ---

func TestAzureCosmosdbSqlRoleDefinition_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurecosmosdbsqlroledefinition", "pulumi")
}
func TestAzureCosmosdbSqlRoleDefinition_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurecosmosdbsqlroledefinition", "terraform")
}

// --- Azure Cosmos DB SQL role assignment (composed chain: scenario-local
// SQL account -> custom role definition -> grant to the fixture identity) ---

func TestAzureCosmosdbSqlRoleAssignment_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurecosmosdbsqlroleassignment", "pulumi")
}
func TestAzureCosmosdbSqlRoleAssignment_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurecosmosdbsqlroleassignment", "terraform")
}

// --- Azure Front Door profile (fixture RG -> Standard profile with log
// scrubbing and tags; Front Door is global -- no region on the resource) ---

func TestAzureFrontDoorProfile_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurefrontdoorprofile", "pulumi")
}
func TestAzureFrontDoorProfile_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurefrontdoorprofile", "terraform")
}

// --- Azure Front Door endpoint (fixture RG -> fixture profile -> endpoint;
// verifies the generated *.azurefd.net hostname surfaces as an output) ---

func TestAzureFrontDoorEndpoint_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurefrontdoorendpoint", "pulumi")
}
func TestAzureFrontDoorEndpoint_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurefrontdoorendpoint", "terraform")
}

// --- Azure Front Door origin group (fixture RG -> fixture profile -> group
// with health probe and load-balancing dials) ---

func TestAzureFrontDoorOriginGroup_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurefrontdoororigingroup", "pulumi")
}
func TestAzureFrontDoorOriginGroup_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurefrontdoororigingroup", "terraform")
}

// --- Azure Front Door origin (fixture RG -> fixture profile -> fixture
// group -> origin pointing at a public backend hostname) ---

func TestAzureFrontDoorOrigin_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurefrontdoororigin", "pulumi")
}
func TestAzureFrontDoorOrigin_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurefrontdoororigin", "terraform")
}

// --- Azure Front Door route (the composed chain: fixture RG -> profile ->
// endpoint + origin group -> origin -> route with caching; proves the whole
// traffic-serving graph and the origin_ids ordering seam) ---

func TestAzureFrontDoorRoute_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurefrontdoorroute", "pulumi")
}
func TestAzureFrontDoorRoute_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurefrontdoorroute", "terraform")
}

// --- Azure Front Door rule set (fixture RG -> fixture profile -> a
// three-rule delivery policy: redirect, security headers, caching
// override; the route's rule-set-attach scenario proves the attach seam) ---

func TestAzureFrontDoorRuleSet_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurefrontdoorruleset", "pulumi")
}
func TestAzureFrontDoorRuleSet_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurefrontdoorruleset", "terraform")
}

// --- Azure Front Door custom domain (fixture RG -> fixture profile -> a
// managed-certificate domain in the pending-validation state; proves the
// validation_token challenge surfaces as an output) ---

func TestAzureFrontDoorCustomDomain_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurefrontdoorcustomdomain", "pulumi")
}
func TestAzureFrontDoorCustomDomain_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurefrontdoorcustomdomain", "terraform")
}

// --- Azure Front Door secret (fixture RG -> fixture profile + the Key
// Vault certificate fixture chain -> a BYO-certificate secret wrapping
// the certificate's versionless id) ---

func TestAzureFrontDoorSecret_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurefrontdoorsecret", "pulumi")
}
func TestAzureFrontDoorSecret_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurefrontdoorsecret", "terraform")
}

// --- Azure Front Door firewall policy (fixture RG -> the WAF policy;
// resource-group-scoped, NO profile fixture -- two scenarios cover the
// STANDARD custom-rules smoke and the PREMIUM managed-rules depth) ---

func TestAzureFrontDoorFirewallPolicy_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurefrontdoorfirewallpolicy", "pulumi")
}
func TestAzureFrontDoorFirewallPolicy_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurefrontdoorfirewallpolicy", "terraform")
}

// --- Azure Front Door security policy (the enforcement seam: fixture RG
// -> profile -> endpoint + STANDARD WAF policy -> the association
// protecting the endpoint's default domain) ---

func TestAzureFrontDoorSecurityPolicy_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurefrontdoorsecuritypolicy", "pulumi")
}
func TestAzureFrontDoorSecurityPolicy_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurefrontdoorsecuritypolicy", "terraform")
}

// --- Azure Service Plan (fixture RG -> Basic B1 Linux plan; the compute tier the app kinds run on) ---

func TestAzureServicePlan_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureserviceplan", "pulumi")
}
func TestAzureServicePlan_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureserviceplan", "terraform")
}

// --- Azure Linux Web App (composed: fixture RG -> fixture plan -> Python app with always-on + health probe) ---

func TestAzureLinuxWebApp_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurelinuxwebapp", "pulumi")
}
func TestAzureLinuxWebApp_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurelinuxwebapp", "terraform")
}

// --- Azure Function App (composed: fixture RG -> fixture plan + scenario-local storage account -> Python function app bound by name + access key) ---

func TestAzureFunctionApp_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurefunctionapp", "pulumi")
}
func TestAzureFunctionApp_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurefunctionapp", "terraform")
}

// --- Azure Container App Environment (fixture RG -> consumption-only environment; the boundary the family lives in) ---

func TestAzureContainerAppEnvironment_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurecontainerappenvironment", "pulumi")
}
func TestAzureContainerAppEnvironment_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurecontainerappenvironment", "terraform")
}

// --- Azure Container App (composed: fixture RG -> fixture environment -> quickstart app with external ingress + HTTP scaling) ---

func TestAzureContainerApp_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurecontainerapp", "pulumi")
}
func TestAzureContainerApp_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurecontainerapp", "terraform")
}

func TestAzurePlantonRunner_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureplantonrunner", "pulumi")
}
func TestAzurePlantonRunner_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureplantonrunner", "terraform")
}

// --- Azure Container App Job (composed: fixture RG -> fixture environment -> schedule-triggered job) ---

func TestAzureContainerAppJob_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurecontainerappjob", "pulumi")
}
func TestAzureContainerAppJob_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurecontainerappjob", "terraform")
}

// --- Azure Container App Environment Storage (composed: fixture RG -> fixture environment + scenario-local account -> share -> SMB registration) ---

func TestAzureContainerAppEnvironmentStorage_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurecontainerappenvironmentstorage", "pulumi")
}
func TestAzureContainerAppEnvironmentStorage_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurecontainerappenvironmentstorage", "terraform")
}

// --- Azure Container App Environment Dapr Component (composed: fixture RG -> fixture environment -> backendless cron-binding component) ---

func TestAzureContainerAppEnvironmentDaprComponent_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurecontainerappenvironmentdaprcomponent", "pulumi")
}
func TestAzureContainerAppEnvironmentDaprComponent_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurecontainerappenvironmentdaprcomponent", "terraform")
}

// --- Azure Log Analytics Workspace (fixture RG -> pay-as-you-go workspace with retention + quota dials) ---

func TestAzureLogAnalyticsWorkspace_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureloganalyticsworkspace", "pulumi")
}
func TestAzureLogAnalyticsWorkspace_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureloganalyticsworkspace", "terraform")
}

// --- Azure Application Insights (composed: fixture RG -> fixture workspace -> workspace-based kind) ---

func TestAzureApplicationInsights_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureapplicationinsights", "pulumi")
}
func TestAzureApplicationInsights_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureapplicationinsights", "terraform")
}

// --- Azure Monitor Diagnostic Setting (composed: fixture RG -> fixture workspace routing its own audit logs into itself) ---

func TestAzureMonitorDiagnosticSetting_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuremonitordiagnosticsetting", "pulumi")
}
func TestAzureMonitorDiagnosticSetting_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuremonitordiagnosticsetting", "terraform")
}

// --- Azure Monitor Action Group (fixture RG -> multi-receiver global notification hub) ---

func TestAzureMonitorActionGroup_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuremonitoractiongroup", "pulumi")
}
func TestAzureMonitorActionGroup_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuremonitoractiongroup", "terraform")
}

// --- Azure Monitor Metric Alert (composed: fixture RG -> fixture action group + scenario-local storage account -> static-threshold rule) ---

func TestAzureMonitorMetricAlert_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuremonitormetricalert", "pulumi")
}
func TestAzureMonitorMetricAlert_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuremonitormetricalert", "terraform")
}

// --- Azure Monitor Autoscale Setting (composed: fixture RG -> scenario-local Standard S1 plan -> CPU rule + weekend recurrence; the shared fixture plan is Basic, which autoscale rejects) ---

func TestAzureMonitorAutoscaleSetting_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuremonitorautoscalesetting", "pulumi")
}
func TestAzureMonitorAutoscaleSetting_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuremonitorautoscalesetting", "terraform")
}

// --- Azure Monitor Data Collection Rule (composed: fixture RG -> fixture workspace -> Linux syslog + perf sources, workspace + metrics destinations) ---

func TestAzureMonitorDataCollectionRule_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuremonitordatacollectionrule", "pulumi")
}
func TestAzureMonitorDataCollectionRule_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuremonitordatacollectionrule", "terraform")
}

// --- Azure Monitor Data Collection Rule Association (composed: fixture VM chain + smoke rule chain -> the target-scoped attachment) ---

func TestAzureMonitorDataCollectionRuleAssociation_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuremonitordatacollectionruleassociation", "pulumi")
}
func TestAzureMonitorDataCollectionRuleAssociation_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuremonitordatacollectionruleassociation", "terraform")
}

// --- Azure Monitor Scheduled Query Alert (composed: fixture RG -> fixture workspace + fixture action group -> row-count KQL rule) ---

func TestAzureMonitorScheduledQueryAlert_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuremonitorscheduledqueryalert", "pulumi")
}
func TestAzureMonitorScheduledQueryAlert_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuremonitorscheduledqueryalert", "terraform")
}

// --- Azure Service Bus family (namespace container + entities + SAS credentials + geo-DR) ---

func TestAzureServiceBusNamespace_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureservicebusnamespace", "pulumi")
}
func TestAzureServiceBusNamespace_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureservicebusnamespace", "terraform")
}

func TestAzureServiceBusQueue_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureservicebusqueue", "pulumi")
}
func TestAzureServiceBusQueue_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureservicebusqueue", "terraform")
}

func TestAzureServiceBusTopic_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureservicebustopic", "pulumi")
}
func TestAzureServiceBusTopic_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureservicebustopic", "terraform")
}

func TestAzureServiceBusSubscription_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureservicebussubscription", "pulumi")
}
func TestAzureServiceBusSubscription_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureservicebussubscription", "terraform")
}

func TestAzureServiceBusAuthorizationRule_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureservicebusauthorizationrule", "pulumi")
}
func TestAzureServiceBusAuthorizationRule_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureservicebusauthorizationrule", "terraform")
}

func TestAzureServiceBusDisasterRecoveryConfig_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureservicebusdisasterrecoveryconfig", "pulumi")
}
func TestAzureServiceBusDisasterRecoveryConfig_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureservicebusdisasterrecoveryconfig", "terraform")
}

// --- Azure Event Grid family (custom topic; domain + its declared domain topics -- the multi-tenant publishers) ---

func TestAzureEventgridTopic_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureeventgridtopic", "pulumi")
}
func TestAzureEventgridTopic_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureeventgridtopic", "terraform")
}

func TestAzureEventgridDomain_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureeventgriddomain", "pulumi")
}
func TestAzureEventgridDomain_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureeventgriddomain", "terraform")
}

func TestAzureEventgridDomainTopic_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureeventgriddomaintopic", "pulumi")
}
func TestAzureEventgridDomainTopic_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureeventgriddomaintopic", "terraform")
}

func TestAzureEventgridSystemTopic_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureeventgridsystemtopic", "pulumi")
}
func TestAzureEventgridSystemTopic_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureeventgridsystemtopic", "terraform")
}

func TestAzureEventgridEventSubscription_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureeventgrideventsubscription", "pulumi")
}
func TestAzureEventgridEventSubscription_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureeventgrideventsubscription", "terraform")
}

func TestAzureEventgridNamespace_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureeventgridnamespace", "pulumi")
}
func TestAzureEventgridNamespace_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureeventgridnamespace", "terraform")
}

func TestAzureEventgridNamespaceTopic_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureeventgridnamespacetopic", "pulumi")
}
func TestAzureEventgridNamespaceTopic_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureeventgridnamespacetopic", "terraform")
}

// --- Azure Event Hubs family (namespace container + hubs + consumer groups + SAS credentials + schema registry + geo-DR + dedicated cluster + CMK) ---

func TestAzureEventHubNamespace_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureeventhubnamespace", "pulumi")
}
func TestAzureEventHubNamespace_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureeventhubnamespace", "terraform")
}

func TestAzureEventHub_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureeventhub", "pulumi")
}
func TestAzureEventHub_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureeventhub", "terraform")
}

func TestAzureEventHubConsumerGroup_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureeventhubconsumergroup", "pulumi")
}
func TestAzureEventHubConsumerGroup_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureeventhubconsumergroup", "terraform")
}

func TestAzureEventHubAuthorizationRule_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureeventhubauthorizationrule", "pulumi")
}
func TestAzureEventHubAuthorizationRule_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureeventhubauthorizationrule", "terraform")
}

func TestAzureEventHubSchemaGroup_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureeventhubschemagroup", "pulumi")
}
func TestAzureEventHubSchemaGroup_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureeventhubschemagroup", "terraform")
}

func TestAzureEventHubDisasterRecoveryConfig_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureeventhubdisasterrecoveryconfig", "pulumi")
}
func TestAzureEventHubDisasterRecoveryConfig_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureeventhubdisasterrecoveryconfig", "terraform")
}

func TestAzureEventHubCluster_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureeventhubcluster", "pulumi")
}
func TestAzureEventHubCluster_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureeventhubcluster", "terraform")
}

func TestAzureEventHubNamespaceCustomerManagedKey_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureeventhubnamespacecustomermanagedkey", "pulumi")
}
func TestAzureEventHubNamespaceCustomerManagedKey_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureeventhubnamespacecustomermanagedkey", "terraform")
}

// --- Azure public DNS (fixture RG -> zone with SOA customization; fixture zone -> typed record sets incl. the alias-A seam) ---

func TestAzureDnsZone_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurednszone", "pulumi")
}
func TestAzureDnsZone_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurednszone", "terraform")
}

func TestAzureDnsRecord_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurednsrecord", "pulumi")
}
func TestAzureDnsRecord_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurednsrecord", "terraform")
}

// --- Azure Container Apps TLS/domain family (fixture environment -> BYO certificate; managed certificate + custom domain are profile-deferred: both block on public-DNS domain validation) ---

func TestAzureContainerAppEnvironmentCertificate_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurecontainerappenvironmentcertificate", "pulumi")
}
func TestAzureContainerAppEnvironmentCertificate_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurecontainerappenvironmentcertificate", "terraform")
}

func TestAzureContainerAppEnvironmentManagedCertificate_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurecontainerappenvironmentmanagedcertificate", "pulumi")
}
func TestAzureContainerAppEnvironmentManagedCertificate_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurecontainerappenvironmentmanagedcertificate", "terraform")
}

func TestAzureContainerAppCustomDomain_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurecontainerappcustomdomain", "pulumi")
}
func TestAzureContainerAppCustomDomain_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurecontainerappcustomdomain", "terraform")
}

// --- Azure Firewall family (IP Group anchor; policy container; rule collection group composed on the policy; the firewall itself on a dedicated AzureFirewallSubnet -- the SLOW lane, ~10-20 min each way) ---

func TestAzureIpGroup_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureipgroup", "pulumi")
}
func TestAzureIpGroup_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureipgroup", "terraform")
}

func TestAzureFirewallPolicy_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurefirewallpolicy", "pulumi")
}
func TestAzureFirewallPolicy_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurefirewallpolicy", "terraform")
}

func TestAzureFirewallPolicyRuleCollectionGroup_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurefirewallpolicyrulecollectiongroup", "pulumi")
}
func TestAzureFirewallPolicyRuleCollectionGroup_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurefirewallpolicyrulecollectiongroup", "terraform")
}

func TestAzureFirewall_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurefirewall", "pulumi")
}
func TestAzureFirewall_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurefirewall", "terraform")
}

// --- Azure Data Factory (fixture RG + scenario-declared fixture storage account -> system-identity factory with the managed VNet and one managed private endpoint; the cloud-side name carries the run-id token -- factory names are globally unique; the endpoint's target-side connection stays Pending by design and the sweep checks the TARGET side) ---

func TestAzureDataFactory_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuredatafactory", "pulumi")
}
func TestAzureDataFactory_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuredatafactory", "terraform")
}

// --- Azure Data Factory Pipeline (composed: factory smoke chain -> a Wait-activity pipeline, the opaque activities_json path proven end to end; free, seconds) ---

func TestAzureDataFactoryPipeline_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuredatafactorypipeline", "pulumi")
}
func TestAzureDataFactoryPipeline_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuredatafactorypipeline", "terraform")
}

// --- Data Factory data flow (chained to the factory smoke instance; the minimal scenario creates the flowlet form as a prerequisite doc, then a mapping flow embedding it -- both provider forms in one lane; authoring metadata, free at rest) ---

func TestAzureDataFactoryDataFlow_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuredatafactorydataflow", "pulumi")
}
func TestAzureDataFactoryDataFlow_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuredatafactorydataflow", "terraform")
}

// --- Data Factory linked service (three scenarios -- web anonymous, key vault via the fixture vault, blob storage via the fixture account's endpoint + managed identity -- all chained to the factory smoke instance and ALL secret-free by design; the other 20 variants are offline-proven per the profile; authoring metadata, free at rest) ---

func TestAzureDataFactoryLinkedService_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuredatafactorylinkedservice", "pulumi")
}
func TestAzureDataFactoryLinkedService_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuredatafactorylinkedservice", "terraform")
}

// --- Data Factory dataset (four scenarios, one per wire form -- delimited text and the flat blob form over the blob linked service instance, the SQL table's linked-service-by-ARM-ID form over a scenario-local connection, and the custom form's raw-JSON seam -- all chained to the factory smoke instance and ALL secret-free by design; the other nine variants are offline-proven per the profile; authoring metadata, free at rest) ---

func TestAzureDataFactoryDataset_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuredatafactorydataset", "pulumi")
}
func TestAzureDataFactoryDataset_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuredatafactorydataset", "terraform")
}

// --- Data Factory trigger (four scenarios, one per variant, all chained to the factory+pipeline smoke instances and deployed STARTED with never-firing conditions -- the Start/Stop lifecycle runs live at zero pipeline-run cost; blob event watches the fixture storage account, custom event subscribes to the Event Grid topic smoke instance) ---

func TestAzureDataFactoryTrigger_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuredatafactorytrigger", "pulumi")
}
func TestAzureDataFactoryTrigger_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuredatafactorytrigger", "terraform")
}

// --- Data Factory integration runtime (three scenarios, one per engine flavor, all chained to the factory smoke instance and ALL secret-free by design -- the managed compute joins the smoke factory's managed virtual network, the self-hosted registration proves the sensitive authorization-key outputs, and the SSIS definition is created STOPPED so it never bills; the deep SSIS arms and the billed interactive-authoring TTL are offline-proven per the profile) ---

func TestAzureDataFactoryIntegrationRuntime_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuredatafactoryintegrationruntime", "pulumi")
}
func TestAzureDataFactoryIntegrationRuntime_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuredatafactoryintegrationruntime", "terraform")
}

// --- Microsoft Fabric Capacity (fixture RG -> the smallest F2 capacity, created-verified-destroyed tightly; bills per hour from create, and the lowercase-alnum-only name class admits NO run-id token -- a crashed run's leftover holds the name and the sweep is the remedy) ---

func TestAzureFabricCapacity_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurefabriccapacity", "pulumi")
}
func TestAzureFabricCapacity_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurefabriccapacity", "terraform")
}

// --- Compute galleries, placement, snapshots (fixture RG -> gallery -> image definition; the image's version scenario chains an OS-typed disk -> snapshot -> version; the availability set and snapshot lanes stand alone -- everything free or pennies at rest) ---

func TestAzureComputeGallery_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurecomputegallery", "pulumi")
}
func TestAzureComputeGallery_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurecomputegallery", "terraform")
}

func TestAzureComputeGalleryImage_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurecomputegalleryimage", "pulumi")
}
func TestAzureComputeGalleryImage_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurecomputegalleryimage", "terraform")
}

func TestAzureAvailabilitySet_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azureavailabilityset", "pulumi")
}
func TestAzureAvailabilitySet_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azureavailabilityset", "terraform")
}

func TestAzureDiskSnapshot_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azuredisksnapshot", "pulumi")
}
func TestAzureDiskSnapshot_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azuredisksnapshot", "terraform")
}

func TestAzureContainerInstance_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurecontainerinstance", "pulumi")
}
func TestAzureContainerInstance_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurecontainerinstance", "terraform")
}

// --- Function App Flex Consumption (two scenarios, each chaining its OWN scenario-local FC1 plan + storage account + blob container by annotation -- the shared plan fixture's B1 SKU is illegal for flex; `minimal` proves connection-string deployment-storage auth with the scale dials and one always-ready warm instance, `identity-auth` proves system-assigned-identity storage auth with no key anywhere; both effectively $0) ---

func TestAzureFunctionAppFlexConsumption_Pulumi(t *testing.T) {
	runAllScenariosForKind(t, "azurefunctionappflexconsumption", "pulumi")
}
func TestAzureFunctionAppFlexConsumption_Terraform(t *testing.T) {
	runAllScenariosForKind(t, "azurefunctionappflexconsumption", "terraform")
}

// runAllScenariosForKind discovers and runs all E2E scenarios for an Azure kind.
func runAllScenariosForKind(t *testing.T, kindDir, engine string) {
	t.Helper()

	if cp, err := profilepkg.LoadKindProfile(repoRoot, "azure", kindDir); err == nil && cp.Spec != nil {
		switch cp.Spec.Status {
		case kindv1.CatalogKindE2EProfileSpec_deferred,
			kindv1.CatalogKindE2EProfileSpec_skip,
			kindv1.CatalogKindE2EProfileSpec_stub,
			// pending_proof: fully authored, offline-validated, awaiting its
			// first live proof. The proving session flips the profile to green
			// immediately before executing the lanes; until then a sweep must
			// never run it.
			kindv1.CatalogKindE2EProfileSpec_pending_proof:
			reason := cp.Spec.DeferredReason
			if reason == "" {
				reason = cp.Spec.Status.String()
			}
			t.Skipf("kind %s E2E profile status is %s: %s", kindDir, cp.Spec.Status, reason)
		}
	}

	moduleDir, err := discovery.ModuleDir(repoRoot, "azure", kindDir, engine)
	if err != nil {
		t.Fatalf("failed to locate %s %s module: %v", kindDir, engine, err)
	}

	if !fileExists(moduleDir) {
		t.Skipf("kind %s %s module not found at %s", kindDir, engine, moduleDir)
	}

	scenarios, err := discovery.DiscoverTestScenarios(repoRoot, "azure", kindDir)
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

	tc := &provider.KindTestContext{
		Kind:                   kindDir,
		Provider:               "azure",
		Engine:                 engine,
		ModuleDir:              moduleDir,
		ManifestPath:           scenario.ManifestPath,
		RepoRoot:               repoRoot,
		RunID:                  runID,
		T:                      t,
		AssertApplyIdempotency: assertApplyIdempotency,
	}

	if engine == "pulumi" {
		// GenerateStackName enforces the length cap uniqueness-preservingly
		// (blind truncation here would collide long kind names' scenarios).
		tc.StackName = runner.GenerateStackName(kindDir+"-"+scenario.Name, runID)
		tc.BackendURL = pulumiBackendURL
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
