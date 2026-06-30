import React from "react";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Badge } from "@/components/ui/badge";
import { Check, X, Minus } from "lucide-react";

export default function CompareSection() {

  const features = [
    { name: "Declarative YAML", planton: true, crossplane: true, terraform: false, pulumi: false },
    { name: "Multi-cloud consistency", planton: true, crossplane: true, terraform: false, pulumi: false },
    { name: "Strong validation", planton: true, crossplane: false, terraform: false, pulumi: false },
    { name: "Multiple backends", planton: true, crossplane: false, terraform: true, pulumi: true },
    { name: "CLI workflow", planton: true, crossplane: false, terraform: true, pulumi: true },
    { name: "Kubernetes native", planton: false, crossplane: true, terraform: false, pulumi: false },
    { name: "Programming languages", planton: false, crossplane: false, terraform: false, pulumi: true },
    { name: "Curated modules", planton: true, crossplane: false, terraform: false, pulumi: false },
  ];

  const FeatureIcon = ({ enabled }: { enabled: boolean | null | undefined }) => {
    if (enabled === true) return <Check className="w-4 h-4 text-green-400" />;
    if (enabled === false) return <X className="w-4 h-4 text-red-400" />;
    return <Minus className="w-4 h-4 text-slate-500" />;
  };

  return (
    <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
      <div className="text-center mb-16">
        <h2 className="text-4xl font-bold mb-6">
          <span className="bg-gradient-to-r from-white to-slate-300 bg-clip-text text-transparent">How Planton Compares</span>
        </h2>
        <p className="text-xl text-slate-400 max-w-3xl mx-auto">Short, factual comparison with other infrastructure tools</p>
      </div>

      <div className="grid lg:grid-cols-3 gap-6 mb-12">
        <Card className="bg-slate-900/30 border-slate-700">
          <CardHeader>
            <CardTitle className="text-lg text-white">vs Cloud Provider CLIs</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-slate-400">
              <strong className="text-red-400">Their way:</strong> Learn AWS, GCP, Azure CLIs separately. Different syntax, mental models, and workflows for every cloud.
            </p>
            <p className="text-slate-400 mt-4">
              <strong className="text-emerald-400">Planton:</strong> One consistent YAML structure and CLI across all clouds. Same workflow everywhere.
            </p>
          </CardContent>
        </Card>
        <Card className="bg-slate-900/30 border-slate-700">
          <CardHeader>
            <CardTitle className="text-lg text-white">vs Kubernetes Controllers</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-slate-400">
              <strong className="text-red-400">Crossplane:</strong> Requires running Kubernetes cluster to manage infrastructure. Controller-based reconciliation.
            </p>
            <p className="text-slate-400 mt-4">
              <strong className="text-emerald-400">Planton:</strong> CLI-driven, runs anywhere. No cluster required. You control when deployments happen.
            </p>
          </CardContent>
        </Card>
        <Card className="bg-slate-900/30 border-slate-700">
          <CardHeader>
            <CardTitle className="text-lg text-white">vs Writing Raw IaC</CardTitle>
          </CardHeader>
          <CardContent>
            <p className="text-slate-400">
              <strong className="text-red-400">Terraform/Pulumi:</strong> Write imperative code or HCL. Learn provider-specific resource syntax for every service.
            </p>
            <p className="text-slate-400 mt-4">
              <strong className="text-emerald-400">Planton:</strong> Write declarative YAML. The framework maintains battle-tested Pulumi and Terraform modules for you.
            </p>
          </CardContent>
        </Card>
      </div>

      <Card className="bg-slate-900/30 border-slate-700">
        <CardHeader>
          <CardTitle className="text-xl text-white">Feature Comparison</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow className="border-slate-800">
                  <TableHead className="text-slate-300">Feature</TableHead>
                  <TableHead className="text-slate-300">Planton</TableHead>
                  <TableHead className="text-slate-300">Crossplane</TableHead>
                  <TableHead className="text-slate-300">Terraform</TableHead>
                  <TableHead className="text-slate-300">Pulumi</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {features.map((feature, index) => (
                  <TableRow key={index} className="border-slate-800">
                    <TableCell className="font-medium text-white">{feature.name}</TableCell>
                    <TableCell>
                      <div className="flex items-center gap-2">
                        <FeatureIcon enabled={feature.planton} />
                        {feature.planton && (
                          <Badge className="bg-emerald-900 text-emerald-200">Yes</Badge>
                        )}
                      </div>
                    </TableCell>
                    <TableCell>
                      <FeatureIcon enabled={feature.crossplane} />
                    </TableCell>
                    <TableCell>
                      <FeatureIcon enabled={feature.terraform} />
                    </TableCell>
                    <TableCell>
                      <FeatureIcon enabled={feature.pulumi} />
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </CardContent>
      </Card>

      {/* Key Differentiators */}
      <div className="grid md:grid-cols-3 gap-6 mt-12">
        <Card className="bg-slate-900/30 border-purple-900/30">
          <CardContent className="p-6">
            <h4 className="text-lg font-bold text-purple-400 mb-3">🔍 Complete Transparency</h4>
            <p className="text-slate-300">
              All Pulumi and Terraform modules are open source. No black boxes. Audit the exact code deploying your infrastructure. Fork and customize if needed.
            </p>
          </CardContent>
        </Card>
        <Card className="bg-slate-900/30 border-purple-900/30">
          <CardContent className="p-6">
            <h4 className="text-lg font-bold text-purple-400 mb-3">⚡ Your Choice of Engine</h4>
            <p className="text-slate-300">
              Use Pulumi OR OpenTofu as your execution engine. Same manifests work with both. Switch engines anytime. Not locked to one IaC provider.
            </p>
          </CardContent>
        </Card>
        <Card className="bg-slate-900/30 border-purple-900/30">
          <CardContent className="p-6">
            <h4 className="text-lg font-bold text-purple-400 mb-3">🔓 Zero Vendor Lock-In</h4>
            <p className="text-slate-300">
              Standalone CLI with zero SaaS dependencies. Your manifests, your credentials, your infrastructure. Runs locally or in your CI/CD. You&apos;re never trapped.
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}


