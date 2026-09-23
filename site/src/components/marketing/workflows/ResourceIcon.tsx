import { WORKFLOW_ICONS } from '../../../data/workflow-icons';
import { workflowDarkTokens as palette } from '../../../theme/workflows';

// Quiet family tints frame the canonical glyph; icon color never encodes a
// deployment status. Keep concrete SVG colors so video matches the browser.
const families: readonly [string, readonly string[]][] = [
  ['#7da3cc', ['GcpVpcNetwork', 'GcpSubnetwork', 'GcpGkeCluster', 'DigitalOceanVpc']],
  ['#d99e66', ['AwsLambda', 'GcpGkeNodePool', 'KubernetesDeployment', 'AzureContainerAppEnvironment', 'AzureContainerApp', 'AzureContainerAppJob', 'DigitalOceanDropletAutoscalePool']],
  ['#7dbfa5', ['AwsS3Bucket', 'CloudflareR2Bucket', 'DigitalOceanBucket']],
  ['#a795c9', ['AwsS3VectorBucket', 'AwsBedrockKnowledgeBase', 'AwsBedrockInferenceProfile', 'KubernetesQdrant', 'AzurePostgresqlFlexibleServer', 'CloudflareD1Database', 'CloudflareKvNamespace', 'DigitalOceanDatabaseCluster', 'DigitalOceanDatabaseUser']],
  ['#c98f8d', ['AwsBedrockGuardrail', 'AzureKeyVault', 'DigitalOceanFirewall', 'DigitalOceanDatabaseFirewall', 'DigitalOceanCertificate', 'KubernetesCertificate']],
  ['#c9b86b', ['AzureServiceBusNamespace', 'AzureServiceBusQueue', 'CloudflareQueue']],
  ['#74bfc8', ['AwsHttpApiGateway', 'KubernetesGateway', 'KubernetesHttpRoute', 'CloudflareWorker', 'DigitalOceanLoadBalancer', 'DigitalOceanCdn']],
  ['#c48db8', ['AwsIamRole', 'AzureUserAssignedIdentity', 'DigitalOceanSpacesKey', 'DigitalOceanSshKey']],
];
const accents = new Map(families.flatMap(([color, kinds]) => kinds.map(kind => [kind, color] as const)));

export function ResourceIcon({ kind, x, y, size = 36 }: { kind: string; x: number; y: number; size?: number }) {
  const accent = accents.get(kind) ?? '#9ca3af';
  const glyph = size - 12;
  return <g aria-hidden="true">
    <rect x={x} y={y} width={size} height={size} rx="7" fill={palette.surface.canvas} />
    <rect x={x} y={y} width={size} height={size} rx="7" fill={accent} fillOpacity="0.18" stroke={accent} strokeOpacity="0.4" strokeWidth="1" />
    <image href={WORKFLOW_ICONS[kind]} x={x + 6} y={y + 6} width={glyph} height={glyph} />
  </g>;
}
