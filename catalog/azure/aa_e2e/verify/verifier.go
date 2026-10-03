// Package verify checks that Azure resources created by an E2E scenario exist
// after DEPLOY and are gone after DESTROY. Each component family has its own
// verifier because Azure verification is service-specific (CheckExistence for a
// resource group, a GET for a VNet, ...). All verifiers run against the same
// ambient credential chain the deploy used, so a verification failure reflects
// real cloud state.
package verify

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/pkg/errors"
)

// Verifier checks a single component's Azure resource for existence/absence.
// Azure resource verification is subscription-scoped, so verifiers take the
// subscription id and the ambient token credential -- not a region, unlike AWS.
type Verifier interface {
	// IDOutputKey is the output key carrying the identifier used to verify
	// the resource (e.g. "resource_group_name").
	IDOutputKey() string
	// VerifyExists returns an error unless the resource exists.
	VerifyExists(ctx context.Context, cred azcore.TokenCredential, subscriptionID, id string) error
	// VerifyAbsent returns an error unless the resource is gone.
	VerifyAbsent(ctx context.Context, cred azcore.TokenCredential, subscriptionID, id string) error
}

// RuntimeCauseVerifier is the optional capability a verifier implements when
// its kind's scenario deploys a workload DESIGNED to fail (the framework's
// expected-runtime-failure lane): it must pin the failure to the expected
// cause with ARM's own evidence, never merely tolerate "some failure".
type RuntimeCauseVerifier interface {
	VerifyRuntimeFailureCause(ctx context.Context, cred azcore.TokenCredential, subscriptionID, id, cause string) error
}

// verifiers maps a component name to its verifier. New Azure components register
// here as they are forged.
var verifiers = map[string]Verifier{
	"azureaifoundry":                                 &aiFoundryVerifier{},
	"azureaifoundryproject":                          &aiFoundryProjectVerifier{},
	"azureakscluster":                                &aksClusterVerifier{},
	"azureaksnodepool":                               &aksNodePoolVerifier{},
	"azureapplicationgateway":                        &applicationGatewayVerifier{},
	"azureapplicationinsights":                       &applicationInsightsVerifier{},
	"azureapplicationinsightsstandardwebtest":        &applicationInsightsStandardWebTestVerifier{},
	"azureapplicationsecuritygroup":                  &applicationSecurityGroupVerifier{},
	"azureavailabilityset":                           &availabilitySetVerifier{},
	"azurebackupcontainerstorageaccount":             &backupContainerStorageAccountVerifier{},
	"azurebackuppolicyfileshare":                     &backupPolicyFileShareVerifier{},
	"azurebackuppolicyvm":                            &backupPolicyVmVerifier{},
	"azurebackupprotectedfileshare":                  &backupProtectedFileShareVerifier{},
	"azurebackupprotectedvm":                         &backupProtectedVmVerifier{},
	"azurebastionhost":                               &bastionHostVerifier{},
	"azurecognitiveaccount":                          &cognitiveAccountVerifier{},
	"azurecognitiveaccountproject":                   &cognitiveAccountProjectVerifier{},
	"azurecognitivedeployment":                       &cognitiveDeploymentVerifier{},
	"azurecomputegallery":                            &computeGalleryVerifier{},
	"azurecomputegalleryimage":                       &computeGalleryImageVerifier{},
	"azurecontainerapp":                              &containerAppVerifier{},
	"azurecontainerappcustomdomain":                  &containerAppCustomDomainVerifier{},
	"azurecontainerappenvironment":                   &containerAppEnvironmentVerifier{},
	"azurecontainerappenvironmentcertificate":        &containerAppEnvironmentCertificateVerifier{},
	"azurecontainerappenvironmentdaprcomponent":      &containerAppEnvironmentDaprComponentVerifier{},
	"azurecontainerappenvironmentmanagedcertificate": &containerAppEnvironmentManagedCertificateVerifier{},
	"azurecontainerappenvironmentstorage":            &containerAppEnvironmentStorageVerifier{},
	"azurecontainerappjob":                           &containerAppJobVerifier{},
	"azurecontainerinstance":                         &containerGroupVerifier{},
	"azurecontainerregistry":                         &containerRegistryVerifier{},
	"azurecosmosdbaccount":                           &cosmosdbAccountVerifier{},
	"azurecosmosdbmongocollection":                   &cosmosdbMongoCollectionVerifier{},
	"azurecosmosdbmongodatabase":                     &cosmosdbMongoDatabaseVerifier{},
	"azurecosmosdbsqlcontainer":                      &cosmosdbSqlContainerVerifier{},
	"azurecosmosdbsqldatabase":                       &cosmosdbSqlDatabaseVerifier{},
	"azurecosmosdbsqlroleassignment":                 &cosmosdbSqlRoleAssignmentVerifier{},
	"azurecosmosdbsqlroledefinition":                 &cosmosdbSqlRoleDefinitionVerifier{},
	"azuredatafactory":                               &dataFactoryVerifier{},
	"azuredatafactorypipeline":                       &dataFactoryPipelineVerifier{},
	"azuredatafactorydataflow":                       &dataFactoryDataFlowVerifier{},
	"azuredatafactorydataset":                        &dataFactoryDatasetVerifier{},
	"azuredatafactoryintegrationruntime":             &dataFactoryIntegrationRuntimeVerifier{},
	"azuredatafactorylinkedservice":                  &dataFactoryLinkedServiceVerifier{},
	"azuredatafactorytrigger":                        &dataFactoryTriggerVerifier{},
	"azuredataprotectionbackupinstance":              &dataProtectionBackupInstanceVerifier{},
	"azuredataprotectionbackuppolicy":                &dataProtectionBackupPolicyVerifier{},
	"azuredataprotectionbackupvault":                 &dataProtectionBackupVaultVerifier{},
	"azuredataprotectionresourceguard":               &dataProtectionResourceGuardVerifier{},
	"azurediskencryptionset":                         &diskEncryptionSetVerifier{},
	"azuredisksnapshot":                              &diskSnapshotVerifier{},
	"azurednsrecord":                                 &dnsRecordVerifier{},
	"azurednszone":                                   &dnsZoneVerifier{},
	"azureeventgriddomain":                           &eventgridDomainVerifier{},
	"azureeventgriddomaintopic":                      &eventgridDomainTopicVerifier{},
	"azureeventgrideventsubscription":                &eventgridEventSubscriptionVerifier{},
	"azureeventgridnamespace":                        &eventgridNamespaceVerifier{},
	"azureeventgridnamespacetopic":                   &eventgridNamespaceTopicVerifier{},
	"azureeventgridsystemtopic":                      &eventgridSystemTopicVerifier{},
	"azureeventgridtopic":                            &eventgridTopicVerifier{},
	"azureeventhub":                                  &eventHubResourceVerifier{component: "azureeventhub", idOutputKey: "event_hub_id"},
	"azureeventhubauthorizationrule":                 &eventHubResourceVerifier{component: "azureeventhubauthorizationrule", idOutputKey: "authorization_rule_id"},
	"azureeventhubcluster":                           &eventHubResourceVerifier{component: "azureeventhubcluster", idOutputKey: "cluster_id"},
	"azureeventhubconsumergroup":                     &eventHubResourceVerifier{component: "azureeventhubconsumergroup", idOutputKey: "consumer_group_id"},
	"azureeventhubdisasterrecoveryconfig":            &eventHubResourceVerifier{component: "azureeventhubdisasterrecoveryconfig", idOutputKey: "disaster_recovery_config_id"},
	"azureeventhubnamespace":                         &eventHubResourceVerifier{component: "azureeventhubnamespace", idOutputKey: "namespace_id"},
	"azureeventhubschemagroup":                       &eventHubResourceVerifier{component: "azureeventhubschemagroup", idOutputKey: "schema_group_id"},
	"azureexpressroutecircuit":                       &expressRouteCircuitVerifier{},
	"azureexpressroutecircuitpeering":                &expressRouteCircuitPeeringVerifier{},
	"azureexpressroutegateway":                       &expressRouteGatewayVerifier{},
	"azureexpressrouteport":                          &expressRoutePortVerifier{},
	"azurefabriccapacity":                            &fabricCapacityVerifier{},
	"azurefederatedidentitycredential":               &federatedIdentityCredentialVerifier{},
	"azurefirewall":                                  &firewallVerifier{},
	"azurefirewallpolicy":                            &firewallPolicyVerifier{},
	"azurefirewallpolicyrulecollectiongroup":         &firewallPolicyRuleCollectionGroupVerifier{},
	"azurefrontdoorcustomdomain":                     &frontDoorCustomDomainVerifier{},
	"azurefrontdoorendpoint":                         &frontDoorEndpointVerifier{},
	"azurefrontdoorfirewallpolicy":                   &frontDoorFirewallPolicyVerifier{},
	"azurefrontdoororigin":                           &frontDoorOriginVerifier{},
	"azurefrontdoororigingroup":                      &frontDoorOriginGroupVerifier{},
	"azurefrontdoorprofile":                          &frontDoorProfileVerifier{},
	"azurefrontdoorroute":                            &frontDoorRouteVerifier{},
	"azurefrontdoorruleset":                          &frontDoorRuleSetVerifier{},
	"azurefrontdoorsecret":                           &frontDoorSecretVerifier{},
	"azurefrontdoorsecuritypolicy":                   &frontDoorSecurityPolicyVerifier{},
	"azurefunctionapp":                               &functionAppVerifier{},
	"azurefunctionappflexconsumption":                &functionAppFlexConsumptionVerifier{},
	"azureipgroup":                                   &ipGroupVerifier{},
	"azurekeyvault":                                  &keyVaultVerifier{},
	"azurekeyvaultcertificate":                       &keyVaultCertificateVerifier{},
	"azurekeyvaultkey":                               &keyVaultKeyVerifier{},
	"azurekeyvaultsecret":                            &keyVaultSecretVerifier{},
	"azurelinuxwebapp":                               &linuxWebAppVerifier{},
	"azureloadbalancer":                              &loadBalancerVerifier{},
	"azurelocalnetworkgateway":                       &localNetworkGatewayVerifier{},
	"azureloganalyticsworkspace":                     &logAnalyticsWorkspaceVerifier{},
	"azuremachinelearningbatchendpoint":              &machineLearningBatchEndpointVerifier{},
	"azuremachinelearningbatchdeployment":            &machineLearningBatchDeploymentVerifier{},
	"azuremachinelearningcomputecluster":             &machineLearningComputeClusterVerifier{},
	"azuremachinelearningcomputeinstance":            &machineLearningComputeInstanceVerifier{},
	"azuremachinelearningonlineendpoint":             &machineLearningOnlineEndpointVerifier{},
	"azuremachinelearningonlinedeployment":           &machineLearningOnlineDeploymentVerifier{},
	"azuremachinelearningdatastore":                  &machineLearningDatastoreVerifier{},
	"azuremachinelearningworkspace":                  &machineLearningWorkspaceVerifier{},
	"azuremanageddisk":                               &managedDiskVerifier{},
	"azuremanagedredis":                              &managedRedisVerifier{},
	"azuremanagedredisaccesspolicyassignment":        &managedRedisAccessPolicyAssignmentVerifier{},
	"azuremanagedredisgeoreplication":                &managedRedisGeoReplicationVerifier{},
	"azuremongocluster":                              &mongoClusterVerifier{},
	"azuremongoclusteruser":                          &mongoClusterUserVerifier{},
	"azuremonitoractiongroup":                        &monitorActionGroupVerifier{},
	"azuremonitoractivitylogalert":                   &monitorActivityLogAlertVerifier{},
	"azuremonitorautoscalesetting":                   &monitorAutoscaleSettingVerifier{},
	"azuremonitordatacollectionrule":                 &monitorDataCollectionRuleVerifier{},
	"azuremonitordatacollectionruleassociation":      &monitorDataCollectionRuleAssociationVerifier{},
	"azuremonitordiagnosticsetting":                  &monitorDiagnosticSettingVerifier{},
	"azuremonitormetricalert":                        &monitorMetricAlertVerifier{},
	"azuremonitorscheduledqueryalert":                &monitorScheduledQueryAlertVerifier{},
	"azuremssqldatabase":                             &mssqlDatabaseVerifier{},
	"azuremssqlelasticpool":                          &mssqlElasticPoolVerifier{},
	"azuremssqlfailovergroup":                        &mssqlFailoverGroupVerifier{},
	"azuremssqlserver":                               &mssqlServerVerifier{},
	"azuremysqlflexibleserver":                       &mysqlFlexibleServerVerifier{},
	"azurenatgateway":                                &natGatewayVerifier{},
	"azurenetworkinterface":                          &networkInterfaceVerifier{},
	"azurenetworksecuritygroup":                      &networkSecurityGroupVerifier{},
	"azurenetworkwatcherflowlog":                     &networkWatcherFlowLogVerifier{},
	"azurepointtositevpngateway":                     &pointToSiteVpnGatewayVerifier{},
	"azureplantonrunner":                             &plantonRunnerVerifier{},
	"azurepostgresqlflexibleserver":                  &postgresqlFlexibleServerVerifier{},
	"azureprivatednsrecord":                          &privateDnsRecordVerifier{},
	"azureprivatednsresolver":                        &privateDnsResolverVerifier{},
	"azureprivatednsresolverforwardingruleset":       &privateDnsResolverForwardingRulesetVerifier{},
	"azureprivatednsresolvervirtualnetworklink":      &privateDnsResolverVirtualNetworkLinkVerifier{},
	"azureprivatednszone":                            &privateDnsZoneVerifier{},
	"azureprivatednszonevirtualnetworklink":          &privateDnsZoneVirtualNetworkLinkVerifier{},
	"azureprivateendpoint":                           &privateEndpointVerifier{},
	"azureprivatelinkservice":                        &privateLinkServiceVerifier{},
	"azurepublicip":                                  &publicIpVerifier{},
	"azurepublicipprefix":                            &publicIpPrefixVerifier{},
	"azurerecoveryservicesvault":                     &recoveryServicesVaultVerifier{},
	"azurerediscache":                                &redisCacheVerifier{},
	"azurerediscacheaccesspolicy":                    &redisCacheAccessPolicyVerifier{},
	"azurerediscacheaccesspolicyassignment":          &redisCacheAccessPolicyAssignmentVerifier{},
	"azureredislinkedserver":                         &redisLinkedServerVerifier{},
	"azureresourcegroup":                             &resourceGroupVerifier{},
	"azureroleassignment":                            &roleAssignmentVerifier{},
	"azureroledefinition":                            &roleDefinitionVerifier{},
	"azureroutetable":                                &routeTableVerifier{},
	"azuresearchservice":                             &searchServiceVerifier{},
	"azureservicebusauthorizationrule":               &serviceBusResourceVerifier{component: "azureservicebusauthorizationrule", idOutputKey: "authorization_rule_id"},
	"azureservicebusdisasterrecoveryconfig":          &serviceBusResourceVerifier{component: "azureservicebusdisasterrecoveryconfig", idOutputKey: "disaster_recovery_config_id"},
	"azureservicebusnamespace":                       &serviceBusResourceVerifier{component: "azureservicebusnamespace", idOutputKey: "namespace_id"},
	"azureservicebusqueue":                           &serviceBusResourceVerifier{component: "azureservicebusqueue", idOutputKey: "queue_id"},
	"azureservicebussubscription":                    &serviceBusResourceVerifier{component: "azureservicebussubscription", idOutputKey: "subscription_id"},
	"azureservicebustopic":                           &serviceBusResourceVerifier{component: "azureservicebustopic", idOutputKey: "topic_id"},
	"azureserviceplan":                               &servicePlanVerifier{},
	"azurestorageaccount":                            &storageAccountVerifier{},
	"azurestoragecontainer":                          &storageContainerVerifier{},
	"azurestoragedatalakegen2filesystem":             &storageDataLakeGen2FilesystemVerifier{},
	"azurestorageencryptionscope":                    &storageEncryptionScopeVerifier{},
	"azurestoragelocaluser":                          &storageLocalUserVerifier{},
	"azurestorageobjectreplication":                  &storageObjectReplicationVerifier{},
	"azurestoragequeue":                              &storageQueueVerifier{},
	"azurestorageshare":                              &storageShareVerifier{},
	"azurestoragetable":                              &storageTableVerifier{},
	"azuresubnet":                                    &subnetVerifier{},
	"azuretrafficmanagerendpoint":                    &trafficManagerEndpointVerifier{},
	"azuretrafficmanagerprofile":                     &trafficManagerProfileVerifier{},
	"azureuserassignedidentity":                      &userAssignedIdentityVerifier{},
	"azurevirtualhub":                                &virtualHubVerifier{},
	"azurevirtualhubconnection":                      &virtualHubConnectionVerifier{},
	"azurevirtualmachine":                            &virtualMachineVerifier{},
	"azurevirtualmachinescaleset":                    &virtualMachineScaleSetVerifier{},
	"azurevirtualnetwork":                            &virtualNetworkVerifier{},
	"azurevirtualnetworkgateway":                     &virtualNetworkGatewayVerifier{},
	"azurevirtualnetworkgatewayconnection":           &virtualNetworkGatewayConnectionVerifier{},
	"azurevirtualnetworkpeering":                     &virtualNetworkPeeringVerifier{},
	"azurevirtualwan":                                &virtualWanVerifier{},
	"azurevpngateway":                                &vpnGatewayVerifier{},
	"azurevpngatewayconnection":                      &vpnGatewayConnectionVerifier{},
	"azurevpnserverconfiguration":                    &vpnServerConfigurationVerifier{},
	"azurevpnsite":                                   &vpnSiteVerifier{},
	"azurewebapplicationfirewallpolicy":              &webApplicationFirewallPolicyVerifier{},
}

// GetVerifier returns the verifier for a component, or an error if none is registered.
func GetVerifier(component string) (Verifier, error) {
	v, ok := verifiers[component]
	if !ok {
		return nil, errors.Errorf("no Azure verifier registered for component %q", component)
	}
	return v, nil
}
