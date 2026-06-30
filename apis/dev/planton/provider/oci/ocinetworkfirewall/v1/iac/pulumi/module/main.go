package module

import (
	"github.com/pkg/errors"
	ocinetworkfirewallv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ocinetworkfirewall/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/oci/pulumiociprovider"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var addressListTypeMap = map[ocinetworkfirewallv1.OciNetworkFirewallSpec_AddressList_AddressListType]string{
	ocinetworkfirewallv1.OciNetworkFirewallSpec_AddressList_ip:   "IP",
	ocinetworkfirewallv1.OciNetworkFirewallSpec_AddressList_fqdn: "FQDN",
}

var serviceTypeMap = map[ocinetworkfirewallv1.OciNetworkFirewallSpec_Service_ServiceType]string{
	ocinetworkfirewallv1.OciNetworkFirewallSpec_Service_tcp_service: "TCP_SERVICE",
	ocinetworkfirewallv1.OciNetworkFirewallSpec_Service_udp_service: "UDP_SERVICE",
}

var actionMap = map[ocinetworkfirewallv1.OciNetworkFirewallSpec_SecurityRule_Action]string{
	ocinetworkfirewallv1.OciNetworkFirewallSpec_SecurityRule_allow:   "ALLOW",
	ocinetworkfirewallv1.OciNetworkFirewallSpec_SecurityRule_drop:    "DROP",
	ocinetworkfirewallv1.OciNetworkFirewallSpec_SecurityRule_reject:  "REJECT",
	ocinetworkfirewallv1.OciNetworkFirewallSpec_SecurityRule_inspect: "INSPECT",
}

var inspectionMap = map[ocinetworkfirewallv1.OciNetworkFirewallSpec_SecurityRule_Inspection]string{
	ocinetworkfirewallv1.OciNetworkFirewallSpec_SecurityRule_intrusion_detection:  "INTRUSION_DETECTION",
	ocinetworkfirewallv1.OciNetworkFirewallSpec_SecurityRule_intrusion_prevention: "INTRUSION_PREVENTION",
}

func Resources(ctx *pulumi.Context, stackInput *ocinetworkfirewallv1.OciNetworkFirewallStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	ociProvider, err := pulumiociprovider.Get(ctx, stackInput.ProviderConfig)
	if err != nil {
		return errors.Wrap(err, "failed to setup oci provider")
	}

	createdPolicy, err := policyResource(ctx, locals, ociProvider)
	if err != nil {
		return errors.Wrap(err, "failed to create firewall policy")
	}

	addressListResources, err := addressListResources(ctx, locals, ociProvider, createdPolicy)
	if err != nil {
		return errors.Wrap(err, "failed to create address lists")
	}

	serviceResources, err := serviceResources(ctx, locals, ociProvider, createdPolicy)
	if err != nil {
		return errors.Wrap(err, "failed to create services")
	}

	serviceListResources, err := serviceListResources(ctx, locals, ociProvider, createdPolicy, serviceResources)
	if err != nil {
		return errors.Wrap(err, "failed to create service lists")
	}

	urlListResources, err := urlListResources(ctx, locals, ociProvider, createdPolicy)
	if err != nil {
		return errors.Wrap(err, "failed to create url lists")
	}

	var allSubResources []pulumi.Resource
	allSubResources = append(allSubResources, addressListResources...)
	allSubResources = append(allSubResources, serviceResources...)
	allSubResources = append(allSubResources, serviceListResources...)
	allSubResources = append(allSubResources, urlListResources...)

	securityRuleResources, err := securityRuleResources(ctx, locals, ociProvider, createdPolicy, allSubResources)
	if err != nil {
		return errors.Wrap(err, "failed to create security rules")
	}

	if err := firewallResource(ctx, locals, ociProvider, createdPolicy, securityRuleResources); err != nil {
		return errors.Wrap(err, "failed to create network firewall")
	}

	return nil
}

func pulumiOciOpt(provider *oci.Provider) pulumi.ResourceOption {
	return pulumi.Provider(provider)
}
