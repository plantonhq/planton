"use client";
import React, { useState } from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Copy, Check, GitBranch } from "lucide-react";
import { Check as CheckIcon } from "lucide-react";

export default function CICDSection() {
  const [copied, setCopied] = useState(false);

  const workflowCode = `name: Deploy with Planton
on:
  push:
    branches: [ main ]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install Planton CLI
        run: brew install plantonhq/tap/planton

      - name: Install Pulumi & OpenTofu
        run: |
          curl -fsSL https://get.pulumi.com | sh
          echo "$HOME/.pulumi/bin" >> $GITHUB_PATH
          curl -L https://github.com/opentofu/opentofu/releases/latest/download/tofu_linux_amd64.zip -o tofu.zip
          sudo unzip -o tofu.zip -d /usr/local/bin

      - name: Plan
        run: planton plan -f infra/manifest.yaml

      - name: Apply
        run: planton apply -f infra/manifest.yaml --stack myorg/myproj/prod`;

  const copyWorkflow = () => {
    navigator.clipboard.writeText(workflowCode);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div className="text-center mb-16">
        <h2 className="text-4xl font-bold mb-6">
          <span className="bg-gradient-to-r from-white to-slate-300 bg-clip-text text-transparent">CI/CD Integration</span>
        </h2>
        <p className="text-xl text-slate-400 max-w-3xl mx-auto">Deploy your infrastructure automatically with GitHub Actions</p>
      </div>

      <Card className="bg-slate-900/30 border-slate-700 max-w-5xl mx-auto">
        <CardHeader className="pb-4">
          <div className="flex items-center justify-between">
            <CardTitle className="text-xl text-white flex items-center gap-2">
              <GitBranch className="w-5 h-5" />
              GitHub Actions Workflow
            </CardTitle>
            <Button
              size="sm"
              variant="outline"
              onClick={copyWorkflow}
              className="border-slate-600 bg-slate-800/50 text-slate-300 hover:text-white hover:bg-slate-700"
            >
              {copied ? (
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
        <CardContent>
          <div className="bg-slate-950 rounded-lg p-6 font-mono text-sm overflow-x-auto max-w-full">
            <pre className="text-slate-300 whitespace-pre-wrap break-words w-full">
              <code className="block">{workflowCode}</code>
            </pre>
          </div>
        </CardContent>
      </Card>

      <div className="grid md:grid-cols-3 gap-6 mt-8">
        <Card className="bg-slate-900/30 border-slate-700">
          <CardContent className="p-6 text-center">
            <div className="w-12 h-12 bg-gradient-to-r from-green-500 to-emerald-400 rounded-xl flex items-center justify-center mx-auto mb-4">
              <CheckIcon className="w-6 h-6 text-white" />
            </div>
            <h3 className="text-lg font-bold text-white mb-2">Deterministic</h3>
            <p className="text-slate-400">Same commands, same results across all environments</p>
          </CardContent>
        </Card>
        <Card className="bg-slate-900/30 border-slate-700">
          <CardContent className="p-6 text-center">
            <div className="w-12 h-12 bg-gradient-to-r from-blue-500 to-cyan-400 rounded-xl flex items-center justify-center mx-auto mb-4">
              <GitBranch className="w-6 h-6 text-white" />
            </div>
            <h3 className="text-lg font-bold text-white mb-2">Version Controlled</h3>
            <p className="text-slate-400">Infrastructure as code with full Git history</p>
          </CardContent>
        </Card>
        <Card className="bg-slate-900/30 border-slate-700">
          <CardContent className="p-6 text-center">
            <div className="w-12 h-12 bg-gradient-to-r from-purple-500 to-pink-400 rounded-xl flex items-center justify-center mx-auto mb-4">
              <Copy className="w-6 h-6 text-white" />
            </div>
            <h3 className="text-lg font-bold text-white mb-2">Portable</h3>
            <p className="text-slate-400">Works with any CI/CD system that supports CLI tools</p>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}


