// Package verify checks that AWS resources created by an E2E scenario exist after
// DEPLOY and are gone after DESTROY. Each component family has its own verifier
// because AWS verification is service-specific (HeadBucket for S3,
// DescribeSubnets for a subnet, ...) -- unlike the single Management-API path a
// SaaS provider uses. All verifiers run against the same ambient credential
// chain the deploy used, so a verification failure reflects real cloud state.
package verify

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/pkg/errors"
)

// Verifier checks a single component's AWS resource for existence/absence.
type Verifier interface {
	// IDOutputKey is the output key carrying the identifier used to verify
	// the resource (e.g. "bucket_id").
	IDOutputKey() string
	// VerifyExists returns an error unless the resource exists.
	VerifyExists(ctx context.Context, cfg aws.Config, id, region string) error
	// VerifyAbsent returns an error unless the resource is gone.
	VerifyAbsent(ctx context.Context, cfg aws.Config, id, region string) error
}

// OutputsVerifier inspects the full output map when a single string id is
// insufficient (e.g. AwsS3ObjectSet verifies HeadObject per key in object_etags).
type OutputsVerifier interface {
	Verifier
	VerifyExistsFromOutputs(ctx context.Context, cfg aws.Config, outputs map[string]interface{}, region string) error
	VerifyAbsentFromOutputs(ctx context.Context, cfg aws.Config, outputs map[string]interface{}, region string) error
}

// RuntimeCauseVerifier is the optional capability a verifier implements when
// its kind's scenario deploys a workload DESIGNED to fail (the framework's
// expected-runtime-failure lane): it must pin the failure to exactly the
// expected cause with AWS's own evidence -- task histories and log events --
// never merely tolerate "some failure".
type RuntimeCauseVerifier interface {
	VerifyRuntimeFailureCause(ctx context.Context, cfg aws.Config, outputs map[string]interface{}, region, cause string) error
}

// verifiers maps a component name to its verifier. New AWS components register
// here as they are forged; today it carries the S3 walking-skeleton only.
var verifiers = map[string]Verifier{
	"awscertmanagercert":              &acmCertificateVerifier{},
	"awscloudfront":                   &cloudFrontDistributionVerifier{},
	"awssqsqueue":                     &sqsQueueVerifier{},
	"awssnstopic":                     &snsTopicVerifier{},
	"awssnssubscription":              &snsSubscriptionVerifier{},
	"awseventbridgebus":               &eventBridgeBusVerifier{},
	"awseventbridgerule":              &eventBridgeRuleVerifier{},
	"awss3bucket":                     &s3Verifier{},
	"awss3objectset":                  &s3ObjectSetVerifier{},
	"awssubnet":                       &subnetVerifier{},
	"awsvpc":                          &vpcVerifier{},
	"awsinternetgateway":              &internetGatewayVerifier{},
	"awsegressonlyinternetgateway":    &egressOnlyInternetGatewayVerifier{},
	"awsvpcendpoint":                  &vpcEndpointVerifier{},
	"awsnatgateway":                   &natGatewayVerifier{},
	"awselasticip":                    &elasticIpVerifier{},
	"awsiampolicy":                    &iamPolicyVerifier{},
	"awsiamrole":                      &iamRoleVerifier{},
	"awsiaminstanceprofile":           &iamInstanceProfileVerifier{},
	"awsiamuser":                      &iamUserVerifier{},
	"awsiamoidcprovider":              &iamOidcProviderVerifier{},
	"awsiamgroup":                     &iamGroupVerifier{},
	"awsiamsamlprovider":              &iamSamlProviderVerifier{},
	"awsiamaccountsettings":           &iamAccountSettingsVerifier{},
	"awsbudget":                       &budgetVerifier{},
	"awscostanomalymonitor":           &costAnomalyMonitorVerifier{},
	"awscostcategory":                 &costCategoryVerifier{},
	"awsalb":                          &loadBalancerVerifier{kind: "awsalb"},
	"awsnlb":                          &loadBalancerVerifier{kind: "awsnlb"},
	"awslbtargetgroup":                &targetGroupVerifier{},
	"awslblistener":                   &listenerVerifier{},
	"awslblistenerrule":               &listenerRuleVerifier{},
	"awslaunchtemplate":               &launchTemplateVerifier{},
	"awsautoscalinggroup":             &autoScalingGroupVerifier{},
	"awsecscluster":                   &ecsClusterVerifier{},
	"awsecstaskdefinition":            &ecsTaskDefinitionVerifier{},
	"awsecsservice":                   &ecsServiceVerifier{},
	"awsplantonrunner":                &plantonRunnerVerifier{},
	"awsrdscluster":                   &rdsClusterVerifier{},
	"awsrdsinstance":                  &rdsInstanceVerifier{},
	"awsdocumentdb":                   &docdbClusterVerifier{},
	"awsneptunecluster":               &neptuneClusterVerifier{},
	"awsredshiftcluster":              &redshiftClusterVerifier{},
	"awsredshiftserverlessnamespace":  &redshiftServerlessNamespaceVerifier{},
	"awsredshiftserverlessworkgroup":  &redshiftServerlessWorkgroupVerifier{},
	"awssesconfigurationset":          &sesConfigurationSetVerifier{},
	"awssesemailidentity":             &sesEmailIdentityVerifier{},
	"awsekscluster":                   &eksClusterVerifier{},
	"awseksnodegroup":                 &eksNodeGroupVerifier{},
	"awseksaddon":                     &eksAddonVerifier{},
	"awseksfargateprofile":            &eksFargateProfileVerifier{},
	"awseksaccessentry":               &eksAccessEntryVerifier{},
	"awsdynamodb":                     &dynamodbTableVerifier{},
	"awskinesisstream":                &kinesisStreamVerifier{},
	"awskinesisstreamconsumer":        &kinesisStreamConsumerVerifier{},
	"awskinesisfirehose":              &kinesisFirehoseVerifier{},
	"awslambda":                       &lambdaFunctionVerifier{},
	"awslambdaeventsourcemapping":     &lambdaEventSourceMappingVerifier{},
	"awskmskey":                       &kmsKeyVerifier{},
	"awselasticacheuser":              &elasticacheUserVerifier{},
	"awselasticacheusergroup":         &elasticacheUserGroupVerifier{},
	"awsrediselasticache":             &elasticacheReplicationGroupVerifier{},
	"awsmemcachedelasticache":         &elasticacheClusterVerifier{},
	"awsserverlesselasticache":        &elasticacheServerlessCacheVerifier{},
	"awssecuritygroup":                &securityGroupVerifier{},
	"awsmskcluster":                   &mskClusterVerifier{},
	"awsmskserverlesscluster":         &mskServerlessClusterVerifier{},
	"awsmwaaenvironment":              &mwaaEnvironmentVerifier{},
	"awsopensearchdomain":             &opensearchDomainVerifier{},
	"awsec2instance":                  &ec2InstanceVerifier{},
	"awsecrrepo":                      &ecrRepoVerifier{},
	"awsroute53zone":                  &route53ZoneVerifier{},
	"awsroute53dnsrecord":             &route53DnsRecordVerifier{},
	"awsroute53healthcheck":           &route53HealthCheckVerifier{},
	"awscloudwatchloggroup":           &cloudwatchLogGroupVerifier{},
	"awscloudwatchalarm":              &cloudwatchAlarmVerifier{},
	"awscloudwatchdashboard":          &cloudwatchDashboardVerifier{},
	"awscloudwatchsynthetics":         &syntheticsVerifier{},
	"awscloudwatchlogdelivery":        &logDeliveryVerifier{},
	"awscloudwatchlogaccountpolicy":   &logAccountPolicyVerifier{},
	"awscloudwatchloganomalydetector": &logAnomalyDetectorVerifier{},
	"awscloudwatchlogresourcepolicy":  &logResourcePolicyVerifier{},
	"awsmanagedprometheus":            &ampWorkspaceVerifier{},
	"awsmanagedprometheusscraper":     &ampScraperVerifier{},
	"awseventbridgepipe":              &eventBridgePipeVerifier{},
	"awseventbridgescheduler":         &eventBridgeSchedulerVerifier{},
	"awseventbridgeapidestination":    &eventBridgeApiDestinationVerifier{},
	"awsvpcpeering":                   &vpcPeeringVerifier{},
	"awsnetworkacl":                   &networkAclVerifier{},
	"awsmanagedprefixlist":            &managedPrefixListVerifier{},
	"awsebsvolume":                    &ebsVolumeVerifier{},
	"awsebssnapshot":                  &ebsSnapshotVerifier{},
	"awsdlmlifecyclepolicy":           &dlmLifecyclePolicyVerifier{},
	"awss3directorybucket":            &s3DirectoryBucketVerifier{},
	"awss3tablebucket":                &s3TableBucketVerifier{},
	"awss3vectorbucket":               &s3VectorBucketVerifier{},
	"awsroute53resolverendpoint":      &resolverEndpointVerifier{},
	"awsroute53resolverfirewall":      &resolverFirewallVerifier{},
	"awsroute53resolverquerylog":      &resolverQueryLogVerifier{},
	"awscloudmapnamespace":            &cloudMapNamespaceVerifier{},
	"awslambdalayer":                  &lambdaLayerVerifier{},
	"awsrdsproxy":                     &rdsProxyVerifier{},
	"awsappsyncapi":                   &appSyncApiVerifier{},
	"awsauroradsql":                   &dsqlClusterVerifier{},
	"awsecrregistrysettings":          &ecrRegistrySettingsVerifier{},
	"awsprivateca":                    &privateCaVerifier{},
	"awscloudwatchcompositealarm":     &cloudwatchCompositeAlarmVerifier{},
	"awsstepfunction":                 &stepFunctionVerifier{},
	"awshttpapigateway":               &httpApiGatewayVerifier{},
	"awshttpapivpclink":               &httpApiVpcLinkVerifier{},
	"awshttpapidomain":                &httpApiDomainVerifier{},
	"awsrestapigateway":               &restApiGatewayVerifier{},
	"awsrestapidomain":                &restApiDomainVerifier{},
	"awsrestapiusageplan":             &restApiUsagePlanVerifier{},
	"awsrestapivpclink":               &restApiVpcLinkVerifier{},
	"awscognitouserpool":              &cognitoUserPoolVerifier{},
	"awscognitouserpoolclient":        &cognitoUserPoolClientVerifier{},
	"awscognitoidentityprovider":      &cognitoIdentityProviderVerifier{},
	"awscognitoresourceserver":        &cognitoResourceServerVerifier{},
	"awselasticfilesystem":            &efsFileSystemVerifier{},
	"awsefsaccesspoint":               &efsAccessPointVerifier{},
	"awswafwebacl":                    &wafWebAclVerifier{},
	"awswafipset":                     &wafIpSetVerifier{},
	"awswafregexpatternset":           &wafRegexPatternSetVerifier{},
	"awsbatchcomputeenvironment":      &batchComputeEnvironmentVerifier{},
	"awsbatchjobqueue":                &batchJobQueueVerifier{},
	"awsbatchschedulingpolicy":        &batchSchedulingPolicyVerifier{},
	"awsbatchjobdefinition":           &batchJobDefinitionVerifier{},

	"awsapprunnerservice":                    &appRunnerServiceVerifier{},
	"awsapprunnerautoscalingconfiguration":   &appRunnerAutoScalingConfigurationVerifier{},
	"awsapprunnervpcconnector":               &appRunnerVpcConnectorVerifier{},
	"awsapprunnerobservabilityconfiguration": &appRunnerObservabilityConfigurationVerifier{},

	"awsbedrockguardrail":             &bedrockGuardrailVerifier{},
	"awsbedrockcustommodel":           &bedrockCustomModelVerifier{},
	"awsbedrockinferenceprofile":      &bedrockInferenceProfileVerifier{},
	"awsbedrockprovisionedthroughput": &bedrockProvisionedThroughputVerifier{},
	"awsbedrockmodelaccess":           &bedrockModelAccessVerifier{},
	"awsbedrockagent":                 &bedrockAgentVerifier{},
	"awsbedrockknowledgebase":         &bedrockKnowledgeBaseVerifier{},
	"awsbedrockflow":                  &bedrockFlowVerifier{},
	"awsbedrockprompt":                &bedrockPromptVerifier{},
	"awsbedrockagentcoreruntime":      &agentCoreRuntimeVerifier{},
	"awsbedrockagentcoregateway":      &agentCoreGatewayVerifier{},
	"awsbedrockagentcorememory":       &agentCoreMemoryVerifier{},
	"awsbedrockagentcoreidentity":     &agentCoreIdentityVerifier{},
	"awsbedrockagentcoretools":        &agentCoreToolsVerifier{},
	"awsbedrockagentcoreevaluation":   &agentCoreEvaluationVerifier{},

	// The settings-singleton class (region/account-scoped settings
	// objects; per-kind destroy contracts in settings_singletons.go).
	"awsapigatewayaccountsettings":  &apiGatewayAccountSettingsVerifier{},
	"awsbedrockinvocationlogging":   &bedrockInvocationLoggingVerifier{},
	"awsbedrockagentcoretokenvault": &agentCoreTokenVaultVerifier{},
	"awssesaccountsettings":         &sesAccountSettingsVerifier{},

	// The governance family (audit/compliance posture; per-kind
	// contracts in governance.go).
	"awscloudtrail":                     &cloudTrailVerifier{},
	"awsconfigrecorder":                 &configRecorderVerifier{},
	"awsconfigrule":                     &configRuleVerifier{},
	"awsguardduty":                      &guardDutyVerifier{},
	"awscloudtraileventdatastore":       &eventDataStoreVerifier{},
	"awsconfigaggregator":               &configAggregatorVerifier{},
	"awsconfigconformancepack":          &conformancePackVerifier{},
	"awsguarddutymalwareprotectionplan": &malwareProtectionPlanVerifier{},

	// The AWS Backup family (data-protection posture; per-kind
	// contracts in backup.go).
	"awsbackupvault":              &backupVaultVerifier{},
	"awsbackupplan":               &backupPlanVerifier{},
	"awsbackupframework":          &backupFrameworkVerifier{},
	"awsbackupreportplan":         &backupReportPlanVerifier{},
	"awsbackuprestoretestingplan": &restoreTestingPlanVerifier{},
	"awsbackupsettings":           &backupSettingsVerifier{},
	"awsssmparameter":             &ssmParameterVerifier{},
	"awsssmdocument":              &ssmDocumentVerifier{},
	"awsssmassociation":           &ssmAssociationVerifier{},
	"awsssmmaintenancewindow":     &ssmMaintenanceWindowVerifier{},
	"awsssmpatchbaseline":         &ssmPatchBaselineVerifier{},

	"awsorganization":        &organizationVerifier{},
	"awsorganizationalunit":  &organizationalUnitVerifier{},
	"awsorganizationaccount": &organizationAccountVerifier{},
	"awsorganizationpolicy":  &organizationPolicyVerifier{},

	"awstransitgateway":              &transitGatewayVerifier{},
	"awstransitgatewayvpcattachment": &transitGatewayVpcAttachmentVerifier{},
	"awstransitgatewayroutetable":    &transitGatewayRouteTableVerifier{},

	"awsathenaworkgroup":     &athenaWorkgroupVerifier{},
	"awscodebuildproject":    &codeBuildProjectVerifier{},
	"awscodepipeline":        &codePipelineVerifier{},
	"awsgluecatalogdatabase": &glueCatalogDatabaseVerifier{},

	"awsmemorydbcluster": &memorydbClusterVerifier{},
	"awsmemorydbuser":    &memorydbUserVerifier{},
	"awsmemorydbacl":     &memorydbAclVerifier{},
	"awsclientvpn":       &clientVpnVerifier{},

	"awsglobalaccelerator": &globalAcceleratorVerifier{},

	"awsfsxlustrefilesystem":           &fsxFileSystemVerifier{kind: "awsfsxlustrefilesystem"},
	"awsfsxopenzfsfilesystem":          &fsxFileSystemVerifier{kind: "awsfsxopenzfsfilesystem"},
	"awsfsxwindowsfilesystem":          &fsxFileSystemVerifier{kind: "awsfsxwindowsfilesystem"},
	"awsfsxontapfilesystem":            &fsxFileSystemVerifier{kind: "awsfsxontapfilesystem"},
	"awsfsxontapstoragevirtualmachine": &fsxStorageVirtualMachineVerifier{},
	"awsfsxontapvolume":                &fsxVolumeVerifier{},
	"awsfsxdatarepositoryassociation":  &fsxDataRepositoryAssociationVerifier{},
	"awssagemakerdomain":               &sageMakerDomainVerifier{},

	"awssagemakermodel":            &sageMakerModelVerifier{},
	"awssagemakerendpoint":         &sageMakerEndpointVerifier{},
	"awssagemakernotebookinstance": &sageMakerNotebookInstanceVerifier{},
	"awssagemakerfeaturegroup":     &sageMakerFeatureGroupVerifier{},
	"awssagemakermodelregistry":    &sageMakerModelRegistryVerifier{},
	"awssagemakerpipeline":         &sageMakerPipelineVerifier{},
	"awssagemakerimage":            &sageMakerImageVerifier{},
	"awssagemakermlflowserver":     &sageMakerMlflowServerVerifier{},
	"awssagemakermlflowapp":        &sageMakerMlflowAppVerifier{},

	"awssecretsmanagersecret":           &secretsManagerSecretVerifier{},
	"awsopensearchserverlesscollection": &openSearchServerlessCollectionVerifier{},
}

// GetVerifier returns the verifier for a component, or an error if none is registered.
func GetVerifier(kind string) (Verifier, error) {
	v, ok := verifiers[kind]
	if !ok {
		return nil, errors.Errorf("no AWS verifier registered for kind %q", kind)
	}
	return v, nil
}
