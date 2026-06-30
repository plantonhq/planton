"use client";
import React, { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Copy, Check } from "lucide-react";

export default function ExampleGallery() {
  const [copiedExamples, setCopiedExamples] = useState<Record<string, boolean>>({});

  const copyCode = (exampleId: string, code: string) => {
    navigator.clipboard.writeText(code);
    setCopiedExamples((prev) => ({ ...prev, [exampleId]: true }));
    setTimeout(() => {
      setCopiedExamples((prev) => ({ ...prev, [exampleId]: false }));
    }, 2000);
  };

  const examples = {
    aws: [
      {
        id: "aws-rds",
        filename: "aws-rds-instance.yaml",
        title: "RDS PostgreSQL Instance",
        manifest: `apiVersion: aws.planton.dev/v1
kind: AwsRdsInstance
metadata:
  name: payments-db
spec:
  subnetIds:
    - value: subnet-abc123
    - value: subnet-def456
  engine: postgres
  engineVersion: "15.4"
  instanceClass: db.t3.medium
  allocatedStorageGb: 100
  storageEncrypted: true
  username: dbadmin
  password: <secret>
  port: 5432
  multiAz: true`,
        deploy: `planton validate aws-rds-instance.yaml
planton apply -f aws-rds-instance.yaml`,
      },
      {
        id: "aws-s3",
        filename: "aws-s3-bucket.yaml",
        title: "S3 Bucket",
        manifest: `apiVersion: aws.planton.dev/v1
kind: AwsS3Bucket
metadata:
  name: app-assets
spec:
  bucketName: my-app-assets-bucket
  versioningEnabled: true
  encryptionEnabled: true
  publicAccessBlock:
    blockPublicAcls: true
    blockPublicPolicy: true
    ignorePublicAcls: true
    restrictPublicBuckets: true`,
        deploy: `planton validate aws-s3-bucket.yaml
planton apply -f aws-s3-bucket.yaml`,
      },
    ],
    gcp: [
      {
        id: "gcp-gke",
        filename: "gcp-gke-cluster.yaml",
        title: "GKE Cluster with Autoscaling",
        manifest: `apiVersion: gcp.planton.dev/v1
kind: GcpGkeCluster
metadata:
  name: main-gke
spec:
  clusterProjectId: <project-id>
  region: us-central1
  zone: us-central1-a
  isWorkloadLogsEnabled: false
  clusterAutoscalingConfig:
    isEnabled: true
    cpuMinCores: 4
    cpuMaxCores: 32
    memoryMinGb: 16
    memoryMaxGb: 128
  nodePools:
    - name: general-pool
      machineType: n2-standard-8
      minNodeCount: 1
      maxNodeCount: 5
      isSpotEnabled: false`,
        deploy: `planton plan -f gcp-gke-cluster.yaml
planton apply -f gcp-gke-cluster.yaml`,
      },
      {
        id: "gcp-cloudrun",
        filename: "gcp-cloud-run.yaml",
        title: "Cloud Run Service",
        manifest: `apiVersion: gcp.planton.dev/v1
kind: GcpCloudRun
metadata:
  name: hello-run
spec:
  projectId: <project-id>
  region: us-central1
  service:
    name: hello
    image: us-docker.pkg.dev/cloudrun/container/hello
    allowUnauthenticated: true`,
        deploy: `planton apply -f gcp-cloud-run.yaml --auto-approve`,
      },
    ],
    azure: [
      {
        id: "azure-aks",
        filename: "azure-aks.yaml",
        title: "AKS Cluster",
        manifest: `apiVersion: azure.planton.dev/v1
kind: AzureAksCluster
metadata:
  name: ops-aks
spec:
  subscriptionId: <subscription>
  resourceGroupName: rg-ops
  region: eastus
  nodePools:
    - name: system
      vmSize: Standard_DS2_v2
      minNodeCount: 1
      maxNodeCount: 3`,
        deploy: `planton apply -f azure-aks.yaml`,
      },
      {
        id: "azure-acr",
        filename: "azure-acr.yaml",
        title: "Container Registry",
        manifest: `apiVersion: azure.planton.dev/v1
kind: AzureContainerRegistry
metadata:
  name: app-registry
spec:
  subscriptionId: <subscription>
  resourceGroupName: rg-ops
  region: eastus
  sku: Basic`,
        deploy: `planton apply -f azure-acr.yaml --auto-approve`,
      },
    ],
    kubernetes: [
      {
        id: "redis-k8s",
        filename: "redis-kubernetes.yaml",
        title: "Redis on Kubernetes",
        manifest: `apiVersion: kubernetes.planton.dev/v1
kind: RedisKubernetes
metadata:
  name: session-cache
spec:
  container:
    replicas: 1
    resources:
      limits:
        cpu: "1000m"
        memory: "2Gi"
      requests:
        cpu: "50m"
        memory: "100Mi"
    persistenceEnabled: true
    diskSize: "5Gi"`,
        deploy: `planton plan -f redis-kubernetes.yaml
planton apply -f redis-kubernetes.yaml`,
      },
      {
        id: "postgres-k8s",
        filename: "postgres-kubernetes.yaml",
        title: "PostgreSQL on Kubernetes",
        manifest: `apiVersion: kubernetes.planton.dev/v1
kind: KubernetesPostgres
metadata:
  name: app-db
spec:
  container:
    replicas: 3
    resources:
      limits:
        cpu: "2000m"
        memory: "4Gi"
      requests:
        cpu: "100m"
        memory: "256Mi"
    diskSize: "100Gi"
  ingress:
    enabled: true
    hostname: db.example.com`,
        deploy: `planton plan -f postgres-kubernetes.yaml
planton apply -f postgres-kubernetes.yaml --auto-approve`,
      },
    ],
  } as const;

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div className="text-center mb-16">
        <h2 className="text-4xl font-bold mb-6">
          <span className="bg-gradient-to-r from-white to-slate-300 bg-clip-text text-transparent">
            Example Gallery
          </span>
        </h2>
        <p className="text-xl text-slate-400 max-w-3xl mx-auto">
          Real manifests from the repo (trimmed for clarity). Replace placeholders like &lt;project-id&gt;.
        </p>
      </div>

      <Tabs defaultValue="aws" className="w-full">
        <TabsList className="grid grid-cols-4 w-full max-w-md mx-auto mb-8 bg-slate-800 border-slate-700">
          <TabsTrigger value="aws" className="data-[state=active]:bg-slate-700">AWS</TabsTrigger>
          <TabsTrigger value="gcp" className="data-[state=active]:bg-slate-700">GCP</TabsTrigger>
          <TabsTrigger value="azure" className="data-[state=active]:bg-slate-700">Azure</TabsTrigger>
          <TabsTrigger value="kubernetes" className="data-[state=active]:bg-slate-700">Kubernetes</TabsTrigger>
        </TabsList>

        {Object.entries(examples).map(([provider, providerExamples]) => (
          <TabsContent key={provider} value={provider} className="space-y-8">
            <div className="grid lg:grid-cols-2 gap-8">
              {providerExamples.map((example) => (
                <Card key={example.id} className="bg-slate-900/30 border-slate-700">
                  <CardHeader className="pb-4">
                    <div className="flex items-center justify-between">
                      <div>
                        <CardTitle className="text-lg text-white">{example.title}</CardTitle>
                        <Badge variant="outline" className="border-slate-600 text-slate-400 mt-2">
                          {example.filename}
                        </Badge>
                      </div>
                      <Button
                        size="sm"
                        variant="outline"
                        onClick={() => copyCode(example.id, example.manifest)}
                        className="border-slate-600 bg-slate-800/50 text-slate-300 hover:text-white hover:bg-slate-700"
                      >
                        {copiedExamples[example.id] ? (
                          <>
                            <Check className="w-3 h-3 mr-1" />
                            Copied
                          </>
                        ) : (
                          <>
                            <Copy className="w-3 h-3 mr-1" />
                            Copy
                          </>
                        )}
                      </Button>
                    </div>
                  </CardHeader>
                  <CardContent className="space-y-4">
                    <div className="bg-slate-950 rounded-lg p-4 font-mono text-xs overflow-x-auto max-w-full">
                      <pre className="text-slate-300 whitespace-pre-wrap break-words w-full">
                        <code className="block">{example.manifest}</code>
                      </pre>
                    </div>
                    <div>
                      <h4 className="text-sm font-bold text-white mb-2">Deploy:</h4>
                      <div className="bg-slate-900 rounded-lg p-3 font-mono text-xs overflow-x-auto max-w-full">
                        <pre className="text-green-400 whitespace-pre-wrap break-words w-full">
                          <code className="block">{example.deploy}</code>
                        </pre>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          </TabsContent>
        ))}
      </Tabs>
    </div>
  );
}


