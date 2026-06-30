"use client";
import React, { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Copy, Check } from "lucide-react";

export default function Quickstart() {
  const [copiedSteps, setCopiedSteps] = useState<Record<string, boolean>>({});

  const copyCode = (stepId: string, code: string) => {
    navigator.clipboard.writeText(code);
    setCopiedSteps((prev) => ({ ...prev, [stepId]: true }));
    setTimeout(() => {
      setCopiedSteps((prev) => ({ ...prev, [stepId]: false }));
    }, 2000);
  };

  const steps = [
    { id: "install", title: "1. Install", code: "brew install plantonhq/tap/planton" },
    {
      id: "manifest",
      title: "2. Create manifest with provisioner label",
      code: `cat > manifest.yaml <<EOF
apiVersion: kubernetes.planton.dev/v1
kind: KubernetesPostgres
metadata:
  name: my-database
  labels:
    planton.dev/provisioner: pulumi  # or tofu
spec:
  container:
    replicas: 1
    resources:
      limits:
        cpu: 500m
        memory: 1Gi
EOF`,
    },
    { id: "validate", title: "3. Validate", code: "planton validate -f manifest.yaml" },
    {
      id: "deploy",
      title: "4. Deploy (kubectl-style! 🚀)",
      code: `# Auto-detects provisioner from label
planton apply -f manifest.yaml

# With field overrides
planton apply -f manifest.yaml --set spec.container.replicas=3`,
    },
    {
      id: "destroy",
      title: "5. Destroy (optional)",
      code: `planton destroy -f manifest.yaml`,
    },
  ];

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div className="text-center mb-16">
        <h2 className="text-4xl font-bold mb-6">
          <span className="bg-gradient-to-r from-white to-slate-300 bg-clip-text text-transparent">
            Get Started in Minutes
          </span>
        </h2>
        <p className="text-xl text-slate-400 max-w-3xl mx-auto">
          Five simple steps to deploy with kubectl-style commands
        </p>
      </div>
      <div className="grid gap-6 max-w-4xl mx-auto">
        {steps.map((step) => (
          <Card key={step.id} className="bg-slate-900/30 border-slate-700">
            <CardHeader className="pb-4">
              <CardTitle className="text-xl text-white flex items-center justify-between">
                {step.title}
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => copyCode(step.id, step.code)}
                  className="border-slate-600 bg-slate-800/50 text-slate-300 hover:text-white hover:bg-slate-700"
                >
                  {copiedSteps[step.id] ? (
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
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="bg-slate-950 rounded-lg p-4 font-mono text-sm overflow-x-auto max-w-full">
                <pre className="text-slate-300 whitespace-pre-wrap break-words w-full">
                  <code className="block">{step.code}</code>
                </pre>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
      <Card className="mt-8 bg-purple-950/30 border-purple-800/50">
        <CardContent className="p-6">
          <h3 className="text-lg font-bold text-white mb-4">✨ What&apos;s New</h3>
          <ul className="text-slate-300 space-y-2">
            <li>• <strong>kubectl-style commands</strong>: Use <code className="text-purple-400">apply</code> and <code className="text-purple-400">destroy</code> for all deployments</li>
            <li>• <strong>Auto-detection</strong>: Provisioner automatically detected from <code className="text-purple-400">planton.dev/provisioner</code> label</li>
            <li>• <strong>Interactive prompts</strong>: If label is missing, CLI prompts you to choose (defaults to Pulumi)</li>
            <li>• <strong>Backward compatible</strong>: Traditional <code className="text-purple-400">pulumi</code> and <code className="text-purple-400">tofu</code> commands still work</li>
          </ul>
        </CardContent>
      </Card>
      <Card className="mt-4 bg-blue-950/30 border-blue-800/50">
        <CardContent className="p-6">
          <h3 className="text-lg font-bold text-white mb-4">Prerequisites</h3>
          <ul className="text-slate-300 space-y-2">
            <li>• Provider credentials are read from environment variables (AWS_ACCESS_KEY_ID, GOOGLE_APPLICATION_CREDENTIALS, KUBECONFIG, etc.)</li>
            <li>• Pulumi and/or OpenTofu CLI must be installed separately</li>
            <li>• Supported provisioners: <code className="text-blue-400">pulumi</code>, <code className="text-blue-400">tofu</code>, <code className="text-blue-400">terraform</code></li>
          </ul>
        </CardContent>
      </Card>
    </div>
  );
}


