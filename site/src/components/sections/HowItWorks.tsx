import React from "react";
import { Card, CardContent } from "@/components/ui/card";
import { ArrowRight, FileText, Search, Package, Play, CheckCircle } from "lucide-react";

export default function HowItWorks() {
  const steps = [
    { icon: FileText, title: "Parse", description: "apiVersion/kind/metadata/spec → Protobuf object" },
    { icon: Search, title: "Validate", description: "Buf ProtoValidate/CEL on spec" },
    { icon: Package, title: "Build", description: "stack-input = { provisioner, pulumi|terraform, target, providerCredential }" },
    { icon: Play, title: "Plan", description: "planton plan (auto-detects provisioner)" },
    { icon: CheckCircle, title: "Apply", description: "planton apply (backends: local|s3|gcs|azurerm)" },
  ];

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div className="text-center mb-16">
        <h2 className="text-4xl font-bold mb-6">
          <span className="bg-gradient-to-r from-white to-slate-300 bg-clip-text text-transparent">How It Works</span>
        </h2>
        <p className="text-xl text-slate-400 max-w-3xl mx-auto mb-6">
          From YAML manifest to deployed infrastructure in five clear steps
        </p>
        <p className="text-lg text-slate-300 max-w-3xl mx-auto mb-12">
          Unlike tools that force you to learn different abstractions or write imperative code, Planton follows the Kubernetes philosophy: <span className="text-purple-400 font-semibold">declarative configuration, strong validation, consistent workflow</span>. Whether you&apos;re deploying to AWS, GCP, Azure, or Kubernetes, the same five-step process applies.
        </p>
      </div>

      <div className="mb-16">
        <div className="max-w-7xl mx-auto flex items-center justify-center gap-3">
          {steps.map((step, index) => {
            const Icon = step.icon;
            return (
              <React.Fragment key={index}>
                <Card className="bg-slate-900/50 border-slate-700 flex-1 max-w-[200px]">
                  <CardContent className="p-6 text-center">
                    <div className="w-14 h-14 bg-gradient-to-r from-[#7a4183] to-purple-500 rounded-xl flex items-center justify-center mx-auto mb-4">
                      <Icon className="w-7 h-7 text-white" />
                    </div>
                    <h3 className="text-lg font-bold text-white">{step.title}</h3>
                  </CardContent>
                </Card>
                {index < steps.length - 1 && (
                  <ArrowRight className="w-5 h-5 text-slate-500 flex-shrink-0" />
                )}
              </React.Fragment>
            );
          })}
        </div>
      </div>

      <div className="grid md:grid-cols-2 gap-8">
        <Card className="bg-slate-900/30 border-slate-700">
          <CardContent className="p-8">
            <h3 className="text-2xl font-bold text-white mb-6">Input: YAML Manifest</h3>
            <div className="bg-slate-900 rounded-lg p-4 font-mono text-sm">
              <div className="text-slate-400"># example: postgres-on-kubernetes.yaml</div>
              <div className="text-blue-400">apiVersion: <span className="text-white">kubernetes.planton.dev/v1</span></div>
              <div className="text-blue-400">kind: <span className="text-white">PostgresKubernetes</span></div>
              <div className="text-blue-400">metadata:</div>
              <div className="text-blue-400 ml-4">name: <span className="text-white">payments-db</span></div>
              <div className="text-blue-400">spec:</div>
              <div className="text-blue-400 ml-4">container:</div>
              <div className="text-blue-400 ml-8">replicas: <span className="text-emerald-400">3</span></div>
              <div className="text-blue-400 ml-8">disk_size: <span className="text-white">&quot;100Gi&quot;</span></div>
            </div>
          </CardContent>
        </Card>
        <Card className="bg-slate-900/30 border-slate-700">
          <CardContent className="p-8">
            <h3 className="text-2xl font-bold text-white mb-6">Output: Status/Results</h3>
            <div className="bg-slate-900 rounded-lg p-4 font-mono text-sm">
              <div className="text-slate-400"># modules export status.outputs</div>
              <div className="text-blue-400">status:</div>
              <div className="text-blue-400 ml-4">outputs:</div>
              <div className="text-blue-400 ml-8">namespace: <span className="text-white">payments-db-main</span></div>
              <div className="text-blue-400 ml-8">service: <span className="text-white">payments-db</span></div>
              <div className="text-blue-400 ml-8">kube_endpoint: <span className="text-white">payments-db.payments-db-main.svc:5432</span></div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}


